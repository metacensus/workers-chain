package vault

import (
	"context"
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/activity"
)

// Activities implements the VaultActivities interface from generated code
type Activities struct {
	client VaultClient
}

// NewActivities creates a new Activities instance
func NewActivities(client VaultClient) *Activities {
	return &Activities{
		client: client,
	}
}

// StoreSecret stores a secret in Vault
func (a *Activities) StoreSecret(ctx context.Context, req *pb.Secret) (*pb.SecretRef, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Storing secret",
		"name", req.GetName(),
		"namespace", req.GetNamespace(),
		"version", req.GetVersion())

	result, err := a.client.StoreSecret(ctx, req)
	if err != nil {
		logger.Error("Failed to store secret", "error", err)
		return nil, fmt.Errorf("vault client error: %w", err)
	}

	logger.Info("Secret stored successfully",
		"name", result.GetName(),
		"version", result.GetVersion())

	return result, nil
}

// FetchSecret retrieves a secret from Vault
func (a *Activities) FetchSecret(ctx context.Context, req *pb.SecretRef) (*pb.Secret, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Fetching secret",
		"name", req.GetName(),
		"namespace", req.GetNamespace())

	result, err := a.client.FetchSecret(ctx, req)
	if err != nil {
		logger.Error("Failed to fetch secret", "error", err)
		return nil, fmt.Errorf("vault client error: %w", err)
	}

	logger.Info("Secret fetched successfully",
		"name", result.GetName(),
		"data_present", len(result.GetData()) > 0)

	return result, nil
}
