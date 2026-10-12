package postgres

import (
	pb "github.com/metacensus/workers-chain/gen/workflows/postgres/v1"
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
	secrets "github.com/metacensus/workers-chain/lib/workflows/secrets/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides postgres v1 activities for dependency injection
var Module = fx.Options(
	// Wrap CertificateStoreConfig in postgres-specific alias type
	fx.Provide(func(certStoreConfig fabricconfig.CertificateStoreConfig) CertStorePostgresConfig {
		return NewCertStorePostgresConfig(certStoreConfig)
	}),

	// Provide ClientProvider that fetches credentials from secrets
	fx.Provide(func(config CertStorePostgresConfig, secretProvider secrets.SecretProvider) ClientProvider {
		return NewPostgresClientProvider(config, secretProvider)
	}),

	// Provide Activities with ClientProvider and SecretProvider
	fx.Provide(func(clientProvider ClientProvider, secretProvider secrets.SecretProvider) pb.PostgresActivities {
		return NewActivities(clientProvider, secretProvider)
	}),
)

func RegisterActivities(w worker.Worker, activities pb.PostgresActivities) {
	pb.RegisterPostgresActivities(w, activities)
}
