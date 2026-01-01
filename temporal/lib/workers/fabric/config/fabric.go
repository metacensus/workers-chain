package config

import (
	"fmt"

	kubev1 "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
)

// Vault KV v2 path prefixes for policy construction
const (
	vaultSecretPrefix         = "secret"
	vaultSecretDataPrefix     = "secret/data"
	vaultSecretMetadataPrefix = "secret/metadata"
)

// FabricConfig provides access to all Fabric-related configuration.
// It acts as a factory for component-specific configs.
type FabricConfig struct {
	orgConfig config.OrgConfig
}

// NewFabricConfig creates a FabricConfig from an OrgConfig.
func NewFabricConfig(orgConfig config.OrgConfig) FabricConfig {
	return FabricConfig{orgConfig: orgConfig}
}

// CertificateStore returns the configuration for the certificate store (shared PostgreSQL).
func (fc FabricConfig) CertificateStore() CertificateStoreConfig {
	return CertificateStoreConfig{OrgConfig: fc.orgConfig}
}

// CertificateAuthority returns the configuration for a specific certificate authority.
// certType should be "tls" or "org".
// caID should be "root" for root CAs, or a version ID (e.g., "v1") for intermediates.
func (fc FabricConfig) CertificateAuthority(certType, caID string) CertificateAuthorityConfig {
	return CertificateAuthorityConfig{
		OrgConfig: fc.orgConfig,
		certType:  certType,
		caID:      caID,
	}
}

// ============= Certificate Store Configuration =============

// CertificateStoreConfig encapsulates all configuration for the certificate store (PostgreSQL).
// This is the shared database used by all certificate authorities in the organization.
// It embeds OrgConfig so it can be used anywhere an OrgConfig is needed.
type CertificateStoreConfig struct {
	config.OrgConfig
}

// ------------ Kubernetes Properties ------------

// ComponentName returns the base component name used in Helm charts and derived identifiers.
func (cs CertificateStoreConfig) ComponentName() string {
	return "certificate-store"
}

// ServiceAccountName returns the Kubernetes service account name.
func (cs CertificateStoreConfig) ServiceAccountName() string {
	return cs.ComponentName()
}

// KubernetesNamespace returns the Kubernetes namespace for deployment.
func (cs CertificateStoreConfig) KubernetesNamespace() string {
	return cs.OrgConfig.KubernetesNamespace()
}

// ServiceHost returns the DNS hostname for the certificate store service.
// Example: "certificate-store.fabric-org1.svc.cluster.local"
func (cs CertificateStoreConfig) ServiceHost() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", cs.ComponentName(), cs.KubernetesNamespace())
}

// Port returns the PostgreSQL port for the certificate store.
func (cs CertificateStoreConfig) Port() int {
	return 5432
}

// AdminDatabase returns the admin database name used for connecting to PostgreSQL
// to perform administrative operations (creating databases, users, etc.).
// PostgreSQL requires connecting to an existing database; "postgres" is the default.
func (cs CertificateStoreConfig) AdminDatabase() string {
	return "postgres"
}

// ResourceIdentifier returns the Kubernetes resource identifier for waiting on deployment readiness.
// The certificate store is deployed as a StatefulSet.
func (cs CertificateStoreConfig) ResourceIdentifier() *kubev1.ResourceIdentifier {
	return &kubev1.ResourceIdentifier{
		ApiVersion: "apps/v1",
		Kind:       "StatefulSet",
		Name:       cs.ComponentName(),
		Namespace:  cs.KubernetesNamespace(),
	}
}

// ------------ Vault Properties ------------

// PolicyName returns the Vault policy name.
// Example: "fabric-org1-certificate-store"
func (cs CertificateStoreConfig) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", cs.OrgID, cs.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (cs CertificateStoreConfig) AuthRoleName() string {
	return cs.PolicyName()
}

