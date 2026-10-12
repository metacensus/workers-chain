package helm

import (
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
)

// CertificateAuthorityValues represents the Helm values for a certificate authority deployment
type CertificateAuthorityValues struct {
	Org            string
	Namespace      string
	CertType       string
	CaId           string
	Image          ImageConfig
	Replicas       int
	Database       CADatabaseConfig
	Vault          CAVaultConfig
	Service        ServiceConfig
	CA             CAConfig
	Resources      ResourceConfig
	LivenessProbe  CAProbeConfig
	ReadinessProbe CAProbeConfig
	CommonLabels   CommonLabels
}

// CADatabaseConfig represents database connection configuration for CA
type CADatabaseConfig struct {
	Host string
	Port int
	Name string
}

// CAVaultConfig represents Vault CSI driver configuration for certificate authority.
// Keys (private) and certs (public) are stored at separate paths.
type CAVaultConfig struct {
	RoleName string
	// Keys (private) - separate paths for private keys and credentials
	SigningKeyPath       string // CA signing private key
	TlsKeyPath           string // CA TLS private key
	AdminCredentialsPath string // Bootstrap admin username/password (under keys/)
	DbCredentialsPath    string // Database credentials (under keys/)
	// Certs (public) - separate paths for certificates + chains
	SigningCertPath string // CA signing cert + chain
	TlsCertPath     string // CA TLS cert + chain
}

// CAConfig represents CA-specific configuration
type CAConfig struct {
	CommonName         string
	OrganizationalUnit string
}

// CAProbeConfig represents probe configuration for CA deployments
type CAProbeConfig struct {
	InitialDelaySeconds int
	PeriodSeconds       int
	TimeoutSeconds      int
}

// NewCertificateAuthorityValues creates CertificateAuthorityValues from CertificateAuthorityConfig
// This ensures consistency between Helm values and workflow/activity configuration
func NewCertificateAuthorityValues(caConfig fabricconfig.CertificateAuthorityConfig) *CertificateAuthorityValues {
	// Get certificate store config for database connection
	certStoreConfig := fabricconfig.NewCertificateStoreConfig(caConfig.OwnerConfig)

	return &CertificateAuthorityValues{
		Org:       caConfig.OwnerID(),
		Namespace: caConfig.KubernetesNamespace(),
		CertType:  caConfig.CertType(),
		CaId:      caConfig.CAID(),
		Image: ImageConfig{
			Repository: "hyperledger/fabric-ca",
			Tag:        "1.5",
			PullPolicy: "IfNotPresent",
		},
		Replicas: 1, // Root CA typically runs as single instance
		Database: CADatabaseConfig{
			Host: certStoreConfig.ServiceHost(),
			Port: certStoreConfig.Port(),
			Name: caConfig.DatabaseName(),
		},
		Vault: CAVaultConfig{
			RoleName: caConfig.AuthRoleName(),
			// Keys (private)
			SigningKeyPath:       "secret/data/" + caConfig.VaultSigningKeyRef().GetRef().Fields["path"].GetStringValue(),
			TlsKeyPath:           "secret/data/" + caConfig.VaultTLSKeyRef().GetRef().Fields["path"].GetStringValue(),
			AdminCredentialsPath: "secret/data/" + caConfig.VaultBootstrapAdminCredentialsRef().GetRef().Fields["path"].GetStringValue(),
			DbCredentialsPath:    "secret/data/" + caConfig.VaultDatabaseUserRef().GetRef().Fields["path"].GetStringValue(),
			// Certs (public) - contain {certificate, chain}
			SigningCertPath: "secret/data/" + caConfig.VaultSigningCertRef().GetRef().Fields["path"].GetStringValue(),
			TlsCertPath:     "secret/data/" + caConfig.VaultTLSCertRef().GetRef().Fields["path"].GetStringValue(),
		},
		Service: ServiceConfig{
			Type: "ClusterIP",
			Port: 7054, // Standard Fabric CA port
		},
		CA: CAConfig{
			CommonName:         caConfig.ComponentName(),
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		Resources: ResourceConfig{
			Requests: ResourceRequirements{
				CPU:    "100m",
				Memory: "256Mi",
			},
			Limits: ResourceRequirements{
				CPU:    "500m",
				Memory: "512Mi",
			},
		},
		LivenessProbe: CAProbeConfig{
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
		},
		ReadinessProbe: CAProbeConfig{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			TimeoutSeconds:      3,
		},
		CommonLabels: DefaultCommonLabels(),
	}
}
