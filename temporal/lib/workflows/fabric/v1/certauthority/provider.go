package certauthority

import (
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
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

// GetSecretRefs returns a map of named secret references
func (c BootstrapConfig) GetSecretRefs() map[string]*pb.SecretRef {
	return map[string]*pb.SecretRef{
		"credentials": c.VaultBootstrapAdminCredentialsRef(),
	}
}

// GetCacheKey returns a unique key for caching the bootstrap client
func (c BootstrapConfig) GetCacheKey() string {
	// Use the Vault path as the cache key
	return c.VaultBootstrapAdminCredentialsRef().GetRef().Fields["path"].GetStringValue()
}

// Hydrate injects secret data (username/password) into a BootstrapClientConfig
func (c BootstrapConfig) Hydrate(secretsMap map[string]secrets.Secret) (BootstrapClientConfig, error) {
	creds, ok := secretsMap["credentials"]
	if !ok {
		return BootstrapClientConfig{}, fmt.Errorf("missing credentials secret")
	}

	username, ok := creds["username"].(string)
	if !ok {
		return BootstrapClientConfig{}, fmt.Errorf("missing or invalid username in secret")
	}

	password, ok := creds["password"].(string)
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

// GetSecretRefs returns a map of named secret references
func (c Config) GetSecretRefs() map[string]*pb.SecretRef {
	return map[string]*pb.SecretRef{
		"identity": c.VaultAdminKeyPairRef(),
	}
}

// GetCacheKey returns a unique key for caching the CA admin client
func (c Config) GetCacheKey() string {
	// Use the Vault path as the cache key
	return c.VaultAdminKeyPairRef().GetRef().Fields["path"].GetStringValue()
}

// Hydrate injects secret data (cert/key) into a ClientConfig
func (c Config) Hydrate(secretsMap map[string]secrets.Secret) (ClientConfig, error) {
	identity, ok := secretsMap["identity"]
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing identity secret")
	}

	// The identity secret should contain both the certificate and private key
	certPEM, ok := identity["certificate"].(string)
	if !ok {
		// Try "cert" as alternative
		certPEM, ok = identity["cert"].(string)
		if !ok {
			return ClientConfig{}, fmt.Errorf("missing or invalid certificate in secret")
		}
	}

	keyPEM, ok := identity["private_key"].(string)
	if !ok {
		// Try "key" as alternative
		keyPEM, ok = identity["key"].(string)
		if !ok {
			return ClientConfig{}, fmt.Errorf("missing or invalid key in secret")
		}
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

// =============================================================================
// ConfigWithAdminRef - For Client with custom admin credentials
// =============================================================================

// ConfigWithAdminRef wraps CertificateAuthorityConfig and allows specifying
// custom admin credentials (e.g., bootstrap admin during CA setup).
// This is used when the permanent admin credentials don't exist yet.
type ConfigWithAdminRef struct {
	fabricconfig.CertificateAuthorityConfig
	AdminCredentialsRef *pb.SecretRef
}

// GetSecretRefs returns a map of named secret references
func (c ConfigWithAdminRef) GetSecretRefs() map[string]*pb.SecretRef {
	return map[string]*pb.SecretRef{
		"identity": c.AdminCredentialsRef,
	}
}

// GetCacheKey returns a unique key for caching the client
func (c ConfigWithAdminRef) GetCacheKey() string {
	// Use the Vault path as the cache key
	return c.AdminCredentialsRef.GetRef().Fields["path"].GetStringValue()
}

// Hydrate injects secret data (cert/key) into a ClientConfig
func (c ConfigWithAdminRef) Hydrate(secretsMap map[string]secrets.Secret) (ClientConfig, error) {
	identity, ok := secretsMap["identity"]
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing identity secret")
	}

	// The identity secret should contain both the certificate and private key
	certPEM, ok := identity["certificate"].(string)
	if !ok {
		// Try "cert" as alternative
		certPEM, ok = identity["cert"].(string)
		if !ok {
			return ClientConfig{}, fmt.Errorf("missing or invalid certificate in secret")
		}
	}

	keyPEM, ok := identity["private_key"].(string)
	if !ok {
		// Try "key" as alternative
		keyPEM, ok = identity["key"].(string)
		if !ok {
			return ClientConfig{}, fmt.Errorf("missing or invalid key in secret")
		}
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

// NewConfigWithAdminRef creates a config that uses specific admin credentials.
func NewConfigWithAdminRef(certAuthConfig fabricconfig.CertificateAuthorityConfig, adminRef *pb.SecretRef) ConfigWithAdminRef {
	return ConfigWithAdminRef{
		CertificateAuthorityConfig: certAuthConfig,
		AdminCredentialsRef:        adminRef,
	}
}
