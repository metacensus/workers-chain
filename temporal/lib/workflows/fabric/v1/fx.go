package fabric

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1temporal"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	"github.com/metacensus/infra/temporal/lib/workflows/fabric/v1/certauthority"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides fabric v1 workflows and activities for dependency injection
//
// Dependency chain for CA client:
//
//	vault.VaultClient (provided by vault module)
//	-> MultiClientProvider (via common.NewMultiClientProvider)
//	-> Activities (via NewActivities)
//
// The ClientProviderFactory caches clients by their Vault path, so repeated
// calls to the same CA return the cached client.
var Module = fx.Options(
	// Provide Workflows implementing pb.FabricWorkflows interface
	fx.Provide(func(fabricConfig fabricconfig.FabricConfig) pb.FabricWorkflows {
		return NewWorkflows(fabricConfig)
	}),

	// Provide CA multi-client provider
	// Uses MultiClientProvider because there are many CAs (tls-root, org-root, etc.)
	// and the specific CA is only known at activity call time.
	// Clients are cached by their Vault path for efficiency.
	fx.Provide(func(vaultClient vault.VaultClient) *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig] {
		return common.NewMultiClientProvider(
			vaultClient,
			certauthority.NewClient, // ClientConstructor
		)
	}),

	// Provide Activities with CA multi-client provider and config
	fx.Provide(func(
		clientProvider *common.MultiClientProvider[certauthority.Client, certauthority.ClientConfig],
		fabricConfig fabricconfig.FabricConfig,
	) *Activities {
		return NewActivities(clientProvider, fabricConfig)
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
