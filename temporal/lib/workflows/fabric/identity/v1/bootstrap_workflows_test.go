package identity

import (
	"errors"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	fabricmocks "github.com/metacensus/infra/temporal/mocks/workflows/fabric/identity/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// Test wrapper function that accepts workflow.Context
func testBootstrapAdminWorkflow(ctx workflow.Context, input *pb.BootstrapAdminWorkflowInput) (*pb.BootstrapAdminOutput, error) {
	w := &bootstrapAdminWorkflow{input: input}
	return w.Execute(ctx)
}

func TestBootstrapAdminWorkflow(t *testing.T) {
	tests := []struct {
		name                     string
		input                    *pb.BootstrapAdminWorkflowInput
		generateSecretRes        *pb.GenerateSecureSecretOutput
		generateSecretErr        error
		createAdminRes           *pb.CreateAdminIdentityOutput
		createAdminErr           error
		expectErr                bool
		expectGenerateSecretCall bool
	}{
		{
			name: "success - with provided secret",
			input: &pb.BootstrapAdminWorkflowInput{
				Req: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "provided-secret",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
					VaultNamespace: "org1",
				},
			},
			createAdminRes: &pb.CreateAdminIdentityOutput{
				AdminIdentity: &pb.AdminIdentity{
					Id:          "admin",
					Secret:      "provided-secret",
					Certificate: []byte("admin-cert"),
					PrivateKey:  []byte("admin-key"),
				},
			},
			expectGenerateSecretCall: false,
		},
		{
			name: "success - generate secret automatically",
			input: &pb.BootstrapAdminWorkflowInput{
				Req: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "", // Empty - should trigger generation
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
					VaultNamespace: "org1",
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{
				Secret: "generated-secret-abc123",
			},
			createAdminRes: &pb.CreateAdminIdentityOutput{
				AdminIdentity: &pb.AdminIdentity{
					Id:          "admin",
					Secret:      "generated-secret-abc123",
					Certificate: []byte("admin-cert"),
					PrivateKey:  []byte("admin-key"),
				},
			},
			expectGenerateSecretCall: true,
		},
		{
			name: "failure - secret generation fails",
			input: &pb.BootstrapAdminWorkflowInput{
				Req: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
			},
			generateSecretErr:        errors.New("activity failed"),
			expectErr:                true,
			expectGenerateSecretCall: true,
		},
		{
			name: "failure - admin creation fails",
			input: &pb.BootstrapAdminWorkflowInput{
				Req: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "secret",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
			},
			createAdminErr:           errors.New("activity failed"),
			expectErr:                true,
			expectGenerateSecretCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)

			// Register GenerateSecureSecret activity
			if tt.expectGenerateSecretCall {
				mockActivities.EXPECT().
					GenerateSecureSecret(mock.Anything, mock.MatchedBy(func(req *pb.GenerateSecureSecretInput) bool {
						if req.GetLength() != 32 {
							t.Errorf("expected length 32, got %d", req.GetLength())
							return false
						}
						return true
					})).
					Return(tt.generateSecretRes, tt.generateSecretErr)
			}

			// Register CreateAdminIdentity activity (only if secret generation doesn't fail)
			if tt.generateSecretErr == nil {
				mockActivities.EXPECT().
					CreateAdminIdentity(mock.Anything, mock.MatchedBy(func(req *pb.CreateAdminIdentityInput) bool {
						if req.GetAdminId() != tt.input.Req.GetAdminId() {
							t.Errorf("admin_id mismatch: got %q, want %q", req.GetAdminId(), tt.input.Req.GetAdminId())
							return false
						}
						return true
					})).
					Return(tt.createAdminRes, tt.createAdminErr)
			}

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testBootstrapAdminWorkflow, tt.input)

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

				var result *pb.BootstrapAdminOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetAdminIdentity().GetId() != tt.createAdminRes.GetAdminIdentity().GetId() {
					t.Errorf("admin_id mismatch: got %q, want %q",
						result.GetAdminIdentity().GetId(),
						tt.createAdminRes.GetAdminIdentity().GetId())
				}

				if !result.GetSuccess() {
					t.Error("expected success to be true")
				}
			}
		})
	}
}
