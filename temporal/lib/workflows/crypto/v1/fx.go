package crypto

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides crypto v1 activities and workflows for dependency injection
var Module = fx.Options(
	// Provide Activities implementing pb.CryptoActivities interface
	fx.Provide(func() pb.CryptoActivities {
		return NewActivities()
	}),

	// Provide Workflows implementing pb.CryptoWorkflows interface
	fx.Provide(func() pb.CryptoWorkflows {
		return NewWorkflows()
	}),
)

// RegisterActivities registers crypto activities with the worker
func RegisterActivities(w worker.Worker, activities pb.CryptoActivities) {
	pb.RegisterCryptoActivities(w, activities)
}

// RegisterWorkflows registers crypto workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.CryptoWorkflows) {
	pb.RegisterCryptoWorkflows(w, workflows)
}
