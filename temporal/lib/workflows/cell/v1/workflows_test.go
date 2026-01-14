package cell

import (
	"errors"
	"strings"
	"testing"

	nexus "github.com/nexus-rpc/sdk-go/nexus"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cell/v1"
	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1nexus"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	temporalpb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	temporalmocks "github.com/metacensus/infra/temporal/mocks/workflows/temporal/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const testOrg = "testorg"

// =============================================================================
// Constructor / Validation Tests
// =============================================================================
// These tests verify input validation without needing a workflow environment.
// They're fast and focused on the validation logic.

func TestNewInitFabricShardWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.InitFabricShardWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - valid input",
			input:   &pb.InitFabricShardWorkflowInput{Req: &pb.InitFabricShardInput{Org: "org1"}},
			wantErr: false,
		},
		{
			name:      "error - empty org",
			input:     &pb.InitFabricShardWorkflowInput{Req: &pb.InitFabricShardInput{Org: ""}},
			wantErr:   true,
			errSubstr: "org is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newInitFabricShardWorkflow(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
				if wf != nil {
					t.Error("expected nil workflow on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if wf == nil {
					t.Fatal("expected non-nil workflow")
				}
			}
		})
	}
}

// =============================================================================
// Workflow Execution Tests
// =============================================================================
// These tests verify workflow orchestration with mocked activities/child workflows.
// All inputs are valid - validation is tested separately above.

// registerChildWorkflowStubs registers child workflow stubs for mocking.
// Stubs return zero values - they're never executed when properly mocked.
func registerChildWorkflowStubs(env *testsuite.TestWorkflowEnvironment) {
	register := func(name string, fn any) {
		env.RegisterWorkflowWithOptions(fn, workflow.RegisterOptions{Name: name})
	}

	register(temporalpb.CreateNamespaceWorkflowName,
		func(workflow.Context, *temporalpb.CreateNamespaceInput) (*temporalpb.CreateNamespaceOutput, error) {
			return nil, nil
		})
	register(clusterpb.CreateNamespaceWorkflowName,
		func(workflow.Context, *clusterpb.CreateNamespaceInput) (*clusterpb.CreateNamespaceOutput, error) {
			return nil, nil
		})
	register(clusterpb.CreateServiceAccountWorkflowName,
		func(workflow.Context, *clusterpb.CreateServiceAccountInput) (*clusterpb.CreateServiceAccountOutput, error) {
			return nil, nil
		})
	register(secretspb.CreateAccessWorkflowName,
		func(workflow.Context, *secretspb.CreateAccessInput) (*secretspb.CreateAccessOutput, error) {
			return nil, nil
		})
	register(clusterpb.CreateDeploymentWorkflowName,
		func(workflow.Context, *clusterpb.CreateDeploymentInput) (*clusterpb.CreateDeploymentOutput, error) {
			return nil, nil
		})
}

// mockChildWorkflow sets up a mock for a child workflow, returning success or the given error.
func mockChildWorkflow[O any](env *testsuite.TestWorkflowEnvironment, name string, err error) {
	if err != nil {
		env.OnWorkflow(name, mock.Anything, mock.Anything).Return(nil, err)
	} else {
		env.OnWorkflow(name, mock.Anything, mock.Anything).Return(new(O), nil)
	}
}

