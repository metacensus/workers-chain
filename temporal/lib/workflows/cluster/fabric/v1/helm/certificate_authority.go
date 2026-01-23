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
	SigningPath          string // Path to CA signing keypair + cert
	SigningChainPath     string // Path to parent CA chain (intermediate CAs only)
	TlsPath              string // Path to CA TLS keypair + cert
	AdminCredentialsPath string // Path to bootstrap admin username/password
	DbCredentialsPath    string // Path to database credentials
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

	// Determine if this is an intermediate CA that needs a chain file
	var signingChainPath string
	if !caConfig.IsRoot() {
		// Intermediate CAs need the parent chain for x509 verification
		signingChainPath = "secret/data/" + caConfig.VaultSigningChainRef().GetRef().Fields["path"].GetStringValue()
	}

	return &CertificateAuthorityValues{
		Org:      caConfig.OwnerID(),
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
			SigningPath:          "secret/data/" + caConfig.VaultSigningRef().GetRef().Fields["path"].GetStringValue(),
			SigningChainPath:     signingChainPath, // Empty for root CAs, set for intermediates
			TlsPath:              "secret/data/" + caConfig.VaultTLSRef().GetRef().Fields["path"].GetStringValue(),
			AdminCredentialsPath: "secret/data/" + caConfig.VaultBootstrapAdminCredentialsRef().GetRef().Fields["path"].GetStringValue(),
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
