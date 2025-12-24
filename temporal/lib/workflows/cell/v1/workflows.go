package cell

import (
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/cell/v1"
	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	temporalpb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/sdk/workflow"
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
// 3. Creation of a service account for the fabric worker (cluster.fabric.v1.CreateServiceAccount) - TODO
// 4. Creation of Vault access policies for the shard (vault.v1.CreateAccess) - TODO
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

	// TODO: Steps 3-7 unimplemented
	// 3. Create service account for fabric worker
	// 4. Create Vault access policies for the shard
	// 5. Deploy fabric worker pod
	// 6. Register Nexus endpoint for BootstrapShard service
	// 7. Call BootstrapShard workflow via Nexus

	logger.Info("InitFabricShard workflow completed", "org", w.input.Req.Org)
	return &pb.InitFabricShardOutput{}, nil
}
