package certauth

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-ca/lib/client/credential"
	x509cred "github.com/hyperledger/fabric-ca/lib/client/credential/x509"
	fabrictls "github.com/hyperledger/fabric-ca/lib/tls"
	"github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-lib-go/bccsp"
)

// Sentinel errors for CA client operations
var (
	ErrClientInit      = errors.New("fabric ca: failed to initialize client")
	ErrRegister        = errors.New("fabric ca: failed to register identity")
	ErrModifyIdentity  = errors.New("fabric ca: failed to modify identity")
	ErrEnroll          = errors.New("fabric ca: failed to enroll identity")
	ErrRevoke          = errors.New("fabric ca: failed to revoke identity")
	ErrInvalidConfig   = errors.New("fabric ca: invalid configuration")
	ErrInvalidResponse = errors.New("fabric ca: invalid response from CA")
	ErrGenCSR          = errors.New("fabric ca: failed to generate CSR")
)

// =============================================================================
// BootstrapClient - Username/Password based authentication
// Used during bootstrap when no keypair/cert exists yet
// =============================================================================

// BootstrapClientConfig represents the configuration for username/password authentication
type BootstrapClientConfig struct {
	URL        string // CA server URL (e.g., "https://ca.fabric-shard-org1.svc.cluster.local:7054")
	Username   string // Username for authentication
	Password   string // Password/secret for authentication
	CAName     string // Optional CA name (for multi-CA servers)
	TLSCertPEM []byte // PEM-encoded TLS certificate to verify CA server (optional, but recommended for self-signed certs)
}

// Validate checks if the BootstrapClientConfig has all required fields
func (c *BootstrapClientConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("%w: URL cannot be empty", ErrInvalidConfig)
	}
	if c.Username == "" {
		return fmt.Errorf("%w: Username cannot be empty", ErrInvalidConfig)
	}
	if c.Password == "" {
		return fmt.Errorf("%w: Password cannot be empty", ErrInvalidConfig)
	}
	return nil
}

// BootstrapClient provides enrollment-only access to a Fabric CA using username/password.
// This is used during initial bootstrap when no admin keypair/cert exists yet.
// After enrollment, the returned certificates should be stored in Vault.
type BootstrapClient interface {
	// Enroll enrolls an identity using username/password and returns certificates.
	// The returned EnrollmentResponse contains the certificate and key material.
	Enroll(ctx context.Context, req *api.EnrollmentRequest) (*lib.EnrollmentResponse, error)

	// EnrollWithCSR enrolls an identity using username/password by submitting an
	// externally-generated CSR. The caller holds the private key; this client
	// never sees it. Returns only public material: the leaf certificate and the
	// CA chain from the enrollment response. Used by the issuer side of the
	// cross-owner CSR flow, where the issuer has the CSR but not the key.
	EnrollWithCSR(ctx context.Context, csrPEM []byte, profile string, hosts []string) (certPEM []byte, caChainPEM []byte, err error)

	// GetCAInfo retrieves information about the CA (certificate chain, etc.)
	GetCAInfo(ctx context.Context) (*lib.GetCAInfoResponse, error)

	// Close cleans up any temporary resources (like TLS cert files)
	// Should be called when the client is no longer needed.
	Close() error
}

// bootstrapClient implements BootstrapClient
type bootstrapClient struct {
	fabricClient *lib.Client
	config       BootstrapClientConfig
}

// Close cleans up temporary resources
func (c *bootstrapClient) Close() error {
	// No resources to clean up - everything is in-memory
	return nil
}

// NewBootstrapClient creates a new bootstrap-only client.
// This client uses username/password authentication and only supports Enroll() and GetCAInfo().
// Use this during initial bootstrap to obtain the admin's keypair/cert for storage in Vault.
func NewBootstrapClient(config BootstrapClientConfig) (BootstrapClient, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	fabricCAConfig := &lib.ClientConfig{
		URL:    config.URL,
		CAName: config.CAName,
		MSPDir: "", // In-memory operations only
	}

	fabricCAClient := &lib.Client{
		HomeDir: "",
		Config:  fabricCAConfig,
	}

	// Configure TLS if certificate is provided
	// Use SetHTTPClient for in-memory TLS configuration (no temporary files)
	if len(config.TLSCertPEM) > 0 {
		httpClient, err := createTLSHTTPClient(config.TLSCertPEM)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create TLS HTTP client: %v", ErrClientInit, err)
		}
		fabricCAClient.SetHTTPClient(httpClient)
	}

	if err := fabricCAClient.Init(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrClientInit, err)
	}

	return &bootstrapClient{
		fabricClient: fabricCAClient,
		config:       config,
	}, nil
}

