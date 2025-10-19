package vault

import (
	"errors"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

var (
	// Test error sentinels for client errors
	errClientConnectionRefused = errors.New("connection refused")
	errClientNotFound          = errors.New("secret not found")
)

func TestStoreSecretActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.Secret
		clientRes *pb.SecretRef
		clientErr error
	}{
		{
			name: "success - parameters passed through",
			input: &pb.Secret{
				Name:      "my-secret",
				Namespace: "production",
				Data: map[string][]byte{
					"key": []byte("value"),
				},
			},
			clientRes: &pb.SecretRef{
				Name:      "my-secret",
				Namespace: "production",
				Version:   1,
			},
		},
		{
			name: "client error propagated",
			input: &pb.Secret{
				Name:      "test-secret",
				Namespace: "test",
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
				StoreSecret(mock.Anything, mock.MatchedBy(func(req *pb.Secret) bool {
					// Verify input is passed through correctly
					if req.Name != tt.input.Name {
						t.Errorf("Name mismatch: got %q, want %q", req.Name, tt.input.Name)
						return false
					}
					if req.Namespace != tt.input.Namespace {
						t.Errorf("Namespace mismatch: got %q, want %q", req.Namespace, tt.input.Namespace)
						return false
					}
					return true
				})).
				Return(tt.clientRes, tt.clientErr)

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
				var result *pb.SecretRef
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result.Name != tt.clientRes.Name || result.Namespace != tt.clientRes.Namespace || result.Version != tt.clientRes.Version {
					t.Errorf("result mismatch: got %+v, want %+v", result, tt.clientRes)
				}
			}
		})
	}
}

func TestFetchSecretActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.SecretRef
		clientRes *pb.Secret
		clientErr error
	}{
		{
			name: "success - parameters passed through",
			input: &pb.SecretRef{
				Name:      "my-secret",
				Namespace: "production",
				Version:   0, // Latest
			},
			clientRes: &pb.Secret{
				Name:      "my-secret",
				Namespace: "production",
				Data: map[string][]byte{
					"key": []byte("value"),
				},
				Version: 5,
			},
		},
		{
			name: "success - specific version requested",
			input: &pb.SecretRef{
				Name:      "my-secret",
				Namespace: "production",
				Version:   3,
			},
			clientRes: &pb.Secret{
				Name:      "my-secret",
				Namespace: "production",
				Version:   3,
			},
		},
		{
			name: "client error propagated",
			input: &pb.SecretRef{
				Name:      "nonexistent",
				Namespace: "test",
			},
			clientErr: errClientNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := vaultmocks.NewMockVaultClient(t)
			mockClient.EXPECT().
				FetchSecret(mock.Anything, mock.MatchedBy(func(req *pb.SecretRef) bool {
					// Verify input is passed through correctly
					if req.Name != tt.input.Name {
						t.Errorf("Name mismatch: got %q, want %q", req.Name, tt.input.Name)
						return false
					}
					if req.Namespace != tt.input.Namespace {
						t.Errorf("Namespace mismatch: got %q, want %q", req.Namespace, tt.input.Namespace)
						return false
					}
					if req.Version != tt.input.Version {
						t.Errorf("Version mismatch: got %d, want %d", req.Version, tt.input.Version)
						return false
					}
					return true
				})).
				Return(tt.clientRes, tt.clientErr)

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.FetchSecret)

			// Execute activity
			val, err := env.ExecuteActivity(activities.FetchSecret, tt.input)

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
				var result *pb.Secret
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result.Name != tt.clientRes.Name || result.Namespace != tt.clientRes.Namespace || result.Version != tt.clientRes.Version {
					t.Errorf("result mismatch: got %+v, want %+v", result, tt.clientRes)
				}
			}
		})
	}
}
