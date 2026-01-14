package temporal

import (
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/client"
)

// NewClient creates a Temporal client from TemporalConfig
func NewClient(cfg *config.TemporalConfig) (client.Client, error) {
	return client.Dial(client.Options{
		HostPort:  cfg.Address,
		Namespace: cfg.Namespace,
	})
}

// NewNamespaceClient creates a Temporal namespace client for namespace management operations
func NewNamespaceClient(cfg *config.TemporalConfig) (client.NamespaceClient, error) {
	return client.NewNamespaceClient(client.Options{
		HostPort: cfg.Address,
	})
}
