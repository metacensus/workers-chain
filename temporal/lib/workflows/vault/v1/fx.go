package vault

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides vault v1 activities for dependency injection
//
// Dependency chain (VaultConfig must be provided by caller):
//
//	VaultConfig (provided by app config layer)
//	NewVaultClient(VaultConfig) -> VaultClient
//	NewActivities(VaultClient) -> pb.VaultActivities
//
// Usage in your main fx.App:
//
//	fx.New(
//	    // Provide VaultConfig from your central config
//	    fx.Provide(func() vaultv1.VaultConfig {
//	        return vaultv1.VaultConfig{
//	            Address: os.Getenv("VAULT_ADDR"),
//	            Token: os.Getenv("VAULT_TOKEN"),
//	        }
//	    }),
//	    vaultv1.Module,
//	    fx.Invoke(vaultv1.RegisterActivities),
//	)
var Module = fx.Options(
	// Provide Vault client (depends on VaultConfig from caller)
	fx.Provide(NewVaultClient),

	// Provide Activities implementing pb.VaultActivities interface (depends on VaultClient)
	fx.Provide(func(client VaultClient) pb.VaultActivities {
		return NewActivities(client)
	}),
)

// RegisterActivities registers vault activities with the worker
func RegisterActivities(w worker.Worker, activities pb.VaultActivities) {
	pb.RegisterVaultActivities(w, activities)
}
