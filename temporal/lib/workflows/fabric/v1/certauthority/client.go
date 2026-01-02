package certauthority

import (
	"context"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
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

// ClientConfig represents the configuration needed to connect to a Fabric CA
type ClientConfig struct {
	URL      string // CA server URL (e.g., "https://ca.fabric-org1.svc.cluster.local:7054")
	Username string // Admin username for authentication
	Password string // Admin password for authentication
	CAName   string // Optional CA name (for multi-CA servers)
}

// Validate checks if the ClientConfig has all required fields
func (c *ClientConfig) Validate() error {
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

// Client defines the interface for Fabric CA operations
// This interface wraps the fabric-ca lib.Client for testability
type Client interface {
	// Register registers a new identity with the CA and returns an enrollment secret
	Register(ctx context.Context, req *api.RegistrationRequest) (*api.RegistrationResponse, error)

	// Enroll enrolls an identity using an enrollment secret and returns certificates
	Enroll(ctx context.Context, req *api.EnrollmentRequest) (*lib.EnrollmentResponse, error)

	// Revoke revokes an identity from the CA
	Revoke(ctx context.Context, req *api.RevocationRequest) (*api.RevocationResponse, error)

	// GetCAInfo retrieves information about the CA (certificate chain, etc.)
	GetCAInfo(ctx context.Context) (*lib.GetCAInfoResponse, error)
}

// client implements Client using the forked fabric-ca lib.Client
// This wrapper provides a clean interface for testing and follows our patterns
type client struct {
	fabricClient  *lib.Client
	config        ClientConfig
	adminIdentity *lib.Identity // Enrolled admin identity used for Register/Revoke operations
}

// NewClient creates a new Fabric CA client with the given configuration
// The client is initialized and ready to use
// Note: The admin is automatically enrolled using the provided credentials
func NewClient(config ClientConfig) (Client, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create fabric-ca ClientConfig
	fabricCAConfig := &lib.ClientConfig{
		URL:    config.URL,
		CAName: config.CAName,
		// MSPDir is intentionally empty - we use in-memory operations only
		MSPDir: "",
	}

	// Create fabric-ca client
	// HomeDir is empty since we're doing in-memory operations only
	fabricCAClient := &lib.Client{
		HomeDir: "",
		Config:  fabricCAConfig,
	}

	// Initialize the client
	if err := fabricCAClient.Init(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrClientInit, err)
	}

	// Enroll the admin identity immediately so we can perform Register/Revoke operations
	// This uses the username/password from the config
	enrollReq := &api.EnrollmentRequest{
		Name:   config.Username,
		Secret: config.Password,
		CAName: config.CAName,
	}

	enrollResp, err := fabricCAClient.Enroll(enrollReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to enroll admin identity: %v", ErrClientInit, err)
	}

	if enrollResp == nil || enrollResp.Identity == nil {
		return nil, fmt.Errorf("%w: admin enrollment returned nil identity", ErrClientInit)
	}

	return &client{
		fabricClient:  fabricCAClient,
		config:        config,
		adminIdentity: enrollResp.Identity,
	}, nil
}

// Register registers a new identity with the CA
// This requires an enrolled admin identity
func (c *client) Register(ctx context.Context, req *api.RegistrationRequest) (*api.RegistrationResponse, error) {
	if c.adminIdentity == nil {
		return nil, fmt.Errorf("%w: admin identity not enrolled", ErrRegister)
	}

	// Perform registration using the admin identity
	resp, err := c.adminIdentity.Register(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRegister, err)
	}

	return resp, nil
}

// Enroll enrolls an identity using an enrollment secret
func (c *client) Enroll(ctx context.Context, req *api.EnrollmentRequest) (*lib.EnrollmentResponse, error) {
	// Perform enrollment
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
// This requires an enrolled admin identity
func (c *client) Revoke(ctx context.Context, req *api.RevocationRequest) (*api.RevocationResponse, error) {
	if c.adminIdentity == nil {
		return nil, fmt.Errorf("%w: admin identity not enrolled", ErrRevoke)
	}

	// Perform revocation using the admin identity
	resp, err := c.adminIdentity.Revoke(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRevoke, err)
	}

	return resp, nil
}

// GetCAInfo retrieves information about the CA
func (c *client) GetCAInfo(ctx context.Context) (*lib.GetCAInfoResponse, error) {
	// Get CA info
	resp, err := c.fabricClient.GetCAInfo(&api.GetCAInfoRequest{
		CAName: c.config.CAName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get CA info: %w", err)
	}

	return resp, nil
}

