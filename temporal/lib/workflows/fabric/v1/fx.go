package fabric

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1temporal"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides fabric v1 deployment orchestration workflows for dependency injection.
//
// CA client access (BootstrapClient, cached Client via MultiClientProvider) lives in
// the certauth module (lib/workflows/fabric/certauth/v1); fabric workflows call
// certauth workflows as children or via Nexus rather than talking to CAs directly.
var Module = fx.Options(
	// Provide Workflows implementing pb.FabricWorkflows interface.
	// Workflows act AS the worker's owner, injected at startup
	// (operator on the fabric worker, org on shard workers).
	fx.Provide(func(owner config.OwnerConfig) pb.FabricWorkflows {
		return NewWorkflows(owner)
	}),
)

// RegisterWorkflows registers fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.FabricWorkflows) {
	pb.RegisterFabricWorkflows(w, workflows)
}

// RegisterNexusService registers the Fabric Nexus service handler with the worker.
// This allows the worker to receive Nexus calls (e.g. BootstrapShard) from the
// central infra worker.
func RegisterNexusService(w worker.Worker) error {
	return fabricv1temporal.RegisterFabricNexusService(w)
}
