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
// This config hydrates from the VaultBootstrapAdminCredentialsRef which contains
// the bootstrap username/password.
type BootstrapConfig struct {
	fabricconfig.CertificateAuthorityConfig
}

// GetSecretRefs returns a map of named secret references
func (c BootstrapConfig) GetSecretRefs() map[string]*pb.SecretRef {
	return map[string]*pb.SecretRef{
		"credentials": c.VaultBootstrapAdminCredentialsRef(),
		"server_tls":  c.VaultTLSRef(), // CA's TLS cert for verifying server
	}
}

// GetCacheKey returns a unique key for caching the bootstrap client
func (c BootstrapConfig) GetCacheKey() string {
	// Use the Vault path as the cache key
	return c.VaultBootstrapAdminCredentialsRef().GetRef().Fields["path"].GetStringValue()
}

// Hydrate injects secret data (username/password) into a BootstrapClientConfig
func (c BootstrapConfig) Hydrate(secretsMap map[string]secrets.Secret) (BootstrapClientConfig, error) {
	credsSecret, ok := secretsMap["credentials"]
	if !ok {
		return BootstrapClientConfig{}, fmt.Errorf("missing credentials secret")
	}

	creds, err := credsSecret.Credentials()
	if err != nil {
		return BootstrapClientConfig{}, fmt.Errorf("failed to extract credentials: %w", err)
	}

	// Get TLS cert for verifying CA server
	var tlsCertPEM []byte
	if serverTLSSecret, ok := secretsMap["server_tls"]; ok {
		serverIdentity, err := serverTLSSecret.IdentityCredentials()
		if err == nil {
			tlsCertPEM = serverIdentity.Certificate
		}
	}

	// Build CA URL from config
	caURL := fmt.Sprintf("https://%s.%s.svc.cluster.local:7054",
		c.ComponentName(),
		c.KubernetesNamespace())

	return BootstrapClientConfig{
		URL:        caURL,
		Username:   creds.Username,
		Password:   creds.Password,
		CAName:     c.ComponentName(),
		TLSCertPEM: tlsCertPEM,
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
// The admin credentials can be specified explicitly via AdminCredentialsRef,
// or if nil, defaults to VaultAdminSigningRef() (the permanent admin's signing identity).
type Config struct {
	fabricconfig.CertificateAuthorityConfig
	AdminCredentialsRef *pb.SecretRef
}

// GetSecretRefs returns a map of named secret references
func (c Config) GetSecretRefs() map[string]*pb.SecretRef {
	return map[string]*pb.SecretRef{
		"identity":   c.getAdminRef(),
		"server_tls": c.VaultTLSRef(), // CA's TLS cert for verifying server
	}
}

// getAdminRef returns the admin credentials ref to use
// If AdminCredentialsRef is provided, use it; otherwise use the default permanent admin signing identity
func (c Config) getAdminRef() *pb.SecretRef {
	if c.AdminCredentialsRef != nil {
		return c.AdminCredentialsRef
	}
	return c.VaultAdminSigningRef()
}

// GetCacheKey returns a unique key for caching the CA admin client
func (c Config) GetCacheKey() string {
	// Use the Vault path as the cache key
	return c.getAdminRef().GetRef().Fields["path"].GetStringValue()
}

// Hydrate injects secret data (cert/key) into a ClientConfig
func (c Config) Hydrate(secretsMap map[string]secrets.Secret) (ClientConfig, error) {
	identitySecret, ok := secretsMap["identity"]
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing identity secret")
	}

	identity, err := identitySecret.IdentityCredentials()
	if err != nil {
		return ClientConfig{}, fmt.Errorf("failed to extract identity credentials: %w", err)
	}

	// Fetch the CA server's TLS certificate for verifying the server
	serverTLSSecret, ok := secretsMap["server_tls"]
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing server_tls secret")
	}

	serverIdentity, err := serverTLSSecret.IdentityCredentials()
	if err != nil {
		return ClientConfig{}, fmt.Errorf("failed to extract server TLS credentials: %w", err)
	}

	// Build CA URL from config
	caURL := fmt.Sprintf("https://%s.%s.svc.cluster.local:7054",
		c.ComponentName(),
		c.KubernetesNamespace())

	return ClientConfig{
		URL:        caURL,
		CAName:     c.ComponentName(),
		AdminCert:  identity.Certificate,
		AdminKey:   identity.PrivateKey,
		TLSCertPEM: serverIdentity.Certificate,
	}, nil
}

// NewConfig creates a config for Client creation with explicit admin credentials.
// If adminRef is nil, uses the default permanent admin credentials at VaultAdminSigningRef().
func NewConfig(certAuthConfig fabricconfig.CertificateAuthorityConfig, adminRef *pb.SecretRef) Config {
	return Config{
		CertificateAuthorityConfig: certAuthConfig,
		AdminCredentialsRef:        adminRef,
	}
}
