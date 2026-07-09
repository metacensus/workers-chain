package certauth

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1/certauthv1temporal"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	// Provide Workflows implementing pb.FabricWorkflows interface
	// Workflows no longer depend on global orgConfig - they get org from input
	fx.Provide(func() pb.CertificateAuthorityWorkflows {
		return NewWorkflows()
	}),

	// Provide Client multi-client provider
	// Clients are cached because they have stable keypair/cert auth from secrets.
	// Uses MultiClientProvider because there are many CAs (tls-root, org-root, etc.)
	// and the specific CA is only known at activity call time.
	fx.Provide(func(secretProvider secrets.SecretProvider) *common.MultiClientProvider[Client, ClientConfig] {
		return common.NewMultiClientProvider(
			secretProvider,
			NewClient,
		)
	}),

	// Provide Activities with CA client providers
	// Activities get:
	// - secretProvider: for fetching/storing secrets with type safety
	// - clientProvider: for cached Client access (Register/Revoke operations)
	// Activities receive owner (org) from their inputs, enabling them to work for any owner
	fx.Provide(func(
		secretProvider secrets.SecretProvider,
		clientProvider *common.MultiClientProvider[Client, ClientConfig],
	) *Activities {
		return NewActivities(secretProvider, clientProvider)
	}),
)

// RegisterWorkflows registers fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.CertificateAuthorityWorkflows) {
	pb.RegisterCertificateAuthorityWorkflows(w, workflows)
}

// RegisterActivities registers fabric activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) {
	pb.RegisterIdentityActivities(w, activities)
}

// RegisterNexusService registers the CertificateAuthority Nexus service handler with the worker.
// This allows the worker owning the shared TLS CAs to receive cross-namespace
// certificate issuance calls (e.g. CreateCertificate) from org workers.
//
// The parameter is worker.Worker (not worker.NexusServiceRegistry) so fx can
// inject the worker.Worker from the container; worker.Worker embeds
// worker.NexusServiceRegistry, which is all the generated registration needs.
func RegisterNexusService(w worker.Worker) error {
	return certauthv1temporal.RegisterCertificateAuthorityNexusService(w)
}
