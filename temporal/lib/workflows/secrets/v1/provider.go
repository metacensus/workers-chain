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
	ErrFetchCredentials  = errors.New("failed to fetch credentials")
	ErrStoreCredentials  = errors.New("failed to store credentials")
	ErrFetchKeyPair      = errors.New("failed to fetch key pair")
	ErrStoreKeyPair      = errors.New("failed to store key pair")
	ErrFetchCertificate  = errors.New("failed to fetch certificate")
	ErrStoreCertificate  = errors.New("failed to store certificate")
)

// =============================================================================
// Secret Data Types
// =============================================================================

// Credentials represents username/password
type Credentials struct {
	Username string
	Password string
}

// KeyPair represents a public/private key pair (PEM-encoded)
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

// Certificate represents a certificate (PEM-encoded)
type Certificate struct {
	Certificate []byte
}

// =============================================================================
// SecretClient Interface
// =============================================================================

// SecretClient provides typed methods for secret operations.
// Each backend (vault, k8s, etc.) implements this interface.
type SecretClient interface {
	// Credentials
	FetchCredentials(ctx context.Context, ref *structpb.Struct) (*Credentials, error)
	StoreCredentials(ctx context.Context, ref *structpb.Struct, creds *Credentials) error

	// KeyPairs
	FetchKeyPair(ctx context.Context, ref *structpb.Struct) (*KeyPair, error)
	StoreKeyPair(ctx context.Context, ref *structpb.Struct, kp *KeyPair) error

	// Certificates
	FetchCertificate(ctx context.Context, ref *structpb.Struct) (*Certificate, error)
	StoreCertificate(ctx context.Context, ref *structpb.Struct, cert *Certificate) error
}

// =============================================================================
// SecretProvider
// =============================================================================

// SecretProvider dispatches secret operations to the appropriate backend client.
type SecretProvider interface {
	// Credentials
	FetchCredentials(ctx context.Context, ref *pb.SecretRef) (*Credentials, error)
	StoreCredentials(ctx context.Context, ref *pb.SecretRef, creds *Credentials) error

	// KeyPairs
	FetchKeyPair(ctx context.Context, ref *pb.SecretRef) (*KeyPair, error)
	StoreKeyPair(ctx context.Context, ref *pb.SecretRef, kp *KeyPair) error

	// Certificates
	FetchCertificate(ctx context.Context, ref *pb.SecretRef) (*Certificate, error)
	StoreCertificate(ctx context.Context, ref *pb.SecretRef, cert *Certificate) error
}

// =============================================================================
// Implementation
// =============================================================================

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

func (p *secretProvider) FetchCredentials(ctx context.Context, ref *pb.SecretRef) (*Credentials, error) {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return nil, err
	}
	creds, err := client.FetchCredentials(ctx, ref.GetRef())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchCredentials, err)
	}
	return creds, nil
}

func (p *secretProvider) StoreCredentials(ctx context.Context, ref *pb.SecretRef, creds *Credentials) error {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return err
	}
	if err := client.StoreCredentials(ctx, ref.GetRef(), creds); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreCredentials, err)
	}
	return nil
}

func (p *secretProvider) FetchKeyPair(ctx context.Context, ref *pb.SecretRef) (*KeyPair, error) {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return nil, err
	}
	kp, err := client.FetchKeyPair(ctx, ref.GetRef())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchKeyPair, err)
	}
	return kp, nil
}

func (p *secretProvider) StoreKeyPair(ctx context.Context, ref *pb.SecretRef, kp *KeyPair) error {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return err
	}
	if err := client.StoreKeyPair(ctx, ref.GetRef(), kp); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreKeyPair, err)
	}
	return nil
}

func (p *secretProvider) FetchCertificate(ctx context.Context, ref *pb.SecretRef) (*Certificate, error) {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return nil, err
	}
	cert, err := client.FetchCertificate(ctx, ref.GetRef())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchCertificate, err)
	}
	return cert, nil
}

func (p *secretProvider) StoreCertificate(ctx context.Context, ref *pb.SecretRef, cert *Certificate) error {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return err
	}
	if err := client.StoreCertificate(ctx, ref.GetRef(), cert); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreCertificate, err)
	}
	return nil
}

var _ SecretProvider = (*secretProvider)(nil)

// =============================================================================
// Helpers
// =============================================================================

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
