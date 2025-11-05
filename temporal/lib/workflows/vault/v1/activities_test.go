package vault

import (
	"errors"
	"fmt"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

var (
	// Test error sentinels for client errors
	errClientConnectionRefused = errors.New("connection refused")
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
