package cell

import (
	"errors"
	"fmt"
	"time"

	pb "github.com/metacensus/workers-chain/gen/workflows/cell/v1"
	clusterpb "github.com/metacensus/workers-chain/gen/workflows/cluster/fabric/v1"
	fabricpb "github.com/metacensus/workers-chain/gen/workflows/fabric/v1"
	"github.com/metacensus/workers-chain/gen/workflows/fabric/v1/fabricv1nexustemporal"
	secretspb "github.com/metacensus/workers-chain/gen/workflows/secrets/v1"
	temporalpb "github.com/metacensus/workers-chain/gen/workflows/temporal/v1"
	"github.com/metacensus/workers-chain/lib/workers/config"
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	// Nexus endpoint names. Per-org endpoints are named by OrgConfig; see
	// config.OrgConfig's NexusEndpoint* methods.
	SecretsCreateAccessNexusEndpoint   = "secrets-CreateAccess" // pragma: allowlist secret
	FabricBootstrapFabricNexusEndpoint = "fabric-BootstrapFabric"
)

// Sentinel errors for workflows
var (
	// BootstrapInfra errors
	ErrCreateSecretsNexusEndpoint = errors.New("create secrets nexus endpoint")

	// InitFabric errors
	ErrNoCertificateAuthorities          = errors.New("no certificate authorities specified")
	ErrCreateFabricTemporalNamespace     = errors.New("create fabric temporal namespace")
	ErrCreateOperatorKubernetesNamespace = errors.New("create operator kubernetes namespace")
	ErrCreateOperatorServiceAccount      = errors.New("create operator service account")
	ErrCreateOperatorVaultAccess         = errors.New("create operator vault access")
	ErrDeployFabricSharedWorker          = errors.New("deploy shared fabric worker")
	ErrCreateBootstrapFabricEndpoint     = errors.New("create BootstrapFabric nexus endpoint")
	ErrBootstrapFabric                   = errors.New("bootstrap fabric via nexus")

	// InitFabricShard errors
	ErrOrgRequired               = errors.New("org is required")
	ErrCreateTemporalNamespace   = errors.New("create temporal namespace")
	ErrCreateKubernetesNamespace = errors.New("create kubernetes namespace")
	ErrCreateServiceAccount      = errors.New("create service account")
	ErrCreateVaultAccess         = errors.New("create vault access")
	ErrDeployFabricWorker        = errors.New("deploy fabric worker")
	ErrCreateNexusEndpoint       = errors.New("create nexus endpoint")
	ErrCreateCertificateEndpoint = errors.New("create CreateCertificate nexus endpoint")
	ErrBootstrapShard            = errors.New("bootstrap shard via nexus")
)

// Workflows implements the CellWorkflows interface.
//
// Cell workflows run on the central infra worker and act ON BEHALF OF
// explicit owners: InitFabric provisions operator-owned shared infrastructure,
// InitFabricShard provisions one org's infrastructure. The operator identity
// is injected at startup; children receive typed Owner messages.
type Workflows struct {
	operator *config.OperatorConfig
}

// NewWorkflows creates a new Workflows instance for the given platform operator.
func NewWorkflows(operator *config.OperatorConfig) *Workflows {
	return &Workflows{operator: operator}
}

// BootstrapInfra implements the BootstrapInfra workflow
func (w *Workflows) BootstrapInfra(ctx workflow.Context, input *pb.BootstrapInfraWorkflowInput) (pb.BootstrapInfraWorkflow, error) {
	return newBootstrapInfraWorkflow(input)
}

type bootstrapInfraWorkflow struct {
	input *pb.BootstrapInfraWorkflowInput
}

// newBootstrapInfraWorkflow creates and validates a new bootstrapInfraWorkflow
func newBootstrapInfraWorkflow(input *pb.BootstrapInfraWorkflowInput) (*bootstrapInfraWorkflow, error) {
	w := &bootstrapInfraWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *bootstrapInfraWorkflow) validate() error {
	// No required field validation - workflow has no required inputs
	return nil
}

// Execute orchestrates the initialization of the infrastructure platform.
func (w *bootstrapInfraWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapInfraOutput, error) {
	// Register Nexus endpoint for secrets.v1.CreateAccess service
	temporalCfg := config.GetTemporal()
	if _, err := temporalpb.CreateNexusEndpoint(ctx, &temporalpb.CreateNexusEndpointInput{
		Name:      SecretsCreateAccessNexusEndpoint,
		Namespace: temporalCfg.Namespace,
		TaskQueue: temporalCfg.TaskQueue,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSecretsNexusEndpoint, err)
	}

	// Initialize operator-level Fabric infrastructure (namespace, vault access, TLS CAs)
	if _, err := pb.InitFabricChild(ctx, &pb.InitFabricInput{
		CertificateAuthorities: buildTLSCertificateAuthorities(),
	}); err != nil {
		return nil, fmt.Errorf("init fabric: %w", err)
	}

	return &pb.BootstrapInfraOutput{}, nil
}

