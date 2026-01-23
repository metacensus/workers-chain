package cluster

import (
	"fmt"
	"strings"
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
	chartNamespace            = "fabric/namespace"
	chartServiceAccount       = "fabric/service-account"
	chartWorker               = "fabric/worker"
	chartCertificateStore     = "fabric/certificate-store"
	chartCertificateAuthority = "fabric/certificate-authority"
	chartCouchDb              = "fabric/couchdb"
	chartPeer                 = "fabric/peer"
	chartOrderer              = "fabric/orderer"

	componentNamespace            = "namespace"
	componentServiceAccount       = "service-account"
	componentWorker               = "worker"
	componentCertificateStore     = "certificate-store"
	componentCertificateAuthority = "certificate-authority"
	componentCouchDb              = "couchdb"
	componentPeer                 = "peer"
	componentOrderer              = "orderer"
)

// Workflows implements the ClusterWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// CreateNamespace implements the CreateNamespace workflow
func (w *Workflows) CreateNamespace(ctx workflow.Context, input *pb.CreateNamespaceWorkflowInput) (pb.CreateNamespaceWorkflow, error) {
	return newCreateNamespaceWorkflow(input.Req)
}

// createNamespaceWorkflow implements pb.CreateNamespaceWorkflow
type createNamespaceWorkflow struct {
	input *pb.CreateNamespaceInput
}

// newCreateNamespaceWorkflow creates and validates a new createNamespaceWorkflow
func newCreateNamespaceWorkflow(input *pb.CreateNamespaceInput) (*createNamespaceWorkflow, error) {
	w := &createNamespaceWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createNamespaceWorkflow) validate() error {
	if w.input.Org == "" {
		return fmt.Errorf("org is required")
	}
	return nil
}

// Execute creates a Kubernetes namespace
func (w *createNamespaceWorkflow) Execute(ctx workflow.Context) (*pb.CreateNamespaceOutput, error) {
	orgCfg := config.OrgConfig{OrgID: w.input.Org}
	values := helm.NewNamespaceValues(orgCfg, w.input.IsOperator)
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
	return newCreateServiceAccountWorkflow(input.Req)
}

// createServiceAccountWorkflow implements pb.CreateServiceAccountWorkflow
type createServiceAccountWorkflow struct {
	input *pb.CreateServiceAccountInput
}

// newCreateServiceAccountWorkflow creates and validates a new createServiceAccountWorkflow
func newCreateServiceAccountWorkflow(input *pb.CreateServiceAccountInput) (*createServiceAccountWorkflow, error) {
	w := &createServiceAccountWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createServiceAccountWorkflow) validate() error {
	if w.input.Org == "" {
		return fmt.Errorf("org is required")
	}
	return nil
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
	return newCreateDeploymentWorkflow(input.Req)
}

// createDeploymentWorkflow implements pb.CreateDeploymentWorkflow
type createDeploymentWorkflow struct {
	input *pb.CreateDeploymentInput
}

// newCreateDeploymentWorkflow creates and validates a new createDeploymentWorkflow
func newCreateDeploymentWorkflow(input *pb.CreateDeploymentInput) (*createDeploymentWorkflow, error) {
	w := &createDeploymentWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createDeploymentWorkflow) validate() error {
	if w.input.Org == "" {
		return fmt.Errorf("org is required")
	}
	if w.input.Component == "" {
		return fmt.Errorf("component is required")
	}
	return nil
}

