package vault

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

var (
	// Test error sentinels for client errors
	errClientConnectionRefused = errors.New("connection refused")
	errClientForbidden         = errors.New("forbidden")
)

func TestStoreSecretActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.StoreSecretInput
		clientErr error
	}{
		{
			name: "success",
			input: &pb.StoreSecretInput{
				Path: "fabric/metacensus/msp/peer-peer0",
				Data: map[string][]byte{
					"key": []byte("value"),
				},
			},
		},
		{
			name: "client error propagated",
			input: &pb.StoreSecretInput{
				Path: "fabric/test/msp/peer-peer1",
			},
			clientErr: errClientConnectionRefused,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := vaultmocks.NewMockVaultClient(t)
			mockClient.EXPECT().
				Write(mock.Anything, mock.Anything, mock.MatchedBy(func(data map[string]any) bool {
					// Verify the data structure contains what we expect
					if _, ok := data["data"]; !ok {
						return false
					}
					return true
				})).
				Return(tt.clientErr).
				Maybe() // Only if test case uses this path

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.StoreSecret)

			// Execute activity
			val, err := env.ExecuteActivity(activities.StoreSecret, tt.input)

			if tt.clientErr != nil {
				// Expect error - activity test environment returns errors from ExecuteActivity
				// Note: Temporal wraps errors multiple levels, so we just verify an error occurred
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				// Expect success
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}
				var result *pb.StoreSecretOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result.SecretRef == nil {
					t.Errorf("result.SecretRef is nil")
				}
				if result.SecretRef.Path != tt.input.Path {
					t.Errorf("result.SecretRef.Path mismatch: got %s, want %s", result.SecretRef.Path, tt.input.Path)
				}
			}
		})
	}
}

func TestFetchSecretActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.FetchSecretInput
		mockData  map[string]any
		clientErr error
	}{
		{
			name: "success - latest version",
			input: &pb.FetchSecretInput{
				Path:    "fabric/metacensus/msp/peer-peer0",
				Version: 0, // Latest
			},
			mockData: map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"key": "value",
					},
					"metadata": map[string]any{
						"version": float64(5),
					},
				},
			},
		},
		{
			name: "success - specific version",
			input: &pb.FetchSecretInput{
				Path:    "fabric/metacensus/msp/peer-peer0",
				Version: 3,
			},
			mockData: map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"key": "value",
					},
					"metadata": map[string]any{
						"version": float64(3),
					},
				},
			},
		},
		{
			name: "client error propagated",
			input: &pb.FetchSecretInput{
				Path:    "fabric/test/msp/peer-peer1",
				Version: 0,
			},
			clientErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := vaultmocks.NewMockVaultClient(t)

			// Setup mock based on whether we expect an error
			if tt.clientErr != nil {
				mockClient.EXPECT().
					Read(mock.Anything, mock.MatchedBy(func(path string) bool {
						return path == fmt.Sprintf("secret/data/%s", tt.input.Path) ||
							path == fmt.Sprintf("secret/data/%s?version=%d", tt.input.Path, tt.input.Version)
					})).
					Return(nil, tt.clientErr)
			} else {
				mockClient.EXPECT().
					Read(mock.Anything, mock.MatchedBy(func(path string) bool {
						// Match either with or without version param
						return path == fmt.Sprintf("secret/data/%s", tt.input.Path) ||
							path == fmt.Sprintf("secret/data/%s?version=%d", tt.input.Path, tt.input.Version)
					})).
					Return(tt.mockData, nil)
			}

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.FetchSecret)

			// Execute activity
			val, err := env.ExecuteActivity(activities.FetchSecret, tt.input)

			if tt.clientErr != nil {
				// Expect error
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				// Expect success
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}
				var result *pb.FetchSecretOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result.SecretRef == nil {
					t.Errorf("result.SecretRef is nil")
				}
				if result.SecretRef.Path != tt.input.Path {
					t.Errorf("result.SecretRef.Path mismatch: got %s, want %s", result.SecretRef.Path, tt.input.Path)
				}
				if len(result.Data) == 0 {
					t.Errorf("result.Data is empty")
				}
			}
		})
	}
}

func TestDeleteSecretActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.DeleteSecretInput
		clientErr error
	}{
		{
			name: "success",
			input: &pb.DeleteSecretInput{
				Path: "fabric/metacensus/msp/peer-peer0",
			},
		},
		{
			name: "client error propagated",
			input: &pb.DeleteSecretInput{
				Path: "fabric/test/msp/peer-peer1",
			},
			clientErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := vaultmocks.NewMockVaultClient(t)
			mockClient.EXPECT().
				Delete(mock.Anything, mock.MatchedBy(func(path string) bool {
					return path == fmt.Sprintf("secret/data/%s", tt.input.Path)
				})).
				Return(tt.clientErr).
				Maybe()

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.DeleteSecret)

			// Execute activity
			val, err := env.ExecuteActivity(activities.DeleteSecret, tt.input)

			if tt.clientErr != nil {
				// Expect error
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				// Expect success
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}
				var result *pb.DeleteSecretOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
			}
		})
	}
}

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
				Organization:        "metacensus",
				ServiceAccountName:  "fabric-metacensus-sa",
				KubernetesNamespace: "fabric-metacensus",
				PolicyNames:         []string{"fabric-foundational", "fabric-metacensus-specific"},
			},
		},
		{
			name: "success - creates role with single policy",
			input: &pb.CreateAuthRoleInput{
				Organization:        "acme-corp",
				ServiceAccountName:  "fabric-acme-corp-sa",
				KubernetesNamespace: "fabric-acme-corp",
				PolicyNames:         []string{"fabric-foundational"},
			},
		},
		{
			name: "error - missing organization",
			input: &pb.CreateAuthRoleInput{
				Organization:        "",
				ServiceAccountName:  "fabric-sa",
				KubernetesNamespace: "fabric-ns",
				PolicyNames:         []string{"policy"},
			},
			expectErr: true,
		},
		{
			name: "error - missing service account name",
			input: &pb.CreateAuthRoleInput{
				Organization:        "metacensus",
				ServiceAccountName:  "",
				KubernetesNamespace: "fabric-metacensus",
				PolicyNames:         []string{"policy"},
			},
			expectErr: true,
		},
		{
			name: "error - missing kubernetes namespace",
			input: &pb.CreateAuthRoleInput{
				Organization:        "metacensus",
				ServiceAccountName:  "fabric-sa",
				KubernetesNamespace: "",
				PolicyNames:         []string{"policy"},
			},
			expectErr: true,
		},
		{
			name: "error - empty policy list",
			input: &pb.CreateAuthRoleInput{
				Organization:        "metacensus",
				ServiceAccountName:  "fabric-sa",
				KubernetesNamespace: "fabric-ns",
				PolicyNames:         []string{},
			},
			expectErr: true,
		},
		{
			name: "error - client connection refused",
			input: &pb.CreateAuthRoleInput{
				Organization:        "metacensus",
				ServiceAccountName:  "fabric-sa",
				KubernetesNamespace: "fabric-ns",
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

			// Setup mock expectations only for valid inputs
			if tt.input.Organization != "" && tt.input.ServiceAccountName != "" &&
				tt.input.KubernetesNamespace != "" && len(tt.input.PolicyNames) > 0 {
				mockClient.EXPECT().
					Write(mock.Anything, mock.MatchedBy(func(path string) bool {
						return path == fmt.Sprintf("auth/kubernetes/role/fabric-%s", tt.input.Organization)
					}), mock.MatchedBy(func(data map[string]any) bool {
						// Verify request structure
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
				expectedRoleName := fmt.Sprintf("fabric-%s", tt.input.Organization)
				if result.Name != expectedRoleName {
					t.Errorf("role name mismatch: got %s, want %s", result.Name, expectedRoleName)
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

			// Setup mock expectations only for valid inputs
			if tt.input.Name != "" && tt.input.Document != nil && len(tt.input.Document.Rules) > 0 {
				// Verify conversion to HCL was successful
				_, conversionErr := policyDocumentToHCL(tt.input.Document)
				if conversionErr == nil {
					mockClient.EXPECT().
						Write(mock.Anything, mock.MatchedBy(func(path string) bool {
							return path == fmt.Sprintf("sys/policies/acl/%s", tt.input.Name)
						}), mock.MatchedBy(func(data map[string]any) bool {
							// Verify HCL policy is present
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

// Helper function for test readability
func TestPolicyDocumentToHCLOutputFormat(t *testing.T) {
	// Test that output is valid HCL with proper formatting
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

	// Verify proper HCL structure
	lines := strings.Split(hcl, "\n")
	if len(lines) < 8 {
		t.Errorf("HCL output too short: %d lines", len(lines))
	}

	// Verify blank lines between rules
	if !strings.Contains(hcl, "}\n\npath") {
		t.Error("expected blank line between rules")
	}

	// Verify no trailing blank lines
	if strings.HasSuffix(hcl, "\n\n") {
		t.Error("HCL should not end with blank line")
	}
}
