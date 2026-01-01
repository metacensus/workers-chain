package cluster

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	helmpb "github.com/metacensus/infra/temporal/gen/workflows/helm/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"github.com/metacensus/infra/temporal/lib/workflows/cluster/fabric/v1/helm"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	chartNamespace        = "fabric/namespace"
	chartServiceAccount   = "fabric/service-account"
	chartWorker           = "fabric/worker"
	chartCertificateStore = "fabric/certificate-store"

	componentNamespace        = "namespace"
	componentServiceAccount   = "service-account"
	componentWorker           = "worker"
	componentCertificateStore = "certificate-store"
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
	orgCfg := config.OrgConfig{OrgID: w.input.Org}
	values := helm.NewNamespaceValues(orgCfg)
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
	orgCfg := config.OrgConfig{OrgID: w.input.Org}
	values := helm.NewServiceAccountValues(orgCfg)
	namespaceName := values.Namespace

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
	var renderOutput *helmpb.RenderChartOutput
	err = workflow.ExecuteActivity(activityCtx, helmpb.RenderChartWithTemplateActivityName, &helmpb.RenderChartWithTemplateInput{
		ChartName: chartServiceAccount,
		Component: componentServiceAccount,
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

// Execute deploys resources to Kubernetes based on the specified component type
func (w *createDeploymentWorkflow) Execute(ctx workflow.Context) (*pb.CreateDeploymentOutput, error) {
	// Validate input
	if w.input.Org == "" {
		return nil, fmt.Errorf("org is required")
	}
	if w.input.Component == "" {
		return nil, fmt.Errorf("component is required")
	}

	var chartName string
	var componentName string
	var values interface{}
	var namespaceName string
	var deploymentName string

	orgCfg := config.OrgConfig{OrgID: w.input.Org}

	switch w.input.Component {
	case componentWorker:
		chartName = chartWorker
		componentName = componentWorker
		workerValues := helm.NewWorkerValues(orgCfg)
		values = workerValues
		namespaceName = orgCfg.KubernetesNamespace()
		deploymentName = orgCfg.WorkerDeploymentName()
	case componentCertificateStore:
		chartName = chartCertificateStore
		componentName = componentCertificateStore
		// Create config objects to get consistent configuration values
		fabricCfg := fabricconfig.NewFabricConfig(orgCfg)
		certStoreConfig := fabricCfg.CertificateStore()
		certStoreValues := helm.NewCertificateStoreValues(certStoreConfig)
		values = certStoreValues
		namespaceName = certStoreConfig.KubernetesNamespace() // Use from certStoreConfig which embeds OrgConfig
		deploymentName = componentCertificateStore            // Matches helm chart: {{ include "certificate-store.name" . }}
	default:
		return nil, fmt.Errorf("unsupported component type: %s", w.input.Component)
	}

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
	var renderOutput *helmpb.RenderChartOutput
	err = workflow.ExecuteActivity(activityCtx, helmpb.RenderChartWithTemplateActivityName, &helmpb.RenderChartWithTemplateInput{
		ChartName: chartName,
		Component: componentName,
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

	return &pb.CreateDeploymentOutput{
		DeploymentName: deploymentName,
	}, nil
}
