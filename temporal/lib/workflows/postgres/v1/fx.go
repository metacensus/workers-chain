package postgres

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides postgres v1 activities for dependency injection
//
// Dependency chain:
//
//	config.OrgConfig (provided by config module)
//	-> ProviderConfig (via NewProviderConfig)
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
	// Provide ProviderConfig (built from OrgConfig)
	fx.Provide(NewProviderConfig),

	// Provide ClientProvider that fetches credentials from Vault
	fx.Provide(NewPostgresClientProvider),

	// Provide Activities implementing pb.PostgresActivities interface
	fx.Provide(NewActivities),
)

// RegisterActivities registers postgres activities with the worker
func RegisterActivities(w worker.Worker, activities pb.PostgresActivities) {
	pb.RegisterPostgresActivities(w, activities)
}
