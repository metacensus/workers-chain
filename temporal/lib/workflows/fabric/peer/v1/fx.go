package peer

import (
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/peer/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides peer activities.
var Module = fx.Options(
	fx.Provide(NewActivities),
)

// RegisterActivities registers peer activities with the worker.
func RegisterActivities(w worker.Worker, activities *Activities) error {
	if activities == nil {
		return fmt.Errorf("peer activities must not be nil")
	}

	pb.RegisterPeerActivities(w, activities)
	return nil
}
