package temporal

import (
	"errors"
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors for temporal workflows
var (
	ErrRegisterNamespaceWorkflow  = errors.New("register namespace workflow failed")
	ErrNamespaceActivationTimeout = errors.New("namespace activation timeout")
)

// Workflows implements the AdminWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// CreateNamespace implements the AdminWorkflows.CreateNamespace method
// Returns a workflow instance that will be executed by Temporal
func (w *Workflows) CreateNamespace(ctx workflow.Context, input *pb.CreateNamespaceWorkflowInput) (pb.CreateNamespaceWorkflow, error) {
	return &createNamespaceWorkflow{
		input: input,
	}, nil
}

// createNamespaceWorkflow implements pb.CreateNamespaceWorkflow
type createNamespaceWorkflow struct {
	input *pb.CreateNamespaceWorkflowInput
}

// Execute calls RegisterNamespace, then polls DescribeNamespace until REGISTERED
func (w *createNamespaceWorkflow) Execute(ctx workflow.Context) (*pb.CreateNamespaceOutput, error) {
	logger := workflow.GetLogger(ctx)
	req := w.input.Req

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

	if err := workflow.ExecuteActivity(ctx, pb.RegisterNamespaceActivityName, &pb.RegisterNamespaceInput{
		Name:      req.Name,
		Retention: req.Retention,
	}).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRegisterNamespaceWorkflow, err)
	}

	// Poll for namespace activation (Temporal's namespace creation is async)
	pollOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
	}
	pollCtx := workflow.WithActivityOptions(ctx, pollOptions)
	pollDeadline := workflow.Now(ctx).Add(2 * time.Minute)
	pollInput := &pb.DescribeNamespaceInput{Name: req.Name}

	for workflow.Now(ctx).Before(pollDeadline) {
		var describeOutput *pb.DescribeNamespaceOutput

		if err := workflow.ExecuteActivity(pollCtx, pb.DescribeNamespaceActivityName, pollInput).Get(pollCtx, &describeOutput); err != nil {
			logger.Warn("Describe namespace failed, retrying",
				"namespace", req.Name,
				"error", err)
		} else if describeOutput != nil && describeOutput.State == pb.NamespaceState_REGISTERED {
			logger.Info("Namespace active and ready",
				"namespace", req.Name)
			return &pb.CreateNamespaceOutput{
				Name: req.Name,
			}, nil
		}

		if err := workflow.Sleep(ctx, 1*time.Second); err != nil {
			return nil, fmt.Errorf("workflow interrupted: %w", err)
		}
	}

	return nil, fmt.Errorf("%w: namespace %s did not reach REGISTERED state within timeout", ErrNamespaceActivationTimeout, req.Name)
}
