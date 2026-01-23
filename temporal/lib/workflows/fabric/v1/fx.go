package fabric

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1temporal"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	"github.com/metacensus/infra/temporal/lib/workflows/fabric/v1/certauthority"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
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

	// Provide Client multi-client provider
	// Clients are cached because they have stable keypair/cert auth from secrets.
	// Uses MultiClientProvider because there are many CAs (tls-root, org-root, etc.)
	// and the specific CA is only known at activity call time.
	fx.Provide(func(secretProvider secrets.SecretProvider) *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig] {
		return common.NewMultiClientProvider(
			secretProvider,
			certauthority.NewClient,
		)
	}),

	// Provide Activities with CA client providers and config
	// Activities get:
	// - secretProvider: for fetching/storing secrets with type safety
	// - clientProvider: for cached Client access (Register/Revoke operations)
	// - orgConfig: for building CA-specific configurations (activities run on fabric worker which has ORG_ID)
	fx.Provide(func(
		secretProvider secrets.SecretProvider,
		clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig],
		orgConfig config.OrgConfig,
	) *Activities {
		return NewActivities(secretProvider, clientProvider, orgConfig)
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