// buildTLSCertificateAuthorities creates the operator-owned TLS CA chain specification.
// This defines the shared TLS certificate authorities (root + intermediate) that will
// be used by all Fabric organizations.
func buildTLSCertificateAuthorities() []*fabricpb.CertificateAuthoritySpec {
	tlsRootRef := &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	tlsIntRef := &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"}

	// 15 year validity for CAs (standard for certificate authorities)
	caValidity := 15 * 365 * 24 * time.Hour

	return []*fabricpb.CertificateAuthoritySpec{
		// TLS Root CA: Self-signed root certificate authority
		{
			Self:             tlsRootRef,
			SigningIssuer:    nil, // Self-signed
			TlsIssuer:        nil, // Self-signed
			ValidityDuration: durationpb.New(caValidity),
		},
		// TLS Intermediate CA: Signed by root, used to issue certificates
		{
			Self:             tlsIntRef,
			SigningIssuer:    tlsRootRef, // Signed by root
			TlsIssuer:        tlsRootRef, // TLS cert from root
			ValidityDuration: durationpb.New(caValidity),
		},
	}
}

// InitFabric implements the InitFabric workflow
func (w *Workflows) InitFabric(ctx workflow.Context, input *pb.InitFabricWorkflowInput) (pb.InitFabricWorkflow, error) {
	return newInitFabricWorkflow(input, w.operator)
}

type initFabricWorkflow struct {
	input    *pb.InitFabricWorkflowInput
	operator *config.OperatorConfig
}

// newInitFabricWorkflow creates and validates a new initFabricWorkflow
func newInitFabricWorkflow(input *pb.InitFabricWorkflowInput, operator *config.OperatorConfig) (*initFabricWorkflow, error) {
	w := &initFabricWorkflow{input: input, operator: operator}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *initFabricWorkflow) validate() error {
	if len(w.input.Req.CertificateAuthorities) == 0 {
		return ErrNoCertificateAuthorities
	}
	return nil
}

