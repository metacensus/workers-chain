package postgres

import (
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
)

// NewPostgresClientProvider creates a ClientProvider for postgres with secret-backed credentials.
// Uses ClientProvider (single-client pattern) because there is exactly one postgres deployment.
func NewPostgresClientProvider(config CertStorePostgresConfig, secretProvider secrets.SecretProvider) *common.ClientProvider[PostgresClient, PostgresConfig] {
	return common.NewClientProvider(
		config,
		secretProvider,
		NewPostgresClient, // ClientConstructor
	)
}

// CertStorePostgresConfig is an alias type that wraps CertificateStoreConfig and implements
// the common.ConfigHydrator interface for PostgreSQL client creation.
// This allows CertificateStoreConfig to remain focused on its core responsibilities while
// the postgres package handles postgres-specific configuration hydration.
type CertStorePostgresConfig struct {
	fabricconfig.CertificateStoreConfig
}

// GetSecretRefs returns a map of named secret references
func (c CertStorePostgresConfig) GetSecretRefs() map[string]*pb.SecretRef {
	return map[string]*pb.SecretRef{
		"credentials": c.VaultAdminCredentialsRef(),
	}
}

// GetCacheKey returns a unique key for caching the postgres client
func (c CertStorePostgresConfig) GetCacheKey() string {
	// Use the Vault path as the cache key
	return c.VaultAdminCredentialsRef().GetRef().Fields["path"].GetStringValue()
}

// Hydrate injects secret data (username/password) into a PostgresConfig
func (c CertStorePostgresConfig) Hydrate(secretsMap map[string]secrets.Secret) (PostgresConfig, error) {
	creds, ok := secretsMap["credentials"]
	if !ok {
		return PostgresConfig{}, fmt.Errorf("missing credentials secret")
	}

	username, ok := creds["username"].(string)
	if !ok {
		return PostgresConfig{}, fmt.Errorf("missing or invalid username in secret")
	}

	password, ok := creds["password"].(string)
	if !ok {
		return PostgresConfig{}, fmt.Errorf("missing or invalid password in secret")
	}

	return PostgresConfig{
		Host:     c.ServiceHost(),
		Port:     c.Port(),
		User:     username,
		Password: password,
		Database: c.AdminDatabase(),
	}, nil
}

// NewCertStorePostgresConfig wraps a CertificateStoreConfig for postgres client provider use
func NewCertStorePostgresConfig(certStoreConfig fabricconfig.CertificateStoreConfig) CertStorePostgresConfig {
	return CertStorePostgresConfig{CertificateStoreConfig: certStoreConfig}
}
