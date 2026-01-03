package fabric

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1temporal"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	"github.com/metacensus/infra/temporal/lib/workflows/fabric/v1/certauthority"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
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
	fx.Provide(func(fabricConfig fabricconfig.FabricConfig) pb.FabricWorkflows {
		return NewWorkflows(fabricConfig)
	}),

	// Provide Client multi-client provider
	// Clients are cached because they have stable keypair/cert auth from Vault.
	// Uses MultiClientProvider because there are many CAs (tls-root, org-root, etc.)
	// and the specific CA is only known at activity call time.
	fx.Provide(func(vaultClient vault.VaultClient) *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig] {
		return common.NewMultiClientProvider(
			vaultClient,
			certauthority.NewClient,
		)
	}),

	// Provide Activities with CA client providers and config
	// Activities get:
	// - secretProvider: for fetching/storing secrets with type safety
	// - clientProvider: for cached Client access (Register/Revoke operations)
	// - fabricConfig: for building CA-specific configurations
	fx.Provide(func(
		secretProvider secrets.SecretProvider,
		clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig],
		fabricConfig fabricconfig.FabricConfig,
	) *Activities {
		return NewActivities(secretProvider, clientProvider, fabricConfig)
	}),
)

// RegisterWorkflows registers fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.FabricWorkflows) {
	pb.RegisterFabricWorkflows(w, workflows)
}

// RegisterActivities registers fabric activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) {
	pb.RegisterFabricActivitiesActivities(w, activities)
}

// RegisterNexusService registers the fabric Nexus service handler with the worker
// This allows the fabric worker to receive Nexus calls from the central worker
func RegisterNexusService(w worker.Worker) error {
	return fabricv1temporal.RegisterFabricNexusService(w)
}
