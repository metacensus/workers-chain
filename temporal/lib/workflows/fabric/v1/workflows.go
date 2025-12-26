package fabric

import (
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"go.temporal.io/sdk/workflow"
)

// Workflows implements the FabricWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// BootstrapShard implements the BootstrapShard workflow
func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return &bootstrapShardWorkflow{input: input}, nil
}

// bootstrapShardWorkflow implements pb.BootstrapShardWorkflow
type bootstrapShardWorkflow struct {
	input *pb.BootstrapShardWorkflowInput
}

// Execute deploys all Fabric components (CAs, peers, orderers) for the organization.
func (w *bootstrapShardWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapShardOutput, error) {
	if w.input.Req.Org == "" {
		return nil, fmt.Errorf("org is required")
	}

	// TODO: Deploy Fabric CA, bootstrap admin, deploy peers/orderers

	return &pb.BootstrapShardOutput{}, nil
}