// Execute deploys resources to Kubernetes based on the specified component type
func (w *createDeploymentWorkflow) Execute(ctx workflow.Context) (*pb.CreateDeploymentOutput, error) {
	var chartName string
	var componentName string
	var values interface{}
	var namespaceName string
	var deploymentName string

	orgCfg := config.OrgConfig{OrgID: w.input.Org}

	switch {
	case w.input.Component == componentWorker:
		chartName = chartWorker
		componentName = componentWorker
		workerValues := helm.NewWorkerValues(orgCfg)
		values = workerValues
		namespaceName = orgCfg.KubernetesNamespace()
		deploymentName = orgCfg.WorkerDeploymentName()
	case w.input.Component == componentCertificateStore:
		chartName = chartCertificateStore
		componentName = componentCertificateStore
		// Create config objects to get consistent configuration values
		certStoreConfig := fabricconfig.NewCertificateStoreConfig(orgCfg)
		certStoreValues := helm.NewCertificateStoreValues(certStoreConfig)
		values = certStoreValues
		namespaceName = certStoreConfig.KubernetesNamespace()
		deploymentName = componentCertificateStore // Matches helm chart: {{ include "certificate-store.name" . }}
	case strings.HasPrefix(w.input.Component, componentCertificateAuthority):
		// Certificate authority component names follow pattern: "certificate-authority-{certType}-{caID}"
		// Examples: "certificate-authority-tls-root", "certificate-authority-org-v1"
		parts := strings.Split(w.input.Component, "-")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid certificate authority component format: %s (expected: certificate-authority-{certType}-{caID})", w.input.Component)
		}
		certType := parts[2] // "tls" or "org"
		caID := parts[3]     // "root", "v1", "v2", etc.

		// Validate cert type
		if certType != "tls" && certType != "org" {
			return nil, fmt.Errorf("invalid cert type: %s (must be 'tls' or 'org')", certType)
		}

		chartName = chartCertificateAuthority
		componentName = componentCertificateAuthority
		// Create config objects to get consistent configuration values
		caConfig := fabricconfig.NewCertificateAuthorityConfig(orgCfg, certType, caID)
		caValues := helm.NewCertificateAuthorityValues(caConfig)
		values = caValues
		namespaceName = caConfig.KubernetesNamespace()
		deploymentName = caConfig.ComponentName() // Matches helm chart: {{ include "certificate-authority.name" . }}
	case strings.HasPrefix(w.input.Component, componentCouchDb):
		// CouchDB component names follow pattern: "couchdb-{peerId}"
		// Examples: "couchdb-peer0", "couchdb-peer1"
		parts := strings.Split(w.input.Component, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid couchdb component format: %s (expected: couchdb-{peerId})", w.input.Component)
		}
		peerID := parts[1] // "peer0", "peer1", etc.

		chartName = chartCouchDb
		componentName = componentCouchDb
		// Create config objects to get consistent configuration values
		couchDbConfig := fabricconfig.NewCouchDbConfig(orgCfg, peerID)
		couchDbValues := helm.NewCouchDbValues(couchDbConfig)
		values = couchDbValues
		namespaceName = couchDbConfig.KubernetesNamespace()
		deploymentName = couchDbConfig.ComponentName() // Matches helm chart: {{ include "couchdb.name" . }}
	case strings.HasPrefix(w.input.Component, componentPeer):
		// Peer component names follow pattern: "peer{id}"
		// Examples: "peer0", "peer1"
		peerID := w.input.Component // The whole component name is the peer ID

		chartName = chartPeer
		componentName = componentPeer
		// Create config objects to get consistent configuration values
		peerConfig := fabricconfig.NewPeerConfig(orgCfg, peerID)
		// Construct CA configs - currently all owned by the org, but TLS CAs
		// will be owned by the operator after shared TLS CA migration
		orgRootCA := fabricconfig.NewCertificateAuthorityConfig(orgCfg, "org", "root")
		orgIntCA := fabricconfig.NewCertificateAuthorityConfig(orgCfg, "org", "v1")
		tlsRootCA := fabricconfig.NewCertificateAuthorityConfig(orgCfg, "tls", "root")
		tlsIntCA := fabricconfig.NewCertificateAuthorityConfig(orgCfg, "tls", "v1")
		peerValues := helm.NewPeerValues(peerConfig, orgRootCA, orgIntCA, tlsRootCA, tlsIntCA)
		values = peerValues
		namespaceName = peerConfig.KubernetesNamespace()
		deploymentName = peerConfig.ComponentName() // Matches helm chart: {{ include "peer.name" . }}
	case strings.HasPrefix(w.input.Component, componentOrderer):
		// Orderer component names follow pattern: "orderer{id}"
		// Examples: "orderer0", "orderer1"
		ordererID := w.input.Component // The whole component name is the orderer ID

		chartName = chartOrderer
		componentName = componentOrderer
		// Create config objects to get consistent configuration values
		ordererConfig := fabricconfig.NewOrdererConfig(orgCfg, ordererID)
		// Construct CA configs - currently all owned by the org, but TLS CAs
		// will be owned by the operator after shared TLS CA migration
		orgRootCA := fabricconfig.NewCertificateAuthorityConfig(orgCfg, "org", "root")
		orgIntCA := fabricconfig.NewCertificateAuthorityConfig(orgCfg, "org", "v1")
		tlsRootCA := fabricconfig.NewCertificateAuthorityConfig(orgCfg, "tls", "root")
		tlsIntCA := fabricconfig.NewCertificateAuthorityConfig(orgCfg, "tls", "v1")
		ordererValues := helm.NewOrdererValues(ordererConfig, orgRootCA, orgIntCA, tlsRootCA, tlsIntCA)
		values = ordererValues
		namespaceName = ordererConfig.KubernetesNamespace()
		deploymentName = ordererConfig.ComponentName() // Matches helm chart: {{ include "orderer.name" . }}
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
