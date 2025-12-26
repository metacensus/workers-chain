package fabric

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides fabric v1 workflows for dependency injection
var Module = fx.Options(
	// Provide Workflows implementing pb.FabricWorkflows interface
	fx.Provide(func() pb.FabricWorkflows {
		return NewWorkflows()
	}),
)

// RegisterWorkflows registers fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.FabricWorkflows) {
	pb.RegisterFabricWorkflows(w, workflows)
}
