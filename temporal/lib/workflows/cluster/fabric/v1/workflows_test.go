package cluster

import (
	"context"
	"errors"
	"strings"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	helmpb "github.com/metacensus/infra/temporal/gen/workflows/helm/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// =============================================================================
// Constructor / Validation Tests
// =============================================================================
// These tests verify input validation without needing a workflow environment.
// They're fast and focused on the validation logic.

func TestNewCreateNamespaceWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateNamespaceInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - valid input",
			input:   &pb.CreateNamespaceInput{Org: "org1"},
			wantErr: false,
		},
		{
			name:      "error - empty org",
			input:     &pb.CreateNamespaceInput{Org: ""},
			wantErr:   true,
			errSubstr: "org is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateNamespaceWorkflow(tt.input)

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

func TestNewCreateServiceAccountWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateServiceAccountInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - valid input",
			input:   &pb.CreateServiceAccountInput{Org: "org1"},
			wantErr: false,
		},
		{
			name:      "error - empty org",
			input:     &pb.CreateServiceAccountInput{Org: ""},
			wantErr:   true,
			errSubstr: "org is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateServiceAccountWorkflow(tt.input)

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

func TestNewCreateDeploymentWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateDeploymentInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - valid worker",
			input:   &pb.CreateDeploymentInput{Org: "org1", Component: "worker"},
			wantErr: false,
		},
		{
			name:    "success - valid certificate-store",
			input:   &pb.CreateDeploymentInput{Org: "org1", Component: "certificate-store"},
			wantErr: false,
		},
		{
			name:    "success - valid certificate-authority",
			input:   &pb.CreateDeploymentInput{Org: "org1", Component: "certificate-authority-tls-root"},
			wantErr: false,
		},
		{
			name:    "success - valid couchdb",
			input:   &pb.CreateDeploymentInput{Org: "org1", Component: "couchdb-peer0"},
			wantErr: false,
		},
		{
			name:    "success - valid peer",
			input:   &pb.CreateDeploymentInput{Org: "org1", Component: "peer0"},
			wantErr: false,
		},
		{
			name:    "success - valid orderer",
			input:   &pb.CreateDeploymentInput{Org: "org1", Component: "orderer0"},
			wantErr: false,
		},
		{
			name:      "error - empty org",
			input:     &pb.CreateDeploymentInput{Org: "", Component: "worker"},
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name:      "error - empty component",
			input:     &pb.CreateDeploymentInput{Org: "org1", Component: ""},
			wantErr:   true,
			errSubstr: "component is required",
		},
		{
			name:      "error - both empty",
			input:     &pb.CreateDeploymentInput{Org: "", Component: ""},
			wantErr:   true,
			errSubstr: "org is required", // Org checked first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateDeploymentWorkflow(tt.input)

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
}

// mockSuccessfulActivities sets up mocks for successful activity execution
func mockSuccessfulActivities(env *testsuite.TestWorkflowEnvironment) {
	env.OnActivity(helmpb.RenderChartWithTemplateActivityName, mock.Anything, mock.Anything).
		Return(&helmpb.RenderChartOutput{
			Resources: map[string]*helmpb.Resource{
				"test": {ManifestYaml: []byte("apiVersion: v1\nkind: ConfigMap")},
			},
		}, nil)
	env.OnActivity(kubepb.ApplyActivityName, mock.Anything, mock.Anything).
		Return(&kubepb.ApplyOutput{
			Summary: &kubepb.ApplySummary{Applied: 1, Failed: 0},
		}, nil)
}

