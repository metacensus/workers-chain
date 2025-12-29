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
//	PostgresConfig (provided by config package or caller)
//	-> PostgresClient (via NewPostgresClient)
//	-> pb.PostgresActivities (via NewActivities)
//
// Usage in your main fx.App:
//
//	fx.New(
//	    postgresv1.Module,
//	    fx.Invoke(postgresv1.RegisterActivities),
//	)
var Module = fx.Options(
	// Provide PostgresClient from config
	fx.Provide(NewPostgresClient),

	// Provide Activities implementing pb.PostgresActivities interface
	fx.Provide(NewActivities),
)

// RegisterActivities registers postgres activities with the worker
func RegisterActivities(w worker.Worker, activities pb.PostgresActivities) {
	pb.RegisterPostgresActivities(w, activities)
}
