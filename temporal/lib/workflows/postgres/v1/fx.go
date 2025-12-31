package postgres

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides postgres v1 activities for dependency injection
//
// Dependency chain:
//
//	fabricconfig.CertificateStoreConfig (provided by fabric worker module)
//	-> CertStorePostgresConfig (via NewCertStorePostgresConfig)
//	vault.VaultClient (provided by vault module)
//	-> ClientProvider (via NewPostgresClientProvider)
//	-> pb.PostgresActivities (via NewActivities)
//
// Usage in your main fx.App:
//
//	fx.New(
//	    postgresv1.Module,
//	    fx.Invoke(postgresv1.RegisterActivities),
//	)
var Module = fx.Options(
	// Wrap CertificateStoreConfig in postgres-specific alias type
	fx.Provide(func(certStoreConfig fabricconfig.CertificateStoreConfig) CertStorePostgresConfig {
		return NewCertStorePostgresConfig(certStoreConfig)
	}),

	// Provide ClientProvider that fetches credentials from Vault
	fx.Provide(func(config CertStorePostgresConfig, vaultClient vault.VaultClient) ClientProvider {
		return NewPostgresClientProvider(config, vaultClient)
	}),

	// Provide Activities implementing pb.PostgresActivities interface
	fx.Provide(func(clientProvider ClientProvider) pb.PostgresActivities {
		return NewActivities(clientProvider)
	}),
)

func RegisterActivities(w worker.Worker, activities pb.PostgresActivities) {
	pb.RegisterPostgresActivities(w, activities)
}
