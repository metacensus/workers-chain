package fabric

import (
	"context"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-ca/api"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
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
	ErrInvalidActivityInput    = errors.New("invalid input")
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
}

// NewActivities creates a new Activities instance
func NewActivities(
	secretProvider secrets.SecretProvider,
	clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig],
) *Activities {
	return &Activities{
		secretProvider: secretProvider,
		clientProvider: clientProvider,
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
	caRef := req.GetCaRef()
	if caRef == nil {
		return nil, fmt.Errorf("%w: ca_ref is required", ErrInvalidActivityInput)
	}
	identityName := req.GetIdentityName()

	// Create owner config from input - enables activities to work for any owner
	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	certType := caRef.GetCertType()
	if certType == "tls" {
		ownerConfig = config.GetOperator()
	} else {
		ownerConfig = config.NewOrgConfig(req.GetOrg())
	}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, caRef)

	// Create config with explicit admin identity (nil = use default permanent admin)
	hydrator := certauthority.NewConfig(caConfig, req.GetAdminIdentity())

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

// EnrollIdentity enrolls a registered identity and retrieves certificates.
// This is the second step in identity creation - enrollment uses the secret from registration.
// The pre-generated keypair is imported into the Fabric CA client's BCCSP so it uses our key
// instead of generating its own, ensuring the certificate matches our stored private key.
//
// Keys and certs are stored at separate Vault paths:
//   - key_ref (keys/ path): private key only
//   - cert_ref (certs/ path): {certificate: leaf, chain: full chain}
//
// The chain is built at enrollment time: [leaf cert] + [issuer's chain].
func (a *Activities) EnrollIdentity(ctx context.Context, req *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
	caRef := req.GetCaRef()
	if caRef == nil {
		return nil, fmt.Errorf("%w: ca_ref is required", ErrInvalidActivityInput)
	}

	identityName := req.GetIdentityName()

	// Create owner config from input - enables activities to work for any owner
	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	certType := caRef.GetCertType()
	if certType == "tls" {
		ownerConfig = config.GetOperator()
	} else {
		ownerConfig = config.NewOrgConfig(req.GetOrg())
	}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, caRef)

	// Get storage refs from identity
	storage := req.GetStorage()
	if storage == nil {
		return nil, fmt.Errorf("%w: storage is required", ErrInvalidActivityInput)
	}

	// Step 1: Fetch the pre-generated keypair from Vault (keys/ path)
	secret, err := a.secretProvider.FetchSecret(ctx, storage.GetKeyRef())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch secret: %w", ErrFetchSecret, err)
	}

	keyPair, err := secret.KeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse key pair: %w", ErrFetchSecret, err)
	}

	// Step 2: Fetch CA's TLS certificate for mTLS connection
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

	// Step 3: Fetch issuer's chain for building the full certificate chain
	var issuerChain []byte
	if req.GetIssuerCertRef() != nil {
		issuerSecret, err := a.secretProvider.FetchSecret(ctx, req.GetIssuerCertRef())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issuer cert: %w", err)
		}
		// Get the chain from the issuer's cert path (certs/ path stores {certificate, chain})
		if chainStr, ok := issuerSecret["chain"].(string); ok {
			issuerChain = []byte(chainStr)
		}
	}

	// Step 4: Create a bootstrap client for enrollment using the enrollment secret
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

	// Step 5: Enroll with our pre-generated key
	// This ensures the certificate matches our private key
	subject := req.GetSubject()
	csrInfo := &api.CSRInfo{
		CN: subject.GetCommonName(),
	}

	// Add DNS SANs if provided
	if len(subject.GetDnsNames()) > 0 {
		csrInfo.Hosts = subject.GetDnsNames()
	}

	// Use the specified signing profile (e.g., "ca" for intermediate CA certificates)
	// Empty string uses the default profile
	enrollResp, err := bootstrapClient.EnrollWithKey(ctx, keyPair.PrivateKey, csrInfo, req.GetProfile())
	if err != nil {
		return nil, fmt.Errorf("%w: enrollment failed: %w", ErrEnrollIdentity, err)
	}

	// Step 6: Extract the certificate and build full chain
	certPEM := enrollResp.Identity.GetECert().Cert()

	// Build full chain: [leaf cert] + [issuer's chain]
	// If no issuer chain is provided, use the CA chain from enrollment response
	var fullChain []byte
	if len(issuerChain) > 0 {
		fullChain = buildCertificateChain(certPEM, issuerChain)
	} else {
		// Fallback to CA chain from enrollment response (for root CA bootstrap)
		fullChain = buildCertificateChain(certPEM, enrollResp.CAInfo.CAChain)
	}

	// Step 7: Store certificate + chain at cert_ref (certs/ path)
	// Both the leaf certificate and full chain are stored for different use cases:
	// - certificate: leaf cert only (for signcerts/)
	// - chain: full chain (for mTLS server.crt)
	certSecret := secrets.Secret{
		"certificate": string(certPEM),
		"chain":       string(fullChain),
	}
	if err := a.secretProvider.UpdateSecret(ctx, storage.GetCertRef(), certSecret); err != nil {
		return nil, fmt.Errorf("%w: failed to store certificate: %w", ErrStoreSecret, err)
	}

	return &pb.EnrollIdentityOutput{
		Identity: storage,
	}, nil
}

