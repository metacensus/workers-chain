package identity

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides fabric identity v1 workflows and activities for dependency injection
//
// Dependency chain (FabricCAConfig must be provided by caller):
//
//	FabricCAConfig (provided by app config layer)
//	NewFabricCAClient(FabricCAConfig) -> FabricCAClient
//	NewActivities(FabricCAClient) -> pb.FabricIdentityActivities
//	NewWorkflows() -> pb.FabricIdentityWorkflows
//
// Usage in your main fx.App:
//
//	fx.New(
//	    // Provide FabricCAConfig from your central config
//	    fx.Provide(func() fabricidentityv1.FabricCAConfig {
//	        return fabricidentityv1.FabricCAConfig{
//	            URL: os.Getenv("FABRIC_CA_URL"),
//	            // ... etc
//	        }
//	    }),
//	    fabricidentityv1.Module,
//	    fx.Invoke(fabricidentityv1.RegisterWorkflows, fabricidentityv1.RegisterActivities),
//	)
var Module = fx.Options(
	// Provide Fabric CA client (depends on FabricCAConfig from caller)
	fx.Provide(NewFabricCAClient),

	// Provide Activities implementing pb.FabricIdentityActivities interface (depends on FabricCAClient)
	fx.Provide(func(client FabricCAClient) pb.FabricIdentityActivities {
		return NewActivities(client)
	}),

	// Provide Workflows implementing pb.FabricIdentityWorkflows interface
	fx.Provide(func() pb.FabricIdentityWorkflows {
		return NewWorkflows()
	}),
)

// RegisterWorkflows registers fabric identity workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.FabricIdentityWorkflows) {
	pb.RegisterFabricIdentityWorkflows(w, workflows)
}

// RegisterActivities registers fabric identity activities with the worker
func RegisterActivities(w worker.Worker, activities pb.FabricIdentityActivities) {
	pb.RegisterFabricIdentityActivities(w, activities)
}
