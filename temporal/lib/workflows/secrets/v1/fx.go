package secrets

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/secrets/v1/secretsv1temporal"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides secrets management for dependency injection
var Module = fx.Options(
	fx.Provide(func(vaultClient vault.VaultClient) SecretProvider {
		provider := NewSecretProvider()
		vaultSecretClient := NewVaultSecretClient(vaultClient)
		provider.RegisterClient(TypeVault, vaultSecretClient)
		return provider
	}),

	fx.Provide(func(provider SecretProvider) *Activities {
		return NewActivities(provider)
	}),

	fx.Provide(func() *Workflows {
		return NewWorkflows()
	}),
)

// RegisterActivities registers secrets activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) {
	pb.RegisterSecretsActivities(w, activities)
}

// RegisterWorkflows registers secrets workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows *Workflows) {
	pb.RegisterSecretsWorkflows(w, workflows)
}

// RegisterNexusService registers the secrets Nexus service handler with the worker
func RegisterNexusService(w worker.Worker) error {
	return secretsv1temporal.RegisterSecretsNexusService(w)
}
