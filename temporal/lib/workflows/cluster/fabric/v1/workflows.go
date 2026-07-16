package cluster

import (
	"fmt"
	"strings"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	fabricv1 "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
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

// Workflows implements the ClusterWorkflows interface from generated code.
//
// Cluster workflows act ON BEHALF OF an explicit typed Owner carried in each
// input: they run on the infra worker (and on fabric/shard workers), and the
// caller states who the provisioned resources belong to. The owner config is
// resolved exactly once at the top of each Execute.
//
// The injected operator config is platform-level context (not owner-guessing):
// org-owned peers and orderers still need the operator's TLS Root CA
// certificate for their trust stores, and per-org shard workers need the
// operator ID in their environment.
type Workflows struct {
	operator *config.OperatorConfig
}

// NewWorkflows creates a new Workflows instance.
// The operator config identifies the platform operator, whose shared TLS root
// CA anchors every component's trust store.
func NewWorkflows(operator *config.OperatorConfig) *Workflows {
	return &Workflows{operator: operator}
}

// CreateNamespace implements the CreateNamespace workflow
func (w *Workflows) CreateNamespace(ctx workflow.Context, input *pb.CreateNamespaceWorkflowInput) (pb.CreateNamespaceWorkflow, error) {
	return newCreateNamespaceWorkflow(input.Req)
}

// createNamespaceWorkflow implements pb.CreateNamespaceWorkflow
type createNamespaceWorkflow struct {
	input *pb.CreateNamespaceInput
	owner config.WorkerOwnerConfig
}

// newCreateNamespaceWorkflow creates and validates a new createNamespaceWorkflow
func newCreateNamespaceWorkflow(input *pb.CreateNamespaceInput) (*createNamespaceWorkflow, error) {
	owner, err := config.NewOwnerConfig(input.GetOwner())
	if err != nil {
		return nil, fmt.Errorf("invalid owner: %w", err)
	}
	return &createNamespaceWorkflow{input: input, owner: owner}, nil
}

// Execute creates a Kubernetes namespace
func (w *createNamespaceWorkflow) Execute(ctx workflow.Context) (*pb.CreateNamespaceOutput, error) {
	// Operator-owned resources live in the shared "fabric" namespace;
	// org-owned resources live in per-org "fabric-shard-{org}" namespaces.
	// Operator namespaces get network policies allowing cross-namespace
	// access from shard namespaces to the TLS CA services.
	values := helm.NewNamespaceValues(w.owner)
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
	owner config.WorkerOwnerConfig
}

// newCreateServiceAccountWorkflow creates and validates a new createServiceAccountWorkflow
func newCreateServiceAccountWorkflow(input *pb.CreateServiceAccountInput) (*createServiceAccountWorkflow, error) {
	owner, err := config.NewOwnerConfig(input.GetOwner())
	if err != nil {
		return nil, fmt.Errorf("invalid owner: %w", err)
	}
	return &createServiceAccountWorkflow{input: input, owner: owner}, nil
}

// Execute creates a Kubernetes service account
func (w *createServiceAccountWorkflow) Execute(ctx workflow.Context) (*pb.CreateServiceAccountOutput, error) {
	// Operator-owned workers live in the shared "fabric" namespace;
	// org-owned shard workers live in per-org "fabric-shard-{org}" namespaces.
	values := helm.NewServiceAccountValues(w.owner)
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
	return newCreateDeploymentWorkflow(input.Req, w.operator)
}

// createDeploymentWorkflow implements pb.CreateDeploymentWorkflow
type createDeploymentWorkflow struct {
	input    *pb.CreateDeploymentInput
	owner    config.WorkerOwnerConfig
	operator *config.OperatorConfig
}

// newCreateDeploymentWorkflow creates and validates a new createDeploymentWorkflow
func newCreateDeploymentWorkflow(input *pb.CreateDeploymentInput, operator *config.OperatorConfig) (*createDeploymentWorkflow, error) {
	owner, err := config.NewOwnerConfig(input.GetOwner())
	if err != nil {
		return nil, fmt.Errorf("invalid owner: %w", err)
	}
	if input.GetComponent() == "" {
		return nil, fmt.Errorf("component is required")
	}
	return &createDeploymentWorkflow{input: input, owner: owner, operator: operator}, nil
}

// Execute deploys resources to Kubernetes based on the specified component type
func (w *createDeploymentWorkflow) Execute(ctx workflow.Context) (*pb.CreateDeploymentOutput, error) {
	var chartName string
	var componentName string
	var values interface{}
	var namespaceName string
	var deploymentName string

	owner := w.owner

	switch {
	case w.input.Component == componentWorker:
		chartName = chartWorker
		componentName = componentWorker
		if owner.IsOperator() {
			// Shared fabric worker in the operator-owned "fabric" namespace
			values = helm.NewOperatorWorkerValues(config.NewOperatorConfig(owner.OwnerID()))
		} else {
			// Per-org shard worker in the org's "fabric-shard-{org}" namespace
			values = helm.NewWorkerValues(*w.operator, config.NewOrgConfig(owner.OwnerID()))
		}
		namespaceName = owner.KubernetesNamespace()
		deploymentName = owner.WorkerDeploymentName()
	case w.input.Component == componentCertificateStore:
		chartName = chartCertificateStore
		componentName = componentCertificateStore
		// Create config objects to get consistent configuration values
		certStoreConfig := fabricconfig.NewCertificateStoreConfig(owner)
		certStoreValues := helm.NewCertificateStoreValues(certStoreConfig)
		values = certStoreValues
		namespaceName = certStoreConfig.KubernetesNamespace()
		deploymentName = componentCertificateStore // Matches helm chart: {{ include "certificate-store.name" . }}
	case strings.HasPrefix(w.input.Component, componentCertificateAuthority):
		// Certificate authority component names follow pattern: "certificate-authority-{certType}-{caID}"
		// Examples: "certificate-authority-tls-root", "certificate-authority-org-int"
		parts := strings.Split(w.input.Component, "-")
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid certificate authority component format: %s (expected: certificate-authority-{certType}-{caID})", w.input.Component)
		}
		certType := parts[2] // "tls" or "org"
		caID := parts[3]     // "root" or "int"
		if certType != "tls" && certType != "org" {
			return nil, fmt.Errorf("invalid cert type: %s (must be 'tls' or 'org')", certType)
		}
		caRef := &fabricv1.CertificateAuthorityRef{
			CertType: certType,
			Id:       caID,
		}

		chartName = chartCertificateAuthority
		componentName = componentCertificateAuthority
		// The CA belongs to the input owner - no cert-type-based owner switching
		caConfig := fabricconfig.NewCertificateAuthorityConfig(owner, caRef)
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
		couchDbConfig := fabricconfig.NewCouchDbConfig(owner, peerID)
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
		peerConfig := fabricconfig.NewPeerConfig(owner, peerID)
		// Construct CA configs for root CA certificates (trust stores)
		// Intermediate CA certs are now included in the chain built at enrollment time
		// The org root CA belongs to the input owner; the TLS root CA is
		// operator-owned shared infrastructure (injected platform context)
		orgRootCA := fabricconfig.NewCertificateAuthorityConfig(owner, &fabricv1.CertificateAuthorityRef{CertType: "org", Id: "root"})
		tlsRootCA := fabricconfig.NewCertificateAuthorityConfig(w.operator, &fabricv1.CertificateAuthorityRef{CertType: "tls", Id: "root"})
		peerValues := helm.NewPeerValues(peerConfig, orgRootCA, tlsRootCA)
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
		ordererConfig := fabricconfig.NewOrdererConfig(owner, ordererID)
		// Construct CA configs for root CA certificates (trust stores)
		// Intermediate CA certs are now included in the chain built at enrollment time
		// The org root CA belongs to the input owner; the TLS root CA is
		// operator-owned shared infrastructure (injected platform context)
		orgRootCA := fabricconfig.NewCertificateAuthorityConfig(owner, &fabricv1.CertificateAuthorityRef{CertType: "org", Id: "root"})
		tlsRootCA := fabricconfig.NewCertificateAuthorityConfig(w.operator, &fabricv1.CertificateAuthorityRef{CertType: "tls", Id: "root"})
		ordererValues := helm.NewOrdererValues(ordererConfig, orgRootCA, tlsRootCA)
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
