package temporal

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	// Provide Activities implementing pb.AdminActivities interface (depends on TemporalAdminClient)
	fx.Provide(func(client client.NamespaceClient) pb.AdminActivities {
		return NewActivities(client)
	}),
)

// RegisterActivities registers temporal admin activities with the worker
func RegisterActivities(w worker.Worker, activities pb.AdminActivities) {
	pb.RegisterAdminActivities(w, activities)
}
