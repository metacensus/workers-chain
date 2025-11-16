package temporal

import (
	"errors"
	"testing"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	mocks "github.com/metacensus/infra/temporal/mocks/workflows/temporal/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/enums/v1"
	namespacev1 "go.temporal.io/api/namespace/v1"
	"go.temporal.io/api/nexus/v1"
	osv1 "go.temporal.io/api/operatorservice/v1"
	wsv1 "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	// Test error sentinels for client errors
	errNamespaceNotFound = errors.New("namespace not found")
	errConnectionRefused = errors.New("connection refused")
	errNexusConflict     = errors.New("nexus endpoint already exists")
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

			mockNamespaceClient := mocks.NewMockNamespaceClient(t)

			if tt.clientErr != nil {
				mockNamespaceClient.EXPECT().Register(mock.Anything, mock.MatchedBy(func(req *wsv1.RegisterNamespaceRequest) bool {
					return req.Namespace == tt.input.Name
				})).Return(tt.clientErr)
			} else {
				mockNamespaceClient.EXPECT().Register(mock.Anything, mock.MatchedBy(func(req *wsv1.RegisterNamespaceRequest) bool {
					if req.Namespace != tt.input.Name {
						return false
					}
					// Verify retention period matches expected
					if req.WorkflowExecutionRetentionPeriod == nil {
						return false
					}
					actualDur := req.WorkflowExecutionRetentionPeriod.AsDuration()
					return actualDur == tt.expectedDuration
				})).Return(nil)
			}

			activities := NewActivities(mockNamespaceClient, nil)
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

			mockNamespaceClient := mocks.NewMockNamespaceClient(t)

			if tt.clientErr != nil {
				mockNamespaceClient.EXPECT().Describe(mock.Anything, tt.input.Name).
					Return(nil, tt.clientErr)
			} else {
				mockNamespaceClient.EXPECT().Describe(mock.Anything, tt.input.Name).
					Return(tt.mockResponse, nil)
			}

			activities := NewActivities(mockNamespaceClient, nil)
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

func TestCreateNexusEndpointActivity(t *testing.T) {
	tests := []struct {
		name         string
		input        *pb.CreateNexusEndpointInput
		mockResponse *osv1.CreateNexusEndpointResponse
		clientErr    error
		expectError  bool
		expectedID   string
	}{
		{
			name: "success - creates endpoint",
			input: &pb.CreateNexusEndpointInput{
				Name:      "create-scoped-identity-access",
				Namespace: "infra",
				TaskQueue: "infra",
			},
			mockResponse: &osv1.CreateNexusEndpointResponse{
				Endpoint: &nexus.Endpoint{
					Id: "endpoint-123",
					Spec: &nexus.EndpointSpec{
						Name: "create-scoped-identity-access",
						Target: &nexus.EndpointTarget{
							Variant: &nexus.EndpointTarget_Worker_{
								Worker: &nexus.EndpointTarget_Worker{
									Namespace: "infra",
									TaskQueue: "infra",
								},
							},
						},
					},
				},
			},
			expectedID: "endpoint-123",
		},
		{
			name: "idempotent - endpoint already exists",
			input: &pb.CreateNexusEndpointInput{
				Name:      "existing-endpoint",
				Namespace: "infra",
				TaskQueue: "infra",
			},
			clientErr:  errNexusConflict,
			expectedID: "", // On conflict, returns empty ID but no error
		},
		{
			name: "error - invalid name",
			input: &pb.CreateNexusEndpointInput{
				Name:      "",
				Namespace: "infra",
				TaskQueue: "infra",
			},
			expectError: true,
		},
		{
			name: "error - connection refused",
			input: &pb.CreateNexusEndpointInput{
				Name:      "endpoint",
				Namespace: "infra",
				TaskQueue: "infra",
			},
			clientErr:   errConnectionRefused,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockOperatorClient := mocks.NewMockOperatorServiceClient(t)

			if tt.input.Name == "" {
				// Skip client setup for invalid name test
			} else if tt.clientErr != nil {
				mockOperatorClient.EXPECT().CreateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *osv1.CreateNexusEndpointRequest) bool {
					return req.Spec.Name == tt.input.Name
				}), mock.Anything).Return(nil, tt.clientErr)
			} else {
				mockOperatorClient.EXPECT().CreateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *osv1.CreateNexusEndpointRequest) bool {
					return req.Spec.Name == tt.input.Name &&
						req.Spec.Target.GetWorker().Namespace == tt.input.Namespace &&
						req.Spec.Target.GetWorker().TaskQueue == tt.input.TaskQueue
				}), mock.Anything).Return(tt.mockResponse, nil)
			}

			activities := NewActivities(nil, mockOperatorClient)
			env.RegisterActivity(activities.CreateNexusEndpoint)

			val, err := env.ExecuteActivity(activities.CreateNexusEndpoint, tt.input)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if tt.clientErr != nil && errors.Is(tt.clientErr, errNexusConflict) {
				// On idempotent conflict, should succeed with empty ID
				if err != nil {
					t.Fatalf("unexpected error on idempotent conflict: %v", err)
				}
				var result *pb.CreateNexusEndpointOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
			} else if tt.clientErr != nil {
				// Other errors should propagate
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.CreateNexusEndpointOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.Id != tt.expectedID {
					t.Errorf("ID mismatch: got %s, want %s", result.Id, tt.expectedID)
				}
			}
		})
	}
}