// RevokeIdentity revokes an identity from a Fabric CA
// Used to revoke bootstrap admin credentials after permanent admin is created
func (a *Activities) RevokeIdentity(ctx context.Context, req *pb.RevokeIdentityInput) (*pb.RevokeIdentityOutput, error) {
	caRef := req.GetCaRef()
	if caRef == nil {
		return nil, fmt.Errorf("%w: ca_ref is required", ErrInvalidActivityInput)
	}
	identityName := req.GetIdentityName()

	// Create owner config from input - enables activities to work for any owner
	ownerConfig := config.NewOrgConfig(req.GetOrg())
	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, caRef)

	// Create config with explicit admin identity (nil = use default permanent admin)
	hydrator := certauthority.NewConfig(caConfig, req.GetAdminIdentity())

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
//
// Keys and certs are stored at separate Vault paths:
//   - storage.key_ref (keys/ path): private key only (already exists, not modified)
//   - storage.cert_ref (certs/ path): {certificate: leaf, chain: full chain}
//
// The chain is built at enrollment time: [leaf cert] + [issuer's chain].
func (a *Activities) BootstrapEnrollIdentity(ctx context.Context, req *pb.BootstrapEnrollIdentityInput) (*pb.BootstrapEnrollIdentityOutput, error) {
	// Get storage refs from identity
	storage := req.GetStorage()
	if storage == nil {
		return nil, fmt.Errorf("%w: storage is required", ErrInvalidActivityInput)
	}

	// Step 1: Fetch the pre-generated keypair from Vault (keys/ path)
	secret, err := a.secretProvider.FetchSecret(ctx, storage.GetKeyRef())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch secret: %w", ErrFetchSecret, err)
	}

	keyPair, err := secret.KeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse key pair: %w", ErrFetchSecret, err)
	}

	// Step 2: Fetch issuer's chain for building the full certificate chain (optional)
	var issuerChain []byte
	if req.GetIssuerCertRef() != nil {
		issuerSecret, err := a.secretProvider.FetchSecret(ctx, req.GetIssuerCertRef())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issuer cert: %w", err)
		}
		// Get the chain from the issuer's cert path (certs/ path stores {certificate, chain})
		if chainStr, ok := issuerSecret["chain"].(string); ok {
			issuerChain = []byte(chainStr)
		}
	}

	// Step 3: Create bootstrap client
	bootstrapClient, err := a.getBootstrapClient(ctx, req.GetBootstrapCredentialsRef(), req.GetCaUrl(), req.GetCaTlsCertRef())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapEnrollIdentity, err)
	}
	defer bootstrapClient.Close()

	// Step 4: Build CSR info from subject
	subject := req.GetSubject()
	csrInfo := &api.CSRInfo{
		CN: subject.GetCommonName(),
	}

	// Add DNS SANs if provided
	if len(subject.GetDnsNames()) > 0 {
		csrInfo.Hosts = subject.GetDnsNames()
	}

	// Step 5: Enroll with our pre-generated key
	// This ensures the certificate matches our private key
	enrollResp, err := bootstrapClient.EnrollWithKey(ctx, keyPair.PrivateKey, csrInfo, req.GetProfile())
	if err != nil {
		return nil, fmt.Errorf("%w: enrollment failed: %w", ErrBootstrapEnrollIdentity, err)
	}

	// Step 6: Extract the certificate and build full chain
	certPEM := enrollResp.Identity.GetECert().Cert()

	// Build full chain: [leaf cert] + [issuer's chain]
	// If no issuer chain is provided, use the CA chain from enrollment response
	var fullChain []byte
	if len(issuerChain) > 0 {
		fullChain = buildCertificateChain(certPEM, issuerChain)
	} else {
		// Fallback to CA chain from enrollment response (for root CA bootstrap)
		fullChain = buildCertificateChain(certPEM, enrollResp.CAInfo.CAChain)
	}

	// Step 7: Store certificate + chain at cert_ref (certs/ path)
	// Both the leaf certificate and full chain are stored for different use cases:
	// - certificate: leaf cert only (for signcerts/)
	// - chain: full chain (for mTLS server.crt)
	certSecret := secrets.Secret{
		"certificate": string(certPEM),
		"chain":       string(fullChain),
	}
	if err := a.secretProvider.UpdateSecret(ctx, storage.GetCertRef(), certSecret); err != nil {
		return nil, fmt.Errorf("%w: failed to store certificate: %w", ErrStoreSecret, err)
	}

	return &pb.BootstrapEnrollIdentityOutput{
		Identity: storage,
	}, nil
}

// buildCertificateChain combines a leaf certificate with the CA chain.
// This creates a full certificate chain that can be used by intermediate CAs
// to verify client certificates they issued (x509 chain validation requires
// the full chain up to a trusted root).
func buildCertificateChain(leafCert []byte, caChain []byte) []byte {
	if len(caChain) == 0 {
		return leafCert
	}
	// Concatenate leaf cert + CA chain (chain already includes intermediate and root)
	return append(leafCert, caChain...)
}
