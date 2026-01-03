package fabric

import (
	"context"
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
	secret, err := a.secretProvider.FetchSecret(ctx, credentialsRef)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bootstrap credentials: %w", err)
	}

	creds, err := secret.Credentials()
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	// Fetch TLS certificate if provided
	var tlsCertPEM []byte
	if tlsCertRef != nil {
		certSecret, err := a.secretProvider.FetchSecret(ctx, tlsCertRef)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch TLS certificate: %w", err)
		}
		cert, err := certSecret.Certificate()
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
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

	// Create config with explicit admin credentials (nil = use default permanent admin)
	hydrator := certauthority.NewConfig(caConfig, req.GetAdminCredentialsRef())

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
// The pre-generated keypair is imported into the Fabric CA client's BCCSP so it uses our key
// instead of generating its own, ensuring the certificate matches our stored private key.
func (a *Activities) EnrollIdentity(ctx context.Context, req *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
	certType := req.GetCertType()
	caID := req.GetCertificateAuthorityId()
	identityName := req.GetIdentityName()

	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidInput, certType)
	}

	caConfig := a.fabricConfig.CertificateAuthority(certType, caID)

	// Step 1: Fetch the pre-generated keypair from Vault
	secret, err := a.secretProvider.FetchSecret(ctx, req.GetKeyRef())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch secret: %w", ErrFetchSecret, err)
	}

	keyPair, err := secret.KeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse key pair: %w", ErrFetchSecret, err)
	}

	// Step 2: Fetch TLS certificate if provided
	var tlsCertPEM []byte
	if req.GetCaTlsCertRef() != nil {
		certSecret, err := a.secretProvider.FetchSecret(ctx, req.GetCaTlsCertRef())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch TLS certificate: %w", err)
		}
		cert, err := certSecret.Certificate()
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}
		tlsCertPEM = cert.Certificate
	}

	// Step 3: Create a bootstrap client for enrollment using the enrollment secret
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

	// Step 4: Enroll with our pre-generated key
	// This ensures the certificate matches our private key
	csrInfo := &api.CSRInfo{
		CN: req.GetCommonName(),
	}

	enrollResp, err := bootstrapClient.EnrollWithKey(ctx, keyPair.PrivateKey, csrInfo)
	if err != nil {
		return nil, fmt.Errorf("%w: enrollment failed: %w", ErrEnrollIdentity, err)
	}

	// Step 5: Extract the certificate
	certPEM := enrollResp.Identity.GetECert().Cert()

	// Step 6: Update the keypair secret with the certificate
	// Store at the same location we fetched the key from
	identityCreds := &secrets.IdentityCredentials{
		PrivateKey:  keyPair.PrivateKey,
		PublicKey:   keyPair.PublicKey,
		Certificate: certPEM,
	}

	identitySecret := secrets.FromIdentityCredentials(identityCreds)
	if err := a.secretProvider.UpdateSecret(ctx, req.GetKeyRef(), identitySecret); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStoreSecret, err)
	}

	return &pb.EnrollIdentityOutput{
		CertRef: req.GetKeyRef(), // Return the same ref - it now contains key + cert
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

	// Create config with explicit admin credentials (nil = use default permanent admin)
	hydrator := certauthority.NewConfig(caConfig, req.GetAdminCredentialsRef())

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
// The pre-generated keypair is imported into the Fabric CA client's BCCSP so it uses our key
// instead of generating its own, ensuring the certificate matches our stored private key.
func (a *Activities) BootstrapEnrollIdentity(ctx context.Context, req *pb.BootstrapEnrollIdentityInput) (*pb.BootstrapEnrollIdentityOutput, error) {
	// Step 1: Fetch the pre-generated keypair from Vault
	secret, err := a.secretProvider.FetchSecret(ctx, req.GetKeyRef())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch secret: %w", ErrFetchSecret, err)
	}

	keyPair, err := secret.KeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse key pair: %w", ErrFetchSecret, err)
	}

	// Step 2: Create bootstrap client
	bootstrapClient, err := a.getBootstrapClient(ctx, req.GetBootstrapCredentialsRef(), req.GetCaUrl(), req.GetCaTlsCertRef())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapEnrollIdentity, err)
	}
	defer bootstrapClient.Close()

	// Step 3: Enroll with our pre-generated key
	// This ensures the certificate matches our private key
	csrInfo := &api.CSRInfo{
		CN: req.GetCommonName(),
	}

	enrollResp, err := bootstrapClient.EnrollWithKey(ctx, keyPair.PrivateKey, csrInfo)
	if err != nil {
		return nil, fmt.Errorf("%w: enrollment failed: %w", ErrBootstrapEnrollIdentity, err)
	}

	certPEM := enrollResp.Identity.GetECert().Cert()

	// Step 4: Store our original keypair + the certificate
	// They match because the CA signed a CSR generated with our key
	identityCreds := &secrets.IdentityCredentials{
		PrivateKey:  keyPair.PrivateKey,
		PublicKey:   keyPair.PublicKey,
		Certificate: certPEM,
	}

	identitySecret := secrets.FromIdentityCredentials(identityCreds)
	if err := a.secretProvider.UpdateSecret(ctx, req.GetCertRef(), identitySecret); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStoreSecret, err)
	}

	return &pb.BootstrapEnrollIdentityOutput{
		CertRef: req.GetCertRef(),
	}, nil
}
