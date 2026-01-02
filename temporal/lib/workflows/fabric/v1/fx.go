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
//	fabricconfig.CertificateAuthorityConfig (provided by fabric worker module)
//	-> CAConfig (via NewCAConfig)
//	vault.VaultClient (provided by vault module)
//	-> ClientProvider (via NewCAClientProvider)
//	-> Activities (via NewActivities)
var Module = fx.Options(
	// Provide Workflows implementing pb.FabricWorkflows interface
	fx.Provide(func(fabricConfig fabricconfig.FabricConfig) pb.FabricWorkflows {
		return NewWorkflows(fabricConfig)
	}),

	// Provide CA client provider factory
	// Note: This provides a factory that can create client providers for different CAs on demand
	// Activities will need to get specific CA configs and create providers as needed
	fx.Provide(func(vaultClient vault.VaultClient) ClientProviderFactory {
		return func(caConfig fabricconfig.CertificateAuthorityConfig) ClientProvider {
			return certauthority.NewClientProvider(certauthority.NewConfig(caConfig), vaultClient)
		}
	}),

	// Provide Activities with CA client provider factory and config
	fx.Provide(func(
		providerFactory ClientProviderFactory,
		fabricConfig fabricconfig.FabricConfig,
	) *Activities {
		return NewActivities(providerFactory, fabricConfig)
	}),
)

// ClientProviderFactory creates a ClientProvider for a specific CA configuration
// This allows activities to dynamically create clients for different CAs (tls-root, org-root, etc.)
type ClientProviderFactory func(caConfig fabricconfig.CertificateAuthorityConfig) ClientProvider

// ClientProvider is a type alias for the common.ClientProvider specialized for CA clients
type ClientProvider = *common.ClientProvider[certauthority.Client, certauthority.ClientConfig]

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
