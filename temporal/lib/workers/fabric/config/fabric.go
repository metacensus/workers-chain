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

// TLS Intermediate CA ID - used for cross-chain TLS cert issuance
const tlsIntermediateCAID = "v1"

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

// CouchDb returns the configuration for a CouchDB instance associated with a peer.
func (fc FabricConfig) CouchDb(peerID string) CouchDbConfig {
	return CouchDbConfig{
		OrgConfig: fc.OrgConfig,
		peerID:    peerID,
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

// TLSCommonName returns the common name for the CA's TLS certificate.
// If the TLS issuer is the same as the signing issuer (e.g., TLS Int CA where
// both certs come from TLS Root), appends "-tls" to avoid registration collision.
// Example: "certificate-authority-tls-v1-tls" (when colliding)
//
//	"certificate-authority-tls-v1" (when different issuers or self-signed)
func (ca CertificateAuthorityConfig) TLSCommonName() string {
	signingIssuer := ca.SigningIssuer()
	tlsIssuer := ca.TLSIssuer()

	// Self-signed cases: no collision (not registering with a CA)
	if signingIssuer == nil || tlsIssuer == nil {
		return ca.ComponentName()
	}

	// Check if same CA issues both - need different identity name
	if signingIssuer.CertType() == tlsIssuer.CertType() && signingIssuer.CAID() == tlsIssuer.CAID() {
		return fmt.Sprintf("%s-tls", ca.ComponentName())
	}

	// Different CAs - can use same name
	return ca.ComponentName()
}

// ------------ Issuer Resolution ------------

// IsRoot returns true if this is a root CA.
func (ca CertificateAuthorityConfig) IsRoot() bool {
	return ca.caID == "root"
}

// IsTLS returns true if this is a TLS CA (not an Org CA).
func (ca CertificateAuthorityConfig) IsTLS() bool {
	return ca.certType == "tls"
}

// withTypeAndID returns a new config with different certType and caID.
func (ca CertificateAuthorityConfig) withTypeAndID(certType, caID string) CertificateAuthorityConfig {
	return CertificateAuthorityConfig{
		OrgConfig: ca.OrgConfig,
		certType:  certType,
		caID:      caID,
	}
}

// withID returns a new config with different caID but same certType.
func (ca CertificateAuthorityConfig) withID(caID string) CertificateAuthorityConfig {
	return ca.withTypeAndID(ca.certType, caID)
}

// SigningIssuer returns who issues this CA's signing certificate.
// Returns nil for self-signed (root CAs).
// Intermediate CAs are signed by their chain's root.
func (ca CertificateAuthorityConfig) SigningIssuer() *CertificateAuthorityConfig {
	if ca.IsRoot() {
		return nil // Self-signed
	}
	issuer := ca.withID("root")
	return &issuer
}

// TLSIssuer returns who issues this CA's TLS certificate.
// Returns nil for self-signed (TLS Root only).
//
// Issuer matrix:
//   - TLS Root CA: self-signed
//   - TLS Int CA: TLS Root CA
//   - Org Root CA: TLS Int CA (cross-chain!)
//   - Org Int CA: TLS Int CA (cross-chain!)
func (ca CertificateAuthorityConfig) TLSIssuer() *CertificateAuthorityConfig {
	switch {
	case ca.IsTLS() && ca.IsRoot():
		return nil // TLS Root CA self-signs its TLS cert
	case ca.IsTLS():
		issuer := ca.withID("root") // TLS Int gets TLS from TLS Root
		return &issuer
	default:
		issuer := ca.withTypeAndID("tls", tlsIntermediateCAID) // Org CAs get TLS from TLS Int
		return &issuer
	}
}

// AdminSigningIssuer returns who issues admin identity certs for this CA.
// All admins get identity certs from the CA they administer.
// This is required because Fabric CA checks its database to verify admin privileges -
// the admin's cert must be in the CA's database for privileged operations to succeed.
//
// Note: The original design (Fabric Certs.csv) had TLS Int CA admins getting identity
// from TLS Root CA, but this doesn't work with Fabric CA's database-centric model.
// Fixed 2026-01-04: All admins now get identity from their own CA.
func (ca CertificateAuthorityConfig) AdminSigningIssuer() CertificateAuthorityConfig {
	// Admin identity always comes from the CA itself
	return ca
}

// AdminTLSIssuer returns who issues admin client TLS certs for this CA.
//
// Issuer matrix:
//   - TLS Root CA Admin: TLS Root CA
//   - TLS Int CA Admin: TLS Int CA
//   - Org Root CA Admin: TLS Int CA
//   - Org Int CA Admin: TLS Int CA
func (ca CertificateAuthorityConfig) AdminTLSIssuer() CertificateAuthorityConfig {
	if ca.IsTLS() && ca.IsRoot() {
		return ca // TLS Root CA admin TLS from TLS Root
	}
	return ca.withTypeAndID("tls", tlsIntermediateCAID) // Everyone else from TLS Int
}

// ------------ Vault Properties: CA Server Certs ------------

// PolicyName returns the Vault policy name.
// Example: "fabric-org1-certificate-authority-tls-root"
func (ca CertificateAuthorityConfig) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", ca.OrgID, ca.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (ca CertificateAuthorityConfig) AuthRoleName() string {
	return ca.PolicyName()
}

// VaultBasePath returns the base Vault path for this CA's secrets.
// Example: "fabric/org1/ca/tls/root"
func (ca CertificateAuthorityConfig) VaultBasePath() string {
	return fmt.Sprintf("fabric/%s/ca/%s/%s", ca.OrgID, ca.certType, ca.caID)
}

// VaultSigningRef returns a SecretRef for the CA's signing keypair + certificate.
// This is used for issuing other certificates (CA signing operations).
// Example path: "fabric/org1/ca/tls/root/signing"
func (ca CertificateAuthorityConfig) VaultSigningRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/signing")
}

