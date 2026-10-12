package vault

import (
	pb "github.com/metacensus/workers-chain/gen/workflows/vault/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides vault v1 activities for dependency injection
var Module = fx.Options(
	// Provide Vault client
	fx.Provide(NewVaultClient),

	// Provide Activities implementing pb.VaultActivities interface
	fx.Provide(func(client VaultClient) pb.VaultActivities {
		return NewActivities(client)
	}),
)

// RegisterActivities registers vault activities with the worker
func RegisterActivities(w worker.Worker, activities pb.VaultActivities) {
	pb.RegisterVaultActivities(w, activities)
}
