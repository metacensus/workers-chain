package secrets

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

// VaultAPI defines the interface for Vault operations.
type VaultAPI interface {
	Write(ctx context.Context, path string, data map[string]any) error
	Read(ctx context.Context, path string) (map[string]any, error)
	Delete(ctx context.Context, path string) error
	FetchSecret(ctx context.Context, path string) (map[string]interface{}, error)
}

// Vault field names
const (
	vaultFieldData   = "data"
	fieldUsername    = "username"
	fieldPassword    = "password"
	fieldPrivateKey  = "private"
	fieldPublicKey   = "public"
	fieldCertificate = "certificate"
	fieldCert        = "cert"      // Alias for certificate (used in Config.Hydrate)
	fieldKey         = "key"       // Alias for private key (used in Config.Hydrate)
	fieldCAChain     = "ca_chain"
)

type vaultSecretClient struct {
	api VaultAPI
}

// NewVaultSecretClient creates a Vault-backed SecretClient
func NewVaultSecretClient(api VaultAPI) SecretClient {
	return &vaultSecretClient{api: api}
}

func buildKVPath(path string) string {
	return fmt.Sprintf("secret/data/%s", path)
}

// extractVaultPath extracts the "path" field from a Vault ref struct
func extractVaultPath(ref *structpb.Struct) (string, error) {
	if ref == nil {
		return "", fmt.Errorf("ref is nil")
	}
	pathValue, ok := ref.Fields["path"]
	if !ok {
		return "", fmt.Errorf("missing 'path' field in ref")
	}
	path := pathValue.GetStringValue()
	if path == "" {
		return "", fmt.Errorf("'path' field is empty")
	}
	return path, nil
}

// =============================================================================
// Credentials
// =============================================================================

func (c *vaultSecretClient) FetchCredentials(ctx context.Context, ref *structpb.Struct) (*Credentials, error) {
	path, err := extractVaultPath(ref)
	if err != nil {
		return nil, err
	}

	secretData, err := c.api.FetchSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	username, ok := secretData[fieldUsername].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid %s in secret", fieldUsername)
	}

	password, ok := secretData[fieldPassword].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid %s in secret", fieldPassword)
	}

	return &Credentials{Username: username, Password: password}, nil
}

func (c *vaultSecretClient) StoreCredentials(ctx context.Context, ref *structpb.Struct, creds *Credentials) error {
	path, err := extractVaultPath(ref)
	if err != nil {
		return err
	}

	body := map[string]any{
		vaultFieldData: map[string]string{
			fieldUsername: creds.Username,
			fieldPassword: creds.Password,
		},
	}
	return c.api.Write(ctx, buildKVPath(path), body)
}

// =============================================================================
// KeyPairs
// =============================================================================

func (c *vaultSecretClient) FetchKeyPair(ctx context.Context, ref *structpb.Struct) (*KeyPair, error) {
	path, err := extractVaultPath(ref)
	if err != nil {
		return nil, err
	}

	secretData, err := c.api.FetchSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	privateKey, ok := secretData[fieldPrivateKey].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid %s in secret", fieldPrivateKey)
	}

	publicKey, ok := secretData[fieldPublicKey].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid %s in secret", fieldPublicKey)
	}

	return &KeyPair{
		PrivateKey: []byte(privateKey),
		PublicKey:  []byte(publicKey),
	}, nil
}

func (c *vaultSecretClient) StoreKeyPair(ctx context.Context, ref *structpb.Struct, kp *KeyPair) error {
	path, err := extractVaultPath(ref)
	if err != nil {
		return err
	}

	body := map[string]any{
		vaultFieldData: map[string]string{
			fieldPrivateKey: string(kp.PrivateKey),
			fieldPublicKey:  string(kp.PublicKey),
		},
	}
	return c.api.Write(ctx, buildKVPath(path), body)
}

// =============================================================================
// Certificates
// =============================================================================

func (c *vaultSecretClient) FetchCertificate(ctx context.Context, ref *structpb.Struct) (*Certificate, error) {
	path, err := extractVaultPath(ref)
	if err != nil {
		return nil, err
	}

	secretData, err := c.api.FetchSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	cert, ok := secretData[fieldCertificate].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid %s in secret", fieldCertificate)
	}

	return &Certificate{Certificate: []byte(cert)}, nil
}

func (c *vaultSecretClient) StoreCertificate(ctx context.Context, ref *structpb.Struct, cert *Certificate) error {
	path, err := extractVaultPath(ref)
	if err != nil {
		return err
	}

	body := map[string]any{
		vaultFieldData: map[string]string{
			fieldCertificate: string(cert.Certificate),
		},
	}
	return c.api.Write(ctx, buildKVPath(path), body)
}

func (c *vaultSecretClient) FetchIdentityCredentials(ctx context.Context, ref *structpb.Struct) (*IdentityCredentials, error) {
	path, err := extractVaultPath(ref)
	if err != nil {
		return nil, err
	}

	secret, err := c.api.FetchSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fetch secret: %w", err)
	}

	// Support both field naming conventions (for compatibility with Config.Hydrate)
	privateKey, ok := secret[fieldKey].(string)
	if !ok {
		privateKey, ok = secret[fieldPrivateKey].(string)
		if !ok {
			return nil, fmt.Errorf("missing private key field")
		}
	}

	publicKey, _ := secret[fieldPublicKey].(string)

	cert, ok := secret[fieldCert].(string)
	if !ok {
		cert, ok = secret[fieldCertificate].(string)
		if !ok {
			return nil, fmt.Errorf("missing certificate field")
		}
	}

	caChain, _ := secret[fieldCAChain].(string)

	return &IdentityCredentials{
		PrivateKey:  []byte(privateKey),
		PublicKey:   []byte(publicKey),
		Certificate: []byte(cert),
		CAChain:     []byte(caChain),
	}, nil
}

func (c *vaultSecretClient) StoreIdentityCredentials(ctx context.Context, ref *structpb.Struct, creds *IdentityCredentials) error {
	path, err := extractVaultPath(ref)
	if err != nil {
		return err
	}

	// Store with both field naming conventions for compatibility
	// Use "key" and "cert" for Config.Hydrate, and also store full names
	data := map[string]string{
		fieldKey:         string(creds.PrivateKey),
		fieldPrivateKey:  string(creds.PrivateKey),
		fieldPublicKey:   string(creds.PublicKey),
		fieldCert:        string(creds.Certificate),
		fieldCertificate: string(creds.Certificate),
	}

	if len(creds.CAChain) > 0 {
		data[fieldCAChain] = string(creds.CAChain)
	}

	body := map[string]any{
		vaultFieldData: data,
	}
	return c.api.Write(ctx, buildKVPath(path), body)
}

var _ SecretClient = (*vaultSecretClient)(nil)
