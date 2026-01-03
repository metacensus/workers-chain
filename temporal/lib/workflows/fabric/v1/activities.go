package fabric

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-ca/api"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	"github.com/metacensus/infra/temporal/lib/workflows/fabric/v1/certauthority"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
)

// Sentinel errors for fabric activities
var (
	ErrRegisterIdentity        = errors.New("register identity")
	ErrEnrollIdentity          = errors.New("enroll identity")
	ErrBootstrapEnrollIdentity = errors.New("bootstrap enroll identity")
	ErrRevokeIdentity          = errors.New("revoke identity")
	ErrInvalidInput            = errors.New("invalid input")
	ErrFetchSecret             = errors.New("failed to fetch secret")
	ErrStoreSecret             = errors.New("failed to store secret")
	ErrGenerateCSR             = errors.New("failed to generate CSR")
	ErrParseKey                = errors.New("failed to parse key")
)

// Activities implements the Fabric activities
type Activities struct {
	// secretProvider provides access to fetch and store secrets
	secretProvider secrets.SecretProvider

	// clientProvider manages cached Clients for Register/Revoke operations
	clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig]

	fabricConfig config.FabricConfig
}

// NewActivities creates a new Activities instance
func NewActivities(
	secretProvider secrets.SecretProvider,
	clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig],
	fabricConfig config.FabricConfig,
) *Activities {
	return &Activities{
		secretProvider: secretProvider,
		clientProvider: clientProvider,
		fabricConfig:   fabricConfig,
	}
}

// getBootstrapClient creates a new BootstrapClient for the specified CA.
// BootstrapClients use username/password auth and only support Enroll.
// These are NOT cached - each call creates a fresh client.
func (a *Activities) getBootstrapClient(ctx context.Context, credentialsRef *secretspb.SecretRef, caURL string, tlsCertRef *secretspb.SecretRef) (certauthority.BootstrapClient, error) {
	// Fetch credentials from secret provider
	creds, err := a.secretProvider.FetchCredentials(ctx, credentialsRef)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bootstrap credentials: %w", err)
	}

	// Fetch TLS certificate if provided
	var tlsCertPEM []byte
	if tlsCertRef != nil {
		cert, err := a.secretProvider.FetchCertificate(ctx, tlsCertRef)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch TLS certificate: %w", err)
		}
		tlsCertPEM = cert.Certificate
	}

	// Create the client
	client, err := certauthority.NewBootstrapClient(certauthority.BootstrapClientConfig{
		URL:        caURL,
		Username:   creds.Username,
		Password:   creds.Password,
		CAName:     "", // Default CA
		TLSCertPEM: tlsCertPEM,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bootstrap client: %w", err)
	}

	return client, nil
}

// RegisterIdentity registers a new identity with a Fabric CA
// This is the first step in identity creation - registration returns an enrollment secret
func (a *Activities) RegisterIdentity(ctx context.Context, req *pb.RegisterIdentityInput) (*pb.RegisterIdentityOutput, error) {
	certType := req.GetCertType()
	caID := req.GetCertificateAuthorityId()
	identityName := req.GetIdentityName()

	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidInput, certType)
	}

	caConfig := a.fabricConfig.CertificateAuthority(certType, caID)

	// Create the appropriate hydrator based on which admin credentials to use
	var hydrator common.ConfigHydrator[certauthority.ClientConfig]
	if req.GetAdminCredentialsRef() != nil {
		// Use specific admin credentials (e.g., bootstrap admin during CA setup)
		hydrator = certauthority.NewConfigWithAdminRef(caConfig, req.GetAdminCredentialsRef())
	} else {
		// Use default permanent admin credentials
		hydrator = certauthority.NewConfig(caConfig)
	}

	// Get the CA client - the clientProvider handles fetching secrets, hydrating config, and caching
	client, err := a.clientProvider.GetClient(ctx, hydrator)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get CA client: %w", ErrRegisterIdentity, err)
	}

	// Build registration request
	registerReq := &api.RegistrationRequest{
		Name:        identityName,
		Type:        req.GetIdentityType(),
		Affiliation: req.GetAffiliation(),
		Attributes:  convertAttributes(req.GetAttributes()),
	}

	// Register the identity with the CA
	resp, err := client.Register(ctx, registerReq)
	if err != nil {
		return nil, fmt.Errorf("%w: registration failed: %w", ErrRegisterIdentity, err)
	}

	if resp == nil || resp.Secret == "" {
		return nil, fmt.Errorf("%w: registration returned empty secret", ErrRegisterIdentity)
	}

	return &pb.RegisterIdentityOutput{
		EnrollmentSecret: resp.Secret,
	}, nil
}

