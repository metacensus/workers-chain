package orderer

import (
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/orderer/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides orderer activities.
var Module = fx.Options(
	fx.Provide(NewActivities),
)

// RegisterActivities registers orderer activities with the worker.
func RegisterActivities(w worker.Worker, activities *Activities) error {
	if activities == nil {
		return fmt.Errorf("orderer activities must not be nil")
	}

	pb.RegisterOrdererActivities(w, activities)
	return nil
}
