package cluster

import (
	"context"
	"errors"
	"strings"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	fabricv1 "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	helmpb "github.com/metacensus/infra/temporal/gen/workflows/helm/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
	kube "github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const testOperatorID = "testoperator"
const testOrg = "testorg"

// Cluster workflows act ON BEHALF OF the explicit typed Owner carried in each
// input. The injected operator config is platform-level context (TLS root
// trust anchors, worker environment).
var testOperatorConfig = workerconfig.NewOperatorConfig(testOperatorID)

// newTestWorkflows builds the Workflows under test with the platform operator injected
func newTestWorkflows() *Workflows {
	return NewWorkflows(&testOperatorConfig)
}

func orgOwner(org string) *fabricv1.Owner {
	return &fabricv1.Owner{Id: &fabricv1.Owner_OrgId{OrgId: org}}
}

func operatorOwner(operatorID string) *fabricv1.Owner {
	return &fabricv1.Owner{Id: &fabricv1.Owner_OperatorId{OperatorId: operatorID}}
}

// =============================================================================
// Constructor / Validation Tests
// =============================================================================
// These tests verify input validation without needing a workflow environment.
// They're fast and focused on the validation logic.

func assertConstructorResult(t *testing.T, wfIsNil bool, err error, wantErr bool, errSubstr string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if errSubstr != "" && !strings.Contains(err.Error(), errSubstr) {
			t.Errorf("expected error containing %q, got: %v", errSubstr, err)
		}
		if !wfIsNil {
			t.Error("expected nil workflow on error")
		}
	} else {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wfIsNil {
			t.Fatal("expected non-nil workflow")
		}
	}
}

func TestNewCreateNamespaceWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateNamespaceInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - org owner",
			input:   &pb.CreateNamespaceInput{Owner: orgOwner(testOrg)},
			wantErr: false,
		},
		{
			name:    "success - operator owner",
			input:   &pb.CreateNamespaceInput{Owner: operatorOwner(testOperatorID)},
			wantErr: false,
		},
		{
			name:      "error - missing owner",
			input:     &pb.CreateNamespaceInput{},
			wantErr:   true,
			errSubstr: "owner must specify operator_id or org_id",
		},
		{
			name:      "error - empty org_id",
			input:     &pb.CreateNamespaceInput{Owner: orgOwner("")},
			wantErr:   true,
			errSubstr: "owner.org_id must not be empty",
		},
		{
			name:      "error - empty operator_id",
			input:     &pb.CreateNamespaceInput{Owner: operatorOwner("")},
			wantErr:   true,
			errSubstr: "owner.operator_id must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateNamespaceWorkflow(tt.input)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

func TestNewCreateServiceAccountWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateServiceAccountInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - org owner",
			input:   &pb.CreateServiceAccountInput{Owner: orgOwner(testOrg)},
			wantErr: false,
		},
		{
			name:    "success - operator owner",
			input:   &pb.CreateServiceAccountInput{Owner: operatorOwner(testOperatorID)},
			wantErr: false,
		},
		{
			name:      "error - missing owner",
			input:     &pb.CreateServiceAccountInput{},
			wantErr:   true,
			errSubstr: "owner must specify operator_id or org_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateServiceAccountWorkflow(tt.input)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

func TestNewCreateDeploymentWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateDeploymentInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - valid worker",
			input:   &pb.CreateDeploymentInput{Owner: orgOwner(testOrg), Component: "worker"},
			wantErr: false,
		},
		{
			name:    "success - valid certificate-store",
			input:   &pb.CreateDeploymentInput{Owner: orgOwner(testOrg), Component: "certificate-store"},
			wantErr: false,
		},
		{
			name:    "success - valid certificate-authority",
			input:   &pb.CreateDeploymentInput{Owner: operatorOwner(testOperatorID), Component: "certificate-authority-tls-root"},
			wantErr: false,
		},
		{
			name:    "success - valid couchdb",
			input:   &pb.CreateDeploymentInput{Owner: orgOwner(testOrg), Component: "couchdb-peer0"},
			wantErr: false,
		},
		{
			name:    "success - valid peer",
			input:   &pb.CreateDeploymentInput{Owner: orgOwner(testOrg), Component: "peer0"},
			wantErr: false,
		},
		{
			name:    "success - valid orderer",
			input:   &pb.CreateDeploymentInput{Owner: orgOwner(testOrg), Component: "orderer0"},
			wantErr: false,
		},
		{
			name:      "error - missing owner",
			input:     &pb.CreateDeploymentInput{Component: "worker"},
			wantErr:   true,
			errSubstr: "owner must specify operator_id or org_id",
		},
		{
			name:      "error - empty component",
			input:     &pb.CreateDeploymentInput{Owner: orgOwner(testOrg), Component: ""},
			wantErr:   true,
			errSubstr: "component is required",
		},
		{
			name:      "error - both empty",
			input:     &pb.CreateDeploymentInput{},
			wantErr:   true,
			errSubstr: "owner must specify operator_id or org_id", // Owner checked first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateDeploymentWorkflow(tt.input, &testOperatorConfig)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

// =============================================================================
// Workflow Execution Tests
// =============================================================================
// These tests verify workflow orchestration with mocked activities.
// All inputs are valid - validation is tested separately above.

// registerActivityStubs registers activity stubs required for testing
func registerActivityStubs(env *testsuite.TestWorkflowEnvironment) {
	env.RegisterActivityWithOptions(
		func(ctx context.Context, input *helmpb.RenderChartWithTemplateInput) (*helmpb.RenderChartOutput, error) {
			return nil, nil
		},
		activity.RegisterOptions{Name: helmpb.RenderChartWithTemplateActivityName},
	)
	env.RegisterActivityWithOptions(
		func(ctx context.Context, input *kubepb.ApplyInput) (*kubepb.ApplyOutput, error) {
			return nil, nil
		},
		activity.RegisterOptions{Name: kubepb.ApplyActivityName},
	)
	env.RegisterActivityWithOptions(
		func(ctx context.Context, input *kube.EnsureTemporalMTLSInput) (*kube.EnsureTemporalMTLSOutput, error) {
			return &kube.EnsureTemporalMTLSOutput{}, nil
		},
		activity.RegisterOptions{Name: "EnsureTemporalMTLSClientCert"},
	)
}

func TestCreateNamespaceWorkflow_Execute(t *testing.T) {
	tests := []struct {
		name           string
		owner          *fabricv1.Owner
		renderErr      error
		renderFailures map[string]string
		applyErr       error
		applyFailed    int32
		wantErr        bool
		errSubstr      string
		wantNamespace  string
	}{
		{
			name:          "success - org namespace created",
			owner:         orgOwner(testOrg),
			wantErr:       false,
			wantNamespace: "fabric-shard-testorg",
		},
		{
			name:          "success - operator namespace created",
			owner:         operatorOwner(testOperatorID),
			wantErr:       false,
			wantNamespace: "fabric",
		},
		{
			name:      "error - render chart fails",
			owner:     orgOwner(testOrg),
			renderErr: errors.New("helm rendering failed"),
			wantErr:   true,
			errSubstr: "render chart",
		},
		{
			name:           "error - render chart has failures",
			owner:          orgOwner(testOrg),
			renderFailures: map[string]string{"template": "invalid syntax"},
			wantErr:        true,
			errSubstr:      "chart rendering failed",
		},
		{
			name:      "error - apply fails",
			owner:     orgOwner(testOrg),
			applyErr:  errors.New("kubectl error"),
			wantErr:   true,
			errSubstr: "apply resources",
		},
		{
			name:        "error - apply has failed resources",
			owner:       orgOwner(testOrg),
			applyFailed: 2,
			wantErr:     true,
			errSubstr:   "failed to apply 2 resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register workflows
			workflows := newTestWorkflows()
			pb.RegisterClusterWorkflows(env, workflows)

			// Register activity stubs
			registerActivityStubs(env)

			// Setup activity mocks based on test case
			if tt.renderErr != nil {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.renderErr)
			} else if tt.renderFailures != nil {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(&helmpb.RenderChartOutput{Failures: tt.renderFailures}, nil)
			} else {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(&helmpb.RenderChartOutput{
						Resources: map[string]*helmpb.Resource{
							"namespace": {ManifestYaml: []byte("apiVersion: v1\nkind: Namespace")},
						},
					}, nil)

				if tt.applyErr != nil {
					env.OnActivity(kubepb.ApplyActivityName, mock.Anything, mock.Anything).
						Return(nil, tt.applyErr)
				} else {
					env.OnActivity(kubepb.ApplyActivityName, mock.Anything, mock.Anything).
						Return(&kubepb.ApplyOutput{
							Summary: &kubepb.ApplySummary{Applied: 1, Failed: tt.applyFailed},
						}, nil)
				}
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateNamespaceInput) (*pb.CreateNamespaceOutput, error) {
				wf, err := workflows.CreateNamespace(ctx, &pb.CreateNamespaceWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateNamespaceInput{Owner: tt.owner})

			// Check results
			err := env.GetWorkflowError()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result pb.CreateNamespaceOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get result: %v", err)
				}
				if result.NamespaceName != tt.wantNamespace {
					t.Errorf("namespace = %q, want %q", result.NamespaceName, tt.wantNamespace)
				}
			}
		})
	}
}

