package secrets

import (
	"errors"
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors
var (
	ErrCreateAccessPolicy   = errors.New("create access policy")
	ErrCreateAccessAuthRole = errors.New("create access auth role")
)

// Workflows implements the SecretsWorkflows interface
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// CreateAccess implements the CreateAccess workflow
func (w *Workflows) CreateAccess(ctx workflow.Context, input *pb.CreateAccessWorkflowInput) (pb.CreateAccessWorkflow, error) {
	return &createAccessWorkflow{input: input}, nil
}

type createAccessWorkflow struct {
	input *pb.CreateAccessWorkflowInput
}

// Execute orchestrates the creation of a Vault policy and Kubernetes auth role
func (w *createAccessWorkflow) Execute(ctx workflow.Context) (*pb.CreateAccessOutput, error) {
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

	req := w.input.Req

	// Step 1: Create the policy
	var policyOutput *vaultpb.CreatePolicyOutput
	err := workflow.ExecuteActivity(activityCtx, vaultpb.CreatePolicyActivityName, &vaultpb.CreatePolicyInput{
		Name:     req.GetPolicyName(),
		Document: req.GetPolicyDocument(),
	}).Get(ctx, &policyOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAccessPolicy, err)
	}

	// Step 2: Create the auth role using the policy
	var authRoleOutput *vaultpb.CreateAuthRoleOutput
	err = workflow.ExecuteActivity(activityCtx, vaultpb.CreateAuthRoleActivityName, &vaultpb.CreateAuthRoleInput{
		RoleName:            req.GetAuthRoleName(),
		ServiceAccountName:  req.GetServiceAccountName(),
		KubernetesNamespace: req.GetKubernetesNamespace(),
		PolicyNames:         []string{policyOutput.GetName()},
	}).Get(ctx, &authRoleOutput)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAccessAuthRole, err)
	}

	return &pb.CreateAccessOutput{
		PolicyName:   policyOutput.GetName(),
		AuthRoleName: authRoleOutput.GetName(),
	}, nil
}