// Enroll enrolls an identity using username/password credentials.
// If req is nil, uses the credentials from the config.
func (c *bootstrapClient) Enroll(ctx context.Context, req *api.EnrollmentRequest) (*lib.EnrollmentResponse, error) {
	// If no request provided, build one from config
	if req == nil {
		req = &api.EnrollmentRequest{
			Name:   c.config.Username,
			Secret: c.config.Password,
			CAName: c.config.CAName,
		}
	}

	resp, err := c.fabricClient.Enroll(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEnroll, err)
	}

	if resp == nil || resp.Identity == nil {
		return nil, fmt.Errorf("%w: enrollment returned nil response or identity", ErrInvalidResponse)
	}

	return resp, nil
}

// EnrollWithCSR enrolls an identity by submitting an externally-generated CSR.
// The caller (key owner) already produced the CSR with GenerateCSR, so this
// client (the issuer) does the enroll HTTP flow without ever touching a private
// key. Only public material is returned.
func (c *bootstrapClient) EnrollWithCSR(ctx context.Context, csrPEM []byte, profile string, hosts []string) ([]byte, []byte, error) {
	if len(csrPEM) == 0 {
		return nil, nil, fmt.Errorf("%w: CSR is empty", ErrInvalidConfig)
	}

	// Build enrollment request wrapping the supplied CSR
	reqNet := &api.EnrollmentRequestNet{
		CAName: c.config.CAName,
	}
	reqNet.SignRequest.Request = string(csrPEM)
	reqNet.SignRequest.Profile = profile
	if len(hosts) > 0 {
		reqNet.SignRequest.Hosts = hosts
	}

	body, err := util.Marshal(reqNet, "SignRequest")
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to marshal enrollment request: %v", ErrEnroll, err)
	}

	// Send enrollment request with basic auth (identity name + secret)
	baseURL, err := lib.NormalizeURL(c.fabricClient.Config.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to parse CA URL: %v", ErrEnroll, err)
	}
	enrollURL := baseURL.String() + "/enroll"

	req, err := http.NewRequestWithContext(ctx, "POST", enrollURL, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to create HTTP request: %v", ErrEnroll, err)
	}
	req.SetBasicAuth(c.config.Username, c.config.Password)

	var result api.EnrollmentResponseNet
	if err := c.fabricClient.SendReq(req, &result); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrEnroll, err)
	}

	certBytes, err := util.B64Decode(result.Cert)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to decode certificate: %v", ErrInvalidResponse, err)
	}

	caChain, err := util.B64Decode(result.ServerInfo.CAChain)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: failed to decode CA chain: %v", ErrInvalidResponse, err)
	}

	return certBytes, caChain, nil
}

// GetCAInfo retrieves information about the CA
func (c *bootstrapClient) GetCAInfo(ctx context.Context) (*lib.GetCAInfoResponse, error) {
	resp, err := c.fabricClient.GetCAInfo(&api.GetCAInfoRequest{
		CAName: c.config.CAName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get CA info: %w", err)
	}

	return resp, nil
}

// =============================================================================
// Client - Keypair/Certificate based authentication from Vault
// Used after bootstrap when admin credentials exist in Vault
// =============================================================================

// ClientConfig represents the configuration for keypair/cert authentication
type ClientConfig struct {
	URL    string // CA server URL
	CAName string // Optional CA name (for multi-CA servers)

	// Admin identity credentials loaded from Vault
	AdminCert []byte // PEM-encoded admin certificate
	AdminKey  []byte // PEM-encoded admin private key

	// TLS certificate for verifying the CA server (optional, but recommended for self-signed certs)
	TLSCertPEM []byte // PEM-encoded TLS certificate of the CA server
}

// Validate checks if the ClientConfig has all required fields
func (c *ClientConfig) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("%w: URL cannot be empty", ErrInvalidConfig)
	}
	if len(c.AdminCert) == 0 {
		return fmt.Errorf("%w: AdminCert cannot be empty", ErrInvalidConfig)
	}
	if len(c.AdminKey) == 0 {
		return fmt.Errorf("%w: AdminKey cannot be empty", ErrInvalidConfig)
	}
	return nil
}

