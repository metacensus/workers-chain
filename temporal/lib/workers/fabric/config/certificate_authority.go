package config

import (
	"fmt"

	kubev1 "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
)

// Re-export OwnerConfig types for convenience
type OwnerConfig = workerconfig.OwnerConfig
type OperatorConfig = workerconfig.OperatorConfig
type OrgConfig = workerconfig.OrgConfig

var NewOperatorConfig = workerconfig.NewOperatorConfig
var NewOrgConfig = workerconfig.NewOrgConfig

// Vault KV v2 path prefixes for policy construction
const (
	vaultSecretPrefix         = "secret"
	vaultSecretDataPrefix     = "secret/data"
	vaultSecretMetadataPrefix = "secret/metadata"
)

// ============= Certificate Store Configuration =============

// CertificateStoreConfig encapsulates all configuration for the certificate store (PostgreSQL).
// This is the shared database used by all certificate authorities owned by this owner.
// It embeds OwnerConfig so methods like KubernetesNamespace() and VaultBasePath() are promoted.
type CertificateStoreConfig struct {
	OwnerConfig // embedded interface - methods promoted
}

// NewCertificateStoreConfig creates a CertificateStoreConfig for the given owner.
func NewCertificateStoreConfig(owner OwnerConfig) CertificateStoreConfig {
	return CertificateStoreConfig{OwnerConfig: owner}
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
	return fmt.Sprintf("fabric-%s-%s", cs.OwnerID(), cs.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (cs CertificateStoreConfig) AuthRoleName() string {
	return cs.PolicyName()
}

// VaultAdminRef returns a SecretRef for the certificate store admin identity.
// Stores username, password for postgres admin.
// Example path: "fabric/org1/certstore/admin"
func (cs CertificateStoreConfig) VaultAdminRef() *pb.SecretRef {
	path := fmt.Sprintf("%s/certstore/admin", cs.VaultBasePath())
	return secrets.NewVaultRef(path)
}

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Used in Vault policy documents with wildcard for all certificate store secrets.
// Example: "secret/data/fabric/org1/certstore/*"
func (cs CertificateStoreConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/%s/certstore/*", vaultSecretDataPrefix, cs.VaultBasePath())
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/certstore/*"
func (cs CertificateStoreConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/%s/certstore/*", vaultSecretMetadataPrefix, cs.VaultBasePath())
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
// It embeds OwnerConfig so methods like KubernetesNamespace() and VaultBasePath() are promoted.
type CertificateAuthorityConfig struct {
	OwnerConfig        // embedded interface - methods promoted
	certType    string // "tls" or "org"
	caID        string // "root" for root CAs, or version ID (e.g., "v1") for intermediates
}

// NewCertificateAuthorityConfig creates a CertificateAuthorityConfig for the given owner.
// certType should be "tls" or "org".
// caID should be "root" for root CAs, or a version ID (e.g., "v1") for intermediates.
func NewCertificateAuthorityConfig(owner OwnerConfig, certType, caID string) CertificateAuthorityConfig {
	return CertificateAuthorityConfig{
		OwnerConfig: owner,
		certType:    certType,
		caID:        caID,
	}
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

// TLSDNSNames returns the DNS Subject Alternative Names for TLS certificates.
// These are required for TLS hostname verification.
func (ca CertificateAuthorityConfig) TLSDNSNames() []string {
	return []string{
		ca.ComponentName(),
		fmt.Sprintf("%s.%s", ca.ComponentName(), ca.KubernetesNamespace()),
		fmt.Sprintf("%s.%s.svc", ca.ComponentName(), ca.KubernetesNamespace()),
		fmt.Sprintf("%s.%s.svc.cluster.local", ca.ComponentName(), ca.KubernetesNamespace()),
	}
}

// ------------ Properties ------------

// IsRoot returns true if this is a root CA.
func (ca CertificateAuthorityConfig) IsRoot() bool {
	return ca.caID == "root"
}

// CertType returns the certificate type ("tls" or "org")
func (ca CertificateAuthorityConfig) CertType() string {
	return ca.certType
}

// CAID returns the CA identifier ("root" for root CAs, "v1", "v2", etc. for intermediates)
func (ca CertificateAuthorityConfig) CAID() string {
	return ca.caID
}

// ------------ Vault Properties: CA Server Certs ------------

// PolicyName returns the Vault policy name.
// Example: "fabric-org1-certificate-authority-tls-root"
func (ca CertificateAuthorityConfig) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", ca.OwnerID(), ca.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (ca CertificateAuthorityConfig) AuthRoleName() string {
	return ca.PolicyName()
}

// VaultCABasePath returns the base Vault path for this CA's secrets.
// Example: "fabric/org1/ca/tls/root"
func (ca CertificateAuthorityConfig) VaultCABasePath() string {
	return fmt.Sprintf("%s/ca/%s/%s", ca.VaultBasePath(), ca.certType, ca.caID)
}

// VaultSigningRef returns a SecretRef for the CA's signing keypair + certificate.
// This is used for issuing other certificates (CA signing operations).
// Example path: "fabric/org1/ca/tls/root/signing"
func (ca CertificateAuthorityConfig) VaultSigningRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/signing")
}

// VaultSigningChainRef returns a SecretRef for the CA's parent certificate chain.
// This is used by intermediate CAs for x509 chain verification.
// The chain file contains parent CA certs needed to verify client certs.
// Root CAs don't have a parent chain (self-signed).
// Example path: "fabric/org1/ca/tls/v1/signing-chain"
func (ca CertificateAuthorityConfig) VaultSigningChainRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/signing-chain")
}

// VaultTLSRef returns a SecretRef for the CA's TLS keypair + certificate.
// This is used for the CA's HTTPS endpoint.
// Example path: "fabric/org1/ca/tls/root/tls"
func (ca CertificateAuthorityConfig) VaultTLSRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/tls")
}

// ------------ Vault Properties: Database User ------------

// VaultDatabaseUserRef returns a SecretRef for the CA's database user identity.
// Stores username, password for database access to the certificate store.
// Each CA has its own database user for accessing the certificate store.
// Example path: "fabric/org1/ca/tls/root/certstore/user"
func (ca CertificateAuthorityConfig) VaultDatabaseUserRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/certstore/user")
}

// ------------ Vault Properties: Bootstrap Admin ------------

// VaultBootstrapAdminCredentialsRef returns a SecretRef for bootstrap admin username/password.
// These credentials are used to initially authenticate with the CA.
// Example path: "fabric/org1/ca/tls/root/bootstrap-admin/credentials"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminCredentialsRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/bootstrap-admin/credentials")
}

// VaultBootstrapAdminSigningRef returns a SecretRef for bootstrap admin's signing identity.
// Example path: "fabric/org1/ca/tls/root/bootstrap-admin/signing"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminSigningRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/bootstrap-admin/signing")
}

// VaultBootstrapAdminTLSRef returns a SecretRef for bootstrap admin's TLS client certificate.
// Example path: "fabric/org1/ca/tls/root/bootstrap-admin/tls"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminTLSRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/bootstrap-admin/tls")
}

// ------------ Vault Properties: Permanent Admin ------------

// VaultAdminSigningRef returns a SecretRef for the permanent admin's signing identity.
// Example path: "fabric/org1/ca/tls/root/admin/signing"
func (ca CertificateAuthorityConfig) VaultAdminSigningRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/admin/signing")
}

// VaultAdminTLSRef returns a SecretRef for the permanent admin's TLS client certificate.
// Example path: "fabric/org1/ca/tls/root/admin/tls"
func (ca CertificateAuthorityConfig) VaultAdminTLSRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCABasePath() + "/admin/tls")
}

// ------------ Vault Properties: Policy Paths ------------

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Used in Vault policy documents with wildcard for all CA secrets.
// Example: "secret/data/fabric/org1/ca/tls/root/*"
func (ca CertificateAuthorityConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, ca.VaultCABasePath())
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/ca/tls/root/*"
func (ca CertificateAuthorityConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, ca.VaultCABasePath())
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
	ownerID := sanitizeForPostgres(ca.OwnerID())
	certType := sanitizeForPostgres(ca.certType)
	caID := sanitizeForPostgres(ca.caID)
	return fmt.Sprintf("fabric_%s_ca_%s_%s", ownerID, certType, caID)
}

// DatabaseUsername returns the PostgreSQL username for this CA's database.
// Each CA gets a dedicated user with access only to its database.
// Example: "fabric_org1_ca_tls_root_user"
func (ca CertificateAuthorityConfig) DatabaseUsername() string {
	return ca.DatabaseName() + "_user"
}

// ------------ Certificate Subject Properties ------------

// Description returns a human-readable description of this CA.
// Useful for logging and debugging.
// Example: "TLS Root CA for org1"
func (ca CertificateAuthorityConfig) Description() string {
	caType := "Root"
	if !ca.IsRoot() {
		caType = "Intermediate"
	}

	certTypeDesc := "TLS"
	if ca.certType == "org" {
		certTypeDesc = "Org"
	}

	return fmt.Sprintf("%s %s CA for %s", certTypeDesc, caType, ca.OwnerID())
}

// OrganizationalUnit returns the OU field for certificate subject.
// Returns "root-ca" for root CAs, or the CA version ID for intermediate CAs.
func (ca CertificateAuthorityConfig) OrganizationalUnit() string {
	if ca.IsRoot() {
		return "root-ca"
	}
	return ca.caID
}

// AdminCommonName returns the common name for the CA admin's signing certificate.
// Example: "admin@certificate-authority-tls-root"
func (ca CertificateAuthorityConfig) AdminCommonName() string {
	return fmt.Sprintf("admin@%s", ca.ComponentName())
}

// BootstrapAdminCommonName returns the common name for the bootstrap admin's signing certificate.
// Example: "bootstrap-admin@certificate-authority-tls-root"
func (ca CertificateAuthorityConfig) BootstrapAdminCommonName() string {
	return fmt.Sprintf("bootstrap-admin@%s", ca.ComponentName())
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
