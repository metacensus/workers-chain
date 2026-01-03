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

// GetVaultPath returns where to find CA bootstrap admin credentials (username/password) in Vault.
// This path is unique per CA and is used as the cache key if caching is enabled.
func (c BootstrapConfig) GetVaultPath() string {
	return c.VaultBootstrapAdminCredentialsPath()
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
// This config hydrates from VaultAdminKeyPairPath which contains the permanent admin's
// private key, public key, and certificate (all stored together after CreateIdentity).
type Config struct {
	fabricconfig.CertificateAuthorityConfig
}

// GetVaultPath returns where to find admin keypair/cert in Vault.
// This is the path where the permanent admin's certificate and key are stored together.
func (c Config) GetVaultPath() string {
	// Use the admin keypair path - this has both key and cert together
	return c.VaultAdminKeyPairPath()
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
