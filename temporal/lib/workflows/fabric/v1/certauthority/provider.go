package certauthority

import (
	"fmt"

	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
)

// =============================================================================
// BootstrapConfig - For BootstrapClient (username/password auth)
// =============================================================================

// BootstrapConfig wraps CertificateAuthorityConfig and implements
// common.ConfigHydrator[BootstrapClientConfig] for creating BootstrapClients.
//
// This config hydrates from the VaultAdminCredentialsPath which contains
// the bootstrap username/password.
type BootstrapConfig struct {
	fabricconfig.CertificateAuthorityConfig
}

// GetVaultPath returns where to find CA admin credentials (username/password) in Vault.
// This path is unique per CA and is used as the cache key if caching is enabled.
func (c BootstrapConfig) GetVaultPath() string {
	return c.VaultAdminCredentialsPath()
}

// Hydrate injects Vault secret data (username/password) into a BootstrapClientConfig
func (c BootstrapConfig) Hydrate(secretData map[string]interface{}) (BootstrapClientConfig, error) {
	username, ok := secretData["username"].(string)
	if !ok {
		return BootstrapClientConfig{}, fmt.Errorf("missing or invalid username in secret")
	}

	password, ok := secretData["password"].(string)
	if !ok {
		return BootstrapClientConfig{}, fmt.Errorf("missing or invalid password in secret")
	}

	// Build CA URL from config
	caURL := fmt.Sprintf("https://%s.%s.svc.cluster.local:7054",
		c.ComponentName(),
		c.KubernetesNamespace())

	return BootstrapClientConfig{
		URL:      caURL,
		Username: username,
		Password: password,
		CAName:   c.ComponentName(),
	}, nil
}

// NewBootstrapConfig wraps a CertificateAuthorityConfig for BootstrapClient creation.
func NewBootstrapConfig(certAuthConfig fabricconfig.CertificateAuthorityConfig) BootstrapConfig {
	return BootstrapConfig{CertificateAuthorityConfig: certAuthConfig}
}

// =============================================================================
// Config - For Client (keypair/cert auth)
// =============================================================================

// Config wraps CertificateAuthorityConfig and implements
// common.ConfigHydrator[ClientConfig] for creating Clients.
//
// This config hydrates from VaultKeyPairPath which contains the admin's
// certificate and private key (stored after initial enrollment).
type Config struct {
	fabricconfig.CertificateAuthorityConfig
}

// GetVaultPath returns where to find admin keypair/cert in Vault.
// This is the path where the admin's certificate and key are stored after enrollment.
func (c Config) GetVaultPath() string {
	// Use the keypair path - this is where we store the admin's cert+key after enrollment
	return c.VaultKeyPairPath()
}

// Hydrate injects Vault secret data (cert/key) into a ClientConfig
func (c Config) Hydrate(secretData map[string]interface{}) (ClientConfig, error) {
	// The keypair secret should contain both the certificate and private key
	certPEM, ok := secretData["cert"].(string)
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing or invalid cert in secret")
	}

	keyPEM, ok := secretData["key"].(string)
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing or invalid key in secret")
	}

	// Build CA URL from config
	caURL := fmt.Sprintf("https://%s.%s.svc.cluster.local:7054",
		c.ComponentName(),
		c.KubernetesNamespace())

	return ClientConfig{
		URL:       caURL,
		CAName:    c.ComponentName(),
		AdminCert: []byte(certPEM),
		AdminKey:  []byte(keyPEM),
	}, nil
}

// NewConfig wraps a CertificateAuthorityConfig for Client creation.
func NewConfig(certAuthConfig fabricconfig.CertificateAuthorityConfig) Config {
	return Config{CertificateAuthorityConfig: certAuthConfig}
}
