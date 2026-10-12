package helm

import (
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
)

// PeerValues represents the Helm values for a Fabric peer deployment
type PeerValues struct {
	Org            string
	Namespace      string
	PeerId         string
	Image          ImageConfig
	Replicas       int
	Msp            PeerMspConfig
	Vault          PeerVaultConfig
	Service        PeerServiceConfig
	Gossip         GossipConfig
	Couchdb        CouchdbConfig
	Persistence    PersistenceConfig
	Logging        LoggingConfig
	Resources      ResourceConfig
	StartupProbe   PeerProbeConfig
	LivenessProbe  PeerProbeConfig
	ReadinessProbe PeerProbeConfig
	CommonLabels   CommonLabels
}

// PeerMspConfig represents MSP configuration for a peer
type PeerMspConfig struct {
	Id string
}

// PeerVaultConfig represents Vault paths for peer certificates and keys.
// Keys (private) and certs (public) are stored at separate paths.
type PeerVaultConfig struct {
	RoleName string
	// Keys (private) - separate paths for private keys
	MspSigningKeyPath string // Peer signing private key
	TlsKeyPath        string // Peer TLS private key
	CouchdbPath       string // CouchDB credentials (under keys/)
	// Certs (public) - separate paths for certificates + chains
	MspSigningCertPath string // Peer signing cert + chain
	TlsCertPath        string // Peer TLS cert + chain
	// CA root certs for trust stores (public, root only)
	MspRootCaCertPath string // Org Root CA cert for MSP cacerts/
	TlsRootCaCertPath string // TLS Root CA cert for TLS trust store
}

// PeerServiceConfig represents service port configuration for a peer
type PeerServiceConfig struct {
	Type           string
	Port           int
	ChaincodePort  int
	OperationsPort int
}

// GossipConfig represents gossip protocol configuration
type GossipConfig struct {
	Bootstrap        string
	ExternalEndpoint string
}

// CouchdbConfig represents CouchDB connection configuration
type CouchdbConfig struct {
	Address string
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level string
}

// PeerProbeConfig represents Kubernetes probe configuration for peers
type PeerProbeConfig struct {
	InitialDelaySeconds int
	PeriodSeconds       int
	TimeoutSeconds      int
	FailureThreshold    int
}

// NewPeerValues creates PeerValues from PeerConfig and CA configurations.
// CA configs are passed explicitly to allow TLS CAs to be owned by a different owner (operator)
// than the org CAs (which are owned by the org).
// This ensures consistency between Helm values and workflow/activity configuration.
func NewPeerValues(
	peerConfig fabricconfig.PeerConfig,
	orgRootCA fabricconfig.CertificateAuthorityConfig,
	tlsRootCA fabricconfig.CertificateAuthorityConfig,
) *PeerValues {
	couchDbConfig := fabricconfig.NewCouchDbConfig(peerConfig.OwnerConfig, peerConfig.PeerID())

	return &PeerValues{
		Org:       peerConfig.OwnerID(),
		Namespace: peerConfig.KubernetesNamespace(),
		PeerId:    peerConfig.PeerID(),
		Image:     DefaultImageConfig("peer", ""),
		Replicas:  1, // Single instance per peer StatefulSet
		Msp: PeerMspConfig{
			Id: peerConfig.MSPID(),
		},
		Vault: PeerVaultConfig{
			RoleName: peerConfig.AuthRoleName(),
			// Keys (private)
			MspSigningKeyPath: "secret/data/" + peerConfig.VaultSigningKeyRef().GetRef().Fields["path"].GetStringValue(),
			TlsKeyPath:        "secret/data/" + peerConfig.VaultTLSKeyRef().GetRef().Fields["path"].GetStringValue(),
			CouchdbPath:       "secret/data/" + couchDbConfig.VaultAdminRef().GetRef().Fields["path"].GetStringValue(),
			// Certs (public) - contain {certificate, chain}
			MspSigningCertPath: "secret/data/" + peerConfig.VaultSigningCertRef().GetRef().Fields["path"].GetStringValue(),
			TlsCertPath:        "secret/data/" + peerConfig.VaultTLSCertRef().GetRef().Fields["path"].GetStringValue(),
			// CA root certs for trust stores (root CA only)
			MspRootCaCertPath: "secret/data/" + orgRootCA.VaultSigningCertRef().GetRef().Fields["path"].GetStringValue(),
			TlsRootCaCertPath: "secret/data/" + tlsRootCA.VaultSigningCertRef().GetRef().Fields["path"].GetStringValue(),
		},
		Service: PeerServiceConfig{
			Type:           "ClusterIP",
			Port:           peerConfig.Port(),
			ChaincodePort:  peerConfig.ChaincodePort(),
			OperationsPort: peerConfig.OperationsPort(),
		},
		Gossip: GossipConfig{
			Bootstrap:        peerConfig.GossipBootstrapAddress(),
			ExternalEndpoint: peerConfig.GossipExternalEndpoint(),
		},
		Couchdb: CouchdbConfig{
			Address: peerConfig.CouchDbAddress(),
		},
		Persistence: PersistenceConfig{
			Enabled:      true,
			StorageClass: "", // Use default storage class
			AccessModes:  []string{"ReadWriteOnce"},
			Size:         "20Gi",
			MountPath:    "/var/hyperledger/production",
		},
		Logging: LoggingConfig{
			// DEBUG level to diagnose health check failures
			// Can be changed back to INFO once issues are resolved
			Level: "DEBUG",
		},
		Resources: DefaultResourceConfig("peer"),
		// Startup probe: runs first, gives container time to initialize
		// Once startup probe succeeds, liveness takes over
		StartupProbe: PeerProbeConfig{
			InitialDelaySeconds: 5,
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
			FailureThreshold:    30, // 30 * 10s = 5 minutes to start
		},
		// Liveness probe: only runs after startup probe succeeds
		LivenessProbe: PeerProbeConfig{
			InitialDelaySeconds: 0, // Startup probe handles initial delay
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
			FailureThreshold:    6, // 6 failures = 60 seconds before restart
		},
		ReadinessProbe: PeerProbeConfig{
			InitialDelaySeconds: 5,
			PeriodSeconds:       5,
			TimeoutSeconds:      3,
			FailureThreshold:    3,
		},
		CommonLabels: DefaultCommonLabels(),
	}
}