// VaultAdminCredentialsPath returns the path where postgres admin credentials are stored.
// Used with vault.StoreSecret and vault.FetchSecret activities.
// Note: Returns path relative to secret mount (activities add "secret/data/" prefix).
// Example: "fabric/org1/certstore/admin/credentials"
func (cs CertificateStoreConfig) VaultAdminCredentialsPath() string {
	return fmt.Sprintf("fabric/%s/certstore/admin/credentials", cs.OrgID)
}

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Used in Vault policy documents with wildcard for all certificate store secrets.
// Example: "secret/data/fabric/org1/certstore/*"
func (cs CertificateStoreConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/fabric/%s/certstore/*", vaultSecretDataPrefix, cs.OrgID)
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/certstore/*"
func (cs CertificateStoreConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/fabric/%s/certstore/*", vaultSecretMetadataPrefix, cs.OrgID)
}

// VaultPolicy creates a Vault policy document for the certificate store.
// The policy grants read access to the certificate store's secrets (postgres credentials).
func (cs CertificateStoreConfig) VaultPolicy() *vaultpb.PolicyDocument {
	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// KV v2 data path - read access to certificate store secrets
			{
				Path: cs.VaultDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// KV v2 metadata path - list and read metadata
			{
				Path: cs.VaultMetadataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
		},
	}
}

// ============= Certificate Authority Configuration =============

// CertificateAuthorityConfig encapsulates all configuration for a certificate authority.
// Each CA (root and intermediate) gets its own dedicated database and vault access.
// It embeds OrgConfig so it can be used anywhere an OrgConfig is needed.
type CertificateAuthorityConfig struct {
	config.OrgConfig
	certType string // "tls" or "org"
	caID     string // "root" for root CAs, or version ID (e.g., "v1") for intermediates
}

// ------------ Kubernetes Properties ------------

// ComponentName returns the base component name used in Helm charts and derived identifiers.
// Examples:
//   - Root TLS CA: "certificate-authority-tls-root"
//   - Intermediate Org CA v1: "certificate-authority-org-v1"
func (ca CertificateAuthorityConfig) ComponentName() string {
	return fmt.Sprintf("certificate-authority-%s-%s", ca.certType, ca.caID)
}

// ServiceAccountName returns the Kubernetes service account name.
func (ca CertificateAuthorityConfig) ServiceAccountName() string {
	return ca.ComponentName()
}

// KubernetesNamespace returns the Kubernetes namespace for deployment.
func (ca CertificateAuthorityConfig) KubernetesNamespace() string {
	return ca.OrgConfig.KubernetesNamespace()
}

// ------------ Vault Properties ------------

// PolicyName returns the Vault policy name.
// Example: "fabric-org1-certificate-authority-tls-root"
func (ca CertificateAuthorityConfig) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", ca.OrgID, ca.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (ca CertificateAuthorityConfig) AuthRoleName() string {
	return ca.PolicyName()
}

// VaultDatabaseCredentialsPath returns the path where database credentials are stored.
// Used with vault.StoreSecret and vault.FetchSecret activities.
// Note: Returns path relative to secret mount (activities add "secret/data/" prefix).
// Example: "fabric/org1/ca/tls/root/certstore/user/credentials"
func (ca CertificateAuthorityConfig) VaultDatabaseCredentialsPath() string {
	return fmt.Sprintf("fabric/%s/ca/%s/%s/certstore/user/credentials",
		ca.OrgID, ca.certType, ca.caID)
}

// VaultAdminCredentialsPath returns the path where CA admin/bootstrap credentials are stored.
// Used with vault.StoreSecret and vault.FetchSecret activities.
// Note: Returns path relative to secret mount (activities add "secret/data/" prefix).
// Note: these credentials are used to bootstrap the CA's initial admin user
// and then are immediately replaced with a new keyPair based admin.
// Example: "fabric/org1/ca/tls/root/admin/credentials"
func (ca CertificateAuthorityConfig) VaultAdminCredentialsPath() string {
	return fmt.Sprintf("fabric/%s/ca/%s/%s/admin/credentials",
		ca.OrgID, ca.certType, ca.caID)
}

// VaultKeyPairPath returns the path where the CA's key pair is stored.
// Used with vault.StoreSecret and vault.FetchSecret activities.
// Note: Returns path relative to secret mount (activities add "secret/data/" prefix).
// Example: "fabric/org1/ca/tls/root/keys"
func (ca CertificateAuthorityConfig) VaultKeyPairPath() string {
	return fmt.Sprintf("fabric/%s/ca/%s/%s/keys",
		ca.OrgID, ca.certType, ca.caID)
}

// VaultCertificatePath returns the path where the CA's certificate is stored.
// Used with vault.StoreSecret and vault.FetchSecret activities.
// Note: Returns path relative to secret mount (activities add "secret/data/" prefix).
// Example: "fabric/org1/ca/tls/root/cert"
func (ca CertificateAuthorityConfig) VaultCertificatePath() string {
	return fmt.Sprintf("fabric/%s/ca/%s/%s/cert",
		ca.OrgID, ca.certType, ca.caID)
}

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Used in Vault policy documents with wildcard for all CA secrets.
// Example: "secret/data/fabric/org1/ca/tls/root/*"
func (ca CertificateAuthorityConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/fabric/%s/ca/%s/%s/*",
		vaultSecretDataPrefix, ca.OrgID, ca.certType, ca.caID)
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/ca/tls/root/*"
func (ca CertificateAuthorityConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/fabric/%s/ca/%s/%s/*",
		vaultSecretMetadataPrefix, ca.OrgID, ca.certType, ca.caID)
}

// VaultPolicy creates a Vault policy document for the certificate authority.
// The policy grants read/write access to the CA's secrets path.
// Works for both root and intermediate CAs.
func (ca CertificateAuthorityConfig) VaultPolicy() *vaultpb.PolicyDocument {
	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// KV v2 data path - read/write access to CA secrets
			{
				Path: ca.VaultDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_CREATE,
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_UPDATE,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// KV v2 metadata path - list and read metadata
			{
				Path: ca.VaultMetadataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
		},
	}
}

// ------------ Database Properties ------------

// DatabaseName returns the PostgreSQL database name for this CA.
// Follows PostgreSQL naming conventions (lowercase, underscores, alphanumeric).
// Examples:
//   - Root TLS CA: "fabric_org1_ca_tls_root"
//   - Intermediate Org CA v1: "fabric_org1_ca_org_v1"
func (ca CertificateAuthorityConfig) DatabaseName() string {
	// Sanitize for PostgreSQL: lowercase, replace hyphens with underscores
	orgID := sanitizeForPostgres(ca.OrgID)
	certType := sanitizeForPostgres(ca.certType)
	caID := sanitizeForPostgres(ca.caID)
	return fmt.Sprintf("fabric_%s_ca_%s_%s", orgID, certType, caID)
}

// DatabaseUsername returns the PostgreSQL username for this CA's database.
// Each CA gets a dedicated user with access only to its database.
// Example: "fabric_org1_ca_tls_root_user"
func (ca CertificateAuthorityConfig) DatabaseUsername() string {
	return ca.DatabaseName() + "_user"
}

// Description returns a human-readable description of this CA.
// Useful for logging and debugging.
// Example: "TLS Root CA for org1"
func (ca CertificateAuthorityConfig) Description() string {
	caType := "Root"
	if ca.caID != "root" {
		caType = "Intermediate"
	}

	certTypeDesc := "TLS"
	if ca.certType == "org" {
		certTypeDesc = "Org"
	}

	return fmt.Sprintf("%s %s CA for %s", certTypeDesc, caType, ca.OrgID)
}

// CertType returns the certificate type ("tls" or "org")
func (ca CertificateAuthorityConfig) CertType() string {
	return ca.certType
}

// CAID returns the CA identifier ("root" for root CAs, "v1", "v2", etc. for intermediates)
func (ca CertificateAuthorityConfig) CAID() string {
	return ca.caID
}

// OrganizationalUnit returns the OU field for certificate subject.
// Returns "root-ca" for root CAs, or the CA version ID for intermediate CAs.
func (ca CertificateAuthorityConfig) OrganizationalUnit() string {
	if ca.caID == "root" {
		return "root-ca"
	}
	return ca.caID
}

// ------------ Helper Functions ------------

// sanitizeForPostgres converts a string to PostgreSQL-safe identifier format
func sanitizeForPostgres(s string) string {
	// PostgreSQL identifiers: lowercase, replace hyphens with underscores
	result := ""
	for _, c := range s {
		if c == '-' {
			result += "_"
		} else {
			result += string(c)
		}
	}
	return result
}