// Client provides full admin access to a Fabric CA using keypair/cert authentication.
// The admin identity is loaded from Vault during construction.
// This client is suitable for caching by MultiClientProvider.
type Client interface {
	// Register registers a new identity with the CA and returns an enrollment secret.
	Register(ctx context.Context, req *api.RegistrationRequest) (*api.RegistrationResponse, error)

	// ModifyIdentity updates an existing identity (e.g. to reset the enrollment
	// secret when a name is already registered). Requires admin privileges.
	ModifyIdentity(ctx context.Context, req *api.ModifyIdentityRequest) (*api.IdentityResponse, error)

	// Enroll enrolls an identity using an enrollment secret and returns certificates.
	// Note: This does not require admin privileges - anyone with a valid secret can enroll.
	Enroll(ctx context.Context, req *api.EnrollmentRequest) (*lib.EnrollmentResponse, error)

	// Revoke revokes an identity from the CA.
	Revoke(ctx context.Context, req *api.RevocationRequest) (*api.RevocationResponse, error)

	// GetCAInfo retrieves information about the CA (certificate chain, etc.)
	GetCAInfo(ctx context.Context) (*lib.GetCAInfoResponse, error)
}

// client implements Client
type client struct {
	fabricClient  *lib.Client
	adminIdentity *lib.Identity
	config        ClientConfig
}

// NewClient creates a new client with keypair/cert authentication.
// The admin identity is constructed from the certificate and key in the config.
// This client is ready to use immediately for Register, Revoke, and Enroll operations.
func NewClient(config ClientConfig) (Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	fabricCAConfig := &lib.ClientConfig{
		URL:    config.URL,
		CAName: config.CAName,
		MSPDir: "", // In-memory operations only
	}

	fabricCAClient := &lib.Client{
		HomeDir: "",
		Config:  fabricCAConfig,
	}

	// Configure TLS if certificate is provided
	// Use SetHTTPClient for in-memory TLS configuration (no temporary files)
	if len(config.TLSCertPEM) > 0 {
		httpClient, err := createTLSHTTPClient(config.TLSCertPEM)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create TLS HTTP client: %v", ErrClientInit, err)
		}
		fabricCAClient.SetHTTPClient(httpClient)
	}

	if err := fabricCAClient.Init(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrClientInit, err)
	}

	// Build the admin identity from the provided cert and key
	identity, err := createIdentityFromCredentials(fabricCAClient, config.AdminCert, config.AdminKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create admin identity from credentials: %v", ErrClientInit, err)
	}

	return &client{
		fabricClient:  fabricCAClient,
		adminIdentity: identity,
		config:        config,
	}, nil
}

// Register registers a new identity with the CA
func (c *client) Register(ctx context.Context, req *api.RegistrationRequest) (*api.RegistrationResponse, error) {
	resp, err := c.adminIdentity.Register(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRegister, err)
	}
	return resp, nil
}

// ModifyIdentity updates an existing identity record on the CA.
// Used to reset the enrollment secret when re-signing for an already-registered
// name (Fabric CA returns 409 on duplicate registration).
func (c *client) ModifyIdentity(ctx context.Context, req *api.ModifyIdentityRequest) (*api.IdentityResponse, error) {
	resp, err := c.adminIdentity.ModifyIdentity(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrModifyIdentity, err)
	}
	return resp, nil
}

// Enroll enrolls an identity using an enrollment secret
func (c *client) Enroll(ctx context.Context, req *api.EnrollmentRequest) (*lib.EnrollmentResponse, error) {
	resp, err := c.fabricClient.Enroll(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEnroll, err)
	}

	if resp == nil || resp.Identity == nil {
		return nil, fmt.Errorf("%w: enrollment returned nil response or identity", ErrInvalidResponse)
	}

	return resp, nil
}

// Revoke revokes an identity from the CA
func (c *client) Revoke(ctx context.Context, req *api.RevocationRequest) (*api.RevocationResponse, error) {
	resp, err := c.adminIdentity.Revoke(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRevoke, err)
	}
	return resp, nil
}