// convertAttributes converts a map of strings to Fabric CA attributes
func convertAttributes(attrs map[string]string) []api.Attribute {
	if len(attrs) == 0 {
		return nil
	}

	result := make([]api.Attribute, 0, len(attrs))
	for name, value := range attrs {
		result = append(result, api.Attribute{
			Name:  name,
			Value: value,
		})
	}
	return result
}

// EnrollIdentity enrolls a registered identity and retrieves certificates
// This is the second step in identity creation - enrollment uses the secret from registration
func (a *Activities) EnrollIdentity(ctx context.Context, req *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
	certType := req.GetCertType()
	caID := req.GetCertificateAuthorityId()
	identityName := req.GetIdentityName()

	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidInput, certType)
	}

	caConfig := a.fabricConfig.CertificateAuthority(certType, caID)

	// Fetch the private key from Vault
	keyPair, err := a.secretProvider.FetchKeyPair(ctx, req.GetKeyRef())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch key pair: %w", ErrFetchSecret, err)
	}

	// Parse the private key for CSR generation
	privateKey, err := parseECDSAPrivateKey(keyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseKey, err)
	}

	// Generate CSR
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         req.GetCommonName(),
			Organization:       []string{req.GetOrganization()},
			OrganizationalUnit: []string{req.GetOrganizationalUnit()},
		},
	}

	_, err = x509.CreateCertificateRequest(rand.Reader, csrTemplate, privateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGenerateCSR, err)
	}

	// Fetch TLS certificate if provided
	var tlsCertPEM []byte
	if req.GetCaTlsCertRef() != nil {
		cert, err := a.secretProvider.FetchCertificate(ctx, req.GetCaTlsCertRef())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch TLS certificate: %w", err)
		}
		tlsCertPEM = cert.Certificate
	}

	// Create a bootstrap client for enrollment using the enrollment secret
	// The enrollment secret is passed in the EnrollmentRequest, not in the client config
	bootstrapClient, err := certauthority.NewBootstrapClient(certauthority.BootstrapClientConfig{
		URL:        caConfig.URL(),
		Username:   identityName,              // Use identity name as username for enrollment
		Password:   req.GetEnrollmentSecret(), // Use enrollment secret as password
		CAName:     "",                        // Default CA
		TLSCertPEM: tlsCertPEM,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create bootstrap client: %w", ErrEnrollIdentity, err)
	}
	defer bootstrapClient.Close()

	// Build enrollment request
	enrollReq := &api.EnrollmentRequest{
		Name:   identityName,
		Secret: req.GetEnrollmentSecret(),
		CSR: &api.CSRInfo{
			CN: req.GetCommonName(),
		},
	}

	// Enroll with the CA
	enrollResp, err := bootstrapClient.Enroll(ctx, enrollReq)
	if err != nil {
		return nil, fmt.Errorf("%w: enrollment failed: %w", ErrEnrollIdentity, err)
	}

	// Extract the certificate
	certPEM := enrollResp.Identity.GetECert().Cert()

	// Store keypair+cert together at a single location
	// For the permanent admin created by CreateIdentity, this will be VaultAdminKeyPairPath()
	// The workflow determines the path based on identity name and CA config
	credentialsRef := secrets.NewVaultRef(fmt.Sprintf("%s/%s/keypair", caConfig.VaultBasePath(), identityName))

	identityCreds := &secrets.IdentityCredentials{
		PrivateKey:  keyPair.PrivateKey,
		PublicKey:   keyPair.PublicKey,
		Certificate: certPEM,
	}

	if err := a.secretProvider.StoreIdentityCredentials(ctx, credentialsRef, identityCreds); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStoreSecret, err)
	}

	return &pb.EnrollIdentityOutput{
		CertRef: credentialsRef,
	}, nil
}

