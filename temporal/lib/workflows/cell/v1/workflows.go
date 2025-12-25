package cell

import (
	"fmt"
	"path"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cell/v1"
	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
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

// Execute orchestrates the complete initialization of a Fabric shard
// This root workflow executes in the central "infra" namespace and coordinates:
// 1. ✅ Creation of a Temporal namespace for the shard (temporal.v1.CreateNamespace)
// 2. ✅ Creation of a Kubernetes namespace for the shard (cluster.fabric.v1.CreateNamespace)
// 3. ✅ Creation of a service account for the fabric worker (cluster.fabric.v1.CreateServiceAccount)
// 4. ✅ Creation of Vault access policies for the shard (vault.v1.CreateAccess)
// 5. Deployment of the fabric worker pod (cluster.fabric.v1.CreateDeployment) - TODO
// 6. Registration of the Nexus endpoint for fabric.v1.BootstrapShard service - TODO
// 7. Cross-namespace Nexus call to BootstrapShard workflow in fabric namespace - TODO
func (w *initFabricShardWorkflow) Execute(ctx workflow.Context) (*pb.InitFabricShardOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("InitFabricShard workflow started", "org", w.input.Req.Org)

	// Validate input
	if w.input.Req.Org == "" {
		return nil, fmt.Errorf("org is required")
	}

	// Derive namespace names from org
	temporalNamespace := fmt.Sprintf("fabric-%s", w.input.Req.Org)

	// Step 1: Create Temporal namespace
	logger.Info("Step 1: Creating Temporal namespace for shard", "namespace", temporalNamespace)
	temporalOutput, err := temporalpb.CreateNamespaceChild(ctx, &temporalpb.CreateNamespaceInput{
		Name: temporalNamespace,
	})
	if err != nil {
		logger.Error("Failed to create Temporal namespace", "error", err)
		return nil, fmt.Errorf("create temporal namespace: %w", err)
	}
	logger.Info("Temporal namespace created", "namespace", temporalOutput.Name)

	// Step 2: Create Kubernetes namespace
	logger.Info("Step 2: Creating Kubernetes namespace for shard", "org", w.input.Req.Org)
	kubeOutput, err := clusterpb.CreateNamespaceChild(ctx, &clusterpb.CreateNamespaceInput{
		Org: w.input.Req.Org,
	})
	if err != nil {
		logger.Error("Failed to create Kubernetes namespace", "error", err)
		return nil, fmt.Errorf("create kubernetes namespace: %w", err)
	}
	logger.Info("Kubernetes namespace created", "namespace", kubeOutput.NamespaceName)

	// Step 3: Create service account for fabric worker
	logger.Info("Step 3: Creating ServiceAccount for fabric worker", "org", w.input.Req.Org)
	saOutput, err := clusterpb.CreateServiceAccountChild(ctx, &clusterpb.CreateServiceAccountInput{
		Org: w.input.Req.Org,
	})
	if err != nil {
		logger.Error("Failed to create ServiceAccount", "error", err)
		return nil, fmt.Errorf("create service account: %w", err)
	}
	logger.Info("ServiceAccount created successfully", "org", w.input.Req.Org)

	// Step 4: Create Vault access policies for the fabric worker
	logger.Info("Step 4: Creating Vault access for fabric worker", "org", w.input.Req.Org)
	fabricNamespace := fmt.Sprintf("fabric-%s", w.input.Req.Org)
	serviceAccountName := fmt.Sprintf("fabric-worker-%s", w.input.Req.Org)
	policyName := fmt.Sprintf("fabric-%s-worker", w.input.Req.Org)

	vaultOutput, err := vaultpb.CreateAccessChild(ctx, &vaultpb.CreateAccessInput{
		PolicyName:          policyName,
		PolicyDocument:      buildFabricWorkerPolicy(w.input.Req.Org),
		Organization:        w.input.Req.Org,
		ServiceAccountName:  serviceAccountName,
		KubernetesNamespace: fabricNamespace,
	})
	if err != nil {
		logger.Error("Failed to create Vault access", "error", err)
		return nil, fmt.Errorf("create vault access: %w", err)
	}
	logger.Info("Vault access created", "policy", vaultOutput.PolicyName, "auth_role", vaultOutput.AuthRoleName)

	// TODO: Steps 5-7 unimplemented
	// 5. Deploy fabric worker pod
	// 6. Register Nexus endpoint for BootstrapShard service
	// 7. Call BootstrapShard workflow via Nexus
	_ = saOutput

	logger.Info("InitFabricShard workflow completed", "org", w.input.Req.Org)
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
