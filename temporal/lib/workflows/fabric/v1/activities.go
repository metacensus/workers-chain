package fabric

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	"github.com/metacensus/infra/temporal/lib/workflows/fabric/v1/certauthority"
)

// Sentinel errors for fabric activities
var (
	ErrRegisterIdentity = errors.New("register identity")
	ErrEnrollIdentity   = errors.New("enroll identity")
	ErrRevokeIdentity   = errors.New("revoke identity")
	ErrInvalidInput     = errors.New("invalid input")
)

// Activities implements the Fabric activities
type Activities struct {
	// vaultClient is used to fetch credentials for creating BootstrapClients on-demand
	vaultClient common.VaultClient

	// clientProvider manages cached Clients for Register/Revoke operations
	clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig]

	fabricConfig config.FabricConfig
}

// NewActivities creates a new Activities instance
func NewActivities(
	vaultClient common.VaultClient,
	clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig],
	fabricConfig config.FabricConfig,
) *Activities {
	return &Activities{
		vaultClient:    vaultClient,
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
func (a *Activities) getBootstrapClient(ctx context.Context, certType, caID string) (certauthority.BootstrapClient, error) {
	caConfig := a.fabricConfig.CertificateAuthority(certType, caID)
	hydrator := certauthority.NewBootstrapConfig(caConfig)

	// Fetch credentials from Vault
	secretData, err := a.vaultClient.FetchSecret(ctx, hydrator.GetVaultPath())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bootstrap credentials for %s-%s: %w", certType, caID, err)
	}

	// Hydrate config with credentials
	clientConfig, err := hydrator.Hydrate(secretData)
	if err != nil {
		return nil, fmt.Errorf("failed to hydrate bootstrap config for %s-%s: %w", certType, caID, err)
	}

	// Create the client
	client, err := certauthority.NewBootstrapClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create bootstrap client for %s-%s: %w", certType, caID, err)
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
