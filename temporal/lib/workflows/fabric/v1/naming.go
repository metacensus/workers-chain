package fabric

import (
	"fmt"
	"path"

	"github.com/metacensus/infra/temporal/lib/workers/config"
)

// Vault KV v2 path prefixes
const (
	vaultSecretDataPrefix     = "secret/data"
	vaultSecretMetadataPrefix = "secret/metadata"
)

// FabricCertificateAuthority encapsulates naming conventions for Fabric certificate authorities.
// This provides consistent, reliable naming for vault policies, auth roles, service accounts,
// and secret paths across all CA-related workflows.
type FabricCertificateAuthority struct {
	orgConfig config.OrgConfig
	certType  string
	ID        string // "root" for root CAs, or a version ID (e.g., "v1") for intermediates
}

// NewRootCertificateAuthority creates a naming helper for a root certificate authority.
// Root CAs use "root" as their ID.
// certType should be "tls" or "org".
func NewRootCertificateAuthority(orgConfig config.OrgConfig, certType string) *FabricCertificateAuthority {
	return &FabricCertificateAuthority{
		orgConfig: orgConfig,
		certType:  certType,
		ID:        "root",
	}
}

// NewIntermediateCertificateAuthority creates a naming helper for an intermediate certificate authority.
// Intermediate CAs require an ID to support rotation/replacement scenarios.
// certType should be "tls" or "org".
func NewIntermediateCertificateAuthority(orgConfig config.OrgConfig, certType string, id string) *FabricCertificateAuthority {
	return &FabricCertificateAuthority{
		orgConfig: orgConfig,
		certType:  certType,
		ID:        id,
	}
}

// ComponentName returns the base component name used in all derived identifiers.
// Examples:
//   - Root TLS CA: "certificate-authority-tls-root"
//   - Intermediate TLS CA with ID "v1": "certificate-authority-tls-v1"
func (ca *FabricCertificateAuthority) ComponentName() string {
	return fmt.Sprintf("certificate-authority-%s-%s", ca.certType, ca.ID)
}

// ServiceAccountName returns the Kubernetes service account name.
// Examples:
//   - Root TLS CA: "certificate-authority-tls-root"
//   - Intermediate Org CA with ID "v1": "certificate-authority-org-v1"
func (ca *FabricCertificateAuthority) ServiceAccountName() string {
	return ca.ComponentName()
}

// PolicyName returns the Vault policy name.
// Examples:
//   - Root TLS CA for org1: "fabric-org1-certificate-authority-tls-root"
//   - Intermediate TLS CA with ID "v1" for org1: "fabric-org1-certificate-authority-tls-v1"
func (ca *FabricCertificateAuthority) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", ca.orgConfig.OrgID, ca.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
// Uses the same naming as PolicyName for consistency.
func (ca *FabricCertificateAuthority) AuthRoleName() string {
	return ca.PolicyName()
}

// KubernetesNamespace returns the Kubernetes namespace where the CA will be deployed.
func (ca *FabricCertificateAuthority) KubernetesNamespace() string {
	return ca.orgConfig.KubernetesNamespace()
}

// VaultSecretBasePath returns the base path for this CA's secrets (without secret/data prefix).
// This is the logical path used in policies and when storing secrets.
// Examples:
//   - Root TLS CA for org1: "fabric/org1/ca/tls/root"
//   - Intermediate Org CA with ID "v1" for org1: "fabric/org1/ca/org/v1"
func (ca *FabricCertificateAuthority) VaultSecretBasePath() string {
	return path.Join("fabric", ca.orgConfig.OrgID, "ca", ca.certType, ca.ID)
}

// VaultSecretDataPath returns the full Vault KV v2 data path pattern for policy rules.
// Includes the secret/data prefix and wildcard for all secrets under this CA.
// Example: "secret/data/fabric/org1/ca/tls/root/*"
func (ca *FabricCertificateAuthority) VaultSecretDataPath() string {
	return path.Join(vaultSecretDataPrefix, ca.VaultSecretBasePath(), "*")
}

// VaultSecretMetadataPath returns the full Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/ca/tls/root/*"
func (ca *FabricCertificateAuthority) VaultSecretMetadataPath() string {
	return path.Join(vaultSecretMetadataPrefix, ca.VaultSecretBasePath(), "*")
}
