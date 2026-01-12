package secrets_test

import (
	"errors"
	"strings"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

func TestCreateAccessWorkflow(t *testing.T) {
	tests := []struct {
		name             string
		input            *pb.CreateAccessInput
		createPolicyErr  error
		createAuthErr    error
		expectSuccess    bool
		expectedPolicy   string
		expectedAuthRole string
	}{
		{
			name: "success - creates policy and auth role",
			input: &pb.CreateAccessInput{
				PolicyName:          "test-policy",
				PolicyDocument:      &vaultpb.PolicyDocument{},
				AuthRoleName:        "test-auth-role",
				ServiceAccountName:  "test-sa",
				KubernetesNamespace: "test-namespace",
			},
			createPolicyErr:  nil,
			createAuthErr:    nil,
			expectSuccess:    true,
			expectedPolicy:   "test-policy",
			expectedAuthRole: "test-auth-role",
		},
		{
			name: "error - create policy fails",
			input: &pb.CreateAccessInput{
				PolicyName:          "fail-policy",
				PolicyDocument:      &vaultpb.PolicyDocument{},
				AuthRoleName:        "fail-auth-role",
				ServiceAccountName:  "test-sa",
				KubernetesNamespace: "test-namespace",
			},
			createPolicyErr: errors.New("vault connection failed"),
			expectSuccess:   false,
		},
		{
			name: "error - create auth role fails",
			input: &pb.CreateAccessInput{
				PolicyName:          "success-policy",
				PolicyDocument:      &vaultpb.PolicyDocument{},
				AuthRoleName:        "fail-auth-role",
				ServiceAccountName:  "test-sa",
				KubernetesNamespace: "test-namespace",
			},
			createPolicyErr: nil,
			createAuthErr:   errors.New("kubernetes auth error"),
			expectSuccess:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register the actual workflows (stateless, no dependencies)
			realWorkflows := secrets.NewWorkflows()
			pb.RegisterSecretsWorkflows(env, realWorkflows)

			// Create and register mock vault activities
			mockVaultActivities := vaultmocks.NewMockVaultActivities(t)

			// Setup CreatePolicy activity expectation
			if tt.createPolicyErr != nil {
				mockVaultActivities.EXPECT().
					CreatePolicy(mock.Anything, mock.MatchedBy(func(req *vaultpb.CreatePolicyInput) bool {
						return req.Name == tt.input.PolicyName
					})).
					Return(nil, tt.createPolicyErr).
					Maybe()
			} else {
				mockVaultActivities.EXPECT().
					CreatePolicy(mock.Anything, mock.MatchedBy(func(req *vaultpb.CreatePolicyInput) bool {
						return req.Name == tt.input.PolicyName
					})).
					Return(&vaultpb.CreatePolicyOutput{
						Name: tt.input.PolicyName,
					}, nil).
					Once()

				// Setup CreateAuthRole activity expectation (only if CreatePolicy succeeds)
				if tt.createAuthErr != nil {
					mockVaultActivities.EXPECT().
						CreateAuthRole(mock.Anything, mock.MatchedBy(func(req *vaultpb.CreateAuthRoleInput) bool {
							return req.RoleName == tt.input.AuthRoleName &&
								req.ServiceAccountName == tt.input.ServiceAccountName &&
								req.KubernetesNamespace == tt.input.KubernetesNamespace &&
								len(req.PolicyNames) == 1 &&
								req.PolicyNames[0] == tt.input.PolicyName
						})).
						Return(nil, tt.createAuthErr).
						Maybe()
				} else {
					mockVaultActivities.EXPECT().
						CreateAuthRole(mock.Anything, mock.MatchedBy(func(req *vaultpb.CreateAuthRoleInput) bool {
							return req.RoleName == tt.input.AuthRoleName &&
								req.ServiceAccountName == tt.input.ServiceAccountName &&
								req.KubernetesNamespace == tt.input.KubernetesNamespace &&
								len(req.PolicyNames) == 1 &&
								req.PolicyNames[0] == tt.input.PolicyName
						})).
						Return(&vaultpb.CreateAuthRoleOutput{
							Name: tt.input.AuthRoleName,
						}, nil).
						Once()
				}
			}

			// Register mock activities with the test environment
			vaultpb.RegisterVaultActivities(env, mockVaultActivities)

			// Execute workflow by name
			env.ExecuteWorkflow(pb.CreateAccessWorkflowName, tt.input)

			if tt.expectSuccess {
				if !env.IsWorkflowCompleted() {
					t.Fatal("workflow did not complete")
				}
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *pb.CreateAccessOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.PolicyName != tt.expectedPolicy {
					t.Errorf("PolicyName mismatch: got %s, want %s", result.PolicyName, tt.expectedPolicy)
				}
				if result.AuthRoleName != tt.expectedAuthRole {
					t.Errorf("AuthRoleName mismatch: got %s, want %s", result.AuthRoleName, tt.expectedAuthRole)
				}
			} else {
				if !env.IsWorkflowCompleted() {
					t.Fatal("workflow did not complete")
				}
				err := env.GetWorkflowError()
				if err == nil {
					t.Fatal("expected workflow error, got nil")
				}
				// Verify error contains expected sentinel error message
				// (Temporal wraps errors, so we check the error string)
				errStr := err.Error()
				if tt.createPolicyErr != nil {
					if !strings.Contains(errStr, secrets.ErrCreateAccessPolicy.Error()) {
						t.Errorf("expected error to contain %q, got: %v", secrets.ErrCreateAccessPolicy.Error(), err)
					}
				} else if tt.createAuthErr != nil {
					if !strings.Contains(errStr, secrets.ErrCreateAccessAuthRole.Error()) {
						t.Errorf("expected error to contain %q, got: %v", secrets.ErrCreateAccessAuthRole.Error(), err)
					}
				}
			}
		})
	}
}
