package certauthority

import (
	"fmt"

	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
)

// Config is an alias type that wraps CertificateAuthorityConfig and implements
// the common.ConfigHydrator interface for Fabric CA client creation.
// This allows CertificateAuthorityConfig to remain focused on its core responsibilities while
// the certauthority package handles CA-specific configuration hydration.
//
// Config implements common.ConfigHydrator[ClientConfig], which means:
// - GetVaultPath() returns the unique Vault path for this CA's credentials
// - Hydrate() creates a ClientConfig from Vault secret data
//
// The GetVaultPath() also serves as the cache key when used with ClientProviderFactory.
type Config struct {
	fabricconfig.CertificateAuthorityConfig
}

// GetVaultPath returns where to find CA admin credentials in Vault.
// This path is unique per CA and is used as the cache key by ClientProviderFactory.
func (c Config) GetVaultPath() string {
	return c.VaultAdminCredentialsPath()
}

// Hydrate injects Vault secret data (username/password) into a ClientConfig
func (c Config) Hydrate(secretData map[string]interface{}) (ClientConfig, error) {
	username, ok := secretData["username"].(string)
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing or invalid username in secret")
	}

	password, ok := secretData["password"].(string)
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing or invalid password in secret")
	}

	// Build CA URL from config
	// Format: https://<component-name>.<namespace>.svc.cluster.local:7054
	caURL := fmt.Sprintf("https://%s.%s.svc.cluster.local:7054",
		c.ComponentName(),
		c.KubernetesNamespace())

	return ClientConfig{
		URL:      caURL,
		Username: username,
		Password: password,
		CAName:   c.ComponentName(), // Use component name as CA name
	}, nil
}

// NewConfig wraps a CertificateAuthorityConfig for use with ClientProviderFactory.
// The returned Config implements common.ConfigHydrator[ClientConfig].
func NewConfig(certAuthConfig fabricconfig.CertificateAuthorityConfig) Config {
	return Config{CertificateAuthorityConfig: certAuthConfig}
}
