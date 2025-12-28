package postgres

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides postgres v1 activities for dependency injection
//
// Note: Activities create PostgreSQL clients from connection strings provided in activity inputs.
// No module-level configuration is required - each activity call provides its own connection string.
//
// Usage in your main fx.App:
//
//	fx.New(
//	    postgresv1.Module,
//	    fx.Invoke(postgresv1.RegisterActivities),
//	)
var Module = fx.Options(
	// Provide Activities implementing pb.PostgresActivities interface
	// Uses default client factory that creates clients from connection strings in inputs
	fx.Provide(func() pb.PostgresActivities {
		return NewActivitiesWithFactory(defaultClientFactory)
	}),
)

// RegisterActivities registers postgres activities with the worker
func RegisterActivities(w worker.Worker, activities pb.PostgresActivities) {
	pb.RegisterPostgresActivities(w, activities)
}

