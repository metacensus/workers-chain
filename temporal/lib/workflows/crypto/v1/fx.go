package crypto

import (
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides activities
var Module = fx.Options(
	// Provide activities
	fx.Provide(NewActivities),
)

// RegisterActivities registers Crypto activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) error {
	if activities == nil {
		return fmt.Errorf("crypto activities must not be nil")
	}

	pb.RegisterCryptoActivities(w, activities)
	return nil
}

