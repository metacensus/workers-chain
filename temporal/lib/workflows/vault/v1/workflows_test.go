package vault_test

import (
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

var errActivityFailed = temporal.NewApplicationError("activity failed", "ActivityFailed")

func TestCreateAccessWorkflow(t *testing.T) {
	tests := []struct {
		name              string
		input             *pb.CreateAccessWorkflowInput
		createPolicyRes   *pb.CreatePolicyOutput
		createPolicyErr   error
		createAuthRoleRes *pb.CreateAuthRoleOutput
		createAuthRoleErr error
		expectErr         bool
	}{
		{
			name: "success - both policy and auth role created",
			input: &pb.CreateAccessWorkflowInput{
				Req: &pb.CreateAccessInput{
					PolicyName: "test-policy",
					PolicyDocument: &pb.PolicyDocument{
						Rules: []*pb.PolicyRule{
							{
								Path:         "secret/data/*",
								Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ, pb.Capability_CAPABILITY_LIST},
							},
						},
					},
					AuthRoleName:        "org1-worker",
					ServiceAccountName:  "test-sa",
					KubernetesNamespace: "default",
				},
			},
			createPolicyRes: &pb.CreatePolicyOutput{
				Name: "test-policy",
			},
			createAuthRoleRes: &pb.CreateAuthRoleOutput{
				Name: "org1-worker",
			},
		},
		{
			name: "failure - CreatePolicy activity fails",
			input: &pb.CreateAccessWorkflowInput{
				Req: &pb.CreateAccessInput{
					PolicyName: "test-policy",
					PolicyDocument: &pb.PolicyDocument{
						Rules: []*pb.PolicyRule{
							{
								Path:         "secret/data/*",
								Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
							},
						},
					},
					AuthRoleName:        "org1-worker",
					ServiceAccountName:  "test-sa",
					KubernetesNamespace: "default",
				},
			},
			createPolicyErr: errActivityFailed,
			expectErr:       true,
		},
		{
			name: "failure - CreateAuthRole activity fails",
			input: &pb.CreateAccessWorkflowInput{
				Req: &pb.CreateAccessInput{
					PolicyName: "test-policy",
					PolicyDocument: &pb.PolicyDocument{
						Rules: []*pb.PolicyRule{
							{
								Path:         "secret/data/*",
								Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
							},
						},
					},
					AuthRoleName:        "org1-worker",
					ServiceAccountName:  "test-sa",
					KubernetesNamespace: "default",
				},
			},
			createPolicyRes: &pb.CreatePolicyOutput{
				Name: "test-policy",
			},
			createAuthRoleErr: errActivityFailed,
			expectErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register the actual vault workflows
			realWorkflows := vault.NewWorkflows()
			pb.RegisterVaultWorkflows(env, realWorkflows)

			// Mock vault activities
			mockActivities := vaultmocks.NewMockVaultActivities(t)

		// Mock CreatePolicy activity
		mockActivities.EXPECT().
			CreatePolicy(mock.Anything, mock.MatchedBy(func(input *pb.CreatePolicyInput) bool {
				return input.GetName() == tt.input.Req.GetPolicyName() &&
					// Document structure is equivalent (same rules and capabilities)
					len(input.GetDocument().GetRules()) == len(tt.input.Req.GetPolicyDocument().GetRules())
			})).
			Return(tt.createPolicyRes, tt.createPolicyErr).
			Maybe()

		// Mock CreateAuthRole activity - only called if CreatePolicy succeeds
		if tt.createPolicyErr == nil {
			mockActivities.EXPECT().
				CreateAuthRole(mock.Anything, mock.MatchedBy(func(input *pb.CreateAuthRoleInput) bool {
					// Verify the policy name from CreatePolicy is included
					expectedPolicyNames := []string{tt.createPolicyRes.GetName()}
					
					if len(input.GetPolicyNames()) != len(expectedPolicyNames) {
						return false
					}
					for i, expected := range expectedPolicyNames {
						if i >= len(input.GetPolicyNames()) || input.GetPolicyNames()[i] != expected {
							return false
						}
					}
					
					return input.GetRoleName() == tt.input.Req.GetAuthRoleName() &&
						input.GetServiceAccountName() == tt.input.Req.GetServiceAccountName() &&
						input.GetKubernetesNamespace() == tt.input.Req.GetKubernetesNamespace()
				})).
				Return(tt.createAuthRoleRes, tt.createAuthRoleErr).
				Maybe()
		}

			pb.RegisterVaultActivities(env, mockActivities)

			// Execute workflow using the test helper function
			env.ExecuteWorkflow(vault.TestCreateAccessWorkflow, tt.input)

			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *pb.CreateAccessOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetPolicyName() != tt.createPolicyRes.GetName() {
					t.Errorf("policy name mismatch: got %q, want %q",
						result.GetPolicyName(), tt.createPolicyRes.GetName())
				}

				if result.GetAuthRoleName() != tt.createAuthRoleRes.GetName() {
					t.Errorf("auth role name mismatch: got %q, want %q",
						result.GetAuthRoleName(), tt.createAuthRoleRes.GetName())
				}
			}
		})
	}
}

