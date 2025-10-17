package v1

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ==================== Core Workflows ====================
// These workflows represent atomic operations that directly interact with Fabric CA

// registerIdentityWorkflow implements pb.RegisterIdentityWorkflow
type registerIdentityWorkflow struct {
	input *pb.RegisterIdentityWorkflowInput
}

// Execute registers a new identity with Fabric CA
func (w *registerIdentityWorkflow) Execute(ctx workflow.Context) (*pb.RegisterIdentityOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("RegisterIdentity workflow started",
		"name", w.input.Req.GetName(),
		"type", w.input.Req.GetType())

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result *pb.RegisterIdentityOutput
	err := workflow.ExecuteActivity(ctx, pb.RegisterIdentityActivityName, w.input.Req).Get(ctx, &result)
	if err != nil {
		logger.Error("RegisterIdentity activity failed", "error", err)
		return nil, fmt.Errorf("failed to register identity: %w", err)
	}

	logger.Info("RegisterIdentity workflow completed",
		"name", result.GetName(),
		"type", result.GetType())

	return result, nil
}

// enrollIdentityWorkflow implements pb.EnrollIdentityWorkflow
type enrollIdentityWorkflow struct {
	input *pb.EnrollIdentityWorkflowInput
}

// Execute enrolls an identity to obtain certificates
func (w *enrollIdentityWorkflow) Execute(ctx workflow.Context) (*pb.EnrollIdentityOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("EnrollIdentity workflow started",
		"name", w.input.Req.GetName(),
		"profile", w.input.Req.GetProfile())

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result *pb.EnrollIdentityOutput
	err := workflow.ExecuteActivity(ctx, pb.EnrollIdentityActivityName, w.input.Req).Get(ctx, &result)
	if err != nil {
		logger.Error("EnrollIdentity activity failed", "error", err)
		return nil, fmt.Errorf("failed to enroll identity: %w", err)
	}

	logger.Info("EnrollIdentity workflow completed",
		"name", w.input.Req.GetName(),
		"cert_present", len(result.GetCertificate()) > 0)

	return result, nil
}

// assembleMspWorkflow implements pb.AssembleMspWorkflow
type assembleMspWorkflow struct {
	input *pb.AssembleMspWorkflowInput
}

// Execute assembles a complete MSP structure from enrollments
func (w *assembleMspWorkflow) Execute(ctx workflow.Context) (*pb.AssembleMspOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("AssembleMsp workflow started")

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

	var result *pb.AssembleMspOutput
	err := workflow.ExecuteActivity(ctx, pb.AssembleMspActivityName, w.input.Req).Get(ctx, &result)
	if err != nil {
		logger.Error("AssembleMsp activity failed", "error", err)
		return nil, fmt.Errorf("failed to assemble MSP: %w", err)
	}

	logger.Info("AssembleMsp workflow completed")
	return result, nil
}
