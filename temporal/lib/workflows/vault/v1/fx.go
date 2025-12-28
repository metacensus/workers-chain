package vault

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/vault/v1/vaultv1temporal"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides vault v1 activities and workflows for dependency injection
//
// Dependency chain:
//
//	config.VaultConfig (provided by config package)
//	-> VaultClient (via NewVaultClient)
//	-> pb.VaultActivities (via NewActivities)
//	pb.VaultWorkflows (via NewWorkflows)
var Module = fx.Options(
	// Provide Vault client
	fx.Provide(NewVaultClient),

	// Provide Activities implementing pb.VaultActivities interface (depends on VaultClient)
	fx.Provide(func(client VaultClient) pb.VaultActivities {
		return NewActivities(client)
	}),

	// Provide Workflows implementing pb.VaultWorkflows interface
	fx.Provide(func() pb.VaultWorkflows {
		return NewWorkflows()
	}),
)

// RegisterActivities registers vault activities with the worker
func RegisterActivities(w worker.Worker, activities pb.VaultActivities) {
	pb.RegisterVaultActivities(w, activities)
}

// RegisterWorkflows registers vault workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.VaultWorkflows) {
	pb.RegisterVaultWorkflows(w, workflows)
}

// RegisterNexusService registers the vault Nexus service handler with the worker
// This allows the infra worker to receive Nexus calls from other namespaces (e.g., fabric shards)
func RegisterNexusService(w worker.Worker) error {
	return vaultv1temporal.RegisterVaultNexusService(w)
}
