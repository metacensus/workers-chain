package common

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/metacensus/workers-chain/gen/workflows/secrets/v1"
	secrets "github.com/metacensus/workers-chain/lib/workflows/secrets/v1"
)

// ConfigHydrator knows how to fetch secrets and inject them into a client config.
type ConfigHydrator[ClientConfigT any] interface {
	// GetSecretRefs returns a map of named secret references.
	// For simple cases with a single secret, use a standard key like "default".
	// For complex cases with multiple secrets, use descriptive keys like "credentials", "tls-cert", etc.
	GetSecretRefs() map[string]*pb.SecretRef

	// GetCacheKey returns a unique key for caching clients.
	// This should uniquely identify the client configuration.
	GetCacheKey() string

	// Hydrate takes a map of named secrets and produces a complete client config
	Hydrate(secrets map[string]secrets.Secret) (ClientConfigT, error)
}

// ClientConstructor creates a client from a hydrated config
type ClientConstructor[ClientT any, ClientConfigT any] func(config ClientConfigT) (ClientT, error)

// -----------------------------------------------------------------------------
// ClientProvider: Single-client pattern
// Use when the hydrator/config is known at construction time (e.g., postgres)
// -----------------------------------------------------------------------------

// ClientProvider provides a single client instance, created on-demand with secrets.
// Use this when there is exactly one client needed and the configuration is known at construction time.
type ClientProvider[ClientT any, ClientConfigT any] struct {
	secretProvider    secrets.SecretProvider
	hydrator          ConfigHydrator[ClientConfigT]
	clientConstructor ClientConstructor[ClientT, ClientConfigT]

	clientMu sync.Mutex
	client   *ClientT // Pointer so we can check if initialized
}

// NewClientProvider creates a new client provider for a single client.
// The hydrator is fixed at construction time.
func NewClientProvider[ClientT any, ClientConfigT any](
	hydrator ConfigHydrator[ClientConfigT],
	secretProvider secrets.SecretProvider,
	clientConstructor ClientConstructor[ClientT, ClientConfigT],
) *ClientProvider[ClientT, ClientConfigT] {
	return &ClientProvider[ClientT, ClientConfigT]{
		secretProvider:    secretProvider,
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
	client, err := createClient(ctx, p.secretProvider, p.hydrator, p.clientConstructor)
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
// Clients are cached by their hydrator's GetCacheKey(), so repeated calls
// with the same configuration return the cached client.
type MultiClientProvider[ClientT any, ClientConfigT any] struct {
	secretProvider    secrets.SecretProvider
	clientConstructor ClientConstructor[ClientT, ClientConfigT]

	clientMu sync.Mutex
	clients  map[string]*ClientT // keyed by hydrator.GetCacheKey()
}

// NewMultiClientProvider creates a new provider for managing multiple clients.
// The hydrator is provided per-call to GetClient.
func NewMultiClientProvider[ClientT any, ClientConfigT any](
	secretProvider secrets.SecretProvider,
	clientConstructor ClientConstructor[ClientT, ClientConfigT],
) *MultiClientProvider[ClientT, ClientConfigT] {
	return &MultiClientProvider[ClientT, ClientConfigT]{
		secretProvider:    secretProvider,
		clientConstructor: clientConstructor,
		clients:           make(map[string]*ClientT),
	}
}

// GetClient returns a client for the given hydrator, creating and caching it on first call.
// The hydrator's GetCacheKey() is used as the cache key.
// Subsequent calls with hydrators that return the same cache key return the cached client.
func (f *MultiClientProvider[ClientT, ClientConfigT]) GetClient(
	ctx context.Context,
	hydrator ConfigHydrator[ClientConfigT],
) (ClientT, error) {
	// Use cache key - it uniquely identifies the client configuration
	key := hydrator.GetCacheKey()

	f.clientMu.Lock()
	defer f.clientMu.Unlock()

	// Return existing client if already created for this key
	if client, ok := f.clients[key]; ok {
		return *client, nil
	}

	// Create client using shared logic
	client, err := createClient(ctx, f.secretProvider, hydrator, f.clientConstructor)
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
	secretProvider secrets.SecretProvider,
	hydrator ConfigHydrator[ClientConfigT],
	clientConstructor ClientConstructor[ClientT, ClientConfigT],
) (ClientT, error) {
	var zero ClientT

	// Fetch all secrets specified by the hydrator
	secretRefs := hydrator.GetSecretRefs()
	fetchedSecrets := make(map[string]secrets.Secret)

	for name, ref := range secretRefs {
		secret, err := secretProvider.FetchSecret(ctx, ref)
		if err != nil {
			return zero, fmt.Errorf("failed to fetch secret %q from %v: %w", name, ref, err)
		}
		fetchedSecrets[name] = secret
	}

	// Hydrate config with fetched secrets
	config, err := hydrator.Hydrate(fetchedSecrets)
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
