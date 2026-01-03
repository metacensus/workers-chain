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

// getClient returns a cached Client for the specified CA.
// Clients use keypair/cert auth loaded from Vault and support Register/Revoke.
func (a *Activities) getClient(ctx context.Context, certType, caID string) (certauthority.Client, error) {
	caConfig := a.fabricConfig.CertificateAuthority(certType, caID)
	hydrator := certauthority.NewConfig(caConfig)

	client, err := a.clientProvider.GetClient(ctx, hydrator)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for %s-%s: %w", certType, caID, err)
	}

	return client, nil
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
	// TODO: Implement identity registration with Fabric CA
	// 1. Connect to Fabric CA server
	// 2. Register identity with attributes
	// 3. Return enrollment secret

	return &pb.RegisterIdentityOutput{}, nil
}

// EnrollIdentity enrolls a registered identity and retrieves certificates
// This is the second step in identity creation - enrollment uses the secret from registration
func (a *Activities) EnrollIdentity(ctx context.Context, req *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
	// TODO: Implement identity enrollment with Fabric CA
	// 1. Connect to Fabric CA server
	// 2. Enroll using enrollment secret
	// 3. Retrieve and return certificates

	return &pb.EnrollIdentityOutput{}, nil
}

// RevokeIdentity revokes an identity from a Fabric CA
// Used to revoke bootstrap admin credentials after permanent admin is created
func (a *Activities) RevokeIdentity(ctx context.Context, req *pb.RevokeIdentityInput) (*pb.RevokeIdentityOutput, error) {
	// TODO: Implement identity revocation with Fabric CA
	// 1. Connect to Fabric CA server
	// 2. Revoke identity by name or serial number
	// 3. Return revocation status

	return &pb.RevokeIdentityOutput{}, nil
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

	cert := &secrets.Certificate{
		Certificate: certPEM,
	}

	if err := a.secretProvider.StoreCertificate(ctx, req.GetCertRef(), cert); err != nil {
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
