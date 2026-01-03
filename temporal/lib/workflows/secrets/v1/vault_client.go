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
	fieldCert        = "cert" // Alias for certificate (used in Config.Hydrate)
	fieldKey         = "key"  // Alias for private key (used in Config.Hydrate)
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
// Generic Secret Operations
// =============================================================================

func (c *vaultSecretClient) FetchSecret(ctx context.Context, ref *structpb.Struct) (Secret, error) {
	path, err := extractVaultPath(ref)
	if err != nil {
		return nil, err
	}

	secretData, err := c.api.FetchSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	// Return as Secret type
	return Secret(secretData), nil
}

func (c *vaultSecretClient) StoreSecret(ctx context.Context, ref *structpb.Struct, secret Secret) error {
	path, err := extractVaultPath(ref)
	if err != nil {
		return err
	}

	body := map[string]any{
		vaultFieldData: secret,
	}
	return c.api.Write(ctx, buildKVPath(path), body)
}

var _ SecretClient = (*vaultSecretClient)(nil)
