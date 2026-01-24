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

// VaultKeysBasePath returns the base Vault path for the certificate store's secrets.
// The certificate store only has credentials, which are private, so they live under keys/.
// Example: "fabric/org1/keys/certstore"
func (cs CertificateStoreConfig) VaultKeysBasePath() string {
	return fmt.Sprintf("%s/keys/certstore", cs.VaultBasePath())
}

// VaultAdminRef returns a SecretRef for the certificate store admin identity.
// Stores {username, password} for postgres admin.
// Example path: "fabric/org1/keys/certstore/admin"
func (cs CertificateStoreConfig) VaultAdminRef() *pb.SecretRef {
	return secrets.NewVaultRef(cs.VaultKeysBasePath() + "/admin")
}

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Used in Vault policy documents with wildcard for all certificate store secrets.
// Example: "secret/data/fabric/org1/keys/certstore/*"
func (cs CertificateStoreConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, cs.VaultKeysBasePath())
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/keys/certstore/*"
func (cs CertificateStoreConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, cs.VaultKeysBasePath())
}

// VaultPolicy creates a Vault policy document for the certificate store.
// The policy grants read access to the certificate store's credentials (postgres admin).
func (cs CertificateStoreConfig) VaultPolicy() *vaultpb.PolicyDocument {
	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// Private: read access to certificate store credentials
			{
				Path: cs.VaultDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// Private: metadata for credentials
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

// ------------ Vault Properties: Base Paths ------------

// PolicyName returns the Vault policy name.
// Example: "fabric-org1-certificate-authority-tls-root"
func (ca CertificateAuthorityConfig) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", ca.OwnerID(), ca.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (ca CertificateAuthorityConfig) AuthRoleName() string {
	return ca.PolicyName()
}

// VaultKeysBasePath returns the base Vault path for this CA's private keys.
// Example: "fabric/org1/keys/ca/tls/root"
func (ca CertificateAuthorityConfig) VaultKeysBasePath() string {
	return fmt.Sprintf("%s/keys/ca/%s/%s", ca.VaultBasePath(), ca.certType, ca.caID)
}

// VaultCertsBasePath returns the base Vault path for this CA's certificates.
// Example: "fabric/org1/certs/ca/tls/root"
func (ca CertificateAuthorityConfig) VaultCertsBasePath() string {
	return fmt.Sprintf("%s/certs/ca/%s/%s", ca.VaultBasePath(), ca.certType, ca.caID)
}

// ------------ Vault Properties: CA Server Keys (Private) ------------

// VaultSigningKeyRef returns a SecretRef for the CA's signing private key.
// Stores {private_key} for CA signing operations.
// Example path: "fabric/org1/keys/ca/tls/root/signing"
func (ca CertificateAuthorityConfig) VaultSigningKeyRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultKeysBasePath() + "/signing")
}

// VaultTLSKeyRef returns a SecretRef for the CA's TLS private key.
// Stores {private_key} for the CA's HTTPS endpoint.
// Example path: "fabric/org1/keys/ca/tls/root/tls"
func (ca CertificateAuthorityConfig) VaultTLSKeyRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultKeysBasePath() + "/tls")
}

// ------------ Vault Properties: CA Server Certs (Public) ------------

// VaultSigningCertRef returns a SecretRef for the CA's signing certificate.
// Stores {certificate, chain} for CA signing operations.
// Example path: "fabric/org1/certs/ca/tls/root/signing"
func (ca CertificateAuthorityConfig) VaultSigningCertRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCertsBasePath() + "/signing")
}

// VaultTLSCertRef returns a SecretRef for the CA's TLS certificate.
// Stores {certificate, chain} for the CA's HTTPS endpoint.
// Example path: "fabric/org1/certs/ca/tls/root/tls"
func (ca CertificateAuthorityConfig) VaultTLSCertRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCertsBasePath() + "/tls")
}

// ------------ Vault Properties: Database User (Private) ------------

// VaultDatabaseUserRef returns a SecretRef for the CA's database user identity.
// Stores {username, password} for database access to the certificate store.
// Each CA has its own database user for accessing the certificate store.
// Example path: "fabric/org1/keys/ca/tls/root/certstore"
func (ca CertificateAuthorityConfig) VaultDatabaseUserRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultKeysBasePath() + "/certstore")
}

// ------------ Vault Properties: Bootstrap Admin Keys (Private) ------------

// VaultBootstrapAdminCredentialsRef returns a SecretRef for bootstrap admin username/password.
// These credentials are used to initially authenticate with the CA.
// Example path: "fabric/org1/keys/ca/tls/root/bootstrap-admin/credentials"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminCredentialsRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultKeysBasePath() + "/bootstrap-admin/credentials")
}

// VaultBootstrapAdminSigningKeyRef returns a SecretRef for bootstrap admin's signing private key.
// Stores {private_key} for the bootstrap admin's signing identity.
// Example path: "fabric/org1/keys/ca/tls/root/bootstrap-admin/signing"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminSigningKeyRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultKeysBasePath() + "/bootstrap-admin/signing")
}

// VaultBootstrapAdminTLSKeyRef returns a SecretRef for bootstrap admin's TLS private key.
// Stores {private_key} for the bootstrap admin's TLS client certificate.
// Example path: "fabric/org1/keys/ca/tls/root/bootstrap-admin/tls"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminTLSKeyRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultKeysBasePath() + "/bootstrap-admin/tls")
}

// ------------ Vault Properties: Bootstrap Admin Certs (Public) ------------

// VaultBootstrapAdminSigningCertRef returns a SecretRef for bootstrap admin's signing certificate.
// Stores {certificate, chain} for the bootstrap admin's signing identity.
// Example path: "fabric/org1/certs/ca/tls/root/bootstrap-admin/signing"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminSigningCertRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCertsBasePath() + "/bootstrap-admin/signing")
}

// VaultBootstrapAdminTLSCertRef returns a SecretRef for bootstrap admin's TLS certificate.
// Stores {certificate, chain} for the bootstrap admin's TLS client certificate.
// Example path: "fabric/org1/certs/ca/tls/root/bootstrap-admin/tls"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminTLSCertRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCertsBasePath() + "/bootstrap-admin/tls")
}

// ------------ Vault Properties: Permanent Admin Keys (Private) ------------

// VaultAdminSigningKeyRef returns a SecretRef for the permanent admin's signing private key.
// Stores {private_key} for the permanent admin's signing identity.
// Example path: "fabric/org1/keys/ca/tls/root/admin/signing"
func (ca CertificateAuthorityConfig) VaultAdminSigningKeyRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultKeysBasePath() + "/admin/signing")
}

// VaultAdminTLSKeyRef returns a SecretRef for the permanent admin's TLS private key.
// Stores {private_key} for the permanent admin's TLS client certificate.
// Example path: "fabric/org1/keys/ca/tls/root/admin/tls"
func (ca CertificateAuthorityConfig) VaultAdminTLSKeyRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultKeysBasePath() + "/admin/tls")
}

// ------------ Vault Properties: Permanent Admin Certs (Public) ------------

// VaultAdminSigningCertRef returns a SecretRef for the permanent admin's signing certificate.
// Stores {certificate, chain} for the permanent admin's signing identity.
// Example path: "fabric/org1/certs/ca/tls/root/admin/signing"
func (ca CertificateAuthorityConfig) VaultAdminSigningCertRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCertsBasePath() + "/admin/signing")
}

// VaultAdminTLSCertRef returns a SecretRef for the permanent admin's TLS certificate.
// Stores {certificate, chain} for the permanent admin's TLS client certificate.
// Example path: "fabric/org1/certs/ca/tls/root/admin/tls"
func (ca CertificateAuthorityConfig) VaultAdminTLSCertRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultCertsBasePath() + "/admin/tls")
}

// ------------ Vault Properties: Policy Paths ------------

// VaultKeysDataPath returns the Vault KV v2 data path pattern for this CA's private keys.
// Used in Vault policy documents with wildcard for all CA key secrets.
// Example: "secret/data/fabric/org1/keys/ca/tls/root/*"
func (ca CertificateAuthorityConfig) VaultKeysDataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, ca.VaultKeysBasePath())
}

// VaultKeysMetadataPath returns the Vault KV v2 metadata path pattern for this CA's private keys.
// Example: "secret/metadata/fabric/org1/keys/ca/tls/root/*"
func (ca CertificateAuthorityConfig) VaultKeysMetadataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, ca.VaultKeysBasePath())
}

// VaultCertsDataPath returns the Vault KV v2 data path pattern for all certificates (public).
// Used in Vault policy documents with wildcard for all certificate secrets.
// Example: "secret/data/fabric/org1/certs/*"
func (ca CertificateAuthorityConfig) VaultCertsDataPath() string {
	return fmt.Sprintf("%s/%s/certs/*", vaultSecretDataPrefix, ca.VaultBasePath())
}

// VaultCertsMetadataPath returns the Vault KV v2 metadata path pattern for all certificates.
// Example: "secret/metadata/fabric/org1/certs/*"
func (ca CertificateAuthorityConfig) VaultCertsMetadataPath() string {
	return fmt.Sprintf("%s/%s/certs/*", vaultSecretMetadataPrefix, ca.VaultBasePath())
}

// VaultPolicy creates a Vault policy document for the certificate authority.
// The policy grants:
// - Read access to ALL certificates (certs/*) - certificates are public
// - Read/write access to this CA's keys (keys/ca/{chain}/{id}/*) - private keys and credentials
// CAs need write access to store generated certificates.
func (ca CertificateAuthorityConfig) VaultPolicy() *vaultpb.PolicyDocument {
	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// Public: read ALL certificates
			{
				Path: ca.VaultCertsDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// Public: metadata for certificates
			{
				Path: ca.VaultCertsMetadataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// Private: read/write this CA's keys (signing key, TLS key, credentials)
			{
				Path: ca.VaultKeysDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_CREATE,
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_UPDATE,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// Private: metadata for keys
			{
				Path: ca.VaultKeysMetadataPath(),
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
