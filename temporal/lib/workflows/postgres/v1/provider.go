package postgres

import (
	"fmt"

	"github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
)

// NewPostgresClientProvider creates a ClientProvider for postgres with Vault-backed credentials
func NewPostgresClientProvider(providerConfig ProviderConfig, vaultClient vault.VaultClient) *common.ClientProvider[PostgresClient, PostgresConfig] {
	// Create and return provider
	// vault.VaultClient directly implements common.VaultClient interface
	return common.NewClientProvider(
		providerConfig,
		vaultClient,
		NewPostgresClient,
	)
}

// ProviderConfig holds the configuration needed to create a postgres client provider.
// It implements common.ConfigHydrator and knows WHERE to find secrets and HOW to inject them.
type ProviderConfig struct {
	Host      string // PostgreSQL host (e.g., "fabric-postgres-org1.fabric-org1.svc.cluster.local")
	Port      int    // PostgreSQL port (e.g., 5432)
	Database  string // PostgreSQL database to connect to (e.g., "postgres")
	VaultPath string // Vault path to credentials (e.g., "secret/fabric/org1/database/cert-store/credentials")
}

// GetVaultPath returns where to find credentials in Vault
func (c ProviderConfig) GetVaultPath() string {
	return c.VaultPath
}

// Hydrate injects secret data into a PostgresConfig
func (c ProviderConfig) Hydrate(secretData map[string]interface{}) (PostgresConfig, error) {
	username, ok := secretData["username"].(string)
	if !ok {
		return PostgresConfig{}, fmt.Errorf("missing or invalid username in secret")
	}

	password, ok := secretData["password"].(string)
	if !ok {
		return PostgresConfig{}, fmt.Errorf("missing or invalid password in secret")
	}

	return PostgresConfig{
		Host:     c.Host,
		Port:     c.Port,
		User:     username,
		Password: password,
		Database: c.Database,
	}, nil
}

// NewProviderConfig creates a ProviderConfig for the given organization
func NewProviderConfig(orgConfig *config.OrgConfig) ProviderConfig {
	org := orgConfig.OrgID
	return ProviderConfig{
		Host:      fmt.Sprintf("fabric-postgres-%s.fabric-%s.svc.cluster.local", org, org),
		Port:      5432,
		Database:  "postgres",
		VaultPath: fmt.Sprintf("secret/fabric/%s/database/cert-store/credentials", org),
	}
}
