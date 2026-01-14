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
		"credentials": c.VaultAdminRef(),
	}
}

// GetCacheKey returns a unique key for caching the postgres client
func (c CertStorePostgresConfig) GetCacheKey() string {
	// Use the Vault path as the cache key
	return c.VaultAdminRef().GetRef().Fields["path"].GetStringValue()
}

// Hydrate injects secret data (username/password) into a PostgresConfig
func (c CertStorePostgresConfig) Hydrate(secretsMap map[string]secrets.Secret) (PostgresConfig, error) {
	credsSecret, ok := secretsMap["credentials"]
	if !ok {
		return PostgresConfig{}, fmt.Errorf("missing credentials secret")
	}

	creds, err := credsSecret.Credentials()
	if err != nil {
		return PostgresConfig{}, fmt.Errorf("failed to extract credentials: %w", err)
	}

	return PostgresConfig{
		Host:     c.ServiceHost(),
		Port:     c.Port(),
		User:     creds.Username,
		Password: creds.Password,
		Database: c.AdminDatabase(),
	}, nil
}

// NewCertStorePostgresConfig wraps a CertificateStoreConfig for postgres client provider use
func NewCertStorePostgresConfig(certStoreConfig fabricconfig.CertificateStoreConfig) CertStorePostgresConfig {
	return CertStorePostgresConfig{CertificateStoreConfig: certStoreConfig}
}