// Execute orchestrates the initialization of operator-level Fabric infrastructure.
// It provisions the shared "fabric" namespaces (Temporal and Kubernetes), deploys the
// shared fabric worker, wires up the Nexus endpoints it serves, and then bootstraps
// the shared infrastructure (certificate store + TLS CA chain) via Nexus.
func (w *initFabricWorkflow) Execute(ctx workflow.Context) (*pb.InitFabricOutput, error) {
	operatorCfg := w.operator
	operatorOwner := operatorCfg.OwnerProto()

	// Create the shared "fabric" Temporal namespace the fabric worker polls
	if _, err := temporalpb.CreateNamespaceChild(ctx, &temporalpb.CreateNamespaceInput{
		Name: operatorCfg.TemporalNamespace(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateFabricTemporalNamespace, err)
	}

	// Create Kubernetes namespace for the operator with network policies
	// (operator ownership enables NetworkPolicy for cross-namespace TLS CA access)
	if _, err := clusterpb.CreateNamespaceChild(ctx, &clusterpb.CreateNamespaceInput{
		Owner: operatorOwner,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOperatorKubernetesNamespace, err)
	}

	// Create service account for the shared fabric worker
	if _, err := clusterpb.CreateServiceAccountChild(ctx, &clusterpb.CreateServiceAccountInput{
		Owner: operatorOwner,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOperatorServiceAccount, err)
	}

	// Create Vault access policies for operator resources
	if _, err := secretspb.CreateAccessChild(ctx, &secretspb.CreateAccessInput{
		PolicyName:          operatorCfg.WorkerVaultPolicyName(),
		PolicyDocument:      fabricconfig.WorkerVaultPolicy(operatorCfg),
		AuthRoleName:        operatorCfg.WorkerVaultAuthRole(),
		ServiceAccountName:  operatorCfg.WorkerServiceAccountName(),
		KubernetesNamespace: operatorCfg.KubernetesNamespace(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOperatorVaultAccess, err)
	}

	// Deploy the shared fabric worker into the "fabric" namespace
	if _, err := clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Owner:     operatorOwner,
		Component: "worker",
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployFabricSharedWorker, err)
	}

	// Register Nexus endpoint for BootstrapFabric, served by the fabric worker
	if _, err := temporalpb.CreateNexusEndpoint(ctx, &temporalpb.CreateNexusEndpointInput{
		Name:      FabricBootstrapFabricNexusEndpoint,
		Namespace: operatorCfg.TemporalNamespace(),
		TaskQueue: "infra",
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapFabricEndpoint, err)
	}

	// Bootstrap shared infrastructure (certificate store + TLS CA chain) via
	// Nexus, executed by the fabric worker in the "fabric" Temporal namespace.
	// The ordered CA specs pass through unchanged - the fabric worker deploys
	// them as the operator.
	nexusClient := fabricv1nexustemporal.NewFabricNexusClient(FabricBootstrapFabricNexusEndpoint)
	if _, err := nexusClient.BootstrapFabric(ctx, &fabricpb.BootstrapFabricInput{
		CertificateAuthorities: w.input.Req.GetCertificateAuthorities(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapFabric, err)
	}

	return &pb.InitFabricOutput{}, nil
}

// InitFabricShard implements the InitFabricShard workflow
func (w *Workflows) InitFabricShard(ctx workflow.Context, input *pb.InitFabricShardWorkflowInput) (pb.InitFabricShardWorkflow, error) {
	return newInitFabricShardWorkflow(input, w.operator)
}

type initFabricShardWorkflow struct {
	input    *pb.InitFabricShardWorkflowInput
	operator *config.OperatorConfig
}

// newInitFabricShardWorkflow creates and validates a new initFabricShardWorkflow
func newInitFabricShardWorkflow(input *pb.InitFabricShardWorkflowInput, operator *config.OperatorConfig) (*initFabricShardWorkflow, error) {
	w := &initFabricShardWorkflow{input: input, operator: operator}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *initFabricShardWorkflow) validate() error {
	if w.input.Req.Org == "" {
		return ErrOrgRequired
	}
	return nil
}

// Execute orchestrates the complete initialization of a Fabric shard,
// acting on behalf of the org being provisioned.
func (w *initFabricShardWorkflow) Execute(ctx workflow.Context) (*pb.InitFabricShardOutput, error) {
	orgCfg := config.NewOrgConfig(w.input.Req.Org)
	orgOwner := orgCfg.OwnerProto()

	// Create Temporal namespace
	if _, err := temporalpb.CreateNamespaceChild(ctx, &temporalpb.CreateNamespaceInput{
		Name: orgCfg.TemporalNamespace(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateTemporalNamespace, err)
	}

	// Create Kubernetes namespace
	if _, err := clusterpb.CreateNamespaceChild(ctx, &clusterpb.CreateNamespaceInput{
		Owner: orgOwner,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateKubernetesNamespace, err)
	}

	// Create service account for the shard worker
	if _, err := clusterpb.CreateServiceAccountChild(ctx, &clusterpb.CreateServiceAccountInput{
		Owner: orgOwner,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateServiceAccount, err)
	}

	// Create Vault access policies for the shard worker
	if _, err := secretspb.CreateAccessChild(ctx, &secretspb.CreateAccessInput{
		PolicyName:          orgCfg.WorkerVaultPolicyName(),
		PolicyDocument:      fabricconfig.WorkerVaultPolicy(orgCfg),
		AuthRoleName:        orgCfg.WorkerVaultAuthRole(),
		ServiceAccountName:  orgCfg.WorkerServiceAccountName(),
		KubernetesNamespace: orgCfg.KubernetesNamespace(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateVaultAccess, err)
	}

	// Deploy shard worker pod
	if _, err := clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Owner:     orgOwner,
		Component: "worker",
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployFabricWorker, err)
	}

	// Register Nexus endpoint for BootstrapShard service
	if _, err := temporalpb.CreateNexusEndpoint(ctx, &temporalpb.CreateNexusEndpointInput{
		Name:      orgCfg.NexusEndpointBootstrapShard(),
		Namespace: orgCfg.TemporalNamespace(),
		TaskQueue: "infra",
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateNexusEndpoint, err)
	}

	// Register this org's endpoint into the operator's certificate authority,
	// served by the shared fabric worker. One endpoint per org is what makes
	// issuance authorizable: the server admits only this org's namespace here,
	// and refuses requests declaring another org as requester. Registered
	// before BootstrapShard, which issues the shard's TLS certificates.
	if _, err := temporalpb.CreateNexusEndpoint(ctx, &temporalpb.CreateNexusEndpointInput{
		Name:      orgCfg.NexusEndpointCreateCertificate(),
		Namespace: w.operator.TemporalNamespace(),
		TaskQueue: "infra",
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateEndpoint, err)
	}

	// Call BootstrapShard workflow via Nexus - the shard worker deploys the
	// org's components as the org (its injected identity)
	nexusClient := fabricv1nexustemporal.NewFabricNexusClient(orgCfg.NexusEndpointBootstrapShard())
	if _, err := nexusClient.BootstrapShard(ctx, &fabricpb.BootstrapShardInput{
		Deployment: w.input.Req.GetDeployment(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapShard, err)
	}

	return &pb.InitFabricShardOutput{}, nil
}