func TestDescribeNexusEndpointActivity(t *testing.T) {
	tests := []struct {
		name         string
		input        *pb.DescribeNexusEndpointInput
		mockResponse *osv1.GetNexusEndpointResponse
		clientErr    error
		expectError  bool
		expected     *pb.Endpoint
	}{
		{
			name: "success - endpoint found",
			input: &pb.DescribeNexusEndpointInput{
				Id: "endpoint-123",
			},
			mockResponse: &osv1.GetNexusEndpointResponse{
				Endpoint: &nexus.Endpoint{
					Id: "endpoint-123",
					Spec: &nexus.EndpointSpec{
						Name: "create-scoped-identity-access",
						Target: &nexus.EndpointTarget{
							Variant: &nexus.EndpointTarget_Worker_{
								Worker: &nexus.EndpointTarget_Worker{
									Namespace: "infra",
									TaskQueue: "infra",
								},
							},
						},
					},
				},
			},
			expected: &pb.Endpoint{
				Id:        "endpoint-123",
				Name:      "create-scoped-identity-access",
				Namespace: "infra",
				TaskQueue: "infra",
			},
		},
		{
			name: "error - empty id",
			input: &pb.DescribeNexusEndpointInput{
				Id: "",
			},
			expectError: true,
		},
		{
			name: "error - endpoint not found",
			input: &pb.DescribeNexusEndpointInput{
				Id: "nonexistent",
			},
			clientErr:   errNamespaceNotFound,
			expectError: true,
		},
		{
			name: "error - connection refused",
			input: &pb.DescribeNexusEndpointInput{
				Id: "endpoint-456",
			},
			clientErr:   errConnectionRefused,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockOperatorClient := mocks.NewMockOperatorServiceClient(t)

			if tt.input.Id != "" {
				if tt.clientErr != nil {
					mockOperatorClient.EXPECT().GetNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *osv1.GetNexusEndpointRequest) bool {
						return req.Id == tt.input.Id
					}), mock.Anything).Return(nil, tt.clientErr)
				} else {
					mockOperatorClient.EXPECT().GetNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *osv1.GetNexusEndpointRequest) bool {
						return req.Id == tt.input.Id
					}), mock.Anything).Return(tt.mockResponse, nil)
				}
			}

			activities := NewActivities(nil, mockOperatorClient)
			env.RegisterActivity(activities.DescribeNexusEndpoint)

			val, err := env.ExecuteActivity(activities.DescribeNexusEndpoint, tt.input)

			if tt.expectError {
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
				var result *pb.DescribeNexusEndpointOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.Endpoint == nil {
					t.Fatal("endpoint is nil")
				}
				if result.Endpoint.Id != tt.expected.Id {
					t.Errorf("ID mismatch: got %s, want %s", result.Endpoint.Id, tt.expected.Id)
				}
				if result.Endpoint.Name != tt.expected.Name {
					t.Errorf("Name mismatch: got %s, want %s", result.Endpoint.Name, tt.expected.Name)
				}
				if result.Endpoint.Namespace != tt.expected.Namespace {
					t.Errorf("Namespace mismatch: got %s, want %s", result.Endpoint.Namespace, tt.expected.Namespace)
				}
				if result.Endpoint.TaskQueue != tt.expected.TaskQueue {
					t.Errorf("TaskQueue mismatch: got %s, want %s", result.Endpoint.TaskQueue, tt.expected.TaskQueue)
				}
			}
		})
	}
}

func TestUpdateNexusEndpointActivity(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.UpdateNexusEndpointInput
		clientErr   error
		expectError bool
	}{
		{
			name: "success - updates endpoint",
			input: &pb.UpdateNexusEndpointInput{
				Endpoint: &pb.Endpoint{
					Id:        "endpoint-123",
					Name:      "create-scoped-identity-access",
					Namespace: "infra",
					TaskQueue: "infra",
				},
			},
			expectError: false,
		},
		{
			name: "error - missing endpoint id",
			input: &pb.UpdateNexusEndpointInput{
				Endpoint: &pb.Endpoint{
					Id:        "",
					Name:      "endpoint",
					Namespace: "infra",
					TaskQueue: "infra",
				},
			},
			expectError: true,
		},
		{
			name: "error - missing endpoint name",
			input: &pb.UpdateNexusEndpointInput{
				Endpoint: &pb.Endpoint{
					Id:        "endpoint-123",
					Name:      "",
					Namespace: "infra",
					TaskQueue: "infra",
				},
			},
			expectError: true,
		},
		{
			name: "error - client error",
			input: &pb.UpdateNexusEndpointInput{
				Endpoint: &pb.Endpoint{
					Id:        "endpoint-456",
					Name:      "endpoint",
					Namespace: "infra",
					TaskQueue: "infra",
				},
			},
			clientErr:   errConnectionRefused,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockOperatorClient := mocks.NewMockOperatorServiceClient(t)

			if tt.input.Endpoint.Id != "" && tt.input.Endpoint.Name != "" {
				if tt.clientErr != nil {
					mockOperatorClient.EXPECT().UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *osv1.UpdateNexusEndpointRequest) bool {
						return req.Spec.Name == tt.input.Endpoint.Name
					}), mock.Anything).Return(nil, tt.clientErr)
				} else {
					mockOperatorClient.EXPECT().UpdateNexusEndpoint(mock.Anything, mock.MatchedBy(func(req *osv1.UpdateNexusEndpointRequest) bool {
						return req.Spec.Name == tt.input.Endpoint.Name &&
							req.Spec.Target.GetWorker().Namespace == tt.input.Endpoint.Namespace &&
							req.Spec.Target.GetWorker().TaskQueue == tt.input.Endpoint.TaskQueue
					}), mock.Anything).Return(&osv1.UpdateNexusEndpointResponse{}, nil)
				}
			}

			activities := NewActivities(nil, mockOperatorClient)
			env.RegisterActivity(activities.UpdateNexusEndpoint)

			val, err := env.ExecuteActivity(activities.UpdateNexusEndpoint, tt.input)

			if tt.expectError {
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
				var result *pb.UpdateNexusEndpointOutput
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
