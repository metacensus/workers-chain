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
)

const (
	// Vault KV v2 path prefixes
	vaultSecretDataPrefix     = "secret/data"
	vaultSecretMetadataPrefix = "secret/metadata"

	// Fabric secret path structure
	fabricSecretBase = "fabric"

	// Nexus endpoint name for secrets.CreateAccess
	secretsCreateAccessNexusEndpoint = "secrets-CreateAccess"
)

// Sentinel errors for workflows
var (
	// BootstrapInfra errors
	ErrCreateSecretsNexusEndpoint = errors.New("create secrets nexus endpoint")

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
	return &bootstrapInfraWorkflow{input: input}, nil
}

type bootstrapInfraWorkflow struct {
	input *pb.BootstrapInfraWorkflowInput
}

// Execute orchestrates the initialization of the infrastructure platform.
func (w *bootstrapInfraWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapInfraOutput, error) {
	// Register Nexus endpoint for secrets.v1.CreateAccess service
	temporalCfg := config.GetTemporal()
	if _, err := temporalpb.CreateNexusEndpoint(ctx, &temporalpb.CreateNexusEndpointInput{
		Name:      secretsCreateAccessNexusEndpoint,
		Namespace: temporalCfg.Namespace,
		TaskQueue: temporalCfg.TaskQueue,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateSecretsNexusEndpoint, err)
	}

	return &pb.BootstrapInfraOutput{}, nil
}

// InitFabricShard implements the InitFabricShard workflow
func (w *Workflows) InitFabricShard(ctx workflow.Context, input *pb.InitFabricShardWorkflowInput) (pb.InitFabricShardWorkflow, error) {
	return &initFabricShardWorkflow{input: input}, nil
}

type initFabricShardWorkflow struct {
	input *pb.InitFabricShardWorkflowInput
}

// Execute orchestrates the complete initialization of a Fabric shard.
func (w *initFabricShardWorkflow) Execute(ctx workflow.Context) (*pb.InitFabricShardOutput, error) {
	if w.input.Req.Org == "" {
		return nil, ErrOrgRequired
	}

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
	bootstrapOutput, err := nexusClient.BootstrapShard(ctx, &fabricpb.BootstrapShardInput{}, workflow.NexusOperationOptions{
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
