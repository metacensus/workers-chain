package helm

import (
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
)

// CertificateAuthorityValues represents the Helm values for a certificate authority deployment
type CertificateAuthorityValues struct {
	Org            string
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

// CAVaultConfig represents Vault CSI driver configuration for certificate authority
type CAVaultConfig struct {
	RoleName             string
	KeyPairPath          string
	CertificatePath      string
	AdminCredentialsPath string
	DbCredentialsPath    string
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
	fabricCfg := fabricconfig.NewFabricConfig(caConfig.OrgConfig)
	certStoreConfig := fabricCfg.CertificateStore()

	return &CertificateAuthorityValues{
		Org:      caConfig.OrgID,
		CertType: caConfig.CertType(),
		CaId:     caConfig.CAID(),
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
			RoleName:             caConfig.AuthRoleName(),
			KeyPairPath:          "secret/data/" + caConfig.VaultServerRef().GetRef().Fields["path"].GetStringValue(),
			CertificatePath:      "secret/data/" + caConfig.VaultServerRef().GetRef().Fields["path"].GetStringValue(),
			AdminCredentialsPath: "secret/data/" + caConfig.VaultBootstrapAdminRef().GetRef().Fields["path"].GetStringValue(),
			DbCredentialsPath:    "secret/data/" + caConfig.VaultDatabaseUserRef().GetRef().Fields["path"].GetStringValue(),
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
