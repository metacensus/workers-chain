package vault

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	StoreSecretName = "vault.StoreSecret"
	FetchSecretName = "vault.FetchSecret"
)

// Workflows contains all vault-related workflows
type Workflows struct {
	client VaultClient
}

// NewWorkflows creates a new Workflows instance
func NewWorkflows(client VaultClient) *Workflows {
	return &Workflows{
		client: client,
	}
}

// StoreSecretWorkflow stores secrets in Vault
func (w *Workflows) StoreSecretWorkflow(ctx workflow.Context, input StoreSecretInput) (*StoreSecretOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Vault secret storage workflow started",
		"name", input.Name,
		"version", input.Version)

	activities := NewActivities(w.client)

	var result StoreSecretOutput
	err := workflow.ExecuteActivity(ctx, activities.StoreSecretActivity, input).Get(ctx, &result)
	if err != nil {
		logger.Error("Vault secret storage activity failed", "error", err)
		return nil, fmt.Errorf("vault secret storage failed: %w", err)
	}

	logger.Info("Vault secret storage workflow completed",
		"name", result.Name,
		"version", result.Version)

	return &result, nil
}

// FetchSecretWorkflow fetches secrets from Vault
func (w *Workflows) FetchSecretWorkflow(ctx workflow.Context, input FetchSecretInput) (*FetchSecretOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Vault secret fetch workflow started",
		"name", input.Name)

	activities := NewActivities(w.client)

	var result FetchSecretOutput
	err := workflow.ExecuteActivity(ctx, activities.FetchSecretActivity, input).Get(ctx, &result)
	if err != nil {
		logger.Error("Vault secret fetch activity failed", "error", err)
		return nil, fmt.Errorf("vault secret fetch failed: %w", err)
	}

	logger.Info("Vault secret fetch workflow completed",
		"name", result.Name,
		"found", result.Found)

	return &result, nil
}
