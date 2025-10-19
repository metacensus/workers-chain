package identity

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"

	"github.com/cloudflare/cfssl/csr"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-ca/util"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
)

// FabricCAConfig represents Fabric CA configuration
type FabricCAConfig struct {
	URL     string // CA server URL (e.g., "http://localhost:7054")
	CAName  string // CA name (optional, for multi-CA servers)
	HomeDir string // Client home directory for state (e.g., "/tmp/fabric-ca-client")
}

// FabricCAClient defines the interface for Fabric CA operations
type FabricCAClient interface {
	// Register registers a new identity with the CA
	Register(ctx context.Context, input *pb.RegisterIdentityInput) (*pb.RegisterIdentityOutput, error)

	// Enroll enrolls an identity to obtain certificates
	Enroll(ctx context.Context, input *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error)

	// Revoke revokes an identity's certificate
	Revoke(ctx context.Context, name string) error

	// CreateAffiliation creates an affiliation in the CA
	CreateAffiliation(ctx context.Context, name string) error

	// GetCAInfo retrieves CA information (CA certificate chain)
	GetCAInfo(ctx context.Context) ([]byte, error)

	// GenerateSecret generates a cryptographically secure random secret of the specified length
	GenerateSecret(length int32) (string, error)
}

// fabricCAClient implements FabricCAClient
type fabricCAClient struct {
	config FabricCAConfig
	client *lib.Client
}

// NewFabricCAClient creates a new Fabric CA client with the given configuration
// This is the constructor that fx will use to provide FabricCAClient
func NewFabricCAClient(config FabricCAConfig) (FabricCAClient, error) {
	client := &lib.Client{
		Config: &lib.ClientConfig{
			URL: config.URL,
		},
		HomeDir: config.HomeDir,
	}

	if err := client.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize Fabric CA client: %w", err)
	}

	return &fabricCAClient{
		config: config,
		client: client,
	}, nil
}

// Register registers a new identity with Fabric CA
func (c *fabricCAClient) Register(ctx context.Context, input *pb.RegisterIdentityInput) (*pb.RegisterIdentityOutput, error) {
	// Convert proto attributes to Fabric CA format
	var attributes []api.Attribute
	for _, attr := range input.GetAttributes() {
		attributes = append(attributes, api.Attribute{
			Name:  attr.GetName(),
			Value: attr.GetValue(),
			ECert: attr.GetEcert(),
		})
	}

	// Create registration request
	req := &api.RegistrationRequest{
		Name:           input.GetName(),
		Type:           input.GetType(),
		Secret:         input.GetSecret(),
		Affiliation:    input.GetAffiliation(),
		Attributes:     attributes,
		MaxEnrollments: int(input.GetMaxEnrollments()),
		CAName:         c.config.CAName,
	}

	// Note: We need an admin identity to register. This should be handled by
	// the calling activity which will enroll as admin first.
	// For now, we'll use the default admin credentials to get an identity.
	adminReq := &api.EnrollmentRequest{
		Name:   "admin",
		Secret: "adminpw", // TODO: Make this configurable
		CAName: c.config.CAName,
	}

	adminResp, err := c.client.Enroll(adminReq)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll as admin: %w", err)
	}

	// Register the new identity using admin identity
	resp, err := adminResp.Identity.Register(req)
	if err != nil {
		return nil, fmt.Errorf("failed to register identity: %w", err)
	}

	// Build output preserving request params + CA response
	output := &pb.RegisterIdentityOutput{
		Name:        input.GetName(),
		Type:        input.GetType(),
		Secret:      resp.Secret, // CA-generated or provided secret
		Affiliation: input.GetAffiliation(),
		Attributes:  input.GetAttributes(),
	}

	return output, nil
}

// Enroll enrolls an identity with Fabric CA
func (c *fabricCAClient) Enroll(ctx context.Context, input *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
	// Convert CSRInfo from proto to Fabric CA format
	csrInfo := c.protoCSRToFabricCSR(input.GetCsr())

	// Generate CSR and capture the private key (MUST be done before enrollment)
	_, key, err := c.client.GenCSR(csrInfo, input.GetName())
	if err != nil {
		return nil, fmt.Errorf("failed to generate CSR: %w", err)
	}

	// Convert private key to PEM format
	keyBytes, err := util.PrivateKeyToPEM(key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert private key to PEM: %w", err)
	}

	// Create enrollment request
	req := &api.EnrollmentRequest{
		Name:    input.GetName(),
		Secret:  input.GetSecret(),
		Profile: input.GetProfile(),
		CSR:     csrInfo,
		CAName:  c.config.CAName,
	}

	// Add attribute requests if provided
	if len(input.GetAttrReqs()) > 0 {
		req.AttrReqs = make([]*api.AttributeRequest, 0, len(input.GetAttrReqs()))
		for _, attrReq := range input.GetAttrReqs() {
			req.AttrReqs = append(req.AttrReqs, &api.AttributeRequest{
				Name:     attrReq.GetName(),
				Optional: attrReq.GetOptional(),
			})
		}
	}

	// Enroll the identity
	resp, err := c.client.Enroll(req)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity: %w", err)
	}

	// Extract certificate from response
	identity := resp.Identity
	signer := identity.GetECert()
	if signer == nil {
		return nil, fmt.Errorf("no signer found in enrollment response")
	}

	cert := signer.GetX509Cert()
	if cert == nil {
		return nil, fmt.Errorf("no certificate found in signer")
	}

	// Marshal certificate to DER then to PEM-like bytes
	certBytes := cert.Raw

	// Get CA certificate chain
	caCertChain := resp.CAInfo.CAChain

	output := &pb.EnrollIdentityOutput{
		Certificate: certBytes,
		PrivateKey:  keyBytes,
		CaCert:      caCertChain,
		Profile:     input.GetProfile(),
	}

	return output, nil
}

