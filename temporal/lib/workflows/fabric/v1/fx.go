package fabric

import (
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1nexus"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1temporal"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/nexus-rpc/sdk-go/nexus"
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

// RegisterBootstrapFabricNexusService registers only BootstrapFabric. Used by
// the shared fabric worker; other fabric.v1 ops stay local (child/activity).
//
// Takes worker.Worker so uber/fx can inject the provided Worker.
func RegisterBootstrapFabricNexusService(w worker.Worker) error {
	return registerFabricNexusOps(w, func(h *fabricv1temporal.FabricNexusHandler) nexus.RegisterableOperation {
		return h.BootstrapFabric(fabricv1nexus.FabricBootstrapFabricOperationName)
	})
}

// RegisterBootstrapShardNexusService registers only BootstrapShard. Used by
// per-org shard workers.
func RegisterBootstrapShardNexusService(w worker.Worker) error {
	return registerFabricNexusOps(w, func(h *fabricv1temporal.FabricNexusHandler) nexus.RegisterableOperation {
		return h.BootstrapShard(fabricv1nexus.FabricBootstrapShardOperationName)
	})
}

func registerFabricNexusOps(r worker.NexusServiceRegistry, ops ...func(*fabricv1temporal.FabricNexusHandler) nexus.RegisterableOperation) error {
	h := &fabricv1temporal.FabricNexusHandler{}
	svc := nexus.NewService(fabricv1nexus.FabricServiceName)
	registered := make([]nexus.RegisterableOperation, 0, len(ops))
	for _, op := range ops {
		registered = append(registered, op(h))
	}
	if err := svc.Register(registered...); err != nil {
		return fmt.Errorf("register fabric nexus ops: %w", err)
	}
	r.RegisterNexusService(svc)
	return nil
}
