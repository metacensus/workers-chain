package workers

import (
	"github.com/metacensus/infra/temporal/lib/env"
	"github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	"github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// NewTemporalClient creates a Temporal client from environment configuration
func NewTemporalClient() (client.Client, error) {
	return client.Dial(client.Options{
		HostPort: env.TemporalAddress(),
	})
}

// NewNamespaceClient creates a Temporal namespace client for namespace management operations
func NewNamespaceClient() (client.NamespaceClient, error) {
	return client.NewNamespaceClient(client.Options{
		HostPort: env.TemporalAddress(),
	})
}

// NewWorker creates a Temporal worker from the client and environment configuration
func NewWorker(c client.Client) worker.Worker {
	return worker.New(c, env.TaskQueue(), worker.Options{})
}

// NewVaultConfig creates VaultConfig from environment variables
func NewVaultConfig() vault.VaultConfig {
	return vault.VaultConfig{
		Address: env.VaultAddress(),
		Token:   env.VaultToken(),
	}
}

// NewKubeConfig creates KubeConfig from environment variables
func NewKubeConfig() kube.KubeConfig {
	return kube.KubeConfig{
		KubeconfigPath: env.KubeconfigPath(),
	}
}