// GetCAInfo retrieves information about the CA
func (c *client) GetCAInfo(ctx context.Context) (*lib.GetCAInfoResponse, error) {
	resp, err := c.fabricClient.GetCAInfo(&api.GetCAInfoRequest{
		CAName: c.config.CAName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get CA info: %w", err)
	}
	return resp, nil
}

// =============================================================================
// Helper functions
// =============================================================================

// GenerateCSR builds a PEM-encoded CSR from a PEM-encoded ECDSA private key
// (PKCS8, as produced by secrets/v1/generator.go) and the given CSR info.
//
// It runs entirely locally on the KEY OWNER's side: the private key never
// leaves this process and only the public CSR is produced (parse PKCS8 -> SEC1
// DER -> BCCSP import -> GenCSRUsingKey).
func GenerateCSR(privateKeyPEM []byte, csrInfo *api.CSRInfo) ([]byte, error) {
	if len(privateKeyPEM) == 0 {
		return nil, fmt.Errorf("%w: private key is empty", ErrInvalidConfig)
	}
	if csrInfo == nil {
		return nil, fmt.Errorf("%w: csr info is required", ErrInvalidConfig)
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("%w: failed to decode PEM private key", ErrInvalidConfig)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse PKCS8 private key: %v", ErrInvalidConfig, err)
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w: key is not an ECDSA key, got %T", ErrInvalidConfig, key)
	}

	// BCCSP's ECDSAPrivateKeyImportOpts expects SEC1 (EC) DER format, not PKCS8
	ecDER, err := x509.MarshalECPrivateKey(ecKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal EC private key: %v", ErrInvalidConfig, err)
	}

	// A minimal client provides the BCCSP used to import the key and build the
	// CSR. Use a unique temp home dir so concurrent calls don't collide on the
	// msp keystore that Init() materializes on disk.
	homeDir, err := os.MkdirTemp("", "fabric-ca-csr-")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create temp dir: %v", ErrGenCSR, err)
	}
	defer os.RemoveAll(homeDir)

	cspClient := &lib.Client{
		HomeDir: homeDir,
		Config:  &lib.ClientConfig{},
	}
	if err := cspClient.Init(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrClientInit, err)
	}

	bccspKey, err := cspClient.GetCSP().KeyImport(ecDER, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to import key into BCCSP: %v", ErrGenCSR, err)
	}

	csrPEM, _, err := cspClient.GenCSRUsingKey(csrInfo, csrInfo.CN, bccspKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenCSR, err)
	}

	return csrPEM, nil
}

// createTLSHTTPClient creates an HTTP client configured to verify the CA server's certificate.
// This allows in-memory TLS configuration without temporary files.
func createTLSHTTPClient(serverCACert []byte) (*http.Client, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(serverCACert) {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	tlsConfig := &tls.Config{
		RootCAs:      certPool,
		CipherSuites: fabrictls.DefaultCipherSuites,
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}, nil
}

// createIdentityFromCredentials creates a lib.Identity from PEM-encoded cert and key.
// This allows us to use an identity that was previously enrolled and stored in Vault.
//
// Implementation approach (from IN_MEMORY_CA_CLIENT.md):
// 1. Parse the certificate to extract the Common Name (identity name)
// 2. Import the private key into BCCSP (Blockchain Crypto Service Provider)
// 3. Create a Signer from the BCCSP key and certificate
// 4. Create an X509 Credential and set it from the Signer
// 5. Create an Identity with the credential
func createIdentityFromCredentials(client *lib.Client, certPEM, keyPEM []byte) (*lib.Identity, error) {
	// Step 1: Parse the certificate to extract identity name (Common Name)
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	identityName := cert.Subject.CommonName
	if identityName == "" {
		return nil, fmt.Errorf("certificate has no Common Name")
	}

	// Step 2: Decode the private key PEM to get DER bytes
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	// Step 3: Import the private key into BCCSP
	// BCCSP expects the raw DER-encoded bytes (keyBlock.Bytes), not the parsed key object
	csp := client.GetCSP()
	if csp == nil {
		return nil, fmt.Errorf("failed to get BCCSP from client")
	}

	// Import the DER bytes as a BCCSP key
	// BCCSP will handle parsing the PKCS8 format internally
	bccspKey, err := csp.KeyImport(keyBlock.Bytes, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, fmt.Errorf("failed to import private key into BCCSP: %w", err)
	}

	// Step 4: Create a Signer from the BCCSP key and certificate
	signer, err := x509cred.NewSigner(bccspKey, certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// Step 5: Create an X509 Credential
	// We use empty strings for certFile and keyFile because we're setting the credential from a Signer
	// The x509cred.Client interface is satisfied by lib.Client
	cred := x509cred.NewCredential("", "", client)

	// Set the credential value from the Signer (this is in-memory, no file I/O)
	if err := cred.SetCredentialFromSigner(signer); err != nil {
		return nil, fmt.Errorf("failed to set credential from signer: %w", err)
	}

	// Step 6: Create the Identity with the credential
	identity := lib.NewIdentity(client, identityName, []credential.Credential{cred})

	return identity, nil
}
