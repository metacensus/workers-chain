package vault

import (
	"errors"
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors for vault workflows
var (
	ErrCreatePolicyWorkflow   = errors.New("create policy workflow failed")
	ErrCreateAuthRoleWorkflow = errors.New("create auth role workflow failed")
)

// Workflows implements the VaultWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// CreateAccess implements the CreateAccess workflow
func (w *Workflows) CreateAccess(ctx workflow.Context, input *pb.CreateAccessWorkflowInput) (pb.CreateAccessWorkflow, error) {
	return &createAccessWorkflow{input: input}, nil
}

// createAccessWorkflow implements pb.CreateAccessWorkflow
type createAccessWorkflow struct {
	input *pb.CreateAccessWorkflowInput
}

// Execute orchestrates the creation of a Vault policy and Kubernetes auth role
func (w *createAccessWorkflow) Execute(ctx workflow.Context) (*pb.CreateAccessOutput, error) {
	// Set up activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Step 1: Create the policy
	var policyOutput *pb.CreatePolicyOutput
	err := workflow.ExecuteActivity(activityCtx, pb.CreatePolicyActivityName,
		&pb.CreatePolicyInput{
			Name:     w.input.Req.GetPolicyName(),
			Document: w.input.Req.GetPolicyDocument(),
		}).Get(ctx, &policyOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreatePolicyWorkflow, err)
	}

	// Step 2: Create the auth role using the policy name from step 1
	var authRoleOutput *pb.CreateAuthRoleOutput
	err = workflow.ExecuteActivity(activityCtx, pb.CreateAuthRoleActivityName,
		&pb.CreateAuthRoleInput{
			RoleName:            w.input.Req.GetAuthRoleName(),
			ServiceAccountName:  w.input.Req.GetServiceAccountName(),
			KubernetesNamespace: w.input.Req.GetKubernetesNamespace(),
			PolicyNames:         []string{policyOutput.GetName()},
		}).Get(ctx, &authRoleOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAuthRoleWorkflow, err)
	}

	return &pb.CreateAccessOutput{
		PolicyName:   policyOutput.GetName(),
		AuthRoleName: authRoleOutput.GetName(),
	}, nil
}

// TestCreateAccessWorkflow is a test helper function that allows testing the workflow
// from external test packages (e.g., vault_test)
func TestCreateAccessWorkflow(ctx workflow.Context, input *pb.CreateAccessWorkflowInput) (*pb.CreateAccessOutput, error) {
	w := &createAccessWorkflow{input: input}
	return w.Execute(ctx)
}
