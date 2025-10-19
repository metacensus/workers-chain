package vault

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides vault v1 workflows and activities for dependency injection
//
// Dependency chain (VaultConfig must be provided by caller):
//
//	VaultConfig (provided by app config layer)
//	NewVaultClient(VaultConfig) -> VaultClient
//	NewActivities(VaultClient) -> pb.VaultActivities
//	NewWorkflows(VaultClient) -> pb.VaultWorkflows
//
// Usage in your main fx.App:
//
//	fx.New(
//	    // Provide VaultConfig from your central config
//	    fx.Provide(func() vaultv1.VaultConfig {
//	        return vaultv1.VaultConfig{
//	            Address: os.Getenv("VAULT_ADDR"),
//	            // ... etc
//	        }
//	    }),
//	    vaultv1.Module,
//	    fx.Invoke(vaultv1.RegisterWorkflows, vaultv1.RegisterActivities),
//	)
var Module = fx.Options(
	// Provide Vault client (depends on VaultConfig from caller)
	fx.Provide(NewVaultClient),

	// Provide Activities implementing pb.VaultActivities interface (depends on VaultClient)
	fx.Provide(func(client VaultClient) pb.VaultActivities {
		return NewActivities(client)
	}),

	// Provide Workflows implementing pb.VaultWorkflows interface (depends on VaultClient)
	fx.Provide(func(client VaultClient) pb.VaultWorkflows {
		return NewWorkflows(client)
	}),
)

// RegisterWorkflows registers vault workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.VaultWorkflows) {
	pb.RegisterVaultWorkflows(w, workflows)
}

// RegisterActivities registers vault activities with the worker
func RegisterActivities(w worker.Worker, activities pb.VaultActivities) {
	pb.RegisterVaultActivities(w, activities)
}
