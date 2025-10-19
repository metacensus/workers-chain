package workflows

import (
	"github.com/metacensus/infra/temporal/env"
	"github.com/metacensus/infra/temporal/workflows/fabric/identity/v1"
	"github.com/metacensus/infra/temporal/workflows/founder"
	"github.com/metacensus/infra/temporal/workflows/vault/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides all workflow infrastructure including Temporal client, worker,
// and all workflow/activity modules
var Module = fx.Options(
	// Provide Temporal client from environment
	fx.Provide(NewTemporalClient),

	// Provide worker with dependencies
	fx.Provide(NewWorker),

	// Provide Vault configuration from environment
	fx.Provide(NewVaultConfig),

	// Provide Fabric CA configuration from environment
	fx.Provide(NewFabricCAConfig),

	// Include vault workflows and activities module
	vault.Module,

	// Include fabric identity workflows and activities module
	identity.Module,

	// Register all workflows and activities independently
	fx.Invoke(vault.RegisterWorkflows),
	fx.Invoke(vault.RegisterActivities),
	fx.Invoke(identity.RegisterWorkflows),
	fx.Invoke(identity.RegisterActivities),
	fx.Invoke(founder.RegisterWorkflows),
)

// NewTemporalClient creates a Temporal client from environment configuration
func NewTemporalClient() (client.Client, error) {
	return client.Dial(client.Options{
		HostPort: env.TemporalAddress(),
	})
}

// NewWorker creates a Temporal worker from the client and environment configuration
func NewWorker(c client.Client) worker.Worker {
	return worker.New(c, env.InfraTaskQueue(), worker.Options{})
}

// NewVaultConfig creates VaultConfig from environment variables
func NewVaultConfig() vault.VaultConfig {
	return vault.VaultConfig{
		Address:    env.VaultAddress(),
		Token:      env.VaultToken(),
		MountPath:  env.VaultMountPath(),
		EngineType: env.VaultEngineType(),
	}
}

// NewFabricCAConfig creates FabricCAConfig from environment variables
func NewFabricCAConfig() identity.FabricCAConfig {
	return identity.FabricCAConfig{
		URL:     env.FabricCAURL(),
		CAName:  env.FabricCAName(),
		HomeDir: env.FabricCAHomeDir(),
	}
}
