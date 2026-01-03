// Package secrets provides backend-agnostic secret management for activities.
package secrets

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// Secret type constants for SecretRef.Type
const (
	TypeVault = "vault"
)

// Sentinel errors
var (
	ErrUnknownSecretType = errors.New("unknown secret type")
	ErrFetchSecret       = errors.New("failed to fetch secret")
	ErrStoreSecret       = errors.New("failed to store secret")
)

// SecretClient provides low-level generic secret operations.
// Each backend (vault, k8s, etc.) implements this interface.
// This is the minimal interface that backends must implement.
type SecretClient interface {
	// FetchSecretByRef fetches generic secret data from the backend
	FetchSecretByRef(ctx context.Context, ref map[string]interface{}) (map[string]interface{}, error)

	// StoreSecretByRef stores generic secret data to the backend (complete replacement)
	StoreSecretByRef(ctx context.Context, ref map[string]interface{}, secret map[string]interface{}) error

	// UpdateSecretByRef merges new fields with existing secret data
	// If the secret doesn't exist, creates it with the provided fields
	// If the secret exists, merges new fields (new fields override existing ones)
	UpdateSecretByRef(ctx context.Context, ref map[string]interface{}, partialSecret map[string]interface{}) error
}

// SecretProvider provides generic secret operations.
// It dispatches to the appropriate backend client based on SecretRef.Type.
//
// Design Philosophy:
// - FetchSecret returns Secret (generic map) that exposes typed conversion methods
// - StoreSecret/UpdateSecret accept Secret maps
// - Typed operations (Credentials, KeyPair, etc.) are methods on Secret type
type SecretProvider interface {
	// FetchSecret fetches generic secret data
	FetchSecret(ctx context.Context, ref *pb.SecretRef) (Secret, error)

	// StoreSecret stores complete secret data (replaces existing)
	StoreSecret(ctx context.Context, ref *pb.SecretRef, secret Secret) error

	// UpdateSecret merges new fields with existing secret data
	UpdateSecret(ctx context.Context, ref *pb.SecretRef, partialSecret Secret) error
}

type secretProvider struct {
	mu      sync.RWMutex
	clients map[string]SecretClient
}

// NewSecretProvider creates a new SecretProvider with no clients registered.
func NewSecretProvider() *secretProvider {
	return &secretProvider{
		clients: make(map[string]SecretClient),
	}
}

// RegisterClient registers a SecretClient for the given type.
func (p *secretProvider) RegisterClient(secretType string, client SecretClient) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clients[secretType] = client
}

func (p *secretProvider) getClient(secretType string) (SecretClient, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	client, ok := p.clients[secretType]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownSecretType, secretType)
	}
	return client, nil
}

func (p *secretProvider) FetchSecret(ctx context.Context, ref *pb.SecretRef) (Secret, error) {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return nil, err
	}
	secret, err := client.FetchSecretByRef(ctx, ref.GetRef().AsMap())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchSecret, err)
	}
	return Secret(secret), nil
}

func (p *secretProvider) StoreSecret(ctx context.Context, ref *pb.SecretRef, secret Secret) error {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return err
	}
	if err := client.StoreSecretByRef(ctx, ref.GetRef().AsMap(), secret); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreSecret, err)
	}
	return nil
}

func (p *secretProvider) UpdateSecret(ctx context.Context, ref *pb.SecretRef, partialSecret Secret) error {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return err
	}
	if err := client.UpdateSecretByRef(ctx, ref.GetRef().AsMap(), partialSecret); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreSecret, err)
	}
	return nil
}

var _ SecretProvider = (*secretProvider)(nil)

// NewVaultRef creates a SecretRef for a Vault path
func NewVaultRef(path string) *pb.SecretRef {
	ref, _ := structpb.NewStruct(map[string]interface{}{
		"path": path,
	})
	return &pb.SecretRef{
		Type: TypeVault,
		Ref:  ref,
	}
}

// NewVaultRefWithVersion creates a SecretRef for a Vault path with version
func NewVaultRefWithVersion(path string, version int64) *pb.SecretRef {
	ref, _ := structpb.NewStruct(map[string]interface{}{
		"path":    path,
		"version": version,
	})
	return &pb.SecretRef{
		Type: TypeVault,
		Ref:  ref,
	}
}
