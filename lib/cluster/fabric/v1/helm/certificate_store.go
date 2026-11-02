package helm

import (
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
)

// CertificateStoreValues represents the Helm values for a certificate store (PostgreSQL) deployment
type CertificateStoreValues struct {
	Org            string               `json:"org"`
	Namespace      string               `json:"namespace"`
	Image          ImageConfig          `json:"image"`
	Replicas       int                  `json:"replicas"`
	Database       DatabaseConfig       `json:"database"`
	Vault          CertStoreVaultConfig `json:"vault"`
	Service        ServiceConfig        `json:"service"`
	Persistence    PersistenceConfig    `json:"persistence"`
	Resources      ResourceConfig       `json:"resources"`
	LivenessProbe  DBProbeConfig        `json:"livenessProbe"`
	ReadinessProbe DBProbeConfig        `json:"readinessProbe"`
	CommonLabels   CommonLabels         `json:"commonLabels"`
}

// DatabaseConfig represents PostgreSQL database configuration
type DatabaseConfig struct {
	Name string `json:"name"`
}

// CertStoreVaultConfig represents Vault CSI driver configuration for certificate store
type CertStoreVaultConfig struct {
	RoleName   string `json:"roleName"`
	SecretPath string `json:"secretPath"`
}

// PersistenceConfig represents persistent volume configuration
type PersistenceConfig struct {
	Enabled      bool     `json:"enabled"`
	StorageClass string   `json:"storageClass"`
	AccessModes  []string `json:"accessModes"`
	Size         string   `json:"size"`
	MountPath    string   `json:"mountPath"`
}

// DBProbeConfig represents database probe configuration
type DBProbeConfig struct {
	InitialDelaySeconds int `json:"initialDelaySeconds"`
	PeriodSeconds       int `json:"periodSeconds"`
	TimeoutSeconds      int `json:"timeoutSeconds"`
}

// NewCertificateStoreValues creates CertificateStoreValues from CertificateStoreConfig
// This ensures consistency between Helm values and workflow/activity configuration
func NewCertificateStoreValues(certStoreConfig fabricconfig.CertificateStoreConfig) *CertificateStoreValues {
	return &CertificateStoreValues{
		Org:       certStoreConfig.OwnerID(),
		Namespace: certStoreConfig.KubernetesNamespace(),
		Image: ImageConfig{
			Repository: "postgres",
			Tag:        "16-alpine",
			PullPolicy: "IfNotPresent",
		},
		Replicas: 1, // Start with single instance; can scale later
		Database: DatabaseConfig{
			Name: certStoreConfig.AdminDatabase(),
		},
		Vault: CertStoreVaultConfig{
			RoleName:   certStoreConfig.AuthRoleName(),
			SecretPath: "secret/data/" + certStoreConfig.VaultAdminRef().GetRef().Fields["path"].GetStringValue(),
		},
		Service: ServiceConfig{
			Type: "ClusterIP",
			Port: certStoreConfig.Port(),
		},
		Persistence: PersistenceConfig{
			Enabled:      true,
			StorageClass: "", // Use default storage class
			AccessModes:  []string{"ReadWriteOnce"},
			Size:         "10Gi",
			MountPath:    "/var/lib/postgresql/data",
		},
		Resources: ResourceConfig{
			Requests: ResourceRequirements{
				CPU:    "250m",
				Memory: "512Mi",
			},
			Limits: ResourceRequirements{
				CPU:    "1000m",
				Memory: "2Gi",
			},
		},
		LivenessProbe: DBProbeConfig{
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
		},
		ReadinessProbe: DBProbeConfig{
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			TimeoutSeconds:      3,
		},
		CommonLabels: DefaultCommonLabels(),
	}
}