// VaultSigningChainRef returns a SecretRef for the CA's parent certificate chain.
// This is used by intermediate CAs for x509 chain verification.
// The chain file contains parent CA certs needed to verify client certs.
// Root CAs don't have a parent chain (self-signed).
// Example path: "fabric/org1/ca/tls/v1/signing-chain"
func (ca CertificateAuthorityConfig) VaultSigningChainRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/signing-chain")
}

// VaultTLSRef returns a SecretRef for the CA's TLS keypair + certificate.
// This is used for the CA's HTTPS endpoint.
// Example path: "fabric/org1/ca/tls/root/tls"
func (ca CertificateAuthorityConfig) VaultTLSRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/tls")
}

// ------------ Vault Properties: Database User ------------

// VaultDatabaseUserRef returns a SecretRef for the CA's database user identity.
// Stores username, password for database access to the certificate store.
// Each CA has its own database user for accessing the certificate store.
// Example path: "fabric/org1/ca/tls/root/certstore/user"
func (ca CertificateAuthorityConfig) VaultDatabaseUserRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/certstore/user")
}

// ------------ Vault Properties: Bootstrap Admin ------------

// VaultBootstrapAdminCredentialsRef returns a SecretRef for bootstrap admin username/password.
// These credentials are used to initially authenticate with the CA.
// Example path: "fabric/org1/ca/tls/root/bootstrap-admin/credentials"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminCredentialsRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/bootstrap-admin/credentials")
}

// VaultBootstrapAdminSigningRef returns a SecretRef for bootstrap admin's signing identity.
// Example path: "fabric/org1/ca/tls/root/bootstrap-admin/signing"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminSigningRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/bootstrap-admin/signing")
}

// VaultBootstrapAdminTLSRef returns a SecretRef for bootstrap admin's TLS client certificate.
// Example path: "fabric/org1/ca/tls/root/bootstrap-admin/tls"
func (ca CertificateAuthorityConfig) VaultBootstrapAdminTLSRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/bootstrap-admin/tls")
}

// ------------ Vault Properties: Permanent Admin ------------

// VaultAdminSigningRef returns a SecretRef for the permanent admin's signing identity.
// Example path: "fabric/org1/ca/tls/root/admin/signing"
func (ca CertificateAuthorityConfig) VaultAdminSigningRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/admin/signing")
}

// VaultAdminTLSRef returns a SecretRef for the permanent admin's TLS client certificate.
// Example path: "fabric/org1/ca/tls/root/admin/tls"
func (ca CertificateAuthorityConfig) VaultAdminTLSRef() *pb.SecretRef {
	return secrets.NewVaultRef(ca.VaultBasePath() + "/admin/tls")
}

