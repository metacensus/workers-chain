package vault

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	pb "github.com/metacensus/workers-chain/gen/workflows/vault/v1"
	vaultmocks "github.com/metacensus/workers-chain/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

var (
	errClientConnectionRefused = errors.New("connection refused")
	errClientForbidden         = errors.New("forbidden")
)

func TestCreateAuthRoleActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateAuthRoleInput
		clientErr error
		expectErr bool
	}{
		{
			name: "success - creates role with policies",
			input: &pb.CreateAuthRoleInput{
				RoleName:            "metacensus-worker",
				ServiceAccountName:  "metacensus-sa",
				KubernetesNamespace: "metacensus",
				PolicyNames:         []string{"policy-foundational", "policy-specific"},
			},
		},
		{
			name: "success - creates role with single policy",
			input: &pb.CreateAuthRoleInput{
				RoleName:            "acme-corp-worker",
				ServiceAccountName:  "acme-corp-sa",
				KubernetesNamespace: "acme-corp",
				PolicyNames:         []string{"policy-foundational"},
			},
		},
		{
			name: "error - missing role name",
			input: &pb.CreateAuthRoleInput{
				RoleName:            "",
				ServiceAccountName:  "my-sa",
				KubernetesNamespace: "my-ns",
				PolicyNames:         []string{"policy"},
			},
			expectErr: true,
		},
		{
			name: "error - missing service account name",
			input: &pb.CreateAuthRoleInput{
				RoleName:            "my-role",
				ServiceAccountName:  "",
				KubernetesNamespace: "my-ns",
				PolicyNames:         []string{"policy"},
			},
			expectErr: true,
		},
		{
			name: "error - missing kubernetes namespace",
			input: &pb.CreateAuthRoleInput{
				RoleName:            "my-role",
				ServiceAccountName:  "my-sa",
				KubernetesNamespace: "",
				PolicyNames:         []string{"policy"},
			},
			expectErr: true,
		},
		{
			name: "error - empty policy list",
			input: &pb.CreateAuthRoleInput{
				RoleName:            "my-role",
				ServiceAccountName:  "my-sa",
				KubernetesNamespace: "my-ns",
				PolicyNames:         []string{},
			},
			expectErr: true,
		},
		{
			name: "error - client connection refused",
			input: &pb.CreateAuthRoleInput{
				RoleName:            "my-role",
				ServiceAccountName:  "my-sa",
				KubernetesNamespace: "my-ns",
				PolicyNames:         []string{"policy"},
			},
			clientErr: errClientConnectionRefused,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := vaultmocks.NewMockVaultClient(t)

			if tt.input.RoleName != "" && tt.input.ServiceAccountName != "" &&
				tt.input.KubernetesNamespace != "" && len(tt.input.PolicyNames) > 0 {
				mockClient.EXPECT().
					Write(mock.Anything, mock.MatchedBy(func(path string) bool {
						return path == fmt.Sprintf("auth/kubernetes/role/%s", tt.input.RoleName)
					}), mock.MatchedBy(func(data map[string]any) bool {
						if _, ok := data["bound_service_account_names"]; !ok {
							return false
						}
						if _, ok := data["bound_service_account_namespaces"]; !ok {
							return false
						}
						if _, ok := data["policies"]; !ok {
							return false
						}
						return true
					})).
					Return(tt.clientErr).
					Maybe()
			}

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.CreateAuthRole)

			val, err := env.ExecuteActivity(activities.CreateAuthRole, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if tt.clientErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.CreateAuthRoleOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.Name != tt.input.RoleName {
					t.Errorf("role name mismatch: got %s, want %s", result.Name, tt.input.RoleName)
				}
			}
		})
	}
}

func TestDeleteAuthRoleActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.DeleteAuthRoleInput
		clientErr error
		expectErr bool
	}{
		{
			name: "success - deletes role",
			input: &pb.DeleteAuthRoleInput{
				Name: "fabric-metacensus",
			},
		},
		{
			name: "error - missing role name",
			input: &pb.DeleteAuthRoleInput{
				Name: "",
			},
			expectErr: true,
		},
		{
			name: "error - client not found",
			input: &pb.DeleteAuthRoleInput{
				Name: "fabric-nonexistent",
			},
			clientErr: ErrNotFound,
			expectErr: true,
		},
		{
			name: "error - client connection refused",
			input: &pb.DeleteAuthRoleInput{
				Name: "fabric-metacensus",
			},
			clientErr: errClientConnectionRefused,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := vaultmocks.NewMockVaultClient(t)

			if tt.input.Name != "" {
				mockClient.EXPECT().
					Delete(mock.Anything, mock.MatchedBy(func(path string) bool {
						return path == fmt.Sprintf("auth/kubernetes/role/%s", tt.input.Name)
					})).
					Return(tt.clientErr).
					Maybe()
			}

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.DeleteAuthRole)

			val, err := env.ExecuteActivity(activities.DeleteAuthRole, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if tt.clientErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.DeleteAuthRoleOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
			}
		})
	}
}

func TestCreatePolicyActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreatePolicyInput
		clientErr error
		expectErr bool
	}{
		{
			name: "success - single rule with multiple capabilities",
			input: &pb.CreatePolicyInput{
				Name: "fabric-foundational",
				Document: &pb.PolicyDocument{
					Rules: []*pb.PolicyRule{
						{
							Path:         "secret/data/fabric/*",
							Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ, pb.Capability_CAPABILITY_LIST},
						},
					},
				},
			},
		},
		{
			name: "success - multiple rules",
			input: &pb.CreatePolicyInput{
				Name: "fabric-complex",
				Document: &pb.PolicyDocument{
					Rules: []*pb.PolicyRule{
						{
							Path:         "secret/data/fabric/metacensus/msp/*",
							Capabilities: []pb.Capability{pb.Capability_CAPABILITY_CREATE, pb.Capability_CAPABILITY_READ, pb.Capability_CAPABILITY_UPDATE},
						},
						{
							Path:         "auth/kubernetes/login",
							Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
						},
					},
				},
			},
		},
		{
			name: "error - missing policy name",
			input: &pb.CreatePolicyInput{
				Name: "",
				Document: &pb.PolicyDocument{
					Rules: []*pb.PolicyRule{
						{
							Path:         "secret/data/*",
							Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "error - nil policy document",
			input: &pb.CreatePolicyInput{
				Name:     "fabric-policy",
				Document: nil,
			},
			expectErr: true,
		},
		{
			name: "error - empty rules",
			input: &pb.CreatePolicyInput{
				Name:     "fabric-policy",
				Document: &pb.PolicyDocument{Rules: []*pb.PolicyRule{}},
			},
			expectErr: true,
		},
		{
			name: "error - rule with empty path",
			input: &pb.CreatePolicyInput{
				Name: "fabric-policy",
				Document: &pb.PolicyDocument{
					Rules: []*pb.PolicyRule{
						{
							Path:         "",
							Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "error - rule with unspecified capability",
			input: &pb.CreatePolicyInput{
				Name: "fabric-policy",
				Document: &pb.PolicyDocument{
					Rules: []*pb.PolicyRule{
						{
							Path:         "secret/data/*",
							Capabilities: []pb.Capability{pb.Capability_CAPABILITY_UNSPECIFIED},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "error - client connection refused",
			input: &pb.CreatePolicyInput{
				Name: "fabric-policy",
				Document: &pb.PolicyDocument{
					Rules: []*pb.PolicyRule{
						{
							Path:         "secret/data/*",
							Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
						},
					},
				},
			},
			clientErr: errClientConnectionRefused,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := vaultmocks.NewMockVaultClient(t)

			if tt.input.Name != "" && tt.input.Document != nil && len(tt.input.Document.Rules) > 0 {
				_, conversionErr := policyDocumentToHCL(tt.input.Document)
				if conversionErr == nil {
					mockClient.EXPECT().
						Write(mock.Anything, mock.MatchedBy(func(path string) bool {
							return path == fmt.Sprintf("sys/policies/acl/%s", tt.input.Name)
						}), mock.MatchedBy(func(data map[string]any) bool {
							_, ok := data["policy"]
							return ok
						})).
						Return(tt.clientErr).
						Maybe()
				}
			}

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.CreatePolicy)

			val, err := env.ExecuteActivity(activities.CreatePolicy, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if tt.clientErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.CreatePolicyOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.Name != tt.input.Name {
					t.Errorf("policy name mismatch: got %s, want %s", result.Name, tt.input.Name)
				}
			}
		})
	}
}

func TestDeletePolicyActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.DeletePolicyInput
		clientErr error
		expectErr bool
	}{
		{
			name: "success - deletes policy",
			input: &pb.DeletePolicyInput{
				Name: "fabric-foundational",
			},
		},
		{
			name: "error - missing policy name",
			input: &pb.DeletePolicyInput{
				Name: "",
			},
			expectErr: true,
		},
		{
			name: "error - client not found",
			input: &pb.DeletePolicyInput{
				Name: "fabric-nonexistent",
			},
			clientErr: ErrNotFound,
			expectErr: true,
		},
		{
			name: "error - client forbidden",
			input: &pb.DeletePolicyInput{
				Name: "fabric-policy",
			},
			clientErr: errClientForbidden,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := vaultmocks.NewMockVaultClient(t)

			if tt.input.Name != "" {
				mockClient.EXPECT().
					Delete(mock.Anything, mock.MatchedBy(func(path string) bool {
						return path == fmt.Sprintf("sys/policies/acl/%s", tt.input.Name)
					})).
					Return(tt.clientErr).
					Maybe()
			}

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.DeletePolicy)

			val, err := env.ExecuteActivity(activities.DeletePolicy, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if tt.clientErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.DeletePolicyOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
			}
		})
	}
}

func TestPolicyDocumentToHCL(t *testing.T) {
	tests := []struct {
		name      string
		doc       *pb.PolicyDocument
		expectErr bool
		checkHCL  func(string) bool
	}{
		{
			name: "success - single rule with single capability",
			doc: &pb.PolicyDocument{
				Rules: []*pb.PolicyRule{
					{
						Path:         "secret/data/fabric/*",
						Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
					},
				},
			},
			checkHCL: func(hcl string) bool {
				return strings.Contains(hcl, `path "secret/data/fabric/*"`) &&
					strings.Contains(hcl, `"read"`)
			},
		},
		{
			name: "success - single rule with multiple capabilities",
			doc: &pb.PolicyDocument{
				Rules: []*pb.PolicyRule{
					{
						Path:         "secret/data/fabric/metacensus/*",
						Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ, pb.Capability_CAPABILITY_LIST},
					},
				},
			},
			checkHCL: func(hcl string) bool {
				return strings.Contains(hcl, `path "secret/data/fabric/metacensus/*"`) &&
					strings.Contains(hcl, `"read"`) &&
					strings.Contains(hcl, `"list"`)
			},
		},
		{
			name: "success - multiple rules with various capabilities",
			doc: &pb.PolicyDocument{
				Rules: []*pb.PolicyRule{
					{
						Path:         "secret/data/fabric/*",
						Capabilities: []pb.Capability{pb.Capability_CAPABILITY_CREATE, pb.Capability_CAPABILITY_READ},
					},
					{
						Path:         "auth/kubernetes/login",
						Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
					},
				},
			},
			checkHCL: func(hcl string) bool {
				return strings.Contains(hcl, `path "secret/data/fabric/*"`) &&
					strings.Contains(hcl, `path "auth/kubernetes/login"`) &&
					strings.Contains(hcl, `"create"`) &&
					strings.Contains(hcl, `"read"`)
			},
		},
		{
			name: "success - all capability types",
			doc: &pb.PolicyDocument{
				Rules: []*pb.PolicyRule{
					{
						Path: "secret/data/all-caps/*",
						Capabilities: []pb.Capability{
							pb.Capability_CAPABILITY_CREATE,
							pb.Capability_CAPABILITY_READ,
							pb.Capability_CAPABILITY_UPDATE,
							pb.Capability_CAPABILITY_PATCH,
							pb.Capability_CAPABILITY_DELETE,
							pb.Capability_CAPABILITY_LIST,
							pb.Capability_CAPABILITY_SUDO,
							pb.Capability_CAPABILITY_DENY,
							pb.Capability_CAPABILITY_SUBSCRIBE,
							pb.Capability_CAPABILITY_RECOVER,
						},
					},
				},
			},
			checkHCL: func(hcl string) bool {
				caps := []string{"create", "read", "update", "patch", "delete", "list", "sudo", "deny", "subscribe", "recover"}
				for _, cap := range caps {
					if !strings.Contains(hcl, fmt.Sprintf(`"%s"`, cap)) {
						return false
					}
				}
				return true
			},
		},
		{
			name:      "error - nil document",
			doc:       nil,
			expectErr: true,
		},
		{
			name:      "error - empty rules",
			doc:       &pb.PolicyDocument{Rules: []*pb.PolicyRule{}},
			expectErr: true,
		},
		{
			name: "error - rule with empty path",
			doc: &pb.PolicyDocument{
				Rules: []*pb.PolicyRule{
					{
						Path:         "",
						Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "error - rule with no capabilities",
			doc: &pb.PolicyDocument{
				Rules: []*pb.PolicyRule{
					{
						Path:         "secret/data/*",
						Capabilities: []pb.Capability{},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "error - rule with unspecified capability",
			doc: &pb.PolicyDocument{
				Rules: []*pb.PolicyRule{
					{
						Path:         "secret/data/*",
						Capabilities: []pb.Capability{pb.Capability_CAPABILITY_UNSPECIFIED},
					},
				},
			},
			expectErr: true,
		},
		{
			name: "error - mixed valid and unspecified capability",
			doc: &pb.PolicyDocument{
				Rules: []*pb.PolicyRule{
					{
						Path:         "secret/data/*",
						Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ, pb.Capability_CAPABILITY_UNSPECIFIED},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hcl, err := policyDocumentToHCL(tt.doc)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if hcl == "" {
					t.Fatal("HCL output is empty")
				}
				if tt.checkHCL != nil && !tt.checkHCL(hcl) {
					t.Errorf("HCL output does not match expected format:\n%s", hcl)
				}
			}
		})
	}
}

func TestPolicyDocumentToHCLOutputFormat(t *testing.T) {
	doc := &pb.PolicyDocument{
		Rules: []*pb.PolicyRule{
			{
				Path:         "secret/data/fabric/*",
				Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ, pb.Capability_CAPABILITY_LIST},
			},
			{
				Path:         "auth/kubernetes/login",
				Capabilities: []pb.Capability{pb.Capability_CAPABILITY_READ},
			},
		},
	}

	hcl, err := policyDocumentToHCL(doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(hcl, "\n")
	if len(lines) < 8 {
		t.Errorf("HCL output too short: %d lines", len(lines))
	}

	if !strings.Contains(hcl, "}\n\npath") {
		t.Error("expected blank line between rules")
	}

	if strings.HasSuffix(hcl, "\n\n") {
		t.Error("HCL should not end with blank line")
	}
}