func TestCreateServiceAccountWorkflow_Execute(t *testing.T) {
	tests := []struct {
		name           string
		owner          *fabricv1.Owner
		renderErr      error
		renderFailures map[string]string
		applyErr       error
		applyFailed    int32
		wantErr        bool
		errSubstr      string
	}{
		{
			name:    "success - org service account created",
			owner:   orgOwner(testOrg),
			wantErr: false,
		},
		{
			name:    "success - operator service account created",
			owner:   operatorOwner(testOperatorID),
			wantErr: false,
		},
		{
			name:      "error - render chart fails",
			owner:     orgOwner(testOrg),
			renderErr: errors.New("helm rendering failed"),
			wantErr:   true,
			errSubstr: "render chart",
		},
		{
			name:           "error - render chart has failures",
			owner:          orgOwner(testOrg),
			renderFailures: map[string]string{"template": "invalid syntax"},
			wantErr:        true,
			errSubstr:      "chart rendering failed",
		},
		{
			name:      "error - apply fails",
			owner:     orgOwner(testOrg),
			applyErr:  errors.New("kubectl error"),
			wantErr:   true,
			errSubstr: "apply resources",
		},
		{
			name:        "error - apply has failed resources",
			owner:       orgOwner(testOrg),
			applyFailed: 1,
			wantErr:     true,
			errSubstr:   "failed to apply 1 resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register workflows
			workflows := newTestWorkflows()
			pb.RegisterClusterWorkflows(env, workflows)

			// Register activity stubs
			registerActivityStubs(env)

			// Setup activity mocks based on test case
			if tt.renderErr != nil {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.renderErr)
			} else if tt.renderFailures != nil {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(&helmpb.RenderChartOutput{Failures: tt.renderFailures}, nil)
			} else {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(&helmpb.RenderChartOutput{
						Resources: map[string]*helmpb.Resource{
							"sa": {ManifestYaml: []byte("apiVersion: v1\nkind: ServiceAccount")},
						},
					}, nil)

				if tt.applyErr != nil {
					env.OnActivity(kubepb.ApplyActivityName, mock.Anything, mock.Anything).
						Return(nil, tt.applyErr)
				} else {
					env.OnActivity(kubepb.ApplyActivityName, mock.Anything, mock.Anything).
						Return(&kubepb.ApplyOutput{
							Summary: &kubepb.ApplySummary{Applied: 1, Failed: tt.applyFailed},
						}, nil)
				}
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateServiceAccountInput) (*pb.CreateServiceAccountOutput, error) {
				wf, err := workflows.CreateServiceAccount(ctx, &pb.CreateServiceAccountWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateServiceAccountInput{Owner: tt.owner})

			// Check results
			err := env.GetWorkflowError()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCreateDeploymentWorkflow_Execute(t *testing.T) {
	tests := []struct {
		name               string
		owner              *fabricv1.Owner
		component          string
		renderErr          error
		renderFailures     map[string]string
		applyErr           error
		applyFailed        int32
		wantErr            bool
		errSubstr          string
		wantDeploymentName string
	}{
		{
			name:               "success - shard worker deployed",
			owner:              orgOwner(testOrg),
			component:          "worker",
			wantErr:            false,
			wantDeploymentName: "fabric-worker-testorg",
		},
		{
			name:               "success - shared fabric worker deployed",
			owner:              operatorOwner(testOperatorID),
			component:          "worker",
			wantErr:            false,
			wantDeploymentName: "fabric-worker-" + testOperatorID,
		},
		{
			name:               "success - org certificate-store deployed",
			owner:              orgOwner(testOrg),
			component:          "certificate-store",
			wantErr:            false,
			wantDeploymentName: "certificate-store",
		},
		{
			name:               "success - operator certificate-store deployed",
			owner:              operatorOwner(testOperatorID),
			component:          "certificate-store",
			wantErr:            false,
			wantDeploymentName: "certificate-store",
		},
		{
			name:               "success - certificate-authority-tls-root deployed (operator-owned)",
			owner:              operatorOwner(testOperatorID),
			component:          "certificate-authority-tls-root",
			wantErr:            false,
			wantDeploymentName: "certificate-authority-tls-root",
		},
		{
			name:               "success - certificate-authority-org-int deployed (org-owned)",
			owner:              orgOwner(testOrg),
			component:          "certificate-authority-org-int",
			wantErr:            false,
			wantDeploymentName: "certificate-authority-org-int",
		},
		{
			name:               "success - couchdb-peer0 deployed",
			owner:              orgOwner(testOrg),
			component:          "couchdb-peer0",
			wantErr:            false,
			wantDeploymentName: "couchdb-peer0",
		},
		{
			name:               "success - peer0 deployed",
			owner:              orgOwner(testOrg),
			component:          "peer0",
			wantErr:            false,
			wantDeploymentName: "peer0",
		},
		{
			name:               "success - orderer0 deployed",
			owner:              orgOwner(testOrg),
			component:          "orderer0",
			wantErr:            false,
			wantDeploymentName: "orderer0",
		},
		{
			name:      "error - unsupported component",
			owner:     orgOwner(testOrg),
			component: "unknown",
			wantErr:   true,
			errSubstr: "unsupported component type",
		},
		{
			name:      "error - invalid certificate-authority format (too few parts)",
			owner:     orgOwner(testOrg),
			component: "certificate-authority-tls",
			wantErr:   true,
			errSubstr: "invalid certificate authority component format",
		},
		{
			name:      "error - invalid certificate-authority format (too many parts)",
			owner:     orgOwner(testOrg),
			component: "certificate-authority-tls-root-extra",
			wantErr:   true,
			errSubstr: "invalid certificate authority component format",
		},
		{
			name:      "error - invalid cert type",
			owner:     orgOwner(testOrg),
			component: "certificate-authority-invalid-root",
			wantErr:   true,
			errSubstr: "invalid cert type",
		},
		{
			name:      "error - invalid couchdb format",
			owner:     orgOwner(testOrg),
			component: "couchdb-peer0-extra",
			wantErr:   true,
			errSubstr: "invalid couchdb component format",
		},
		{
			name:      "error - render chart fails",
			owner:     orgOwner(testOrg),
			component: "worker",
			renderErr: errors.New("helm rendering failed"),
			wantErr:   true,
			errSubstr: "render chart",
		},
		{
			name:           "error - render chart has failures",
			owner:          orgOwner(testOrg),
			component:      "worker",
			renderFailures: map[string]string{"template": "invalid syntax"},
			wantErr:        true,
			errSubstr:      "chart rendering failed",
		},
		{
			name:      "error - apply fails",
			owner:     orgOwner(testOrg),
			component: "worker",
			applyErr:  errors.New("kubectl error"),
			wantErr:   true,
			errSubstr: "apply resources",
		},
		{
			name:        "error - apply has failed resources",
			owner:       orgOwner(testOrg),
			component:   "worker",
			applyFailed: 3,
			wantErr:     true,
			errSubstr:   "failed to apply 3 resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register workflows
			workflows := newTestWorkflows()
			pb.RegisterClusterWorkflows(env, workflows)

			// Register activity stubs
			registerActivityStubs(env)
			env.OnActivity("EnsureTemporalMTLSClientCert", mock.Anything, mock.Anything).
				Return(&kube.EnsureTemporalMTLSOutput{}, nil)

			// Setup activity mocks based on test case
			if tt.renderErr != nil {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.renderErr)
			} else if tt.renderFailures != nil {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(&helmpb.RenderChartOutput{Failures: tt.renderFailures}, nil)
			} else {
				env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
					Return(&helmpb.RenderChartOutput{
						Resources: map[string]*helmpb.Resource{
							"deployment": {ManifestYaml: []byte("apiVersion: apps/v1\nkind: Deployment")},
						},
					}, nil)

				if tt.applyErr != nil {
					env.OnActivity(kubepb.ApplyActivityName, mock.Anything, mock.Anything).
						Return(nil, tt.applyErr)
				} else {
					env.OnActivity(kubepb.ApplyActivityName, mock.Anything, mock.Anything).
						Return(&kubepb.ApplyOutput{
							Summary: &kubepb.ApplySummary{Applied: 1, Failed: tt.applyFailed},
						}, nil)
				}
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateDeploymentInput) (*pb.CreateDeploymentOutput, error) {
				wf, err := workflows.CreateDeployment(ctx, &pb.CreateDeploymentWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateDeploymentInput{Owner: tt.owner, Component: tt.component})

			// Check results
			err := env.GetWorkflowError()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result pb.CreateDeploymentOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get result: %v", err)
				}
				if result.DeploymentName != tt.wantDeploymentName {
					t.Errorf("deployment = %q, want %q", result.DeploymentName, tt.wantDeploymentName)
				}
			}
		})
	}
}
