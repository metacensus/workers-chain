package cell

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/cell/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/cell/v1/cellv1temporal"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides cell v1 workflows for dependency injection
var Module = fx.Options(
	// Provide Workflows implementing pb.CellWorkflows interface
	fx.Provide(func() pb.CellWorkflows {
		return NewWorkflows()
	}),
)

// RegisterWorkflows registers cell workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.CellWorkflows) {
	pb.RegisterCellWorkflows(w, workflows)
}

// RegisterNexusService registers the cell Nexus service handler with the worker
// This allows the infra worker to receive Nexus calls from other namespaces (e.g., fabric shards)
func RegisterNexusService(w worker.Worker) error {
	return cellv1temporal.RegisterCellNexusService(w)
}
