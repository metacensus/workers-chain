package cell

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	nexus "github.com/nexus-rpc/sdk-go/nexus"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cell/v1"
	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1nexus"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	temporalpb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
	temporalmocks "github.com/metacensus/infra/temporal/mocks/workflows/temporal/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

const testOrg = "testorg"
const testOperatorID = "testoperator"

// Cell workflows run on the infra worker and act on behalf of explicit
// owners; the injected operator identifies the platform operator.
var testOperatorConfig = workerconfig.NewOperatorConfig(testOperatorID)

// newTestWorkflows builds the Workflows under test with the platform operator injected
func newTestWorkflows() *Workflows {
	return NewWorkflows(&testOperatorConfig)
}

// testCASpecs is a minimal ordered TLS CA chain spec for InitFabric inputs
func testCASpecs() []*fabricpb.CertificateAuthoritySpec {
	tlsRootRef := &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	tlsIntRef := &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"}
	validity := durationpb.New(15 * 365 * 24 * time.Hour)
	return []*fabricpb.CertificateAuthoritySpec{
		{Self: tlsRootRef, ValidityDuration: validity},
		{Self: tlsIntRef, SigningIssuer: tlsRootRef, TlsIssuer: tlsRootRef, ValidityDuration: validity},
	}
}

// =============================================================================
// Constructor / Validation Tests
// =============================================================================
// These tests verify input validation without needing a workflow environment.
// They're fast and focused on the validation logic.

func TestNewInitFabricWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.InitFabricWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - valid input",
			input:   &pb.InitFabricWorkflowInput{Req: &pb.InitFabricInput{CertificateAuthorities: testCASpecs()}},
			wantErr: false,
		},
		{
			name:      "error - no certificate authorities",
			input:     &pb.InitFabricWorkflowInput{Req: &pb.InitFabricInput{}},
			wantErr:   true,
			errSubstr: "no certificate authorities specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newInitFabricWorkflow(tt.input, &testOperatorConfig)

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
			wf, err := newInitFabricShardWorkflow(tt.input, &testOperatorConfig)

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
				workflows := newTestWorkflows()
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

// InitFabricShard registers one certificate-issuance endpoint per org, into the
// operator's namespace. That per-org endpoint is what the Temporal server keys
// authorization on: only this org's namespace may use it, so a shard cannot
// request another org's identity.
func TestInitFabricShardRegistersPerOrgCertificateEndpoint(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	registerChildWorkflowStubs(env)
	mockChildWorkflow[temporalpb.CreateNamespaceOutput](env, temporalpb.CreateNamespaceWorkflowName, nil)
	mockChildWorkflow[clusterpb.CreateNamespaceOutput](env, clusterpb.CreateNamespaceWorkflowName, nil)
	mockChildWorkflow[clusterpb.CreateServiceAccountOutput](env, clusterpb.CreateServiceAccountWorkflowName, nil)
	mockChildWorkflow[secretspb.CreateAccessOutput](env, secretspb.CreateAccessWorkflowName, nil)
	mockChildWorkflow[clusterpb.CreateDeploymentOutput](env, clusterpb.CreateDeploymentWorkflowName, nil)

	endpoints := map[string]string{} // endpoint name -> target Temporal namespace
	mockActivities := temporalmocks.NewMockAdminActivities(t)
	mockActivities.EXPECT().CreateNexusEndpoint(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, input *temporalpb.CreateNexusEndpointInput) (*temporalpb.CreateNexusEndpointOutput, error) {
			endpoints[input.GetName()] = input.GetNamespace()
			return &temporalpb.CreateNexusEndpointOutput{}, nil
		}).Maybe()
	temporalpb.RegisterAdminActivities(env, mockActivities)

	nexusSvc, err := fabricv1nexus.NewFabricNexusService(&fabricv1nexus.UnimplementedFabricNexusHandler{})
	if err != nil {
		t.Fatalf("failed to create nexus service: %v", err)
	}
	env.RegisterNexusService(nexusSvc)
	env.OnNexusOperation(nexusSvc, fabricv1nexus.FabricBootstrapShardOperationName, mock.Anything, mock.Anything).
		Return(&nexus.HandlerStartOperationResultSync[*fabricpb.BootstrapShardOutput]{Value: &fabricpb.BootstrapShardOutput{}}, nil)

	env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.InitFabricShardInput) (*pb.InitFabricShardOutput, error) {
		workflows := newTestWorkflows()
		wf, err := workflows.InitFabricShard(ctx, &pb.InitFabricShardWorkflowInput{Req: input})
		if err != nil {
			return nil, err
		}
		return wf.Execute(ctx)
	}, &pb.InitFabricShardInput{Org: testOrg})

	var result *pb.InitFabricShardOutput
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("unexpected workflow error: %v", err)
	}

	orgCfg := workerconfig.NewOrgConfig(testOrg)
	certEndpoint := orgCfg.NexusEndpointCreateCertificate()
	gotNamespace, ok := endpoints[certEndpoint]
	if !ok {
		t.Fatalf("endpoint %q was not registered; registered: %v", certEndpoint, endpoints)
	}
	if want := testOperatorConfig.TemporalNamespace(); gotNamespace != want {
		t.Errorf("endpoint %q targets namespace %q, want the operator's %q", certEndpoint, gotNamespace, want)
	}
	if _, ok := endpoints[orgCfg.NexusEndpointBootstrapShard()]; !ok {
		t.Errorf("endpoint %q was not registered; registered: %v", orgCfg.NexusEndpointBootstrapShard(), endpoints)
	}
}

