package secrets

import (
	"errors"
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors
var (
	ErrCreateAccessPolicy   = errors.New("create access policy")
	ErrCreateAccessAuthRole = errors.New("create access auth role")
	ErrInvalidInput         = errors.New("invalid input")
	ErrPolicyOutOfScope     = errors.New("policy path outside the target namespace's Vault subtree")
)

// Workflows implements the SecretsWorkflows interface
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// CreateAccess implements the CreateAccess workflow
func (w *Workflows) CreateAccess(ctx workflow.Context, input *pb.CreateAccessWorkflowInput) (pb.CreateAccessWorkflow, error) {
	return newCreateAccessWorkflow(input)
}

type createAccessWorkflow struct {
	input *pb.CreateAccessWorkflowInput
}

// newCreateAccessWorkflow creates and validates a new createAccessWorkflow
func newCreateAccessWorkflow(input *pb.CreateAccessWorkflowInput) (*createAccessWorkflow, error) {
	w := &createAccessWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate scopes policy paths to the k8s namespace being bound (Nexus
// handlers cannot identify the caller).
func (w *createAccessWorkflow) validate() error {
	req := w.input.Req
	if req.GetPolicyName() == "" {
		return fmt.Errorf("%w: policy_name is required", ErrInvalidInput)
	}
	if req.GetAuthRoleName() == "" {
		return fmt.Errorf("%w: auth_role_name is required", ErrInvalidInput)
	}
	if req.GetServiceAccountName() == "" {
		return fmt.Errorf("%w: service_account_name is required", ErrInvalidInput)
	}
	namespace := req.GetKubernetesNamespace()
	if namespace == "" {
		return fmt.Errorf("%w: kubernetes_namespace is required", ErrInvalidInput)
	}
	if !config.IsOwnerKubernetesNamespace(namespace) {
		return fmt.Errorf("%w: %q owns no Vault subtree", ErrPolicyOutOfScope, namespace)
	}

	for _, rule := range req.GetPolicyDocument().GetRules() {
		if !config.VaultPathAllowedForKubernetesNamespace(namespace, rule.GetPath()) {
			return fmt.Errorf("%w: %q cannot be granted to %q", ErrPolicyOutOfScope, rule.GetPath(), namespace)
		}
	}
	return nil
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
