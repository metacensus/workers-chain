package cell

import (
	"errors"
	"fmt"
	"path"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cell/v1"
	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1nexustemporal"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	temporalpb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	// Vault KV v2 path prefixes
	vaultSecretDataPrefix     = "secret/data"
	vaultSecretMetadataPrefix = "secret/metadata"

	// Fabric secret path structure
	fabricSecretBase = "fabric"

	// Nexus endpoint name for secrets.CreateAccess
	SecretsCreateAccessNexusEndpoint     = "secrets-CreateAccess"
	FabricCreateCertificateNexusEndpoint = "fabric-CreateCertificate"
)

// Sentinel errors for workflows
var (
	// BootstrapInfra errors
	ErrCreateSecretsNexusEndpoint = errors.New("create secrets nexus endpoint")

	// InitFabric errors
	ErrNoCertificateAuthorities           = errors.New("no certificate authorities specified")
	ErrBootstrapTlsCertificateAuthorities = errors.New("bootstrap TLS certificate authorities")
	ErrCreateOperatorKubernetesNamespace  = errors.New("create operator kubernetes namespace")
	ErrCreateOperatorServiceAccount       = errors.New("create operator service account")
	ErrCreateOperatorVaultAccess          = errors.New("create operator vault access")
	ErrCreateTLSCANexusEndpoint           = errors.New("create TLS CA nexus endpoint")

	// InitFabricShard errors
	ErrOrgRequired               = errors.New("org is required")
	ErrCreateTemporalNamespace   = errors.New("create temporal namespace")
	ErrCreateKubernetesNamespace = errors.New("create kubernetes namespace")
	ErrCreateServiceAccount      = errors.New("create service account")
	ErrCreateVaultAccess         = errors.New("create vault access")
	ErrDeployFabricWorker        = errors.New("deploy fabric worker")
	ErrCreateNexusEndpoint       = errors.New("create nexus endpoint")
	ErrBootstrapShard            = errors.New("bootstrap shard via nexus")
)

// Workflows implements the CellWorkflows interface
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
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
// This defines the shared TLS certificate authorities (root + v1 intermediate) that will
// be used by all Fabric organizations.
func buildTLSCertificateAuthorities() []*fabricpb.CertificateAuthoritySpec {
	tlsRootRef := &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	tlsIntRef := &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "v1"}

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
		// TLS Intermediate CA (v1): Signed by root, used to issue certificates
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
	return newInitFabricWorkflow(input)
}

type initFabricWorkflow struct {
	input *pb.InitFabricWorkflowInput
}

