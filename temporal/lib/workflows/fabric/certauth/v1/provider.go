package certauth

import (
	"fmt"

	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
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
func (c BootstrapConfig) GetSecretRefs() map[string]*secretspb.SecretRef {
	return map[string]*secretspb.SecretRef{
		"credentials": c.VaultBootstrapAdminCredentialsRef(),
		"server_tls":  c.VaultTLSCertRef(), // CA's TLS cert for verifying server
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
	// The server_tls path is a certs/ path containing {certificate, chain}
	var tlsCertPEM []byte
	if serverTLSSecret, ok := secretsMap["server_tls"]; ok {
		serverCert, err := serverTLSSecret.Certificate()
		if err == nil {
			tlsCertPEM = serverCert.Certificate
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
// The admin credentials require both a private key (for signing) and certificate (for identity).
// Per cert-design.md, keys and certs are stored at separate paths:
//   - keys/... contains {private_key, public_key}
//   - certs/... contains {certificate, chain}
//
// AdminIdentity contains both the key and cert refs for the admin identity.
type Config struct {
	fabricconfig.CertificateAuthorityConfig
	AdminIdentity *fabricpb.IdentityRef // Admin's key + cert for CA authentication
}

// GetSecretRefs returns a map of named secret references
func (c Config) GetSecretRefs() map[string]*secretspb.SecretRef {
	adminIdentity := c.getAdminIdentity()

	return map[string]*secretspb.SecretRef{
		"identity_key":  adminIdentity.GetKeyRef(),
		"identity_cert": adminIdentity.GetCertRef(),
		"server_tls":    c.VaultTLSCertRef(), // CA's TLS cert for verifying server
	}
}

// getAdminIdentity returns the admin identity to use
// If AdminIdentity is provided, use it; otherwise use the default permanent admin signing identity
func (c Config) getAdminIdentity() *fabricpb.IdentityRef {
	if c.AdminIdentity != nil {
		return c.AdminIdentity
	}
	// Default to permanent admin identity
	return &fabricpb.IdentityRef{
		KeyRef:  c.VaultAdminSigningKeyRef(),
		CertRef: c.VaultAdminSigningCertRef(),
	}
}

// GetCacheKey returns a unique key for caching the CA admin client
func (c Config) GetCacheKey() string {
	// Use the cert path as the cache key (unique per admin identity)
	return c.getAdminIdentity().GetCertRef().GetRef().Fields["path"].GetStringValue()
}

// Hydrate injects secret data (cert/key) into a ClientConfig
// Combines key from keys/ path and cert from certs/ path
func (c Config) Hydrate(secretsMap map[string]secrets.Secret) (ClientConfig, error) {
	// Get private key from keys/ path
	keySecret, ok := secretsMap["identity_key"]
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing identity_key secret")
	}

	keyPair, err := keySecret.KeyPair()
	if err != nil {
		return ClientConfig{}, fmt.Errorf("failed to extract key pair: %w", err)
	}

	// Get certificate from certs/ path
	certSecret, ok := secretsMap["identity_cert"]
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing identity_cert secret")
	}

	cert, err := certSecret.Certificate()
	if err != nil {
		return ClientConfig{}, fmt.Errorf("failed to extract certificate: %w", err)
	}

	// Fetch the CA server's TLS certificate for verifying the server
	serverTLSSecret, ok := secretsMap["server_tls"]
	if !ok {
		return ClientConfig{}, fmt.Errorf("missing server_tls secret")
	}

	serverCert, err := serverTLSSecret.Certificate()
	if err != nil {
		return ClientConfig{}, fmt.Errorf("failed to extract server TLS certificate: %w", err)
	}

	// Build CA URL from config
	caURL := fmt.Sprintf("https://%s.%s.svc.cluster.local:7054",
		c.ComponentName(),
		c.KubernetesNamespace())

	return ClientConfig{
		URL:        caURL,
		CAName:     c.ComponentName(),
		AdminCert:  cert.Certificate,
		AdminKey:   keyPair.PrivateKey,
		TLSCertPEM: serverCert.Certificate,
	}, nil
}

// NewConfig creates a config for Client creation with explicit admin identity.
// If adminIdentity is nil, uses the default permanent admin identity.
func NewConfig(certAuthConfig fabricconfig.CertificateAuthorityConfig, adminIdentity *fabricpb.IdentityRef) Config {
	return Config{
		CertificateAuthorityConfig: certAuthConfig,
		AdminIdentity:              adminIdentity,
	}
}
