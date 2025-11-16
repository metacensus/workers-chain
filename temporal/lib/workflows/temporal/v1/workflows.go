package temporal

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// CreateNamespaceWorkflow orchestrates async namespace creation via RegisterNamespace
// then polls DescribeNamespace until the namespace reaches REGISTERED state
type CreateNamespaceWorkflow struct {
	activities pb.AdminActivities
}

// NewCreateNamespaceWorkflow creates a new workflow instance
func NewCreateNamespaceWorkflow(activities pb.AdminActivities) *CreateNamespaceWorkflow {
	return &CreateNamespaceWorkflow{
		activities: activities,
	}
}

// Execute calls RegisterNamespace, then polls DescribeNamespace until REGISTERED
func (w *CreateNamespaceWorkflow) Execute(ctx workflow.Context, input *pb.CreateNamespaceInput) (*pb.CreateNamespaceOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CreateNamespace workflow started",
		"namespace", input.Name)

	// Register namespace with activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger.Info("Registering namespace", "namespace", input.Name)
	if err := workflow.ExecuteActivity(ctx, w.activities.RegisterNamespace, &pb.RegisterNamespaceInput{
		Name:      input.Name,
		Retention: input.Retention,
	}).Get(ctx, nil); err != nil {
		logger.Error("Failed to register namespace", "error", err)
		return nil, fmt.Errorf("register namespace: %w", err)
	}

	// Poll for namespace activation (Temporal's namespace creation is async)
	pollOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
	}
	pollCtx := workflow.WithActivityOptions(ctx, pollOptions)
	pollDeadline := workflow.Now(ctx).Add(2 * time.Minute)
	pollInput := &pb.DescribeNamespaceInput{Name: input.Name}

	for workflow.Now(ctx).Before(pollDeadline) {
		var describeOutput *pb.DescribeNamespaceOutput

		if err := workflow.ExecuteActivity(pollCtx, w.activities.DescribeNamespace, pollInput).Get(pollCtx, &describeOutput); err != nil {
			logger.Warn("Describe namespace failed, retrying",
				"namespace", input.Name,
				"error", err)
		} else if describeOutput != nil && describeOutput.State == pb.NamespaceState_REGISTERED {
			logger.Info("Namespace active and ready",
				"namespace", input.Name)
			return &pb.CreateNamespaceOutput{
				Name: input.Name,
			}, nil
		}

		if err := workflow.Sleep(ctx, 1*time.Second); err != nil {
			return nil, fmt.Errorf("workflow interrupted: %w", err)
		}
	}

	logger.Error("Timeout waiting for namespace to activate",
		"namespace", input.Name)

	return nil, fmt.Errorf("namespace %s did not reach REGISTERED state within timeout", input.Name)
}
