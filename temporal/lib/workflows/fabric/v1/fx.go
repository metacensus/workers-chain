package fabric

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1temporal"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides fabric v1 workflows and activities for dependency injection
var Module = fx.Options(
	// Provide Workflows implementing pb.FabricWorkflows interface
	fx.Provide(func(orgConfig *config.OrgConfig, fabricConfig fabricconfig.FabricConfig) pb.FabricWorkflows {
		return NewWorkflows(*orgConfig, fabricConfig)
	}),
	// Provide Activities
	fx.Provide(NewActivities),
)

// RegisterWorkflows registers fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.FabricWorkflows) {
	pb.RegisterFabricWorkflows(w, workflows)
}

// RegisterActivities registers fabric activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) {
	pb.RegisterFabricActivitiesActivities(w, activities)
}

// RegisterNexusService registers the fabric Nexus service handler with the worker
// This allows the fabric worker to receive Nexus calls from the central worker
func RegisterNexusService(w worker.Worker) error {
	return fabricv1temporal.RegisterFabricNexusService(w)
}
