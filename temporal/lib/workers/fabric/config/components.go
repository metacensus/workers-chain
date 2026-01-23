package config

import (
	"fmt"

	kubev1 "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
)

// ============= CouchDB Configuration =============

// CouchDbConfig encapsulates all configuration for a CouchDB instance.
// Each peer gets its own CouchDB for state database storage.
// It embeds OwnerConfig so methods like KubernetesNamespace() and VaultBasePath() are promoted.
type CouchDbConfig struct {
	OwnerConfig // embedded interface - methods promoted
	peerID      string
}

// NewCouchDbConfig creates a CouchDbConfig for the given owner and peer.
func NewCouchDbConfig(owner OwnerConfig, peerID string) CouchDbConfig {
	return CouchDbConfig{
		OwnerConfig: owner,
		peerID:      peerID,
	}
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
	return fmt.Sprintf("fabric-%s-%s", c.OwnerID(), c.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (c CouchDbConfig) AuthRoleName() string {
	return c.PolicyName()
}

// VaultCouchDbBasePath returns the base Vault path for this CouchDB's secrets.
// Example: "fabric/org1/couchdb/peer0"
func (c CouchDbConfig) VaultCouchDbBasePath() string {
	return fmt.Sprintf("%s/couchdb/%s", c.VaultBasePath(), c.peerID)
}

// VaultAdminRef returns a SecretRef for the CouchDB admin credentials.
// Stores username, password for CouchDB admin.
// Example path: "fabric/org1/couchdb/peer0/admin"
func (c CouchDbConfig) VaultAdminRef() *pb.SecretRef {
	return secrets.NewVaultRef(c.VaultCouchDbBasePath() + "/admin")
}

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Example: "secret/data/fabric/org1/couchdb/peer0/*"
func (c CouchDbConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, c.VaultCouchDbBasePath())
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/couchdb/peer0/*"
func (c CouchDbConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, c.VaultCouchDbBasePath())
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

// ============= Peer Configuration =============

// PeerConfig encapsulates all configuration for a Hyperledger Fabric peer.
// Each peer needs identity certs (from Org Int CA) and TLS certs (from TLS Int CA).
// It embeds OwnerConfig so methods like KubernetesNamespace() and VaultBasePath() are promoted.
type PeerConfig struct {
	OwnerConfig // embedded interface - methods promoted
	peerID      string
}

// NewPeerConfig creates a PeerConfig for the given owner and peer ID.
func NewPeerConfig(owner OwnerConfig, peerID string) PeerConfig {
	return PeerConfig{
		OwnerConfig: owner,
		peerID:      peerID,
	}
}

// ------------ Kubernetes Properties ------------

// ComponentName returns the base component name used in Helm charts and derived identifiers.
// Example: "peer0"
func (p PeerConfig) ComponentName() string {
	return p.peerID
}

// ServiceAccountName returns the Kubernetes service account name.
// Example: "peer0"
func (p PeerConfig) ServiceAccountName() string {
	return p.ComponentName()
}

// Port returns the peer gRPC port.
func (p PeerConfig) Port() int {
	return 7051
}

// ChaincodePort returns the chaincode listen port.
func (p PeerConfig) ChaincodePort() int {
	return 7052
}

// OperationsPort returns the operations/metrics port.
func (p PeerConfig) OperationsPort() int {
	return 9443
}

// ResourceIdentifier returns the Kubernetes resource identifier for waiting on deployment readiness.
// Peers are deployed as StatefulSets for stable network identity and persistent storage.
func (p PeerConfig) ResourceIdentifier() *kubev1.ResourceIdentifier {
	return &kubev1.ResourceIdentifier{
		ApiVersion: "apps/v1",
		Kind:       "StatefulSet",
		Name:       p.ComponentName(),
		Namespace:  p.KubernetesNamespace(),
	}
}

// PeerID returns the peer identifier.
func (p PeerConfig) PeerID() string {
	return p.peerID
}

// MSPID returns the organization's MSP ID.
// Example: "org1MSP"
func (p PeerConfig) MSPID() string {
	return p.OwnerID() + "MSP"
}

// TLSDNSNames returns the DNS Subject Alternative Names for TLS certificates.
// These are required for TLS hostname verification.
func (p PeerConfig) TLSDNSNames() []string {
	return []string{
		p.ComponentName(),
		fmt.Sprintf("%s.%s", p.ComponentName(), p.KubernetesNamespace()),
		fmt.Sprintf("%s.%s.svc", p.ComponentName(), p.KubernetesNamespace()),
		fmt.Sprintf("%s.%s.svc.cluster.local", p.ComponentName(), p.KubernetesNamespace()),
	}
}

// CouchDbAddress returns the CouchDB address for this peer's state database.
// Example: "couchdb-peer0.fabric-org1.svc.cluster.local:5984"
func (p PeerConfig) CouchDbAddress() string {
	return fmt.Sprintf("couchdb-%s.%s.svc.cluster.local:5984", p.peerID, p.KubernetesNamespace())
}

// GossipBootstrapAddress returns the gossip bootstrap address.
// For the first peer, this is empty. For subsequent peers, it points to peer0.
func (p PeerConfig) GossipBootstrapAddress() string {
	if p.peerID == "peer0" {
		return "" // First peer has no bootstrap peer
	}
	return fmt.Sprintf("peer0.%s.svc.cluster.local:7051", p.KubernetesNamespace())
}

// GossipExternalEndpoint returns the external endpoint for gossip service discovery.
func (p PeerConfig) GossipExternalEndpoint() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:7051", p.ComponentName(), p.KubernetesNamespace())
}

// ------------ Vault Properties ------------

// PolicyName returns the Vault policy name.
// Example: "fabric-org1-peer0"
func (p PeerConfig) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", p.OwnerID(), p.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (p PeerConfig) AuthRoleName() string {
	return p.PolicyName()
}

// VaultPeerBasePath returns the base Vault path for this peer's secrets.
// Example: "fabric/org1/peer/peer0"
func (p PeerConfig) VaultPeerBasePath() string {
	return fmt.Sprintf("%s/peer/%s", p.VaultBasePath(), p.peerID)
}

// VaultSigningRef returns a SecretRef for the peer's signing identity (org cert).
// Example path: "fabric/org1/peer/peer0/signing"
func (p PeerConfig) VaultSigningRef() *pb.SecretRef {
	return secrets.NewVaultRef(p.VaultPeerBasePath() + "/signing")
}

// VaultTLSRef returns a SecretRef for the peer's TLS identity.
// Example path: "fabric/org1/peer/peer0/tls"
func (p PeerConfig) VaultTLSRef() *pb.SecretRef {
	return secrets.NewVaultRef(p.VaultPeerBasePath() + "/tls")
}

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Example: "secret/data/fabric/org1/peer/peer0/*"
func (p PeerConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, p.VaultPeerBasePath())
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/peer/peer0/*"
func (p PeerConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, p.VaultPeerBasePath())
}

// VaultPolicy creates a Vault policy document for the peer.
// The policy grants read access to:
// - The peer's own secrets (identity and TLS certs)
// - The Org Root CA signing cert (for MSP cacerts)
// - The Org Int CA signing cert (for MSP intermediatecerts)
// - The TLS Root CA signing cert (for TLS CA chain)
// - The TLS Int CA signing cert (for TLS CA chain)
// - CouchDB credentials (for state database access)
func (p PeerConfig) VaultPolicy() *vaultpb.PolicyDocument {
	// Build paths for CA certificates the peer needs to read
	orgRootCASigningPath := fmt.Sprintf("%s/%s/ca/org/root/signing", vaultSecretDataPrefix, p.VaultBasePath())
	orgIntCASigningPath := fmt.Sprintf("%s/%s/ca/org/v1/signing", vaultSecretDataPrefix, p.VaultBasePath())
	tlsRootCASigningPath := fmt.Sprintf("%s/%s/ca/tls/root/signing", vaultSecretDataPrefix, p.VaultBasePath())
	tlsIntCASigningPath := fmt.Sprintf("%s/%s/ca/tls/v1/signing", vaultSecretDataPrefix, p.VaultBasePath())
	couchDbPath := fmt.Sprintf("%s/%s/couchdb/%s/*", vaultSecretDataPrefix, p.VaultBasePath(), p.peerID)

	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// KV v2 data path - read access to peer's own secrets
			{
				Path: p.VaultDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// KV v2 metadata path - list and read metadata for peer secrets
			{
				Path: p.VaultMetadataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// Org Root CA signing cert - for MSP cacerts directory
			{
				Path: orgRootCASigningPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// Org Int CA signing cert - for MSP intermediatecerts directory
			{
				Path: orgIntCASigningPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// TLS Root CA signing cert - for TLS CA chain
			{
				Path: tlsRootCASigningPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// TLS Int CA signing cert - for TLS CA chain
			{
				Path: tlsIntCASigningPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// CouchDB credentials - for state database access
			{
				Path: couchDbPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
		},
	}
}

// ------------ Identity Configuration ------------

// IdentityName returns the identity name used for certificate registration.
// Example: "peer0"
func (p PeerConfig) IdentityName() string {
	return p.peerID
}

// CommonName returns the common name for the peer's certificates.
// Example: "peer0@org1"
func (p PeerConfig) CommonName() string {
	return fmt.Sprintf("%s@%s", p.peerID, p.OwnerID())
}

// OrganizationalUnit returns the OU field for certificate subject.
// Peers use "peer" as their OU.
func (p PeerConfig) OrganizationalUnit() string {
	return "peer"
}

// ============= Orderer Configuration =============

// OrdererConfig encapsulates all configuration for a Hyperledger Fabric orderer.
// Each orderer needs identity certs (from Org Int CA) and TLS certs (from TLS Int CA).
// It embeds OwnerConfig so methods like KubernetesNamespace() and VaultBasePath() are promoted.
type OrdererConfig struct {
	OwnerConfig // embedded interface - methods promoted
	ordererID   string
}

// NewOrdererConfig creates an OrdererConfig for the given owner and orderer ID.
func NewOrdererConfig(owner OwnerConfig, ordererID string) OrdererConfig {
	return OrdererConfig{
		OwnerConfig: owner,
		ordererID:   ordererID,
	}
}

// ------------ Kubernetes Properties ------------

// ComponentName returns the base component name used in Helm charts and derived identifiers.
// Example: "orderer0"
func (o OrdererConfig) ComponentName() string {
	return o.ordererID
}

// ServiceAccountName returns the Kubernetes service account name.
// Example: "orderer0"
func (o OrdererConfig) ServiceAccountName() string {
	return o.ComponentName()
}

// Port returns the orderer gRPC port for transaction ordering.
func (o OrdererConfig) Port() int {
	return 7050
}

// AdminPort returns the orderer admin API port for osnadmin CLI.
func (o OrdererConfig) AdminPort() int {
	return 7053
}

// OperationsPort returns the operations/metrics port.
func (o OrdererConfig) OperationsPort() int {
	return 9443
}

// ResourceIdentifier returns the Kubernetes resource identifier for waiting on deployment readiness.
// Orderers are deployed as StatefulSets for stable network identity and persistent storage.
func (o OrdererConfig) ResourceIdentifier() *kubev1.ResourceIdentifier {
	return &kubev1.ResourceIdentifier{
		ApiVersion: "apps/v1",
		Kind:       "StatefulSet",
		Name:       o.ComponentName(),
		Namespace:  o.KubernetesNamespace(),
	}
}

// OrdererID returns the orderer identifier.
func (o OrdererConfig) OrdererID() string {
	return o.ordererID
}

// MSPID returns the organization's MSP ID.
// Example: "org1MSP"
func (o OrdererConfig) MSPID() string {
	return o.OwnerID() + "MSP"
}

// TLSDNSNames returns the DNS Subject Alternative Names for TLS certificates.
// These are required for TLS hostname verification.
func (o OrdererConfig) TLSDNSNames() []string {
	return []string{
		o.ComponentName(),
		fmt.Sprintf("%s.%s", o.ComponentName(), o.KubernetesNamespace()),
		fmt.Sprintf("%s.%s.svc", o.ComponentName(), o.KubernetesNamespace()),
		fmt.Sprintf("%s.%s.svc.cluster.local", o.ComponentName(), o.KubernetesNamespace()),
	}
}

// ClusterAddress returns the address for Raft cluster communication.
// Example: "orderer0.fabric-org1.svc.cluster.local:7050"
func (o OrdererConfig) ClusterAddress() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", o.ComponentName(), o.KubernetesNamespace(), o.Port())
}

// ------------ Vault Properties ------------

// PolicyName returns the Vault policy name.
// Example: "fabric-org1-orderer0"
func (o OrdererConfig) PolicyName() string {
	return fmt.Sprintf("fabric-%s-%s", o.OwnerID(), o.ComponentName())
}

// AuthRoleName returns the Vault Kubernetes auth role name.
func (o OrdererConfig) AuthRoleName() string {
	return o.PolicyName()
}

// VaultOrdererBasePath returns the base Vault path for this orderer's secrets.
// Example: "fabric/org1/orderer/orderer0"
func (o OrdererConfig) VaultOrdererBasePath() string {
	return fmt.Sprintf("%s/orderer/%s", o.VaultBasePath(), o.ordererID)
}

// VaultSigningRef returns a SecretRef for the orderer's signing identity (org cert).
// Example path: "fabric/org1/orderer/orderer0/signing"
func (o OrdererConfig) VaultSigningRef() *pb.SecretRef {
	return secrets.NewVaultRef(o.VaultOrdererBasePath() + "/signing")
}

// VaultTLSRef returns a SecretRef for the orderer's TLS identity.
// Example path: "fabric/org1/orderer/orderer0/tls"
func (o OrdererConfig) VaultTLSRef() *pb.SecretRef {
	return secrets.NewVaultRef(o.VaultOrdererBasePath() + "/tls")
}

// VaultDataPath returns the Vault KV v2 data path pattern for policy rules.
// Example: "secret/data/fabric/org1/orderer/orderer0/*"
func (o OrdererConfig) VaultDataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, o.VaultOrdererBasePath())
}

// VaultMetadataPath returns the Vault KV v2 metadata path pattern for policy rules.
// Example: "secret/metadata/fabric/org1/orderer/orderer0/*"
func (o OrdererConfig) VaultMetadataPath() string {
	return fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, o.VaultOrdererBasePath())
}

// VaultPolicy creates a Vault policy document for the orderer.
// The policy grants read access to:
// - The orderer's own secrets (identity and TLS certs)
// - The Org Root CA signing cert (for MSP cacerts)
// - The Org Int CA signing cert (for MSP intermediatecerts)
// - The TLS Root CA signing cert (for TLS CA chain)
// - The TLS Int CA signing cert (for TLS CA chain)
func (o OrdererConfig) VaultPolicy() *vaultpb.PolicyDocument {
	// Build paths for CA certificates the orderer needs to read
	orgRootCASigningPath := fmt.Sprintf("%s/%s/ca/org/root/signing", vaultSecretDataPrefix, o.VaultBasePath())
	orgIntCASigningPath := fmt.Sprintf("%s/%s/ca/org/v1/signing", vaultSecretDataPrefix, o.VaultBasePath())
	tlsRootCASigningPath := fmt.Sprintf("%s/%s/ca/tls/root/signing", vaultSecretDataPrefix, o.VaultBasePath())
	tlsIntCASigningPath := fmt.Sprintf("%s/%s/ca/tls/v1/signing", vaultSecretDataPrefix, o.VaultBasePath())

	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// KV v2 data path - read access to orderer's own secrets
			{
				Path: o.VaultDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// KV v2 metadata path - list and read metadata for orderer secrets
			{
				Path: o.VaultMetadataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// Org Root CA signing cert - for MSP cacerts directory
			{
				Path: orgRootCASigningPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// Org Int CA signing cert - for MSP intermediatecerts directory
			{
				Path: orgIntCASigningPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// TLS Root CA signing cert - for TLS CA chain
			{
				Path: tlsRootCASigningPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
			// TLS Int CA signing cert - for TLS CA chain
			{
				Path: tlsIntCASigningPath,
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
		},
	}
}

// ------------ Identity Configuration ------------

// IdentityName returns the identity name used for certificate registration.
// Example: "orderer0"
func (o OrdererConfig) IdentityName() string {
	return o.ordererID
}

// CommonName returns the common name for the orderer's certificates.
// Example: "orderer0@org1"
func (o OrdererConfig) CommonName() string {
	return fmt.Sprintf("%s@%s", o.ordererID, o.OwnerID())
}

// OrganizationalUnit returns the OU field for certificate subject.
// Orderers use "orderer" as their OU.
func (o OrdererConfig) OrganizationalUnit() string {
	return "orderer"
}
