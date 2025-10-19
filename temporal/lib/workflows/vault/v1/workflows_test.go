package vault

import (
	"errors"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

var (
	// Test error sentinels for error comparison
	errVaultUnavailable = errors.New("vault unavailable")
	errSecretNotFound   = errors.New("secret not found")
)

// testStoreSecretWorkflow wraps the workflow execution for testing
func testStoreSecretWorkflow(ctx workflow.Context, input *pb.StoreSecretWorkflowInput) (*pb.SecretRef, error) {
	w := &storeSecretWorkflow{
		input: input,
	}
	return w.Execute(ctx)
}

// testFetchSecretWorkflow wraps the workflow execution for testing
func testFetchSecretWorkflow(ctx workflow.Context, input *pb.FetchSecretWorkflowInput) (*pb.Secret, error) {
	w := &fetchSecretWorkflow{
		input: input,
	}
	return w.Execute(ctx)
}

func TestStoreSecretWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.Secret
		activityRes *pb.SecretRef
		activityErr error
	}{
		{
			name: "success",
			input: &pb.Secret{
				Name:      "my-secret",
				Namespace: "production",
				Data: map[string][]byte{
					"key": []byte("value"),
				},
			},
			activityRes: &pb.SecretRef{
				Name:      "my-secret",
				Namespace: "production",
				Version:   1,
			},
		},
		{
			name: "activity error propagated",
			input: &pb.Secret{
				Name:      "test-secret",
				Namespace: "test-namespace",
			},
			activityErr: errVaultUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := vaultmocks.NewMockVaultActivities(t)
			mockActivities.EXPECT().
				StoreSecret(mock.Anything, tt.input).
				Return(tt.activityRes, tt.activityErr)

			pb.RegisterVaultActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testStoreSecretWorkflow, &pb.StoreSecretWorkflowInput{Req: tt.input})

			// Verify completion
			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.activityErr != nil {
				// Expect error to be propagated
				// Note: Temporal wraps errors multiple levels deep, so we just verify an error occurred
				if env.GetWorkflowError() == nil {
					t.Fatal("expected workflow error, got nil")
				}
			} else {
				// Expect success
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				var result *pb.SecretRef
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
				if result.Name != tt.activityRes.Name || result.Namespace != tt.activityRes.Namespace || result.Version != tt.activityRes.Version {
					t.Errorf("result mismatch: got %+v, want %+v", result, tt.activityRes)
				}
			}
		})
	}
}

func TestFetchSecretWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.SecretRef
		activityRes *pb.Secret
		activityErr error
	}{
		{
			name: "success - latest version",
			input: &pb.SecretRef{
				Name:      "my-secret",
				Namespace: "production",
				Version:   0, // Latest
			},
			activityRes: &pb.Secret{
				Name:      "my-secret",
				Namespace: "production",
				Data: map[string][]byte{
					"key": []byte("value"),
				},
				Version: 5,
			},
		},
		{
			name: "success - specific version",
			input: &pb.SecretRef{
				Name:      "my-secret",
				Namespace: "production",
				Version:   3,
			},
			activityRes: &pb.Secret{
				Name:      "my-secret",
				Namespace: "production",
				Version:   3,
			},
		},
		{
			name: "activity error propagated",
			input: &pb.SecretRef{
				Name:      "nonexistent",
				Namespace: "test",
			},
			activityErr: errSecretNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := vaultmocks.NewMockVaultActivities(t)
			mockActivities.EXPECT().
				FetchSecret(mock.Anything, tt.input).
				Return(tt.activityRes, tt.activityErr)

			pb.RegisterVaultActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testFetchSecretWorkflow, &pb.FetchSecretWorkflowInput{Req: tt.input})

			// Verify completion
			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.activityErr != nil {
				// Expect error to be propagated
				// Note: Temporal wraps errors multiple levels deep, so we just verify an error occurred
				if env.GetWorkflowError() == nil {
					t.Fatal("expected workflow error, got nil")
				}
			} else {
				// Expect success
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				var result *pb.Secret
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
				if result.Name != tt.activityRes.Name || result.Namespace != tt.activityRes.Namespace || result.Version != tt.activityRes.Version {
					t.Errorf("result mismatch: got %+v, want %+v", result, tt.activityRes)
				}
			}
		})
	}
}
