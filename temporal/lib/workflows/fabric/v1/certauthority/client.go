package certauthority

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-ca/lib/tls"
)

// Sentinel errors for CA client operations
var (
	ErrClientInit      = errors.New("fabric ca: failed to initialize client")
	ErrRegister        = errors.New("fabric ca: failed to register identity")
	ErrEnroll          = errors.New("fabric ca: failed to enroll identity")
	ErrRevoke          = errors.New("fabric ca: failed to revoke identity")
	ErrInvalidConfig   = errors.New("fabric ca: invalid configuration")
	ErrInvalidResponse = errors.New("fabric ca: invalid response from CA")
)

// =============================================================================
// BootstrapClient - Username/Password based authentication
// Used during bootstrap when no keypair/cert exists yet
// =============================================================================

// BootstrapClientConfig represents the configuration for username/password authentication
type BootstrapClientConfig struct {
	URL         string // CA server URL (e.g., "https://ca.fabric-org1.svc.cluster.local:7054")
	Username    string // Username for authentication
	Password    string // Password/secret for authentication
	CAName      string // Optional CA name (for multi-CA servers)
	TLSCertPEM  []byte // PEM-encoded TLS certificate to verify CA server (optional, but recommended for self-signed certs)
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
	tlsCertFile  string // Path to temporary TLS cert file (empty if not using TLS)
}

// Close cleans up temporary resources (like TLS cert file)
func (c *bootstrapClient) Close() error {
	if c.tlsCertFile != "" {
		return os.Remove(c.tlsCertFile)
	}
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

	// Configure TLS if certificate is provided
	// We write the cert to a temporary file because fabric-ca library expects file paths
	var tlsCertFile string
	if len(config.TLSCertPEM) > 0 {
		tmpFile, err := os.CreateTemp("", "ca-tls-cert-*.pem")
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create temp file for TLS cert: %v", ErrClientInit, err)
		}
		tlsCertFile = tmpFile.Name()

		if _, err := tmpFile.Write(config.TLSCertPEM); err != nil {
			tmpFile.Close()
			os.Remove(tlsCertFile)
			return nil, fmt.Errorf("%w: failed to write TLS cert to temp file: %v", ErrClientInit, err)
		}
		tmpFile.Close()

		fabricCAConfig.TLS = tls.ClientTLSConfig{
			Enabled:   true,
			CertFiles: []string{tlsCertFile},
		}
	}

	fabricCAClient := &lib.Client{
		HomeDir: "",
		Config:  fabricCAConfig,
	}

	if err := fabricCAClient.Init(); err != nil {
		// Clean up temp file on error
		if tlsCertFile != "" {
			os.Remove(tlsCertFile)
		}
		return nil, fmt.Errorf("%w: %v", ErrClientInit, err)
	}

	return &bootstrapClient{
		fabricClient: fabricCAClient,
		config:       config,
		tlsCertFile:  tlsCertFile,
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

	if err := fabricCAClient.Init(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrClientInit, err)
	}

	// Build the admin identity from the provided cert and key
	// TODO: Create identity from cert/key bytes using fabric-ca lib
	// For now, this is a placeholder - need to investigate the fabric-ca API
	// to create an Identity from PEM bytes without enrolling
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

// createIdentityFromCredentials creates a lib.Identity from PEM-encoded cert and key.
// This allows us to use an identity that was previously enrolled and stored in Vault.
func createIdentityFromCredentials(client *lib.Client, certPEM, keyPEM []byte) (*lib.Identity, error) {
	// TODO: Implement using fabric-ca library
	// Need to investigate the correct API:
	// - lib.Client.NewX509Identity?
	// - x509cred.NewCredential + x509cred.NewSigner?
	// - Loading cert/key and building Identity manually?
	//
	// For now, return nil with an error indicating this needs implementation
	return nil, errors.New("createIdentityFromCredentials not yet implemented - need to investigate fabric-ca API")
}
