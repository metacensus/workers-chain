package temporal_test

import (
	"errors"
	"testing"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	temporalv1 "github.com/metacensus/infra/temporal/lib/workflows/temporal/v1"
	temporalmocks "github.com/metacensus/infra/temporal/mocks/workflows/temporal/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestCreateNamespaceWorkflow(t *testing.T) {
	tests := []struct {
		name              string
		input             *pb.CreateNamespaceInput
		registerErr       error
		describeResponses []*pb.DescribeNamespaceOutput
		describeErrors    []error
		expectSuccess     bool
		expectedNamespace string
	}{
		{
			name: "success - namespace becomes registered immediately",
			input: &pb.CreateNamespaceInput{
				Name:      "fabric-org1",
				Retention: durationpb.New(30 * 24 * time.Hour),
			},
			registerErr: nil,
			describeResponses: []*pb.DescribeNamespaceOutput{
				{
					Name:  "fabric-org1",
					State: pb.NamespaceState_REGISTERED,
				},
			},
			expectSuccess:     true,
			expectedNamespace: "fabric-org1",
		},
		{
			name: "success - namespace becomes registered after retries",
			input: &pb.CreateNamespaceInput{
				Name:      "fabric-org2",
				Retention: nil, // Use default
			},
			registerErr: nil,
			describeResponses: []*pb.DescribeNamespaceOutput{
				{
					Name:  "fabric-org2",
					State: pb.NamespaceState_UNSPECIFIED,
				},
				{
					Name:  "fabric-org2",
					State: pb.NamespaceState_UNSPECIFIED,
				},
				{
					Name:  "fabric-org2",
					State: pb.NamespaceState_REGISTERED,
				},
			},
			expectSuccess:     true,
			expectedNamespace: "fabric-org2",
		},
		{
			name: "failure - register namespace fails",
			input: &pb.CreateNamespaceInput{
				Name:      "fabric-error",
				Retention: durationpb.New(7 * 24 * time.Hour),
			},
			registerErr:   errors.New("connection failed"),
			expectSuccess: false,
		},
		{
			name: "failure - describe never reaches registered (timeout)",
			input: &pb.CreateNamespaceInput{
				Name:      "fabric-timeout",
				Retention: nil,
			},
			registerErr: nil,
			describeResponses: []*pb.DescribeNamespaceOutput{
				{
					Name:  "fabric-timeout",
					State: pb.NamespaceState_UNSPECIFIED,
				},
			},
			expectSuccess: false,
		},
		{
			name: "success - describe transient errors, then succeeds",
			input: &pb.CreateNamespaceInput{
				Name:      "fabric-retry",
				Retention: durationpb.New(1 * 24 * time.Hour),
			},
			registerErr: nil,
			describeErrors: []error{
				errors.New("temporary failure"),
				nil,
			},
			describeResponses: []*pb.DescribeNamespaceOutput{
				nil, // Ignored due to error
				{
					Name:  "fabric-retry",
					State: pb.NamespaceState_REGISTERED,
				},
			},
			expectSuccess:     true,
			expectedNamespace: "fabric-retry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register the actual workflows (stateless, no dependencies)
			realWorkflows := temporalv1.NewWorkflows()
			pb.RegisterAdminWorkflows(env, realWorkflows)

			// Mock activities
			mockActivities := temporalmocks.NewMockAdminActivities(t)

			// Setup register namespace expectation
			if tt.registerErr != nil {
				// Activity may be retried, so allow multiple calls
				mockActivities.EXPECT().
					RegisterNamespace(mock.Anything, mock.MatchedBy(func(req *pb.RegisterNamespaceInput) bool {
						return req.Name == tt.input.Name
					})).
					Return(nil, tt.registerErr).
					Maybe()
			} else {
				mockActivities.EXPECT().
					RegisterNamespace(mock.Anything, mock.MatchedBy(func(req *pb.RegisterNamespaceInput) bool {
						return req.Name == tt.input.Name
					})).
					Return(&pb.RegisterNamespaceOutput{}, nil).
					Once()

				// Setup describe namespace expectations for polling
				// Handle the case where we have both errors and responses (transient failure case)
				numCalls := len(tt.describeErrors)
				if numCalls < len(tt.describeResponses) {
					numCalls = len(tt.describeResponses)
				}

				for i := 0; i < numCalls; i++ {
					// Determine error and response for this call
					var respErr error
					var resp *pb.DescribeNamespaceOutput

					if i < len(tt.describeErrors) {
						respErr = tt.describeErrors[i]
					}

					if respErr == nil && i < len(tt.describeResponses) {
						resp = tt.describeResponses[i]
					}

					// Set up the mock for this specific call
					if respErr != nil {
						mockActivities.EXPECT().
							DescribeNamespace(mock.Anything, mock.MatchedBy(func(req *pb.DescribeNamespaceInput) bool {
								return req.Name == tt.input.Name
							})).
							Return(nil, respErr).
							Once()
					} else {
						mockActivities.EXPECT().
							DescribeNamespace(mock.Anything, mock.MatchedBy(func(req *pb.DescribeNamespaceInput) bool {
								return req.Name == tt.input.Name
							})).
							Return(resp, nil).
							Once()
					}
				}

				// For all cases, allow unlimited additional calls with a reasonable response
				// This handles polling retries beyond what we've explicitly mocked
				if len(tt.describeResponses) > 0 {
					// Use the last response configured for unlimited retries
					lastResponse := tt.describeResponses[len(tt.describeResponses)-1]
					mockActivities.EXPECT().
						DescribeNamespace(mock.Anything, mock.MatchedBy(func(req *pb.DescribeNamespaceInput) bool {
							return req.Name == tt.input.Name
						})).
						Return(lastResponse, nil).
						Maybe()
				}
			}

			// Register mock activities - resolved by name at runtime
			pb.RegisterAdminActivities(env, mockActivities)

			// Execute workflow by name
			env.ExecuteWorkflow(pb.CreateNamespaceWorkflowName, tt.input)

			if tt.expectSuccess {
				if !env.IsWorkflowCompleted() {
					t.Fatal("workflow did not complete")
				}
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *pb.CreateNamespaceOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.Name != tt.expectedNamespace {
					t.Errorf("result.Name mismatch: got %s, want %s", result.Name, tt.expectedNamespace)
				}
			} else {
				if !env.IsWorkflowCompleted() {
					t.Fatal("workflow did not complete")
				}
				if err := env.GetWorkflowError(); err == nil {
					t.Fatal("expected workflow error, got nil")
				}
			}
		})
	}
}