// newInitFabricWorkflow creates and validates a new initFabricWorkflow
func newInitFabricWorkflow(input *pb.InitFabricWorkflowInput) (*initFabricWorkflow, error) {
	w := &initFabricWorkflow{input: input}
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
func (w *initFabricWorkflow) Execute(ctx workflow.Context) (*pb.InitFabricOutput, error) {
	operatorCfg := config.GetOperator()

	// Create Kubernetes namespace for operator with network policies
	// IsOperator=true enables NetworkPolicy for cross-namespace TLS CA access
	if _, err := clusterpb.CreateNamespaceChild(ctx, &clusterpb.CreateNamespaceInput{
		Org:        operatorCfg.OwnerID(),
		IsOperator: true,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOperatorKubernetesNamespace, err)
	}

	// Create service account for operator fabric components
	if _, err := clusterpb.CreateServiceAccountChild(ctx, &clusterpb.CreateServiceAccountInput{
		Org: operatorCfg.OwnerID(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOperatorServiceAccount, err)
	}

	// Create Vault access policies for operator resources
	if _, err := secretspb.CreateAccessChild(ctx, &secretspb.CreateAccessInput{
		PolicyName:          operatorCfg.WorkerVaultPolicyName(),
		PolicyDocument:      buildOperatorVaultPolicy(*operatorCfg),
		AuthRoleName:        operatorCfg.WorkerVaultAuthRole(),
		ServiceAccountName:  operatorCfg.WorkerServiceAccountName(),
		KubernetesNamespace: operatorCfg.KubernetesNamespace(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateOperatorVaultAccess, err)
	}

	// Register Nexus endpoint for operator TLS CA certificate issuance
	temporalCfg := config.GetTemporal()
	if _, err := temporalpb.CreateNexusEndpoint(ctx, &temporalpb.CreateNexusEndpointInput{
		Name:      FabricCreateCertificateNexusEndpoint,
		Namespace: temporalCfg.Namespace,
		TaskQueue: "infra",
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateTLSCANexusEndpoint, err)
	}

	// Deploy certificate store and TLS CA chain
	// Pass operator ID as the "org" since this is operator-owned infrastructure
	// Extract validity durations from the CA specs
	var tlsRootValidity, tlsIntValidity *durationpb.Duration
	for _, caSpec := range w.input.Req.GetCertificateAuthorities() {
		if caSpec.GetSelf().GetCertType() == "tls" {
			if caSpec.GetSelf().GetId() == "root" {
				tlsRootValidity = caSpec.GetValidityDuration()
			} else if caSpec.GetSelf().GetId() == "int" {
				tlsIntValidity = caSpec.GetValidityDuration()
			}
		}
	}

	if _, err := fabricpb.BootstrapTlsCertificateAuthoritiesChild(ctx, &fabricpb.BootstrapTlsCertificateAuthoritiesInput{
		Org:                     operatorCfg.OwnerID(),
		TlsRootValidity:         tlsRootValidity,
		TlsIntermediateValidity: tlsIntValidity,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapTlsCertificateAuthorities, err)
	}

	return &pb.InitFabricOutput{}, nil
}

// buildOperatorVaultPolicy creates a Vault policy document for operator resources
func buildOperatorVaultPolicy(operatorCfg config.OperatorConfig) *vaultpb.PolicyDocument {
	dataPathPattern := path.Join(vaultSecretDataPrefix, fabricSecretBase, operatorCfg.OwnerID(), "*")
	metadataPathPattern := path.Join(vaultSecretMetadataPrefix, fabricSecretBase, operatorCfg.OwnerID(), "*")

	// Allow writing certificates to all org namespaces (for TLS cert issuance via Nexus)
	allOrgCertsPattern := path.Join(vaultSecretDataPrefix, fabricSecretBase, "*", "certs/*")
	allOrgCertsMetadataPattern := path.Join(vaultSecretMetadataPrefix, fabricSecretBase, "*", "certs/*")

	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// Operator namespace: full access
			{
				Path: dataPathPattern,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_CREATE,
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_UPDATE,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			{
				Path: metadataPathPattern,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// All org namespaces: write access to certs/* only (for TLS issuance)
			{
				Path: allOrgCertsPattern,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_CREATE,
					vaultpb.Capability_CAPABILITY_UPDATE,
				},
			},
			{
				Path: allOrgCertsMetadataPattern,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
		},
	}
}

// InitFabricShard implements the InitFabricShard workflow
func (w *Workflows) InitFabricShard(ctx workflow.Context, input *pb.InitFabricShardWorkflowInput) (pb.InitFabricShardWorkflow, error) {
	return newInitFabricShardWorkflow(input)
}

type initFabricShardWorkflow struct {
	input *pb.InitFabricShardWorkflowInput
}

// newInitFabricShardWorkflow creates and validates a new initFabricShardWorkflow
func newInitFabricShardWorkflow(input *pb.InitFabricShardWorkflowInput) (*initFabricShardWorkflow, error) {
	w := &initFabricShardWorkflow{input: input}
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

// Execute orchestrates the complete initialization of a Fabric shard.
func (w *initFabricShardWorkflow) Execute(ctx workflow.Context) (*pb.InitFabricShardOutput, error) {
	orgCfg := config.OrgConfig{OrgID: w.input.Req.Org}

	// Create Temporal namespace
	if _, err := temporalpb.CreateNamespaceChild(ctx, &temporalpb.CreateNamespaceInput{
		Name: orgCfg.TemporalNamespace(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateTemporalNamespace, err)
	}

	// Create Kubernetes namespace
	if _, err := clusterpb.CreateNamespaceChild(ctx, &clusterpb.CreateNamespaceInput{
		Org: w.input.Req.Org,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateKubernetesNamespace, err)
	}

	// Create service account for fabric worker
	if _, err := clusterpb.CreateServiceAccountChild(ctx, &clusterpb.CreateServiceAccountInput{
		Org: w.input.Req.Org,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateServiceAccount, err)
	}

	// Create Vault access policies for the fabric worker
	if _, err := secretspb.CreateAccessChild(ctx, &secretspb.CreateAccessInput{
		PolicyName:          orgCfg.WorkerVaultPolicyName(),
		PolicyDocument:      buildFabricWorkerPolicy(orgCfg),
		AuthRoleName:        orgCfg.WorkerVaultAuthRole(),
		ServiceAccountName:  orgCfg.WorkerServiceAccountName(),
		KubernetesNamespace: orgCfg.KubernetesNamespace(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateVaultAccess, err)
	}

	// Deploy fabric worker pod
	if _, err := clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.input.Req.Org,
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

	// Call BootstrapShard workflow via Nexus
	nexusClient := fabricv1nexustemporal.NewFabricNexusClient(orgCfg.NexusEndpointBootstrapShard())
	bootstrapOutput, err := nexusClient.BootstrapShard(ctx, &fabricpb.BootstrapShardInput{
		Org:        w.input.Req.Org,
		Deployment: w.input.Req.GetDeployment(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapShard, err)
	}
	_ = bootstrapOutput

	return &pb.InitFabricShardOutput{}, nil
}

// buildFabricWorkerPolicy creates a Vault policy document for a fabric worker
func buildFabricWorkerPolicy(orgCfg config.OrgConfig) *vaultpb.PolicyDocument {
	dataPathPattern := path.Join(vaultSecretDataPrefix, fabricSecretBase, orgCfg.OrgID, "*")
	metadataPathPattern := path.Join(vaultSecretMetadataPrefix, fabricSecretBase, orgCfg.OrgID, "*")

	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			{
				Path: dataPathPattern,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_CREATE,
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_UPDATE,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			{
				Path: metadataPathPattern,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
		},
	}
}
