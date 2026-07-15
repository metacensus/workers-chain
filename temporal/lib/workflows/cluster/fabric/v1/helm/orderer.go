package helm

import (
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
)

// OrdererValues represents the Helm values for a Fabric orderer deployment
type OrdererValues struct {
	Org            string
	Namespace      string
	OrdererId      string
	Image          ImageConfig
	Replicas       int
	Msp            OrdererMspConfig
	Vault          OrdererVaultConfig
	Service        OrdererServiceConfig
	Persistence    PersistenceConfig
	Logging        LoggingConfig
	Resources      ResourceConfig
	LivenessProbe  OrdererProbeConfig
	ReadinessProbe OrdererProbeConfig
	CommonLabels   CommonLabels
}

// OrdererMspConfig represents MSP configuration for an orderer
type OrdererMspConfig struct {
	Id string
}

// OrdererVaultConfig represents Vault paths for orderer certificates and keys.
// Keys (private) and certs (public) are stored at separate paths.
type OrdererVaultConfig struct {
	RoleName string
	// Keys (private) - separate paths for private keys
	MspSigningKeyPath string // Orderer signing private key
	TlsKeyPath        string // Orderer TLS private key
	// Certs (public) - separate paths for certificates + chains
	MspSigningCertPath string // Orderer signing cert + chain
	TlsCertPath        string // Orderer TLS cert + chain
	// CA root certs for trust stores (public, root only)
	MspRootCaCertPath string // Org Root CA cert for MSP cacerts/
	TlsRootCaCertPath string // TLS Root CA cert for TLS trust store
}

// OrdererServiceConfig represents service port configuration for an orderer
type OrdererServiceConfig struct {
	Type           string
	Port           int
	AdminPort      int
	OperationsPort int
}

// OrdererProbeConfig represents Kubernetes probe configuration for orderers
type OrdererProbeConfig struct {
	InitialDelaySeconds int
	PeriodSeconds       int
	TimeoutSeconds      int
}

// NewOrdererValues creates OrdererValues from OrdererConfig and CA configurations.
// The CA configs are passed explicitly to support shared TLS CAs owned by the operator
// while org CAs remain owned by the organization.
//
// Parameters:
//   - ordererConfig: Configuration for the orderer itself
//   - orgRootCA: Org Root CA config (for MSP cacerts/)
//   - tlsRootCA: TLS Root CA config (for TLS trust store)
func NewOrdererValues(
	ordererConfig fabricconfig.OrdererConfig,
	orgRootCA fabricconfig.CertificateAuthorityConfig,
	tlsRootCA fabricconfig.CertificateAuthorityConfig,
) *OrdererValues {
	return &OrdererValues{
		Org:       ordererConfig.OwnerID(),
		Namespace: ordererConfig.KubernetesNamespace(),
		OrdererId: ordererConfig.OrdererID(),
		Image:     DefaultImageConfig("orderer", ""),
		Replicas:  1, // Single instance per orderer StatefulSet
		Msp: OrdererMspConfig{
			Id: ordererConfig.MSPID(),
		},
		Vault: OrdererVaultConfig{
			RoleName: ordererConfig.AuthRoleName(),
			// Keys (private)
			MspSigningKeyPath: "secret/data/" + ordererConfig.VaultSigningKeyRef().GetRef().Fields["path"].GetStringValue(),
			TlsKeyPath:        "secret/data/" + ordererConfig.VaultTLSKeyRef().GetRef().Fields["path"].GetStringValue(),
			// Certs (public) - contain {certificate, chain}
			MspSigningCertPath: "secret/data/" + ordererConfig.VaultSigningCertRef().GetRef().Fields["path"].GetStringValue(),
			TlsCertPath:        "secret/data/" + ordererConfig.VaultTLSCertRef().GetRef().Fields["path"].GetStringValue(),
			// CA root certs for trust stores (root CA only)
			MspRootCaCertPath: "secret/data/" + orgRootCA.VaultSigningCertRef().GetRef().Fields["path"].GetStringValue(),
			TlsRootCaCertPath: "secret/data/" + tlsRootCA.VaultSigningCertRef().GetRef().Fields["path"].GetStringValue(),
		},
		Service: OrdererServiceConfig{
			Type:           "ClusterIP",
			Port:           ordererConfig.Port(),
			AdminPort:      ordererConfig.AdminPort(),
			OperationsPort: ordererConfig.OperationsPort(),
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
		Resources: DefaultResourceConfig("orderer"),
		LivenessProbe: OrdererProbeConfig{
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
		},
		ReadinessProbe: OrdererProbeConfig{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			TimeoutSeconds:      3,
		},
		CommonLabels: DefaultCommonLabels(),
	}
}
