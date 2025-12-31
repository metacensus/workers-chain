package postgres

import (
	"fmt"

	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
)

// NewPostgresClientProvider creates a ClientProvider for postgres with Vault-backed credentials
func NewPostgresClientProvider(config CertStorePostgresConfig, vaultClient vault.VaultClient) *common.ClientProvider[PostgresClient, PostgresConfig] {
	return common.NewClientProvider(
		config,
		vaultClient,
		NewPostgresClient,
	)
}

// CertStorePostgresConfig is an alias type that wraps CertificateStoreConfig and implements
// the common.ConfigHydrator interface for PostgreSQL client creation.
// This allows CertificateStoreConfig to remain focused on its core responsibilities while
// the postgres package handles postgres-specific configuration hydration.
type CertStorePostgresConfig struct {
	fabricconfig.CertificateStoreConfig
}

// GetVaultPath returns where to find PostgreSQL admin credentials in Vault
func (c CertStorePostgresConfig) GetVaultPath() string {
	return c.VaultAdminCredentialsPath()
}

// Hydrate injects Vault secret data (username/password) into a PostgresConfig
func (c CertStorePostgresConfig) Hydrate(secretData map[string]interface{}) (PostgresConfig, error) {
	username, ok := secretData["username"].(string)
	if !ok {
		return PostgresConfig{}, fmt.Errorf("missing or invalid username in secret")
	}

	password, ok := secretData["password"].(string)
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
