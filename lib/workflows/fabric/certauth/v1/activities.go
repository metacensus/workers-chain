package certauth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-ca/api"
	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/certauth/v1"
	secretspb "github.com/metacensus/workers-chain/gen/workflows/secrets/v1"
	"github.com/metacensus/workers-chain/lib/workers/config"
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
	"github.com/metacensus/workers-chain/lib/workflows/common"
	secrets "github.com/metacensus/workers-chain/lib/workflows/secrets/v1"
)

// Sentinel errors for fabric activities
var (
	ErrRegisterIdentity     = errors.New("register identity")
	ErrEnrollIdentity       = errors.New("enroll identity")
	ErrBootstrapEnrollCsr   = errors.New("bootstrap enroll csr")
	ErrRevokeIdentity       = errors.New("revoke identity")
	ErrInvalidActivityInput = errors.New("invalid input")
	ErrFetchSecret          = errors.New("failed to fetch secret")
	ErrStoreSecret          = errors.New("failed to store secret")
	ErrGenerateCSR          = errors.New("failed to generate CSR")
)

// Activities implements the Fabric activities.
//
// Activities act AS their worker's owner (injected at startup): every CA
// referenced in an activity input belongs to the executing worker's owner,
// so all CA paths derive from the injected owner.
type Activities struct {
	// secretProvider provides access to fetch and store secrets
	secretProvider secrets.SecretProvider

	// clientProvider manages cached Clients for Register/Revoke operations
	clientProvider *common.MultiClientProvider[Client, ClientConfig]

	// owner identifies whose CAs this worker operates
	owner config.OwnerConfig
}

// NewActivities creates a new Activities instance acting as the given owner
func NewActivities(
	secretProvider secrets.SecretProvider,
	clientProvider *common.MultiClientProvider[Client, ClientConfig],
	owner config.OwnerConfig,
) *Activities {
	return &Activities{
		secretProvider: secretProvider,
		clientProvider: clientProvider,
		owner:          owner,
	}
}

