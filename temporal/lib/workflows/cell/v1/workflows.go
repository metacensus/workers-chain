package cell

import (
	"fmt"
	"path"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cell/v1"
	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
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
		return nil, fmt.Errorf("org is required")
	}

	temporalNamespace := fmt.Sprintf("fabric-%s", w.input.Req.Org)

	// Create Temporal namespace
	if _, err := temporalpb.CreateNamespaceChild(ctx, &temporalpb.CreateNamespaceInput{
		Name: temporalNamespace,
	}); err != nil {
		return nil, fmt.Errorf("create temporal namespace: %w", err)
	}

	// Create Kubernetes namespace
	if _, err := clusterpb.CreateNamespaceChild(ctx, &clusterpb.CreateNamespaceInput{
		Org: w.input.Req.Org,
	}); err != nil {
		return nil, fmt.Errorf("create kubernetes namespace: %w", err)
	}

	// Create service account for fabric worker
	if _, err := clusterpb.CreateServiceAccountChild(ctx, &clusterpb.CreateServiceAccountInput{
		Org: w.input.Req.Org,
	}); err != nil {
		return nil, fmt.Errorf("create service account: %w", err)
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
		return nil, fmt.Errorf("create vault access: %w", err)
	}

	// Deploy fabric worker pod
	if _, err := clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.input.Req.Org,
		Component: "worker",
	}); err != nil {
		return nil, fmt.Errorf("deploy fabric worker: %w", err)
	}

	// Register Nexus endpoint for BootstrapShard service
	nexusEndpointName := fmt.Sprintf("BootstrapShard-%s", w.input.Req.Org)
	if _, err := temporalpb.CreateNexusEndpoint(ctx, &temporalpb.CreateNexusEndpointInput{
		Name:      nexusEndpointName,
		Namespace: temporalNamespace,
		TaskQueue: "infra",
	}); err != nil {
		return nil, fmt.Errorf("create nexus endpoint: %w", err)
	}

	// Call BootstrapShard workflow via Nexus
	nexusClient := workflow.NewNexusClient(nexusEndpointName, "Fabric")
	bootstrapFut := nexusClient.ExecuteOperation(ctx, "BootstrapShard", &fabricpb.BootstrapShardInput{
		Org: w.input.Req.Org,
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 30 * time.Minute,
	})

	var bootstrapOutput fabricpb.BootstrapShardOutput
	if err := bootstrapFut.Get(ctx, &bootstrapOutput); err != nil {
		return nil, fmt.Errorf("bootstrap shard via nexus: %w", err)
	}

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