// ------------ Vault Properties: Policy Paths ------------

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Used in Vault policy documents with wildcard for all CA secrets.
// Example: "secret/data/fabric/org1/ca/tls/root/*"
func (ca CertificateAuthorityConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, ca.VaultBasePath())
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/ca/tls/root/*"
func (ca CertificateAuthorityConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, ca.VaultBasePath())
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

// AdminTLSCommonName returns the common name for the CA admin's TLS certificate.
// If the TLS issuer is the same as the signing issuer, appends "-tls" to avoid
// registration collision (can't register same identity name twice with same CA).
// Example: "admin-tls@certificate-authority-tls-root" (when colliding)
//
//	"admin@certificate-authority-tls-root" (when different CAs)
func (ca CertificateAuthorityConfig) AdminTLSCommonName() string {
	signingIssuer := ca.AdminSigningIssuer()
	tlsIssuer := ca.AdminTLSIssuer()

	// Check if same CA issues both - need different identity name to avoid collision
	if signingIssuer.CertType() == tlsIssuer.CertType() && signingIssuer.CAID() == tlsIssuer.CAID() {
		return fmt.Sprintf("admin-tls@%s", ca.ComponentName())
	}
	// Different CAs - can use same name (registered with different CAs)
	return ca.AdminCommonName()
}

// BootstrapAdminCommonName returns the common name for the bootstrap admin's signing certificate.
// Example: "bootstrap-admin@certificate-authority-tls-root"
func (ca CertificateAuthorityConfig) BootstrapAdminCommonName() string {
	return fmt.Sprintf("bootstrap-admin@%s", ca.ComponentName())
}

// BootstrapAdminTLSCommonName returns the common name for the bootstrap admin's TLS certificate.
// If the TLS issuer is the same as the signing issuer, appends "-tls" to avoid
// registration collision. For bootstrap enrollment, this matters less since we're
// enrolling with existing credentials, but we use consistent naming.
func (ca CertificateAuthorityConfig) BootstrapAdminTLSCommonName() string {
	signingIssuer := ca.AdminSigningIssuer()
	tlsIssuer := ca.AdminTLSIssuer()

	if signingIssuer.CertType() == tlsIssuer.CertType() && signingIssuer.CAID() == tlsIssuer.CAID() {
		return fmt.Sprintf("bootstrap-admin-tls@%s", ca.ComponentName())
	}
	return ca.BootstrapAdminCommonName()
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

// ============= CouchDB Configuration =============

// CouchDbConfig encapsulates all configuration for a CouchDB instance.
// Each peer gets its own CouchDB for state database storage.
// It embeds OrgConfig so it can be used anywhere an OrgConfig is needed.
type CouchDbConfig struct {
	config.OrgConfig
	peerID string
}

// ------------ Kubernetes Properties ------------

// ComponentName returns the base component name used in Helm charts and derived identifiers.
// Example: "couchdb-peer0"
func (c CouchDbConfig) ComponentName() string {
	return fmt.Sprintf("couchdb-%s", c.peerID)
}

// ServiceAccountName returns the Kubernetes service account name.
func (c CouchDbConfig) ServiceAccountName() string {
	return c.ComponentName()
}

// KubernetesNamespace returns the Kubernetes namespace for deployment.
func (c CouchDbConfig) KubernetesNamespace() string {
	return c.OrgConfig.KubernetesNamespace()
}

// Port returns the CouchDB HTTP port.
func (c CouchDbConfig) Port() int {
	return 5984
}

// ResourceIdentifier returns the Kubernetes resource identifier for waiting on deployment readiness.
// CouchDB is deployed as a StatefulSet for persistent storage.
func (c CouchDbConfig) ResourceIdentifier() *kubev1.ResourceIdentifier {
	return &kubev1.ResourceIdentifier{
		ApiVersion: "apps/v1",
		Kind:       "StatefulSet",
		Name:       c.ComponentName(),
		Namespace:  c.KubernetesNamespace(),
	}
}

// PeerID returns the peer ID this CouchDB is associated with.
func (c CouchDbConfig) PeerID() string {
	return c.peerID
}

// ------------ Vault Properties ------------

// PolicyName returns the Vault policy name.
// Example: "fabric-org1-couchdb-peer0"
func (c CouchDbConfig) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", c.OrgID, c.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (c CouchDbConfig) AuthRoleName() string {
	return c.PolicyName()
}

// VaultBasePath returns the base Vault path for this CouchDB's secrets.
// Example: "fabric/org1/couchdb/peer0"
func (c CouchDbConfig) VaultBasePath() string {
	return fmt.Sprintf("fabric/%s/couchdb/%s", c.OrgID, c.peerID)
}

// VaultAdminRef returns a SecretRef for the CouchDB admin credentials.
// Stores username, password for CouchDB admin.
// Example path: "fabric/org1/couchdb/peer0/admin"
func (c CouchDbConfig) VaultAdminRef() *pb.SecretRef {
	return secrets.NewVaultRef(c.VaultBasePath() + "/admin")
}

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Example: "secret/data/fabric/org1/couchdb/peer0/*"
func (c CouchDbConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, c.VaultBasePath())
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/couchdb/peer0/*"
func (c CouchDbConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, c.VaultBasePath())
}

// VaultPolicy creates a Vault policy document for the CouchDB instance.
// The policy grants read access to the CouchDB's secrets (admin credentials).
func (c CouchDbConfig) VaultPolicy() *vaultpb.PolicyDocument {
	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// KV v2 data path - read access to CouchDB secrets
			{
				Path: c.VaultDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// KV v2 metadata path - list and read metadata
			{
				Path: c.VaultMetadataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
		},
	}
}
