package cell

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/cell/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides cell v1 workflows for dependency injection
var Module = fx.Options(
	// Provide Workflows implementing pb.CellWorkflows interface.
	// Cell workflows act on behalf of explicit owners; the injected operator
	// identifies the platform operator for InitFabric.
	fx.Provide(func(operator *config.OperatorConfig) pb.CellWorkflows {
		return NewWorkflows(operator)
	}),
)

// RegisterWorkflows registers cell workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.CellWorkflows) {
	pb.RegisterCellWorkflows(w, workflows)
}
