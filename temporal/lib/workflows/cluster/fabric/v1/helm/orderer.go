package helm

import (
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
)

// OrdererValues represents the Helm values for a Fabric orderer deployment
type OrdererValues struct {
	Org            string
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

// OrdererVaultConfig represents Vault paths for orderer certificates
type OrdererVaultConfig struct {
	RoleName          string
	MspSigningPath    string
	MspRootCaCertPath string // Org Root CA cert for cacerts/
	MspIntCaCertPath  string // Org Int CA cert for intermediatecerts/
	TlsPath           string
	TlsRootCaCertPath string // TLS Root CA cert
	TlsIntCaCertPath  string // TLS Int CA cert
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
//   - orgIntCA: Org Intermediate CA config (for MSP intermediatecerts/)
//   - tlsRootCA: TLS Root CA config (for TLS trust chain)
//   - tlsIntCA: TLS Intermediate CA config (for TLS certificates)
func NewOrdererValues(
	ordererConfig fabricconfig.OrdererConfig,
	orgRootCA fabricconfig.CertificateAuthorityConfig,
	orgIntCA fabricconfig.CertificateAuthorityConfig,
	tlsRootCA fabricconfig.CertificateAuthorityConfig,
	tlsIntCA fabricconfig.CertificateAuthorityConfig,
) *OrdererValues {
	return &OrdererValues{
		Org:       ordererConfig.OwnerID(),
		OrdererId: ordererConfig.OrdererID(),
		Image:     DefaultImageConfig("orderer", ""),
		Replicas:  1, // Single instance per orderer StatefulSet
		Msp: OrdererMspConfig{
			Id: ordererConfig.MSPID(),
		},
		Vault: OrdererVaultConfig{
			RoleName:          ordererConfig.AuthRoleName(),
			MspSigningPath:    "secret/data/" + ordererConfig.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			MspRootCaCertPath: "secret/data/" + orgRootCA.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			MspIntCaCertPath:  "secret/data/" + orgIntCA.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			TlsPath:           "secret/data/" + ordererConfig.VaultTLSRef().GetRef().Fields["path"].GetStringValue(),
			TlsRootCaCertPath: "secret/data/" + tlsRootCA.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			TlsIntCaCertPath:  "secret/data/" + tlsIntCA.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
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
