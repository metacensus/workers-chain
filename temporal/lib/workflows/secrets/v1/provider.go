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

// Secret is generic secret data (untyped key-value pairs)
type Secret map[string]any

// Credentials extracts username/password credentials from the secret
func (s Secret) Credentials() (*Credentials, error) {
	username, ok := s["username"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid username in secret")
	}
	password, ok := s["password"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid password in secret")
	}
	return &Credentials{Username: username, Password: password}, nil
}

// KeyPair extracts a keypair from the secret
func (s Secret) KeyPair() (*KeyPair, error) {
	privateKey, ok := s["private_key"].(string)
	if !ok {
		// Try "key" as alternative
		privateKey, ok = s["key"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid private_key in secret")
		}
	}
	publicKey, ok := s["public_key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid public_key in secret")
	}
	return &KeyPair{
		PrivateKey: []byte(privateKey),
		PublicKey:  []byte(publicKey),
	}, nil
}

// Certificate extracts a certificate from the secret
func (s Secret) Certificate() (*Certificate, error) {
	certificate, ok := s["certificate"].(string)
	if !ok {
		// Try "cert" as alternative
		certificate, ok = s["cert"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid certificate in secret")
		}
	}
	return &Certificate{Certificate: []byte(certificate)}, nil
}

// IdentityCredentials extracts identity credentials (keypair + certificate) from the secret
func (s Secret) IdentityCredentials() (*IdentityCredentials, error) {
	privateKey, ok := s["private_key"].(string)
	if !ok {
		// Try "key" as alternative
		privateKey, ok = s["key"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid private_key in secret")
		}
	}

	publicKey, ok := s["public_key"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid public_key in secret")
	}

	certificate, ok := s["certificate"].(string)
	if !ok {
		// Try "cert" as alternative
		certificate, ok = s["cert"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid certificate in secret")
		}
	}

	// CA chain is optional
	caChain := ""
	if val, ok := s["ca_chain"].(string); ok {
		caChain = val
	}

	return &IdentityCredentials{
		PrivateKey:  []byte(privateKey),
		PublicKey:   []byte(publicKey),
		Certificate: []byte(certificate),
		CAChain:     []byte(caChain),
	}, nil
}

// Sentinel errors
var (
	ErrUnknownSecretType        = errors.New("unknown secret type")
	ErrFetchSecret              = errors.New("failed to fetch secret")
	ErrStoreSecret              = errors.New("failed to store secret")
	ErrFetchCredentials         = errors.New("failed to fetch credentials")
	ErrStoreCredentials         = errors.New("failed to store credentials")
	ErrFetchKeyPair             = errors.New("failed to fetch key pair")
	ErrStoreKeyPair             = errors.New("failed to store key pair")
	ErrFetchCertificate         = errors.New("failed to fetch certificate")
	ErrStoreCertificate         = errors.New("failed to store certificate")
	ErrFetchIdentityCredentials = errors.New("failed to fetch identity credentials")
	ErrStoreIdentityCredentials = errors.New("failed to store identity credentials")
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

// IdentityCredentials represents a complete identity (keypair + certificate)
// This is used for storing admin credentials where both key and cert are needed together
type IdentityCredentials struct {
	PrivateKey  []byte
	PublicKey   []byte
	Certificate []byte
	CAChain     []byte // Optional CA certificate chain
}

// =============================================================================
// SecretClient Interface
// =============================================================================

// SecretClient provides low-level generic secret operations.
// Each backend (vault, k8s, etc.) implements this interface.
// This is the minimal interface that backends must implement.
type SecretClient interface {
	// FetchSecret fetches generic secret data from the backend
	FetchSecret(ctx context.Context, ref *structpb.Struct) (Secret, error)

	// StoreSecret stores generic secret data to the backend
	StoreSecret(ctx context.Context, ref *structpb.Struct, secret Secret) error
}

// =============================================================================
// SecretProvider
// =============================================================================

// SecretProvider provides both generic and typed secret operations.
// It dispatches to the appropriate backend client based on SecretRef.Type.
type SecretProvider interface {
	// Generic secret operations
	FetchSecret(ctx context.Context, ref *pb.SecretRef) (Secret, error)
	StoreSecret(ctx context.Context, ref *pb.SecretRef, secret Secret) error

	// Typed convenience methods for common secret types
	// These use FetchSecret/StoreSecret internally with type conversion

	// Credentials
	FetchCredentials(ctx context.Context, ref *pb.SecretRef) (*Credentials, error)
	StoreCredentials(ctx context.Context, ref *pb.SecretRef, creds *Credentials) error

	// KeyPairs
	FetchKeyPair(ctx context.Context, ref *pb.SecretRef) (*KeyPair, error)
	StoreKeyPair(ctx context.Context, ref *pb.SecretRef, kp *KeyPair) error

	// Certificates
	FetchCertificate(ctx context.Context, ref *pb.SecretRef) (*Certificate, error)
	StoreCertificate(ctx context.Context, ref *pb.SecretRef, cert *Certificate) error

	// Identity Credentials (keypair + certificate together)
	FetchIdentityCredentials(ctx context.Context, ref *pb.SecretRef) (*IdentityCredentials, error)
	StoreIdentityCredentials(ctx context.Context, ref *pb.SecretRef, creds *IdentityCredentials) error
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

// =============================================================================
// Generic Secret Operations
// =============================================================================

func (p *secretProvider) FetchSecret(ctx context.Context, ref *pb.SecretRef) (Secret, error) {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return nil, err
	}
	secret, err := client.FetchSecret(ctx, ref.GetRef())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchSecret, err)
	}
	return secret, nil
}

