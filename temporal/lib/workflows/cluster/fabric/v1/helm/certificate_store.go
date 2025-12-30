package helm

import "github.com/metacensus/infra/temporal/lib/workers/config"

// CertificateStoreValues represents the Helm values for a certificate store (PostgreSQL) deployment
type CertificateStoreValues struct {
	Org            string
	Image          ImageConfig
	Replicas       int
	Database       DatabaseConfig
	Service        ServiceConfig
	Persistence    PersistenceConfig
	Resources      ResourceConfig
	LivenessProbe  DBProbeConfig
	ReadinessProbe DBProbeConfig
	CommonLabels   CommonLabels
}

// DatabaseConfig represents PostgreSQL database configuration
type DatabaseConfig struct {
	Name     string
	User     string
	Password string
}

// PersistenceConfig represents persistent volume configuration
type PersistenceConfig struct {
	Enabled      bool
	StorageClass string
	AccessModes  []string
	Size         string
	MountPath    string
}

// DBProbeConfig represents database probe configuration
type DBProbeConfig struct {
	InitialDelaySeconds int
	PeriodSeconds       int
	TimeoutSeconds      int
}

// NewCertificateStoreValues creates CertificateStoreValues with sensible defaults for PostgreSQL
func NewCertificateStoreValues(orgCfg config.OrgConfig) *CertificateStoreValues {
	return &CertificateStoreValues{
		Org: orgCfg.OrgID,
		Image: ImageConfig{
			Repository: "postgres",
			Tag:        "16-alpine",
			PullPolicy: "IfNotPresent",
		},
		Replicas: 1, // Start with single instance; can scale later
		Database: DatabaseConfig{
			Name:     "fabricca",
			User:     "fabricca",
			Password: "changeme", // TODO: Generate secure password via crypto workflow
		},
		Service: ServiceConfig{
			Type: "ClusterIP",
			Port: 5432,
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
