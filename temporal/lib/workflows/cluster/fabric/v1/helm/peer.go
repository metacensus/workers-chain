package helm

import (
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
)

// PeerValues represents the Helm values for a Fabric peer deployment
type PeerValues struct {
	Org            string
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
	LivenessProbe  PeerProbeConfig
	ReadinessProbe PeerProbeConfig
	CommonLabels   CommonLabels
}

// PeerMspConfig represents MSP configuration for a peer
type PeerMspConfig struct {
	Id string
}

// PeerVaultConfig represents Vault paths for peer certificates
type PeerVaultConfig struct {
	RoleName              string
	MspSigningPath        string
	MspRootCaCertPath     string // Org Root CA cert for cacerts/
	MspIntCaCertPath      string // Org Int CA cert for intermediatecerts/
	TlsPath               string
	TlsRootCaCertPath     string // TLS Root CA cert
	TlsIntCaCertPath      string // TLS Int CA cert
	CouchdbPath           string
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
}

// NewPeerValues creates PeerValues from PeerConfig
// This ensures consistency between Helm values and workflow/activity configuration
func NewPeerValues(peerConfig fabricconfig.PeerConfig) *PeerValues {
	fabricCfg := fabricconfig.NewFabricConfig(peerConfig.OrgConfig)
	couchDbConfig := fabricCfg.CouchDb(peerConfig.PeerID())

	// Get the CA configs for cert paths
	// MSP needs: Root CA in cacerts/, Intermediate CA in intermediatecerts/
	orgRootCA := fabricCfg.CertificateAuthority("org", "root")
	orgIntCA := fabricCfg.CertificateAuthority("org", "v1")
	tlsRootCA := fabricCfg.CertificateAuthority("tls", "root")
	tlsIntCA := fabricCfg.CertificateAuthority("tls", "v1")

	return &PeerValues{
		Org:      peerConfig.OrgID,
		PeerId:   peerConfig.PeerID(),
		Image:    DefaultImageConfig("peer", ""),
		Replicas: 1, // Single instance per peer StatefulSet
		Msp: PeerMspConfig{
			Id: peerConfig.MSPID(),
		},
		Vault: PeerVaultConfig{
			RoleName:          peerConfig.AuthRoleName(),
			MspSigningPath:    "secret/data/" + peerConfig.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			MspRootCaCertPath: "secret/data/" + orgRootCA.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			MspIntCaCertPath:  "secret/data/" + orgIntCA.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			TlsPath:           "secret/data/" + peerConfig.VaultTLSRef().GetRef().Fields["path"].GetStringValue(),
			TlsRootCaCertPath: "secret/data/" + tlsRootCA.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			TlsIntCaCertPath:  "secret/data/" + tlsIntCA.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			CouchdbPath:       "secret/data/" + couchDbConfig.VaultAdminRef().GetRef().Fields["path"].GetStringValue(),
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
			Level: "INFO",
		},
		Resources: DefaultResourceConfig("peer"),
		LivenessProbe: PeerProbeConfig{
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
		},
		ReadinessProbe: PeerProbeConfig{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			TimeoutSeconds:      3,
		},
		CommonLabels: DefaultCommonLabels(),
	}
}

