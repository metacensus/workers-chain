package common

import (
	"context"
	"fmt"
	"sync"
)

// ConfigHydrator knows how to fetch secrets and inject them into a client config.
// The GetVaultPath is used both for fetching secrets and as the cache key for ClientProviderFactory.
type ConfigHydrator[ClientConfigT any] interface {
	// GetVaultPath returns the Vault path where secrets are stored.
	// This path also serves as the unique cache key for ClientProviderFactory.
	GetVaultPath() string

	// Hydrate takes secret data from Vault and produces a complete client config
	Hydrate(secretData map[string]interface{}) (ClientConfigT, error)
}

// ClientConstructor creates a client from a hydrated config
type ClientConstructor[ClientT any, ClientConfigT any] func(config ClientConfigT) (ClientT, error)

// VaultClient is the interface for fetching secrets (avoids circular dependency)
// This matches the vault.VaultClient.FetchSecret method signature
type VaultClient interface {
	FetchSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// -----------------------------------------------------------------------------
// ClientProvider: Single-client pattern
// Use when the hydrator/config is known at construction time (e.g., postgres)
// -----------------------------------------------------------------------------

// ClientProvider provides a single client instance, created on-demand with Vault credentials.
// Use this when there is exactly one client needed and the configuration is known at construction time.
type ClientProvider[ClientT any, ClientConfigT any] struct {
	vaultClient       VaultClient
	hydrator          ConfigHydrator[ClientConfigT]
	clientConstructor ClientConstructor[ClientT, ClientConfigT]

	clientMu sync.Mutex
	client   *ClientT // Pointer so we can check if initialized
}

// NewClientProvider creates a new client provider for a single client.
// The hydrator is fixed at construction time.
func NewClientProvider[ClientT any, ClientConfigT any](
	hydrator ConfigHydrator[ClientConfigT],
	vaultClient VaultClient,
	clientConstructor ClientConstructor[ClientT, ClientConfigT],
) *ClientProvider[ClientT, ClientConfigT] {
	return &ClientProvider[ClientT, ClientConfigT]{
		vaultClient:       vaultClient,
		hydrator:          hydrator,
		clientConstructor: clientConstructor,
	}
}

// GetClient returns the client, creating it on first call.
// Subsequent calls return the cached client.
func (p *ClientProvider[ClientT, ClientConfigT]) GetClient(ctx context.Context) (ClientT, error) {
	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	// Return existing client if already created
	if p.client != nil {
		return *p.client, nil
	}

	// Create client using shared logic
	client, err := createClient(ctx, p.vaultClient, p.hydrator, p.clientConstructor)
	if err != nil {
		var zero ClientT
		return zero, err
	}

	p.client = &client
	return client, nil
}

// -----------------------------------------------------------------------------
// MultiClientProvider: Multi-client pattern
// Use when the hydrator/config is only known at call time (e.g., certauthority)
// -----------------------------------------------------------------------------

// MultiClientProvider manages multiple clients keyed by their configuration.
// Use this when multiple clients may be needed and the specific configuration
// is only known at call time (determined by activity input).
//
// Clients are cached by their hydrator's GetVaultPath(), so repeated calls
// with the same configuration return the cached client.
type MultiClientProvider[ClientT any, ClientConfigT any] struct {
	vaultClient       VaultClient
	clientConstructor ClientConstructor[ClientT, ClientConfigT]

	clientMu sync.Mutex
	clients  map[string]*ClientT // keyed by hydrator.GetVaultPath()
}

// NewMultiClientProvider creates a new provider for managing multiple clients.
// The hydrator is provided per-call to GetClient.
func NewMultiClientProvider[ClientT any, ClientConfigT any](
	vaultClient VaultClient,
	clientConstructor ClientConstructor[ClientT, ClientConfigT],
) *MultiClientProvider[ClientT, ClientConfigT] {
	return &MultiClientProvider[ClientT, ClientConfigT]{
		vaultClient:       vaultClient,
		clientConstructor: clientConstructor,
		clients:           make(map[string]*ClientT),
	}
}

// GetClient returns a client for the given hydrator, creating and caching it on first call.
// The hydrator's GetVaultPath() is used as the cache key.
// Subsequent calls with hydrators that return the same VaultPath return the cached client.
func (f *MultiClientProvider[ClientT, ClientConfigT]) GetClient(
	ctx context.Context,
	hydrator ConfigHydrator[ClientConfigT],
) (ClientT, error) {
	// Use VaultPath as cache key - it uniquely identifies the client configuration
	key := hydrator.GetVaultPath()

	f.clientMu.Lock()
	defer f.clientMu.Unlock()

	// Return existing client if already created for this key
	if client, ok := f.clients[key]; ok {
		return *client, nil
	}

	// Create client using shared logic
	client, err := createClient(ctx, f.vaultClient, hydrator, f.clientConstructor)
	if err != nil {
		var zero ClientT
		return zero, err
	}

	f.clients[key] = &client
	return client, nil
}

// -----------------------------------------------------------------------------
// Shared client creation logic
// -----------------------------------------------------------------------------

// createClient handles the common logic of fetching secrets, hydrating config, and creating a client
func createClient[ClientT any, ClientConfigT any](
	ctx context.Context,
	vaultClient VaultClient,
	hydrator ConfigHydrator[ClientConfigT],
	clientConstructor ClientConstructor[ClientT, ClientConfigT],
) (ClientT, error) {
	var zero ClientT

	// Fetch secrets from Vault
	vaultPath := hydrator.GetVaultPath()
	secretData, err := vaultClient.FetchSecret(ctx, vaultPath)
	if err != nil {
		return zero, fmt.Errorf("failed to fetch secrets from %s: %w", vaultPath, err)
	}

	// Hydrate config with secrets
	config, err := hydrator.Hydrate(secretData)
	if err != nil {
		return zero, fmt.Errorf("failed to hydrate config: %w", err)
	}

	// Create client
	client, err := clientConstructor(config)
	if err != nil {
		return zero, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}
