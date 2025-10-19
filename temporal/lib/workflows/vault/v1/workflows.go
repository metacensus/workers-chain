package vault

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Workflows implements the VaultWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// StoreSecret implements the StoreSecret workflow
func (w *Workflows) StoreSecret(ctx workflow.Context, input *pb.StoreSecretWorkflowInput) (pb.StoreSecretWorkflow, error) {
	return &storeSecretWorkflow{
		input: input,
	}, nil
}

// FetchSecret implements the FetchSecret workflow
func (w *Workflows) FetchSecret(ctx workflow.Context, input *pb.FetchSecretWorkflowInput) (pb.FetchSecretWorkflow, error) {
	return &fetchSecretWorkflow{
		input: input,
	}, nil
}

// storeSecretWorkflow implements pb.StoreSecretWorkflow
type storeSecretWorkflow struct {
	input *pb.StoreSecretWorkflowInput
}

// Execute stores a secret in Vault
func (w *storeSecretWorkflow) Execute(ctx workflow.Context) (*pb.SecretRef, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("StoreSecret workflow started",
		"name", w.input.Req.GetName(),
		"namespace", w.input.Req.GetNamespace())

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

	var result *pb.SecretRef
	err := workflow.ExecuteActivity(ctx, pb.StoreSecretActivityName, w.input.Req).Get(ctx, &result)
	if err != nil {
		logger.Error("StoreSecret activity failed", "error", err)
		return nil, fmt.Errorf("failed to store secret: %w", err)
	}

	logger.Info("StoreSecret workflow completed",
		"name", result.GetName(),
		"version", result.GetVersion())

	return result, nil
}

// fetchSecretWorkflow implements pb.FetchSecretWorkflow
type fetchSecretWorkflow struct {
	input *pb.FetchSecretWorkflowInput
}

// Execute fetches a secret from Vault
func (w *fetchSecretWorkflow) Execute(ctx workflow.Context) (*pb.Secret, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("FetchSecret workflow started",
		"name", w.input.Req.GetName(),
		"namespace", w.input.Req.GetNamespace())

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

	var result *pb.Secret
	err := workflow.ExecuteActivity(ctx, pb.FetchSecretActivityName, w.input.Req).Get(ctx, &result)
	if err != nil {
		logger.Error("FetchSecret activity failed", "error", err)
		return nil, fmt.Errorf("failed to fetch secret: %w", err)
	}

	logger.Info("FetchSecret workflow completed",
		"name", result.GetName(),
		"found", result.GetData() != nil)

	return result, nil
}
