package config

import (
	"fmt"

	kubev1 "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
)

// Vault KV v2 path prefixes for policy construction
const (
	vaultSecretPrefix         = "secret"
	vaultSecretDataPrefix     = "secret/data"
	vaultSecretMetadataPrefix = "secret/metadata"
)

// FabricConfig provides access to all Fabric-related configuration.
// It acts as a factory for component-specific configs.
// It embeds OrgConfig to provide direct access to org-level configuration.
type FabricConfig struct {
	config.OrgConfig
}

// NewFabricConfig creates a FabricConfig from an OrgConfig.
func NewFabricConfig(orgConfig config.OrgConfig) FabricConfig {
	return FabricConfig{OrgConfig: orgConfig}
}

// CertificateStore returns the configuration for the certificate store (shared PostgreSQL).
func (fc FabricConfig) CertificateStore() CertificateStoreConfig {
	return CertificateStoreConfig{OrgConfig: fc.OrgConfig}
}

// CertificateAuthority returns the configuration for a specific certificate authority.
// certType should be "tls" or "org".
// caID should be "root" for root CAs, or a version ID (e.g., "v1") for intermediates.
func (fc FabricConfig) CertificateAuthority(certType, caID string) CertificateAuthorityConfig {
	return CertificateAuthorityConfig{
		OrgConfig: fc.OrgConfig,
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

// VaultAdminRef returns a SecretRef for the certificate store admin identity.
// Stores username, password for postgres admin.
// Example path: "fabric/org1/certstore/admin"
func (cs CertificateStoreConfig) VaultAdminRef() *pb.SecretRef {
	path := fmt.Sprintf("fabric/%s/certstore/admin", cs.OrgID)
	return secrets.NewVaultRef(path)
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

// URL returns the Fabric CA server URL for client connections.
// Uses the Kubernetes service DNS name with the CA port (7054).
// Example: "https://certificate-authority-tls-root.fabric-org1.svc.cluster.local:7054"
func (ca CertificateAuthorityConfig) URL() string {
	return fmt.Sprintf("https://%s.%s.svc.cluster.local:7054",
		ca.ComponentName(), ca.KubernetesNamespace())
}

// ServiceAccountName returns the Kubernetes service account name.
func (ca CertificateAuthorityConfig) ServiceAccountName() string {
	return ca.ComponentName()
}

// KubernetesNamespace returns the Kubernetes namespace for deployment.
func (ca CertificateAuthorityConfig) KubernetesNamespace() string {
	return ca.OrgConfig.KubernetesNamespace()
}

// ResourceIdentifier returns the Kubernetes resource identifier for waiting on deployment readiness.
// The certificate authority is deployed as a Deployment.
func (ca CertificateAuthorityConfig) ResourceIdentifier() *kubev1.ResourceIdentifier {
	return &kubev1.ResourceIdentifier{
		ApiVersion: "apps/v1",
		Kind:       "Deployment",
		Name:       ca.ComponentName(),
		Namespace:  ca.KubernetesNamespace(),
	}
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

// VaultDatabaseUserRef returns a SecretRef for the CA's database user identity.
// Stores username, password for database access to the certificate store.
// Each CA has its own database user for accessing the certificate store.
// Example path: "fabric/org1/ca/tls/root/certstore/user"
func (ca CertificateAuthorityConfig) VaultDatabaseUserRef() *pb.SecretRef {
	path := fmt.Sprintf("fabric/%s/ca/%s/%s/certstore/user",
		ca.OrgID, ca.certType, ca.caID)
	return secrets.NewVaultRef(path)
}

// VaultServerRef returns a SecretRef for the CA server identity.
// Stores private_key, public_key, certificate for the CA's signing identity.
// Example path: "fabric/org1/ca/tls/root/server"
func (ca CertificateAuthorityConfig) VaultServerRef() *pb.SecretRef {
	path := fmt.Sprintf("fabric/%s/ca/%s/%s/server",
		ca.OrgID, ca.certType, ca.caID)
	return secrets.NewVaultRef(path)
}

// VaultBootstrapAdminRef returns a SecretRef for the CA bootstrap admin identity.
// Stores username, password, private_key, public_key, certificate.
// The bootstrap admin is created during CA deployment with username/password,
// enrolled to get keypair+cert, used to create permanent admin, then revoked.
// Example path: "fabric/org1/ca/tls/root/bootstrap-admin"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminRef() *pb.SecretRef {
	path := fmt.Sprintf("fabric/%s/ca/%s/%s/bootstrap-admin",
		ca.OrgID, ca.certType, ca.caID)
	return secrets.NewVaultRef(path)
}

// VaultAdminRef returns a SecretRef for the permanent admin identity.
// Stores private_key, public_key, certificate.
// This is the long-lived admin identity used for all CA administrative operations.
// This is distinct from VaultServerRef, which stores the CA's own signing key.
// Example path: "fabric/org1/ca/tls/root/admin"
func (ca CertificateAuthorityConfig) VaultAdminRef() *pb.SecretRef {
	path := fmt.Sprintf("fabric/%s/ca/%s/%s/admin",
		ca.OrgID, ca.certType, ca.caID)
	return secrets.NewVaultRef(path)
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

// AdminCommonName returns the common name for the CA admin certificate.
// Example: "admin@certificate-authority-tls-root"
func (ca CertificateAuthorityConfig) AdminCommonName() string {
	return fmt.Sprintf("admin@%s", ca.ComponentName())
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
