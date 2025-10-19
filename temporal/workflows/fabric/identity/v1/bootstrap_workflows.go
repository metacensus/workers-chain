package identity

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ==================== Bootstrap Workflows ====================
// These workflows handle initial setup and admin identity creation

// bootstrapAdminWorkflow implements pb.BootstrapAdminWorkflow
type bootstrapAdminWorkflow struct {
	input *pb.BootstrapAdminWorkflowInput
}

// Execute creates a new admin identity
func (w *bootstrapAdminWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapAdminOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("BootstrapAdmin workflow started",
		"admin_id", w.input.Req.GetAdminId(),
		"vault_namespace", w.input.Req.GetVaultNamespace())

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Generate secure secret for new admin if not provided
	adminSecret := w.input.Req.GetAdminSecret()
	if adminSecret == "" {
		logger.Info("Generating secure secret for admin")
		var secretOutput *pb.GenerateSecureSecretOutput
		err := workflow.ExecuteActivity(ctx, pb.GenerateSecureSecretActivityName, &pb.GenerateSecureSecretInput{
			Length: 32,
		}).Get(ctx, &secretOutput)
		if err != nil {
			logger.Error("Failed to generate admin secret", "error", err)
			return nil, fmt.Errorf("failed to generate admin secret: %w", err)
		}
		adminSecret = secretOutput.GetSecret()
	}

	// Step 2: Create the new admin identity
	logger.Info("Creating admin identity")
	var adminOutput *pb.CreateAdminIdentityOutput
	err := workflow.ExecuteActivity(ctx, pb.CreateAdminIdentityActivityName, &pb.CreateAdminIdentityInput{
		AdminId:        w.input.Req.GetAdminId(),
		AdminSecret:    adminSecret,
		DefaultAdminId: w.input.Req.GetDefaultAdminId(),
		DefaultSecret:  w.input.Req.GetDefaultSecret(),
		Attributes:     w.input.Req.GetAttributes(),
	}).Get(ctx, &adminOutput)
	if err != nil {
		logger.Error("Failed to create admin identity", "error", err)
		return nil, fmt.Errorf("failed to create admin identity: %w", err)
	}

	logger.Info("BootstrapAdmin workflow completed", "admin_id", adminOutput.GetAdminIdentity().GetId())
	return &pb.BootstrapAdminOutput{
		AdminIdentity: adminOutput.GetAdminIdentity(),
		Success:       true,
		Message:       "Admin identity created successfully",
	}, nil
}
