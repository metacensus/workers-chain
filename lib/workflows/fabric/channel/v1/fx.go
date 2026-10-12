package channel

import (
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/channel/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides channel activities.
var Module = fx.Options(
	fx.Provide(NewActivities),
)

// RegisterActivities registers channel activities with the worker.
func RegisterActivities(w worker.Worker, activities *Activities) error {
	if activities == nil {
		return fmt.Errorf("channel activities must not be nil")
	}

	pb.RegisterChannelActivities(w, activities)
	return nil
}
