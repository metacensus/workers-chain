package fabric

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1temporal"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides fabric v1 workflows and activities for dependency injection
//
// Two types of CA clients are available:
//
//  1. BootstrapClient - Username/password auth for initial bootstrap enrollment
//     Created on-demand, not cached (each enrollment is unique)
//
//  2. Client - Keypair/cert auth for admin operations (Register, Revoke)
//     Cached by MultiClientProvider using Vault path as key
//
// Dependency chain:
//
//	vault.VaultClient (provided by vault module)
//	-> MultiClientProvider[Client] (cached by Vault path)
//	-> Activities
var Module = fx.Options(
	// Provide Workflows implementing pb.FabricWorkflows interface
	// Workflows no longer depend on global orgConfig - they get org from input
	fx.Provide(func() pb.FabricWorkflows {
		return NewWorkflows()
	}),
)

// RegisterWorkflows registers fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.FabricWorkflows) {
	pb.RegisterFabricWorkflows(w, workflows)
}

// RegisterNexusService registers the fabric Nexus service handler with the worker
// This allows the fabric worker to receive Nexus calls from the central worker
func RegisterNexusService(w worker.Worker) error {
	return fabricv1temporal.RegisterFabricNexusService(w)
}