// RevokeIdentity revokes an identity from a Fabric CA
// Used to revoke bootstrap admin credentials after permanent admin is created
func (a *Activities) RevokeIdentity(ctx context.Context, req *pb.RevokeIdentityInput) (*pb.RevokeIdentityOutput, error) {
	certType := req.GetCertType()
	caID := req.GetCertificateAuthorityId()
	identityName := req.GetIdentityName()

	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidInput, certType)
	}

	caConfig := a.fabricConfig.CertificateAuthority(certType, caID)

	// Create the appropriate hydrator based on which admin credentials to use
	var hydrator common.ConfigHydrator[certauthority.ClientConfig]
	if req.GetAdminCredentialsRef() != nil {
		// Use specific admin credentials (e.g., permanent admin revoking bootstrap admin)
		hydrator = certauthority.NewConfigWithAdminRef(caConfig, req.GetAdminCredentialsRef())
	} else {
		// Use default permanent admin credentials
		hydrator = certauthority.NewConfig(caConfig)
	}

	// Get the CA client - the clientProvider handles fetching secrets, hydrating config, and caching
	client, err := a.clientProvider.GetClient(ctx, hydrator)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get CA client: %w", ErrRevokeIdentity, err)
	}

	// Build revocation request
	revokeReq := &api.RevocationRequest{
		Name:   identityName,
		Reason: req.GetReason(),
	}

	// Revoke the identity
	_, err = client.Revoke(ctx, revokeReq)
	if err != nil {
		return nil, fmt.Errorf("%w: revocation failed: %w", ErrRevokeIdentity, err)
	}

	return &pb.RevokeIdentityOutput{
		Revoked: true,
	}, nil
}

// BootstrapEnrollIdentity enrolls an identity using bootstrap credentials (username/password).
// This is used ONLY during CA initialization to obtain the bootstrap admin's certificate.
func (a *Activities) BootstrapEnrollIdentity(ctx context.Context, req *pb.BootstrapEnrollIdentityInput) (*pb.BootstrapEnrollIdentityOutput, error) {
	keyPair, err := a.secretProvider.FetchKeyPair(ctx, req.GetKeyRef())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch key pair: %w", ErrFetchSecret, err)
	}

	privateKey, err := parseECDSAPrivateKey(keyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseKey, err)
	}

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         req.GetCommonName(),
			Organization:       []string{req.GetOrganization()},
			OrganizationalUnit: []string{req.GetOrganizationalUnit()},
		},
	}

	_, err = x509.CreateCertificateRequest(rand.Reader, csrTemplate, privateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGenerateCSR, err)
	}

	bootstrapClient, err := a.getBootstrapClient(ctx, req.GetBootstrapCredentialsRef(), req.GetCaUrl(), req.GetCaTlsCertRef())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapEnrollIdentity, err)
	}

	// Clean up temporary TLS cert file when done
	defer bootstrapClient.Close()

	creds, err := a.secretProvider.FetchCredentials(ctx, req.GetBootstrapCredentialsRef())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch credentials: %w", ErrFetchSecret, err)
	}

	enrollReq := &api.EnrollmentRequest{
		Name:   creds.Username,
		Secret: creds.Password,
		CSR: &api.CSRInfo{
			CN: req.GetCommonName(),
		},
	}

	enrollResp, err := bootstrapClient.Enroll(ctx, enrollReq)
	if err != nil {
		return nil, fmt.Errorf("%w: enrollment failed: %w", ErrBootstrapEnrollIdentity, err)
	}

	certPEM := enrollResp.Identity.GetECert().Cert()

	// Store keypair+cert together so Config.Hydrate can find both
	identityCreds := &secrets.IdentityCredentials{
		PrivateKey:  keyPair.PrivateKey,
		PublicKey:   keyPair.PublicKey,
		Certificate: certPEM,
	}

	if err := a.secretProvider.StoreIdentityCredentials(ctx, req.GetCertRef(), identityCreds); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStoreSecret, err)
	}

	return &pb.BootstrapEnrollIdentityOutput{
		CertRef: req.GetCertRef(),
	}, nil
}

// parseECDSAPrivateKey parses a PEM-encoded ECDSA private key in PKCS8 format.
// All keys in this system are generated in PKCS8 format (see secrets/v1/generator.go).
func parseECDSAPrivateKey(privateKeyPEM []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse PKCS8 format (standard format used by our key generation)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}

	// Verify it's an ECDSA key
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("PKCS8 key is not an ECDSA key, got %T", key)
	}

	return ecKey, nil
}
