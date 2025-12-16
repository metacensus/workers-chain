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
	logger := workflow.GetLogger(ctx)
	logger.Info("CreateAccess workflow started",
		"policy_name", w.input.Req.GetPolicyName(),
		"organization", w.input.Req.GetOrganization(),
		"service_account", w.input.Req.GetServiceAccountName())

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
	logger.Info("Creating Vault policy", "policy_name", w.input.Req.GetPolicyName())
	var policyOutput *pb.CreatePolicyOutput
	err := workflow.ExecuteActivity(activityCtx, pb.CreatePolicyActivityName,
		&pb.CreatePolicyInput{
			Name:     w.input.Req.GetPolicyName(),
			Document: w.input.Req.GetPolicyDocument(),
		}).Get(ctx, &policyOutput)
	if err != nil {
		logger.Error("Failed to create policy", "error", err)
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	// Step 2: Create the auth role using the policy name from step 1
	logger.Info("Creating Kubernetes auth role", "policy_name", policyOutput.GetName())
	
	// Build the list of policy names: include the created policy plus any additional ones
	policyNames := []string{policyOutput.GetName()}
	policyNames = append(policyNames, w.input.Req.GetAdditionalPolicyNames()...)

	var authRoleOutput *pb.CreateAuthRoleOutput
	err = workflow.ExecuteActivity(activityCtx, pb.CreateAuthRoleActivityName,
		&pb.CreateAuthRoleInput{
			Organization:        w.input.Req.GetOrganization(),
			ServiceAccountName:  w.input.Req.GetServiceAccountName(),
			KubernetesNamespace: w.input.Req.GetKubernetesNamespace(),
			PolicyNames:         policyNames,
		}).Get(ctx, &authRoleOutput)
	if err != nil {
		logger.Error("Failed to create auth role", "error", err)
		return nil, fmt.Errorf("failed to create auth role: %w", err)
	}

	logger.Info("CreateAccess workflow completed successfully",
		"policy_name", policyOutput.GetName(),
		"auth_role_name", authRoleOutput.GetName())

	return &pb.CreateAccessOutput{
		PolicyName:    policyOutput.GetName(),
		AuthRoleName:  authRoleOutput.GetName(),
	}, nil
}

// TestCreateAccessWorkflow is a test helper function that allows testing the workflow
// from external test packages (e.g., vault_test)
func TestCreateAccessWorkflow(ctx workflow.Context, input *pb.CreateAccessWorkflowInput) (*pb.CreateAccessOutput, error) {
	w := &createAccessWorkflow{input: input}
	return w.Execute(ctx)
}