func TestInitFabricShardWorkflow(t *testing.T) {
	tests := []struct {
		name                         string
		org                          string
		createTemporalNamespaceErr   error
		createKubernetesNamespaceErr error
		createServiceAccountErr      error
		createVaultAccessErr         error
		createDeploymentErr          error
		createNexusEndpointErr       error
		bootstrapShardErr            error
		expectedErr                  error // nil = success
	}{
		// Note: "org is required" validation is tested in TestNewInitFabricShardWorkflow above
		{name: "failure - create temporal namespace fails", org: testOrg,
			createTemporalNamespaceErr: errors.New("temporal namespace error"), expectedErr: ErrCreateTemporalNamespace},
		{name: "failure - create kubernetes namespace fails", org: testOrg,
			createKubernetesNamespaceErr: errors.New("k8s namespace error"), expectedErr: ErrCreateKubernetesNamespace},
		{name: "failure - create service account fails", org: testOrg,
			createServiceAccountErr: errors.New("sa error"), expectedErr: ErrCreateServiceAccount},
		{name: "failure - create vault access fails", org: testOrg,
			createVaultAccessErr: errors.New("vault error"), expectedErr: ErrCreateVaultAccess},
		{name: "failure - deploy fabric worker fails", org: testOrg,
			createDeploymentErr: errors.New("deploy error"), expectedErr: ErrDeployFabricWorker},
		{name: "failure - create nexus endpoint fails", org: testOrg,
			createNexusEndpointErr: errors.New("nexus error"), expectedErr: ErrCreateNexusEndpoint},
		{name: "failure - bootstrap shard via nexus fails", org: testOrg,
			bootstrapShardErr: errors.New("bootstrap error"), expectedErr: ErrBootstrapShard},
		{name: "success - all steps complete", org: testOrg, expectedErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register and mock child workflows
			registerChildWorkflowStubs(env)
			mockChildWorkflow[temporalpb.CreateNamespaceOutput](env, temporalpb.CreateNamespaceWorkflowName, tt.createTemporalNamespaceErr)
			mockChildWorkflow[clusterpb.CreateNamespaceOutput](env, clusterpb.CreateNamespaceWorkflowName, tt.createKubernetesNamespaceErr)
			mockChildWorkflow[clusterpb.CreateServiceAccountOutput](env, clusterpb.CreateServiceAccountWorkflowName, tt.createServiceAccountErr)
			mockChildWorkflow[secretspb.CreateAccessOutput](env, secretspb.CreateAccessWorkflowName, tt.createVaultAccessErr)
			mockChildWorkflow[clusterpb.CreateDeploymentOutput](env, clusterpb.CreateDeploymentWorkflowName, tt.createDeploymentErr)

			// Mock CreateNexusEndpoint activity
			mockActivities := temporalmocks.NewMockAdminActivities(t)
			if tt.createNexusEndpointErr != nil {
				mockActivities.EXPECT().CreateNexusEndpoint(mock.Anything, mock.Anything).Return(nil, tt.createNexusEndpointErr).Maybe()
			} else {
				mockActivities.EXPECT().CreateNexusEndpoint(mock.Anything, mock.Anything).Return(&temporalpb.CreateNexusEndpointOutput{}, nil).Maybe()
			}
			temporalpb.RegisterAdminActivities(env, mockActivities)

			// Mock BootstrapShard Nexus operation
			nexusSvc, err := fabricv1nexus.NewFabricNexusService(&fabricv1nexus.UnimplementedFabricNexusHandler{})
			if err != nil {
				t.Fatalf("failed to create nexus service: %v", err)
			}
			env.RegisterNexusService(nexusSvc)
			if tt.bootstrapShardErr != nil {
				env.OnNexusOperation(nexusSvc, fabricv1nexus.FabricBootstrapShardOperationName, mock.Anything, mock.Anything).
					Return(nil, tt.bootstrapShardErr)
			} else {
				env.OnNexusOperation(nexusSvc, fabricv1nexus.FabricBootstrapShardOperationName, mock.Anything, mock.Anything).
					Return(&nexus.HandlerStartOperationResultSync[*fabricpb.BootstrapShardOutput]{Value: &fabricpb.BootstrapShardOutput{}}, nil)
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.InitFabricShardInput) (*pb.InitFabricShardOutput, error) {
				workflows := NewWorkflows()
				wf, err := workflows.InitFabricShard(ctx, &pb.InitFabricShardWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.InitFabricShardInput{Org: tt.org})

			var result *pb.InitFabricShardOutput
			err = env.GetWorkflowResult(&result)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				// Temporal serializes workflow errors, breaking Go's error chain, so we check the message.
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}
