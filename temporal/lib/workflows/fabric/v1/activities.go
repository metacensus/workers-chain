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
	// clientProvider manages multiple CA clients, caching them by Vault path
	clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig]
	fabricConfig   config.FabricConfig
}

// NewActivities creates a new Activities instance with CA multi-client provider
func NewActivities(
	clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig],
	fabricConfig config.FabricConfig,
) *Activities {
	return &Activities{
		clientProvider: clientProvider,
		fabricConfig:   fabricConfig,
	}
}

// getCAClient builds a CA client for the specified cert type and CA ID.
// The client is cached by the provider using the Vault path as the key,
// so repeated calls with the same certType/caID return the cached client.
func (a *Activities) getCAClient(ctx context.Context, certType, caID string) (certauthority.Client, error) {
	// Build the CA config for this specific CA
	caConfig := a.fabricConfig.CertificateAuthority(certType, caID)

	// Wrap in hydrator for the provider
	hydrator := certauthority.NewConfig(caConfig)

	// Get the client (provider caches by Vault path, so this is efficient for repeated calls)
	client, err := a.clientProvider.GetClient(ctx, hydrator)
	if err != nil {
		return nil, fmt.Errorf("failed to get CA client for %s-%s: %w", certType, caID, err)
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