func TestCreateNamespaceWorkflow_Execute(t *testing.T) {
	tests := []struct {
		name           string
		org            string
		renderErr      error
		renderFailures map[string]string
		applyErr       error
		applyFailed    int32
		wantErr        bool
		errSubstr      string
		wantNamespace  string
	}{
		{
			name:          "success - namespace created",
			org:           "testorg",
			wantErr:       false,
			wantNamespace: "fabric-testorg",
		},
		{
			name:      "error - render chart fails",
			org:       "testorg",
			renderErr: errors.New("helm rendering failed"),
			wantErr:   true,
			errSubstr: "render chart",
		},
		{
			name:           "error - render chart has failures",
			org:            "testorg",
			renderFailures: map[string]string{"template": "invalid syntax"},
			wantErr:        true,
			errSubstr:      "chart rendering failed",
		},
		{
			name:      "error - apply fails",
			org:       "testorg",
			applyErr:  errors.New("kubectl error"),
			wantErr:   true,
			errSubstr: "apply resources",
		},
		{
			name:        "error - apply has failed resources",
			org:         "testorg",
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
			workflows := NewWorkflows()
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
			}, &pb.CreateNamespaceInput{Org: tt.org})

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
		org            string
		renderErr      error
		renderFailures map[string]string
		applyErr       error
		applyFailed    int32
		wantErr        bool
		errSubstr      string
	}{
		{
			name:    "success - service account created",
			org:     "testorg",
			wantErr: false,
		},
		{
			name:      "error - render chart fails",
			org:       "testorg",
			renderErr: errors.New("helm rendering failed"),
			wantErr:   true,
			errSubstr: "render chart",
		},
		{
			name:           "error - render chart has failures",
			org:            "testorg",
			renderFailures: map[string]string{"template": "invalid syntax"},
			wantErr:        true,
			errSubstr:      "chart rendering failed",
		},
		{
			name:      "error - apply fails",
			org:       "testorg",
			applyErr:  errors.New("kubectl error"),
			wantErr:   true,
			errSubstr: "apply resources",
		},
		{
			name:        "error - apply has failed resources",
			org:         "testorg",
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
			workflows := NewWorkflows()
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
			}, &pb.CreateServiceAccountInput{Org: tt.org})

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
		org                string
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
			name:               "success - worker deployed",
			org:                "testorg",
			component:          "worker",
			wantErr:            false,
			wantDeploymentName: "fabric-worker-testorg",
		},
		{
			name:               "success - certificate-store deployed",
			org:                "testorg",
			component:          "certificate-store",
			wantErr:            false,
			wantDeploymentName: "certificate-store",
		},
		{
			name:               "success - certificate-authority-tls-root deployed",
			org:                "testorg",
			component:          "certificate-authority-tls-root",
			wantErr:            false,
			wantDeploymentName: "certificate-authority-tls-root",
		},
		{
			name:               "success - certificate-authority-org-v1 deployed",
			org:                "testorg",
			component:          "certificate-authority-org-v1",
			wantErr:            false,
			wantDeploymentName: "certificate-authority-org-v1",
		},
		{
			name:               "success - couchdb-peer0 deployed",
			org:                "testorg",
			component:          "couchdb-peer0",
			wantErr:            false,
			wantDeploymentName: "couchdb-peer0",
		},
		{
			name:               "success - peer0 deployed",
			org:                "testorg",
			component:          "peer0",
			wantErr:            false,
			wantDeploymentName: "peer0",
		},
		{
			name:               "success - orderer0 deployed",
			org:                "testorg",
			component:          "orderer0",
			wantErr:            false,
			wantDeploymentName: "orderer0",
		},
		{
			name:      "error - unsupported component",
			org:       "testorg",
			component: "unknown",
			wantErr:   true,
			errSubstr: "unsupported component type",
		},
		{
			name:      "error - invalid certificate-authority format (too few parts)",
			org:       "testorg",
			component: "certificate-authority-tls",
			wantErr:   true,
			errSubstr: "invalid certificate authority component format",
		},
		{
			name:      "error - invalid certificate-authority format (too many parts)",
			org:       "testorg",
			component: "certificate-authority-tls-root-extra",
			wantErr:   true,
			errSubstr: "invalid certificate authority component format",
		},
		{
			name:      "error - invalid cert type",
			org:       "testorg",
			component: "certificate-authority-invalid-root",
			wantErr:   true,
			errSubstr: "invalid cert type",
		},
		{
			name:      "error - invalid couchdb format",
			org:       "testorg",
			component: "couchdb-peer0-extra",
			wantErr:   true,
			errSubstr: "invalid couchdb component format",
		},
		{
			name:      "error - render chart fails",
			org:       "testorg",
			component: "worker",
			renderErr: errors.New("helm rendering failed"),
			wantErr:   true,
			errSubstr: "render chart",
		},
		{
			name:           "error - render chart has failures",
			org:            "testorg",
			component:      "worker",
			renderFailures: map[string]string{"template": "invalid syntax"},
			wantErr:        true,
			errSubstr:      "chart rendering failed",
		},
		{
			name:      "error - apply fails",
			org:       "testorg",
			component: "worker",
			applyErr:  errors.New("kubectl error"),
			wantErr:   true,
			errSubstr: "apply resources",
		},
		{
			name:        "error - apply has failed resources",
			org:         "testorg",
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
			workflows := NewWorkflows()
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
			}, &pb.CreateDeploymentInput{Org: tt.org, Component: tt.component})

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