// getBootstrapClient creates a new BootstrapClient for the specified CA.
// BootstrapClients use username/password auth and only support Enroll.
// These are NOT cached - each call creates a fresh client.
func (a *Activities) getBootstrapClient(ctx context.Context, credentialsRef *secretspb.SecretRef, caURL string, tlsCertRef *secretspb.SecretRef) (BootstrapClient, error) {
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
	client, err := NewBootstrapClient(BootstrapClientConfig{
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

// RevokeIdentity revokes an identity from a Fabric CA
// Used to revoke bootstrap admin credentials after permanent admin is created
func (a *Activities) RevokeIdentity(ctx context.Context, req *pb.RevokeIdentityInput) (*pb.RevokeIdentityOutput, error) {
	caRef := req.GetCaRef()
	if caRef == nil {
		return nil, fmt.Errorf("%w: ca_ref is required", ErrInvalidActivityInput)
	}
	identityName := req.GetIdentityName()

	// Locally-referenced CAs always belong to this worker's injected owner
	caConfig := fabricconfig.NewCertificateAuthorityConfig(a.owner, caRef)

	// Create config with explicit admin identity (nil = use default permanent admin)
	hydrator := NewConfig(caConfig, req.GetAdminIdentity())

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

// BootstrapEnrollCsr runs on the CA OWNER's worker during CA initialization,
// before any registrar (permanent admin) exists. It enrolls an externally
// generated CSR using the CA's preset bootstrap credentials (username/password)
// and returns the leaf certificate plus the full chain.
//
// Unlike SignCsr it performs NO registration: bootstrap credentials are
// pre-provisioned in the CA server config, so there is nothing to register and
// no admin to register as. It is the register-less bootstrap counterpart of
// SignCsr. The private key that produced the CSR never leaves the caller; all
// referenced Vault paths belong to this worker's own owner. Only public
// certificate material is returned.
func (a *Activities) BootstrapEnrollCsr(ctx context.Context, req *pb.BootstrapEnrollCsrInput) (*pb.BootstrapEnrollCsrOutput, error) {
	if len(req.GetCsrPem()) == 0 {
		return nil, fmt.Errorf("%w: csr_pem is required", ErrInvalidActivityInput)
	}

	// Fetch issuer's stored chain for building the full certificate chain
	// (optional). Prefer it over the CA chain from the enrollment response.
	var issuerChain []byte
	if req.GetIssuerCertRef() != nil {
		issuerSecret, err := a.secretProvider.FetchSecret(ctx, req.GetIssuerCertRef())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issuer cert: %w", err)
		}
		if chainStr, ok := issuerSecret["chain"].(string); ok {
			issuerChain = []byte(chainStr)
		}
	}

	// Create a bootstrap client using the CA's preset credentials.
	bootstrapClient, err := a.getBootstrapClient(ctx, req.GetBootstrapCredentialsRef(), req.GetCaUrl(), req.GetCaTlsCertRef())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapEnrollCsr, err)
	}
	defer bootstrapClient.Close()

	// Enroll the externally-generated CSR. No register step: the bootstrap
	// credentials themselves authenticate the enrollment.
	certPEM, caChainPEM, err := bootstrapClient.EnrollWithCSR(ctx, req.GetCsrPem(), req.GetProfile(), req.GetHosts())
	if err != nil {
		return nil, fmt.Errorf("%w: enrollment failed: %w", ErrBootstrapEnrollCsr, err)
	}

	// Build the full chain: [leaf] + [issuer's chain]. Prefer the issuer's
	// stored chain; fall back to the CA chain from the enrollment response
	// (root CA bootstrap case).
	var fullChain []byte
	if len(issuerChain) > 0 {
		fullChain = buildCertificateChain(certPEM, issuerChain)
	} else {
		fullChain = buildCertificateChain(certPEM, caChainPEM)
	}

	return &pb.BootstrapEnrollCsrOutput{
		CertificatePem: certPEM,
		ChainPem:       fullChain,
	}, nil
}

// GenerateCsr runs on the KEY OWNER's worker. It fetches the private key at
// key_ref and returns a CSR built from that key for the given subject.
//
// The private key is ensured up-front by secrets.EnsureKeyPair, so here it is
// only read - never generated or written. It never leaves this worker's Vault
// subtree; only the public CSR is returned.
func (a *Activities) GenerateCsr(ctx context.Context, req *pb.GenerateCsrInput) (*pb.GenerateCsrOutput, error) {
	keyRef := req.GetKeyRef()
	if keyRef == nil {
		return nil, fmt.Errorf("%w: key_ref is required", ErrInvalidActivityInput)
	}
	if req.GetCommonName() == "" {
		return nil, fmt.Errorf("%w: common_name is required", ErrInvalidActivityInput)
	}

	secret, err := a.secretProvider.FetchSecret(ctx, keyRef)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to fetch key: %w", ErrFetchSecret, err)
	}
	keyPair, err := secret.KeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse key pair: %w", ErrFetchSecret, err)
	}

	csrInfo := &api.CSRInfo{CN: req.GetCommonName()}
	if len(req.GetHosts()) > 0 {
		csrInfo.Hosts = req.GetHosts()
	}

	csrPEM, err := GenerateCSR(keyPair.PrivateKey, csrInfo)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGenerateCSR, err)
	}

	return &pb.GenerateCsrOutput{CsrPem: csrPEM}, nil
}

