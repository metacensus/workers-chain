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
	temporalpb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/workflow"
)

const (
	// Vault KV v2 path prefixes
	vaultSecretDataPrefix     = "secret/data"
	vaultSecretMetadataPrefix = "secret/metadata"

	// Fabric secret path structure
	fabricSecretBase = "fabric"
)

// Sentinel errors for InitFabricShard workflow
var (
	ErrOrgRequired               = errors.New("org is required")
	ErrCreateTemporalNamespace   = errors.New("create temporal namespace")
	ErrCreateKubernetesNamespace = errors.New("create kubernetes namespace")
	ErrCreateServiceAccount      = errors.New("create service account")
	ErrCreateVaultAccess         = errors.New("create vault access")
	ErrDeployFabricWorker        = errors.New("deploy fabric worker")
	ErrCreateNexusEndpoint       = errors.New("create nexus endpoint")
	ErrBootstrapShard            = errors.New("bootstrap shard via nexus")
)

// Workflows implements the CellWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// InitFabricShard implements the InitFabricShard workflow
func (w *Workflows) InitFabricShard(ctx workflow.Context, input *pb.InitFabricShardWorkflowInput) (pb.InitFabricShardWorkflow, error) {
	return &initFabricShardWorkflow{input: input}, nil
}

// initFabricShardWorkflow implements pb.InitFabricShardWorkflow
type initFabricShardWorkflow struct {
	input *pb.InitFabricShardWorkflowInput
}

// Execute orchestrates the complete initialization of a Fabric shard.
func (w *initFabricShardWorkflow) Execute(ctx workflow.Context) (*pb.InitFabricShardOutput, error) {
	if w.input.Req.Org == "" {
		return nil, ErrOrgRequired
	}

	temporalNamespace := fmt.Sprintf("fabric-%s", w.input.Req.Org)

	// Create Temporal namespace
	if _, err := temporalpb.CreateNamespaceChild(ctx, &temporalpb.CreateNamespaceInput{
		Name: temporalNamespace,
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
	fabricNamespace := fmt.Sprintf("fabric-%s", w.input.Req.Org)
	serviceAccountName := fmt.Sprintf("fabric-worker-%s", w.input.Req.Org)
	policyName := fmt.Sprintf("fabric-%s-worker", w.input.Req.Org)
	authRoleName := fmt.Sprintf("fabric-%s-worker", w.input.Req.Org)

	if _, err := vaultpb.CreateAccessChild(ctx, &vaultpb.CreateAccessInput{
		PolicyName:          policyName,
		PolicyDocument:      buildFabricWorkerPolicy(w.input.Req.Org),
		AuthRoleName:        authRoleName,
		ServiceAccountName:  serviceAccountName,
		KubernetesNamespace: fabricNamespace,
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
	nexusEndpointName := fmt.Sprintf("fabric-%s-BootstrapShard", w.input.Req.Org)
	if _, err := temporalpb.CreateNexusEndpoint(ctx, &temporalpb.CreateNexusEndpointInput{
		Name:      nexusEndpointName,
		Namespace: temporalNamespace,
		TaskQueue: "infra",
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateNexusEndpoint, err)
	}

	// Call BootstrapShard workflow via Nexus using generated client
	nexusClient := fabricv1nexustemporal.NewFabricNexusClient(nexusEndpointName)
	bootstrapOutput, err := nexusClient.BootstrapShard(ctx, &fabricpb.BootstrapShardInput{}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapShard, err)
	}
	_ = bootstrapOutput // Unused for now

	return &pb.InitFabricShardOutput{}, nil
}

// buildFabricWorkerPolicy creates a Vault policy document for a fabric worker
// The policy grants read/write access to all secrets under secret/fabric/{org}/*
func buildFabricWorkerPolicy(org string) *vaultpb.PolicyDocument {
	// Build path patterns for the org
	dataPathPattern := path.Join(vaultSecretDataPrefix, fabricSecretBase, org, "*")
	metadataPathPattern := path.Join(vaultSecretMetadataPrefix, fabricSecretBase, org, "*")

	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// KV v2 data path - read/write access to all org secrets
			{
				Path: dataPathPattern,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_CREATE,
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_UPDATE,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// KV v2 metadata path - list and read metadata
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
