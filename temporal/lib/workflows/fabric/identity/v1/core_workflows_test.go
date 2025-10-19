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

var (
	// Test error sentinels for activity errors
	errActivityFailed = errors.New("activity failed")
)

// Test wrapper functions that accept workflow.Context
func testRegisterIdentityWorkflow(ctx workflow.Context, input *pb.RegisterIdentityWorkflowInput) (*pb.RegisterIdentityOutput, error) {
	w := &registerIdentityWorkflow{input: input}
	return w.Execute(ctx)
}

func testEnrollIdentityWorkflow(ctx workflow.Context, input *pb.EnrollIdentityWorkflowInput) (*pb.EnrollIdentityOutput, error) {
	w := &enrollIdentityWorkflow{input: input}
	return w.Execute(ctx)
}

func testAssembleMspWorkflow(ctx workflow.Context, input *pb.AssembleMspWorkflowInput) (*pb.AssembleMspOutput, error) {
	w := &assembleMspWorkflow{input: input}
	return w.Execute(ctx)
}

func TestRegisterIdentityWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		input         *pb.RegisterIdentityWorkflowInput
		activityRes   *pb.RegisterIdentityOutput
		activityErr   error
		expectErr     bool
		expectErrType error
	}{
		{
			name: "success - peer registration",
			input: &pb.RegisterIdentityWorkflowInput{
				Req: &pb.RegisterIdentityInput{
					Name:        "peer0",
					Type:        "peer",
					Affiliation: "org1",
				},
			},
			activityRes: &pb.RegisterIdentityOutput{
				Name:   "peer0",
				Type:   "peer",
				Secret: "generated-secret",
			},
		},
		{
			name: "success - orderer registration",
			input: &pb.RegisterIdentityWorkflowInput{
				Req: &pb.RegisterIdentityInput{
					Name:        "orderer0",
					Type:        "orderer",
					Affiliation: "org1",
				},
			},
			activityRes: &pb.RegisterIdentityOutput{
				Name:   "orderer0",
				Type:   "orderer",
				Secret: "generated-secret",
			},
		},
		{
			name: "activity failure propagated",
			input: &pb.RegisterIdentityWorkflowInput{
				Req: &pb.RegisterIdentityInput{
					Name: "peer1",
					Type: "peer",
				},
			},
			activityErr: errActivityFailed,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				RegisterIdentity(mock.Anything, tt.input.Req).
				Return(tt.activityRes, tt.activityErr)

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testRegisterIdentityWorkflow, tt.input)

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

				var result *pb.RegisterIdentityOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetName() != tt.activityRes.GetName() {
					t.Errorf("name mismatch: got %q, want %q", result.GetName(), tt.activityRes.GetName())
				}
			}
		})
	}
}

func TestEnrollIdentityWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.EnrollIdentityWorkflowInput
		activityRes *pb.EnrollIdentityOutput
		activityErr error
		expectErr   bool
	}{
		{
			name: "success - enroll with default profile",
			input: &pb.EnrollIdentityWorkflowInput{
				Req: &pb.EnrollIdentityInput{
					Name:   "peer0",
					Secret: "secret123",
				},
			},
			activityRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("cert-data"),
				PrivateKey:  []byte("key-data"),
				CaCert:      []byte("ca-cert-data"),
			},
		},
		{
			name: "success - enroll with TLS profile",
			input: &pb.EnrollIdentityWorkflowInput{
				Req: &pb.EnrollIdentityInput{
					Name:    "peer0",
					Secret:  "secret123",
					Profile: "tls",
				},
			},
			activityRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("tls-cert-data"),
				PrivateKey:  []byte("tls-key-data"),
				CaCert:      []byte("ca-cert-data"),
				Profile:     "tls",
			},
		},
		{
			name: "activity failure propagated",
			input: &pb.EnrollIdentityWorkflowInput{
				Req: &pb.EnrollIdentityInput{
					Name:   "peer1",
					Secret: "invalid",
				},
			},
			activityErr: errActivityFailed,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				EnrollIdentity(mock.Anything, tt.input.Req).
				Return(tt.activityRes, tt.activityErr)

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testEnrollIdentityWorkflow, tt.input)

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

				var result *pb.EnrollIdentityOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if len(result.GetCertificate()) == 0 || len(result.GetPrivateKey()) == 0 {
					t.Error("expected certificate and private key in result")
				}
			}
		})
	}
}

func TestAssembleMspWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.AssembleMspWorkflowInput
		activityRes *pb.AssembleMspOutput
		activityErr error
		expectErr   bool
	}{
		{
			name: "success - assemble MSP",
			input: &pb.AssembleMspWorkflowInput{
				Req: &pb.AssembleMspInput{
					SigningEnrollment: &pb.EnrollIdentityOutput{
						Certificate: []byte("signing-cert"),
						PrivateKey:  []byte("signing-key"),
					},
					TlsEnrollment: &pb.EnrollIdentityOutput{
						Certificate: []byte("tls-cert"),
						PrivateKey:  []byte("tls-key"),
					},
					CaCerts:      [][]byte{[]byte("ca-cert-1"), []byte("ca-cert-2")},
					NodeOuConfig: &pb.NodeOUConfig{Enable: true},
				},
			},
			activityRes: &pb.AssembleMspOutput{
				SigningCert:  []byte("signing-cert"),
				SigningKey:   []byte("signing-key"),
				TlsCert:      []byte("tls-cert"),
				TlsKey:       []byte("tls-key"),
				CaCerts:      [][]byte{[]byte("ca-cert-1"), []byte("ca-cert-2")},
				NodeOuConfig: &pb.NodeOUConfig{Enable: true},
			},
		},
		{
			name: "activity failure propagated",
			input: &pb.AssembleMspWorkflowInput{
				Req: &pb.AssembleMspInput{
					SigningEnrollment: &pb.EnrollIdentityOutput{},
				},
			},
			activityErr: errActivityFailed,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				AssembleMsp(mock.Anything, mock.AnythingOfType("*fabricidentityv1.AssembleMspInput")).
				Return(tt.activityRes, tt.activityErr)

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testAssembleMspWorkflow, tt.input)

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

				var result *pb.AssembleMspOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if len(result.GetSigningCert()) == 0 || len(result.GetTlsCert()) == 0 {
					t.Error("expected signing and TLS certificates in result")
				}
			}
		})
	}
}