func TestInitFabricWorkflow(t *testing.T) {
	tests := []struct {
		name                         string
		createTemporalNamespaceErr   error
		createKubernetesNamespaceErr error
		createServiceAccountErr      error
		createVaultAccessErr         error
		createDeploymentErr          error
		createNexusEndpointErr       error
		bootstrapFabricErr           error
		expectedErr                  error // nil = success
	}{
		// Note: "no certificate authorities" validation is tested in TestNewInitFabricWorkflow above
		{name: "failure - create fabric temporal namespace fails",
			createTemporalNamespaceErr: errors.New("temporal namespace error"), expectedErr: ErrCreateFabricTemporalNamespace},
		{name: "failure - create operator kubernetes namespace fails",
			createKubernetesNamespaceErr: errors.New("k8s namespace error"), expectedErr: ErrCreateOperatorKubernetesNamespace},
		{name: "failure - create operator service account fails",
			createServiceAccountErr: errors.New("sa error"), expectedErr: ErrCreateOperatorServiceAccount},
		{name: "failure - create operator vault access fails",
			createVaultAccessErr: errors.New("vault error"), expectedErr: ErrCreateOperatorVaultAccess},
		{name: "failure - deploy shared fabric worker fails",
			createDeploymentErr: errors.New("deploy error"), expectedErr: ErrDeployFabricSharedWorker},
		{name: "failure - create nexus endpoint fails",
			createNexusEndpointErr: errors.New("nexus error"), expectedErr: ErrCreateBootstrapFabricEndpoint},
		{name: "failure - bootstrap fabric via nexus fails",
			bootstrapFabricErr: errors.New("bootstrap error"), expectedErr: ErrBootstrapFabric},
		{name: "success - all steps complete", expectedErr: nil},
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

			// Mock BootstrapFabric Nexus operation: the ordered CA specs must
			// pass through to the fabric worker unchanged
			nexusSvc, err := fabricv1nexus.NewFabricNexusService(&fabricv1nexus.UnimplementedFabricNexusHandler{})
			if err != nil {
				t.Fatalf("failed to create nexus service: %v", err)
			}
			env.RegisterNexusService(nexusSvc)
			if tt.bootstrapFabricErr != nil {
				env.OnNexusOperation(nexusSvc, fabricv1nexus.FabricBootstrapFabricOperationName, mock.Anything, mock.Anything).
					Return(nil, tt.bootstrapFabricErr)
			} else {
				env.OnNexusOperation(nexusSvc, fabricv1nexus.FabricBootstrapFabricOperationName, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						input, ok := args.Get(1).(*fabricpb.BootstrapFabricInput)
						if !ok {
							t.Errorf("unexpected BootstrapFabric input type %T", args.Get(1))
							return
						}
						if got := len(input.GetCertificateAuthorities()); got != 2 {
							t.Errorf("BootstrapFabric received %d CA specs, want 2", got)
						}
					}).
					Return(&nexus.HandlerStartOperationResultSync[*fabricpb.BootstrapFabricOutput]{Value: &fabricpb.BootstrapFabricOutput{}}, nil)
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.InitFabricInput) (*pb.InitFabricOutput, error) {
				workflows := newTestWorkflows()
				wf, err := workflows.InitFabric(ctx, &pb.InitFabricWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.InitFabricInput{CertificateAuthorities: testCASpecs()})

			var result *pb.InitFabricOutput
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