// SignCsr runs on the ISSUER's worker. It registers the identity with a
// locally-owned CA (resetting the enrollment secret via ModifyIdentity when the
// name is already registered), enrolls the supplied CSR, and returns the leaf
// certificate plus the full chain.
//
// The enrollment secret is generated and consumed entirely in-memory: it is
// never written to Vault or workflow history, so secret-reuse risk is nil
// regardless of maxenrollments. All referenced Vault paths (ca_tls_cert_ref,
// issuer_cert_ref) belong to this worker's own owner - no shard/caller path is
// ever read or written. Only public certificate material is returned.
func (a *Activities) SignCsr(ctx context.Context, req *pb.SignCsrInput) (*pb.SignCsrOutput, error) {
	caRef := req.GetCaRef()
	if caRef == nil {
		return nil, fmt.Errorf("%w: ca_ref is required", ErrInvalidActivityInput)
	}
	identityName := req.GetIdentityName()
	if identityName == "" {
		return nil, fmt.Errorf("%w: identity_name is required", ErrInvalidActivityInput)
	}
	if len(req.GetCsrPem()) == 0 {
		return nil, fmt.Errorf("%w: csr_pem is required", ErrInvalidActivityInput)
	}

	// Locally-referenced CAs always belong to this worker's injected owner
	caConfig := fabricconfig.NewCertificateAuthorityConfig(a.owner, caRef)

	// Admin client for register/modify. admin_identity selects the registrar:
	// nil = this CA's default permanent admin (steady state); a non-nil value
	// (e.g. the bootstrap admin registering the permanent admin) overrides it.
	adminClient, err := a.clientProvider.GetClient(ctx, NewConfig(caConfig, req.GetAdminIdentity()))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get CA client: %w", ErrRegisterIdentity, err)
	}

	// In-memory enrollment secret: never persisted, consumed by the immediately
	// following enroll. -1 keeps enrollments unbounded to preserve retry
	// idempotency (each retry yields a valid cert for the same CSR/CN).
	enrollmentSecret, err := secrets.GeneratePassword()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate enrollment secret: %w", ErrRegisterIdentity, err)
	}

	registerReq := &api.RegistrationRequest{
		Name:           identityName,
		Type:           req.GetIdentityType(),
		Attributes:     convertAttributes(req.GetAttributes()),
		Secret:         enrollmentSecret,
		MaxEnrollments: -1,
	}

	if _, err := adminClient.Register(ctx, registerReq); err != nil {
		if !isAlreadyRegistered(err) {
			return nil, fmt.Errorf("%w: registration failed: %w", ErrRegisterIdentity, err)
		}
		// Name already exists (409): reset its secret + attributes so we can
		// enroll our CSR. ModifyIdentity REPLACES the record's mutable fields.
		modifyReq := &api.ModifyIdentityRequest{
			ID:             identityName,
			Type:           req.GetIdentityType(),
			Attributes:     convertAttributes(req.GetAttributes()),
			MaxEnrollments: -1,
			Secret:         enrollmentSecret,
		}
		if _, err := adminClient.ModifyIdentity(ctx, modifyReq); err != nil {
			return nil, fmt.Errorf("%w: failed to reset enrollment secret: %w", ErrModifyIdentity, err)
		}
	}

	// Fetch the CA's own TLS cert for the mTLS connection to the CA server.
	var tlsCertPEM []byte
	if req.GetCaTlsCertRef() != nil {
		certSecret, err := a.secretProvider.FetchSecret(ctx, req.GetCaTlsCertRef())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch CA TLS certificate: %w", err)
		}
		cert, err := certSecret.Certificate()
		if err != nil {
			return nil, fmt.Errorf("failed to parse CA TLS certificate: %w", err)
		}
		tlsCertPEM = cert.Certificate
	}

	// Enroll the externally-generated CSR using the identity name + fresh secret.
	enrollClient, err := NewBootstrapClient(BootstrapClientConfig{
		URL:        caConfig.URL(),
		Username:   identityName,
		Password:   enrollmentSecret,
		CAName:     "",
		TLSCertPEM: tlsCertPEM,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create enroll client: %w", ErrEnrollIdentity, err)
	}
	defer enrollClient.Close()

	certPEM, caChainPEM, err := enrollClient.EnrollWithCSR(ctx, req.GetCsrPem(), req.GetProfile(), req.GetHosts())
	if err != nil {
		return nil, fmt.Errorf("%w: enrollment failed: %w", ErrEnrollIdentity, err)
	}

	// Build the full chain: [leaf] + [issuer's chain]. Prefer the issuer's stored
	// chain (its own certs/ path); fall back to the CA chain from enrollment.
	var issuerChain []byte
	if req.GetIssuerCertRef() != nil {
		issuerSecret, err := a.secretProvider.FetchSecret(ctx, req.GetIssuerCertRef())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issuer cert: %w", err)
		}
		if chainStr, ok := issuerSecret["chain"].(string); ok {
			issuerChain = []byte(chainStr)
		}
	}

	var fullChain []byte
	if len(issuerChain) > 0 {
		fullChain = buildCertificateChain(certPEM, issuerChain)
	} else {
		fullChain = buildCertificateChain(certPEM, caChainPEM)
	}

	return &pb.SignCsrOutput{
		CertificatePem: certPEM,
		ChainPem:       fullChain,
	}, nil
}

// isAlreadyRegistered reports whether err is Fabric CA's duplicate-registration
// error (HTTP 409, error code 74). The client surfaces server errors as a
// formatted string, so we match on both the stable code and the message.
func isAlreadyRegistered(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "is already registered") || strings.Contains(msg, "Error Code: 74")
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
