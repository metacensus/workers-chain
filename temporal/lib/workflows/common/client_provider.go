package common

import (
	"context"
	"fmt"
	"sync"
)

// ConfigHydrator knows how to fetch secrets and inject them into a client config
type ConfigHydrator[ClientConfigT any] interface {
	// GetVaultPath returns the Vault path where secrets are stored
	GetVaultPath() string

	// Hydrate takes secret data from Vault and produces a complete client config
	Hydrate(secretData map[string]interface{}) (ClientConfigT, error)
}

// ClientFactory creates a client from a hydrated config
type ClientFactory[ClientT any, ClientConfigT any] func(config ClientConfigT) (ClientT, error)

// VaultClient is the interface for fetching secrets (avoids circular dependency)
// This matches the vault.VaultClient.FetchSecret method signature
type VaultClient interface {
	FetchSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// ClientProvider provides a single client instance, created on-demand with Vault credentials
type ClientProvider[ClientT any, ClientConfigT any] struct {
	vaultClient   VaultClient
	hydrator      ConfigHydrator[ClientConfigT]
	clientFactory ClientFactory[ClientT, ClientConfigT]

	clientMu sync.Mutex
	client   *ClientT // Pointer so we can check if initialized
}

// NewClientProvider creates a new client provider
func NewClientProvider[ClientT any, ClientConfigT any](
	hydrator ConfigHydrator[ClientConfigT],
	vaultClient VaultClient,
	clientFactory ClientFactory[ClientT, ClientConfigT],
) *ClientProvider[ClientT, ClientConfigT] {
	return &ClientProvider[ClientT, ClientConfigT]{
		vaultClient:   vaultClient,
		hydrator:      hydrator,
		clientFactory: clientFactory,
	}
}

// GetClient returns the client, creating it on first call
func (p *ClientProvider[ClientT, ClientConfigT]) GetClient(ctx context.Context) (ClientT, error) {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	// Return existing client if already created
	if p.client != nil {
		return *p.client, nil
	}

	// Fetch secrets from Vault
	vaultPath := p.hydrator.GetVaultPath()
	secretData, err := p.vaultClient.FetchSecret(ctx, vaultPath)
	if err != nil {
		var zero ClientT
		return zero, fmt.Errorf("failed to fetch secrets from %s: %w", vaultPath, err)
	}

	// Hydrate config with secrets
	config, err := p.hydrator.Hydrate(secretData)
	if err != nil {
		var zero ClientT
		return zero, fmt.Errorf("failed to hydrate config: %w", err)
	}

	// Create client
	client, err := p.clientFactory(config)
	if err != nil {
		var zero ClientT
		return zero, fmt.Errorf("failed to create client: %w", err)
	}

	p.client = &client
	return client, nil
}