func (p *secretProvider) StoreSecret(ctx context.Context, ref *pb.SecretRef, secret Secret) error {
	client, err := p.getClient(ref.GetType())
	if err != nil {
		return err
	}
	if err := client.StoreSecret(ctx, ref.GetRef(), secret); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreSecret, err)
	}
	return nil
}

// =============================================================================
// Typed Convenience Methods
// =============================================================================

func (p *secretProvider) FetchCredentials(ctx context.Context, ref *pb.SecretRef) (*Credentials, error) {
	secret, err := p.FetchSecret(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchCredentials, err)
	}

	creds, err := secret.Credentials()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchCredentials, err)
	}
	return creds, nil
}

func (p *secretProvider) StoreCredentials(ctx context.Context, ref *pb.SecretRef, creds *Credentials) error {
	secret := Secret{
		"username": creds.Username,
		"password": creds.Password,
	}
	if err := p.StoreSecret(ctx, ref, secret); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreCredentials, err)
	}
	return nil
}

func (p *secretProvider) FetchKeyPair(ctx context.Context, ref *pb.SecretRef) (*KeyPair, error) {
	secret, err := p.FetchSecret(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchKeyPair, err)
	}

	kp, err := secret.KeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchKeyPair, err)
	}
	return kp, nil
}

func (p *secretProvider) StoreKeyPair(ctx context.Context, ref *pb.SecretRef, kp *KeyPair) error {
	secret := Secret{
		"private_key": string(kp.PrivateKey),
		"public_key":  string(kp.PublicKey),
	}
	if err := p.StoreSecret(ctx, ref, secret); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreKeyPair, err)
	}
	return nil
}

func (p *secretProvider) FetchCertificate(ctx context.Context, ref *pb.SecretRef) (*Certificate, error) {
	secret, err := p.FetchSecret(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchCertificate, err)
	}

	cert, err := secret.Certificate()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchCertificate, err)
	}
	return cert, nil
}

func (p *secretProvider) StoreCertificate(ctx context.Context, ref *pb.SecretRef, cert *Certificate) error {
	secret := Secret{
		"certificate": string(cert.Certificate),
	}
	if err := p.StoreSecret(ctx, ref, secret); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreCertificate, err)
	}
	return nil
}

func (p *secretProvider) FetchIdentityCredentials(ctx context.Context, ref *pb.SecretRef) (*IdentityCredentials, error) {
	secret, err := p.FetchSecret(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchIdentityCredentials, err)
	}

	idCreds, err := secret.IdentityCredentials()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFetchIdentityCredentials, err)
	}
	return idCreds, nil
}

func (p *secretProvider) StoreIdentityCredentials(ctx context.Context, ref *pb.SecretRef, creds *IdentityCredentials) error {
	secret := Secret{
		"private_key": string(creds.PrivateKey),
		"public_key":  string(creds.PublicKey),
		"certificate": string(creds.Certificate),
	}
	if len(creds.CAChain) > 0 {
		secret["ca_chain"] = string(creds.CAChain)
	}
	if err := p.StoreSecret(ctx, ref, secret); err != nil {
		return fmt.Errorf("%w: %w", ErrStoreIdentityCredentials, err)
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
