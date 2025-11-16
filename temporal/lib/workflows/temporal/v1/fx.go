package temporal

import (
	"go.temporal.io/api/operatorservice/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	// Provide OperatorService client for Nexus endpoint operations
	fx.Provide(func(c client.Client) operatorservice.OperatorServiceClient {
		return c.OperatorService()
	}),

	// Provide Activities implementing pb.AdminActivities interface
	fx.Provide(func(namespaceClient client.NamespaceClient, operatorService operatorservice.OperatorServiceClient) pb.AdminActivities {
		return NewActivities(namespaceClient, operatorService)
	}),
)

// RegisterActivities registers temporal admin activities with the worker
func RegisterActivities(w worker.Worker, activities pb.AdminActivities) {
	pb.RegisterAdminActivities(w, activities)
}
