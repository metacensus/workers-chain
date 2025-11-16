package temporal

import (
	"errors"
	"testing"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	temporalmocks "github.com/metacensus/infra/temporal/mocks/workflows/temporal/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/enums/v1"
	namespacev1 "go.temporal.io/api/namespace/v1"
	wsv1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	// Test error sentinels for client errors
	errNamespaceNotFound = errors.New("namespace not found")
	errConnectionRefused = errors.New("connection refused")
)

func TestRegisterNamespaceActivity(t *testing.T) {
	tests := []struct {
		name             string
		input            *pb.RegisterNamespaceInput
		clientErr        error
		expectedDuration time.Duration
	}{
		{
			name: "success with default retention",
			input: &pb.RegisterNamespaceInput{
				Name:      "fabric-org1",
				Retention: nil, // Use default (30 days)
			},
			expectedDuration: 30 * 24 * time.Hour,
		},
		{
			name: "success with custom retention",
			input: &pb.RegisterNamespaceInput{
				Name:      "fabric-org2",
				Retention: durationpb.New(7 * 24 * time.Hour), // 7 days
			},
			expectedDuration: 7 * 24 * time.Hour,
		},
		{
			name: "success with minimal retention",
			input: &pb.RegisterNamespaceInput{
				Name:      "fabric-test",
				Retention: durationpb.New(1 * time.Hour),
			},
			expectedDuration: 1 * time.Hour,
		},
		{
			name: "client error propagated",
			input: &pb.RegisterNamespaceInput{
				Name:      "fabric-error",
				Retention: nil,
			},
			clientErr: errConnectionRefused,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := temporalmocks.NewMockTemporalAdminClient(t)

			if tt.clientErr != nil {
				mockClient.EXPECT().
					Register(mock.Anything, mock.MatchedBy(func(req *wsv1.RegisterNamespaceRequest) bool {
						return req.Namespace == tt.input.Name
					})).
					Return(tt.clientErr)
			} else {
				mockClient.EXPECT().
					Register(mock.Anything, mock.MatchedBy(func(req *wsv1.RegisterNamespaceRequest) bool {
						if req.Namespace != tt.input.Name {
							return false
						}
						// Verify retention period matches expected
						if req.WorkflowExecutionRetentionPeriod == nil {
							return false
						}
						actualDur := req.WorkflowExecutionRetentionPeriod.AsDuration()
						return actualDur == tt.expectedDuration
					})).
					Return(nil)
			}

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.RegisterNamespace)

			// Execute activity
			val, err := env.ExecuteActivity(activities.RegisterNamespace, tt.input)

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
				var result *pb.RegisterNamespaceOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				// RegisterNamespaceOutput is empty on success, just verify no panic
				if result == nil {
					t.Error("result is nil")
				}
			}
		})
	}
}

func TestDescribeNamespaceActivity(t *testing.T) {
	tests := []struct {
		name          string
		input         *pb.DescribeNamespaceInput
		mockResponse  *wsv1.DescribeNamespaceResponse
		clientErr     error
		expectedState pb.NamespaceState
	}{
		{
			name: "success - registered namespace",
			input: &pb.DescribeNamespaceInput{
				Name: "fabric-org1",
			},
			mockResponse: &wsv1.DescribeNamespaceResponse{
				NamespaceInfo: &namespacev1.NamespaceInfo{
					Name:  "fabric-org1",
					State: enums.NAMESPACE_STATE_REGISTERED,
				},
			},
			expectedState: pb.NamespaceState_REGISTERED,
		},
		{
			name: "success - deprecated namespace",
			input: &pb.DescribeNamespaceInput{
				Name: "old-namespace",
			},
			mockResponse: &wsv1.DescribeNamespaceResponse{
				NamespaceInfo: &namespacev1.NamespaceInfo{
					Name:  "old-namespace",
					State: enums.NAMESPACE_STATE_DEPRECATED,
				},
			},
			expectedState: pb.NamespaceState_DEPRECATED,
		},
		{
			name: "success - unspecified state",
			input: &pb.DescribeNamespaceInput{
				Name: "unknown-state",
			},
			mockResponse: &wsv1.DescribeNamespaceResponse{
				NamespaceInfo: &namespacev1.NamespaceInfo{
					Name:  "unknown-state",
					State: enums.NAMESPACE_STATE_UNSPECIFIED,
				},
			},
			expectedState: pb.NamespaceState_UNSPECIFIED,
		},
		{
			name: "nil namespace info",
			input: &pb.DescribeNamespaceInput{
				Name: "nil-info",
			},
			mockResponse: &wsv1.DescribeNamespaceResponse{
				NamespaceInfo: nil,
			},
			expectedState: pb.NamespaceState_UNSPECIFIED,
		},
		{
			name: "client error - namespace not found",
			input: &pb.DescribeNamespaceInput{
				Name: "does-not-exist",
			},
			clientErr:     errNamespaceNotFound,
			expectedState: pb.NamespaceState_UNSPECIFIED, // Not checked on error path
		},
		{
			name: "client error - connection refused",
			input: &pb.DescribeNamespaceInput{
				Name: "fabric-error",
			},
			clientErr:     errConnectionRefused,
			expectedState: pb.NamespaceState_UNSPECIFIED, // Not checked on error path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := temporalmocks.NewMockTemporalAdminClient(t)

			if tt.clientErr != nil {
				mockClient.EXPECT().
					Describe(mock.Anything, tt.input.Name).
					Return(nil, tt.clientErr)
			} else {
				mockClient.EXPECT().
					Describe(mock.Anything, tt.input.Name).
					Return(tt.mockResponse, nil)
			}

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.DescribeNamespace)

			// Execute activity
			val, err := env.ExecuteActivity(activities.DescribeNamespace, tt.input)

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
				var result *pb.DescribeNamespaceOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.Name != tt.input.Name {
					t.Errorf("result.Name mismatch: got %s, want %s", result.Name, tt.input.Name)
				}
				if result.State != tt.expectedState {
					t.Errorf("result.State mismatch: got %v, want %v", result.State, tt.expectedState)
				}
			}
		})
	}
}
