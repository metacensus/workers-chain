package helm

import (
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
)

// CouchDbValues represents the Helm values for a CouchDB deployment
type CouchDbValues struct {
	Org            string
	PeerId         string
	Image          ImageConfig
	Replicas       int
	Vault          CouchDbVaultConfig
	Service        ServiceConfig
	Persistence    PersistenceConfig
	Resources      ResourceConfig
	LivenessProbe  DBProbeConfig
	ReadinessProbe DBProbeConfig
	CommonLabels   CommonLabels
}

// CouchDbVaultConfig represents Vault CSI driver configuration for CouchDB
type CouchDbVaultConfig struct {
	RoleName   string
	SecretPath string
}

// NewCouchDbValues creates CouchDbValues from CouchDbConfig
// This ensures consistency between Helm values and workflow/activity configuration
func NewCouchDbValues(couchDbConfig fabricconfig.CouchDbConfig) *CouchDbValues {
	return &CouchDbValues{
		Org:      couchDbConfig.OwnerID(),
		PeerId:   couchDbConfig.PeerID(),
		Image:    DefaultImageConfig("couchdb", ""),
		Replicas: 1, // Single instance per peer
		Vault: CouchDbVaultConfig{
			RoleName:   couchDbConfig.AuthRoleName(),
			SecretPath: "secret/data/" + couchDbConfig.VaultAdminRef().GetRef().Fields["path"].GetStringValue(),
		},
		Service: DefaultServiceConfig("couchdb"),
		Persistence: PersistenceConfig{
			Enabled:      true,
			StorageClass: "", // Use default storage class
			AccessModes:  []string{"ReadWriteOnce"},
			Size:         "10Gi",
			MountPath:    "/opt/couchdb/data",
		},
		Resources: DefaultResourceConfig("couchdb"),
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

