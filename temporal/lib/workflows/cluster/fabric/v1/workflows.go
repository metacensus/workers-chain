package cluster

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	helmpb "github.com/metacensus/infra/temporal/gen/workflows/helm/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	"github.com/metacensus/infra/temporal/lib/workflows/cluster/fabric/v1/helm"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	chartNamespace     = "fabric/namespace"
	componentNamespace = "namespace"
)

// Workflows implements the ClusterWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// CreateNamespace implements the CreateNamespace workflow
func (w *Workflows) CreateNamespace(ctx workflow.Context, input *pb.CreateNamespaceWorkflowInput) (pb.CreateNamespaceWorkflow, error) {
	return &createNamespaceWorkflow{input: input.Req}, nil
}

// createNamespaceWorkflow implements pb.CreateNamespaceWorkflow
type createNamespaceWorkflow struct {
	input *pb.CreateNamespaceInput
}

// Execute creates a Kubernetes namespace
func (w *createNamespaceWorkflow) Execute(ctx workflow.Context) (*pb.CreateNamespaceOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CreateNamespace workflow started", "org", w.input.Org)

	// Build values for Helm chart
	values := helm.NewNamespaceValues(w.input.Org)
	namespaceName := values.Name

	// Convert Go struct to protobuf Struct
	valuesStruct, err := ToProtoStruct(values)
	if err != nil {
		return nil, fmt.Errorf("convert values: %w", err)
	}

	// Activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    5,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Step 1: Render Helm chart
	logger.Info("Rendering Helm chart", "chart", chartNamespace, "namespace", namespaceName)
	var renderOutput *helmpb.RenderChartOutput
	err = workflow.ExecuteActivity(activityCtx, helmpb.RenderChartWithTemplateActivityName, &helmpb.RenderChartWithTemplateInput{
		ChartName: chartNamespace,
		Component: componentNamespace,
		Values:    valuesStruct,
		Namespace: &namespaceName,
	}).Get(ctx, &renderOutput)
	if err != nil {
		return nil, fmt.Errorf("render chart: %w", err)
	}

	if len(renderOutput.Failures) > 0 {
		return nil, fmt.Errorf("chart rendering failed: %v", renderOutput.Failures)
	}

	// Step 2: Apply resources to Kubernetes
	logger.Info("Applying resources", "resource_count", len(renderOutput.Resources))
	kubeResources := make(map[string]*kubepb.Resource)
	for id, res := range renderOutput.Resources {
		kubeResources[id] = &kubepb.Resource{
			ManifestYaml: res.ManifestYaml,
		}
	}

	var applyOutput *kubepb.ApplyOutput
	err = workflow.ExecuteActivity(activityCtx, kubepb.ApplyActivityName, &kubepb.ApplyInput{
		Resources: kubeResources,
	}).Get(ctx, &applyOutput)
	if err != nil {
		return nil, fmt.Errorf("apply resources: %w", err)
	}

	if applyOutput.Summary.Failed > 0 {
		return nil, fmt.Errorf("failed to apply %d resources", applyOutput.Summary.Failed)
	}

	logger.Info("Namespace created successfully", "namespace", namespaceName)
	return &pb.CreateNamespaceOutput{
		NamespaceName: namespaceName,
	}, nil
}

// CreateServiceAccount implements the CreateServiceAccount workflow
func (w *Workflows) CreateServiceAccount(ctx workflow.Context, input *pb.CreateServiceAccountWorkflowInput) (pb.CreateServiceAccountWorkflow, error) {
	return &createServiceAccountWorkflow{input: input.Req}, nil
}

// createServiceAccountWorkflow implements pb.CreateServiceAccountWorkflow
type createServiceAccountWorkflow struct {
	input *pb.CreateServiceAccountInput
}

// Execute creates a Kubernetes service account
func (w *createServiceAccountWorkflow) Execute(ctx workflow.Context) (*pb.CreateServiceAccountOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CreateServiceAccount workflow started")

	// TODO: Implementation

	return &pb.CreateServiceAccountOutput{}, nil
}

// CreateDeployment implements the CreateDeployment workflow
func (w *Workflows) CreateDeployment(ctx workflow.Context, input *pb.CreateDeploymentWorkflowInput) (pb.CreateDeploymentWorkflow, error) {
	return &createDeploymentWorkflow{input: input.Req}, nil
}

// createDeploymentWorkflow implements pb.CreateDeploymentWorkflow
type createDeploymentWorkflow struct {
	input *pb.CreateDeploymentInput
}

// Execute deploys resources to Kubernetes
func (w *createDeploymentWorkflow) Execute(ctx workflow.Context) (*pb.CreateDeploymentOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CreateDeployment workflow started")

	// TODO: Implementation

	return &pb.CreateDeploymentOutput{}, nil
}