// Revoke revokes an identity's certificate
func (c *fabricCAClient) Revoke(ctx context.Context, name string) error {
	// First enroll as admin to get identity for revocation
	adminReq := &api.EnrollmentRequest{
		Name:   "admin",
		Secret: "adminpw", // TODO: Make this configurable
		CAName: c.config.CAName,
	}

	adminResp, err := c.client.Enroll(adminReq)
	if err != nil {
		return fmt.Errorf("failed to enroll as admin for revocation: %w", err)
	}

	// Revoke the identity
	revokeReq := &api.RevocationRequest{
		Name:   name,
		CAName: c.config.CAName,
	}

	_, err = adminResp.Identity.Revoke(revokeReq)
	if err != nil {
		return fmt.Errorf("failed to revoke identity: %w", err)
	}

	return nil
}

// CreateAffiliation creates an affiliation in the CA
func (c *fabricCAClient) CreateAffiliation(ctx context.Context, name string) error {
	// First enroll as admin
	adminReq := &api.EnrollmentRequest{
		Name:   "admin",
		Secret: "adminpw", // TODO: Make this configurable
		CAName: c.config.CAName,
	}

	adminResp, err := c.client.Enroll(adminReq)
	if err != nil {
		return fmt.Errorf("failed to enroll as admin for affiliation creation: %w", err)
	}

	// Create affiliation request
	affiliationReq := &api.AddAffiliationRequest{
		Name:   name,
		Force:  false, // Don't force if parent doesn't exist
		CAName: c.config.CAName,
	}

	_, err = adminResp.Identity.AddAffiliation(affiliationReq)
	if err != nil {
		return fmt.Errorf("failed to create affiliation: %w", err)
	}

	return nil
}

// GetCAInfo retrieves CA certificate chain
func (c *fabricCAClient) GetCAInfo(ctx context.Context) ([]byte, error) {
	req := &api.GetCAInfoRequest{
		CAName: c.config.CAName,
	}

	resp, err := c.client.GetCAInfo(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get CA info: %w", err)
	}

	return resp.CAChain, nil
}

// protoCSRToFabricCSR converts proto CSRInfo to Fabric CA CSRInfo
func (c *fabricCAClient) protoCSRToFabricCSR(proto *pb.CSRInfo) *api.CSRInfo {
	if proto == nil {
		return nil
	}

	csrInfo := &api.CSRInfo{
		CN:    proto.GetCn(),
		Hosts: proto.GetHosts(),
	}

	// Convert CSRName array
	if len(proto.GetNames()) > 0 {
		csrInfo.Names = make([]csr.Name, 0, len(proto.GetNames()))
		for _, name := range proto.GetNames() {
			csrInfo.Names = append(csrInfo.Names, csr.Name{
				C:  name.GetC(),
				ST: name.GetSt(),
				L:  name.GetL(),
				O:  name.GetO(),
				OU: name.GetOu(),
			})
		}
	}

	// Convert KeyRequest
	if keyReq := proto.GetKey(); keyReq != nil {
		csrInfo.KeyRequest = &api.KeyRequest{
			Algo: keyReq.GetAlgo(),
			Size: int(keyReq.GetSize()),
		}
	}

	return csrInfo
}

// GenerateSecret generates a cryptographically secure random secret
func (c *fabricCAClient) GenerateSecret(length int32) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid length: %d", length)
	}

	// Use Fabric CA's util package for secure random generation
	// Note: util.RandomString returns only the string (no error)
	secret := util.RandomString(int(length))

	return secret, nil
}

// Helper to convert ECDSA private key (commonly used in Fabric)
func privateKeyToPEM(key interface{}) ([]byte, error) {
	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		keyBytes, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal EC private key: %w", err)
		}
		return keyBytes, nil
	default:
		return util.PrivateKeyToPEM(key, nil)
	}
}
