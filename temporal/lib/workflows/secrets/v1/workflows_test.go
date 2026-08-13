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
				KubernetesNamespace: "fabric-shard-org1",
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
				KubernetesNamespace: "fabric-shard-org1",
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
				KubernetesNamespace: "fabric-shard-org1",
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

// CreateAccess is reachable over Nexus and cannot identify its caller, so the
// Kubernetes namespace being granted access is what bounds the Vault paths it may
// be handed. Every rejection here must happen before any Vault activity runs.
func TestCreateAccessWorkflow_PolicyScope(t *testing.T) {
	rules := func(paths ...string) *vaultpb.PolicyDocument {
		doc := &vaultpb.PolicyDocument{}
		for _, path := range paths {
			doc.Rules = append(doc.Rules, &vaultpb.PolicyRule{
				Path:         path,
				Capabilities: []vaultpb.Capability{vaultpb.Capability_CAPABILITY_READ},
			})
		}
		return doc
	}

	tests := []struct {
		name      string
		namespace string
		document  *vaultpb.PolicyDocument
		allowed   bool
	}{
		{
			name:      "shard grants its own subtree",
			namespace: "fabric-shard-org1",
			document:  rules("secret/data/fabric/shard/org1/*", "secret/metadata/fabric/shard/org1/*"),
			allowed:   true,
		},
		{
			name:      "operator grants its own subtree",
			namespace: "fabric",
			document:  rules("secret/data/fabric/metacensus/certs/*"),
			allowed:   true,
		},
		{
			name:      "shard cannot reach another org",
			namespace: "fabric-shard-org1",
			document:  rules("secret/data/fabric/shard/org2/keys/*"),
		},
		{
			name:      "shard cannot reach the operator subtree",
			namespace: "fabric-shard-org1",
			document:  rules("secret/data/fabric/metacensus/keys/*"),
		},
		{
			name:      "org prefix is not a subtree boundary",
			namespace: "fabric-shard-org1",
			document:  rules("secret/data/fabric/shard/org10/keys/*"),
		},
		{
			name:      "operator cannot reach a shard subtree",
			namespace: "fabric",
			document:  rules("secret/data/fabric/shard/org1/keys/*"),
		},
		{
			name:      "one bad rule rejects the whole document",
			namespace: "fabric-shard-org1",
			document:  rules("secret/data/fabric/shard/org1/*", "secret/data/fabric/shard/org2/*"),
		},
		{
			name:      "paths outside the KV mount are rejected",
			namespace: "fabric-shard-org1",
			document:  rules("sys/policies/acl/*"),
		},
		{
			name:      "traversal is rejected",
			namespace: "fabric-shard-org1",
			document:  rules("secret/data/fabric/shard/org1/../org2/*"),
		},
		{
			name:      "non-owner namespace owns no subtree",
			namespace: "workers",
			document:  rules("secret/data/fabric/metacensus/*"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()
			pb.RegisterSecretsWorkflows(env, secrets.NewWorkflows())

			mockVaultActivities := vaultmocks.NewMockVaultActivities(t)
			if tt.allowed {
				mockVaultActivities.EXPECT().
					CreatePolicy(mock.Anything, mock.Anything).
					Return(&vaultpb.CreatePolicyOutput{Name: "policy"}, nil).
					Once()
				mockVaultActivities.EXPECT().
					CreateAuthRole(mock.Anything, mock.Anything).
					Return(&vaultpb.CreateAuthRoleOutput{Name: "role"}, nil).
					Once()
			}
			vaultpb.RegisterVaultActivities(env, mockVaultActivities)

			env.ExecuteWorkflow(pb.CreateAccessWorkflowName, &pb.CreateAccessInput{
				PolicyName:          "policy",
				PolicyDocument:      tt.document,
				AuthRoleName:        "role",
				ServiceAccountName:  "sa",
				KubernetesNamespace: tt.namespace,
			})

			if !env.IsWorkflowCompleted() {
				t.Fatal("workflow did not complete")
			}
			err := env.GetWorkflowError()
			if tt.allowed {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected the request to be rejected")
			}
			if !strings.Contains(err.Error(), secrets.ErrPolicyOutOfScope.Error()) {
				t.Errorf("expected %q, got: %v", secrets.ErrPolicyOutOfScope.Error(), err)
			}
		})
	}
}
