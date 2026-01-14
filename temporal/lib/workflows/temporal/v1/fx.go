package temporal

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// ClientModule provides Temporal SDK clients (used by activities)
// Can be used independently of activities/workflows
var ClientModule = fx.Options(
	// Provide Temporal client
	fx.Provide(NewClient),

	// Provide Temporal namespace client (for namespace management)
	fx.Provide(NewNamespaceClient),
)

// Module provides temporal admin activities and workflows along with clients
var Module = fx.Options(
	// Include client providers
	ClientModule,

	// Provide OperatorService client for Nexus endpoint operations
	fx.Provide(func(c client.Client) operatorservice.OperatorServiceClient {
		return c.OperatorService()
	}),

	// Provide Activities implementing pb.AdminActivities interface
	fx.Provide(func(namespaceClient client.NamespaceClient, operatorService operatorservice.OperatorServiceClient) pb.AdminActivities {
		return NewActivities(namespaceClient, operatorService)
	}),

	// Provide Workflows implementing pb.AdminWorkflows interface
	fx.Provide(func() pb.AdminWorkflows {
		return NewWorkflows()
	}),
)

// RegisterActivities registers temporal admin activities with the worker
func RegisterActivities(w worker.Worker, activities pb.AdminActivities) {
	pb.RegisterAdminActivities(w, activities)
}

// RegisterWorkflows registers temporal admin workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.AdminWorkflows) {
	pb.RegisterAdminWorkflows(w, workflows)
}
