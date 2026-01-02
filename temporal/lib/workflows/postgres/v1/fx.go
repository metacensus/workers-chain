package postgres

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides postgres v1 activities for dependency injection
var Module = fx.Options(
	// Wrap CertificateStoreConfig in postgres-specific alias type
	fx.Provide(func(certStoreConfig fabricconfig.CertificateStoreConfig) CertStorePostgresConfig {
		return NewCertStorePostgresConfig(certStoreConfig)
	}),

	// Provide ClientProvider that fetches credentials from Vault
	fx.Provide(func(config CertStorePostgresConfig, vaultClient vault.VaultClient) ClientProvider {
		return NewPostgresClientProvider(config, vaultClient)
	}),

	// Provide Activities with ClientProvider and SecretProvider
	fx.Provide(func(clientProvider ClientProvider, secretProvider secrets.SecretProvider) pb.PostgresActivities {
		return NewActivities(clientProvider, secretProvider)
	}),
)

func RegisterActivities(w worker.Worker, activities pb.PostgresActivities) {
	pb.RegisterPostgresActivities(w, activities)
}
