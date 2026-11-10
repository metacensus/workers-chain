package helm

import (
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
)

// OrdererValues represents the Helm values for a Fabric orderer deployment
type OrdererValues struct {
	Org            string               `json:"org"`
	Namespace      string               `json:"namespace"`
	OrdererId      string               `json:"ordererId"`
	Image          ImageConfig          `json:"image"`
	Replicas       int                  `json:"replicas"`
	Msp            OrdererMspConfig     `json:"msp"`
	Vault          OrdererVaultConfig   `json:"vault"`
	Service        OrdererServiceConfig `json:"service"`
	Persistence    PersistenceConfig    `json:"persistence"`
	Logging        LoggingConfig        `json:"logging"`
	Resources      ResourceConfig       `json:"resources"`
	LivenessProbe  OrdererProbeConfig   `json:"livenessProbe"`
	ReadinessProbe OrdererProbeConfig   `json:"readinessProbe"`
	CommonLabels   CommonLabels         `json:"commonLabels"`
}

// OrdererMspConfig represents MSP configuration for an orderer
type OrdererMspConfig struct {
	Id string `json:"id"`
}

// OrdererVaultConfig represents Vault paths for orderer certificates and keys.
// Keys (private) and certs (public) are stored at separate paths.
type OrdererVaultConfig struct {
	RoleName string `json:"roleName"`
	// Keys (private) - separate paths for private keys
	MspSigningKeyPath string `json:"mspSigningKeyPath"` // Orderer signing private key
	TlsKeyPath        string `json:"tlsKeyPath"`        // Orderer TLS private key
	// Certs (public) - separate paths for certificates + chains
	MspSigningCertPath string `json:"mspSigningCertPath"` // Orderer signing cert + chain
	TlsCertPath        string `json:"tlsCertPath"`        // Orderer TLS cert + chain
	// CA root certs for trust stores (public, root only)
	MspRootCaCertPath string `json:"mspRootCaCertPath"` // Org Root CA cert for MSP cacerts/
	TlsRootCaCertPath string `json:"tlsRootCaCertPath"` // TLS Root CA cert for TLS trust store
}

// OrdererServiceConfig represents service port configuration for an orderer
type OrdererServiceConfig struct {
	Type           string `json:"type"`
	Port           int    `json:"port"`
	AdminPort      int    `json:"adminPort"`
	OperationsPort int    `json:"operationsPort"`
}

// OrdererProbeConfig represents Kubernetes probe configuration for orderers
type OrdererProbeConfig struct {
	InitialDelaySeconds int `json:"initialDelaySeconds"`
	PeriodSeconds       int `json:"periodSeconds"`
	TimeoutSeconds      int `json:"timeoutSeconds"`
}

// NewOrdererValues creates OrdererValues from OrdererConfig and CA configurations.
// The CA configs are passed explicitly to support shared TLS CAs owned by the operator
// while org CAs remain owned by the organization.
//
// Parameters:
//   - ordererConfig: Configuration for the orderer itself
//   - orgRootCA: Org Root CA config (for MSP cacerts/)
func NewOrdererValues(
	ordererConfig fabricconfig.OrdererConfig,
	orgRootCA fabricconfig.CertificateAuthorityConfig,
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
			TlsRootCaCertPath: ordererConfig.TLSRootCA().VaultSigningCertDataPath(), // the path its Vault policy grants
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
