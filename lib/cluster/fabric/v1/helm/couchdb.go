package helm

import (
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
)

// CouchDbValues represents the Helm values for a CouchDB deployment
type CouchDbValues struct {
	Org            string             `json:"org"`
	Namespace      string             `json:"namespace"`
	PeerId         string             `json:"peerId"`
	Image          ImageConfig        `json:"image"`
	Replicas       int                `json:"replicas"`
	Vault          CouchDbVaultConfig `json:"vault"`
	Service        ServiceConfig      `json:"service"`
	Persistence    PersistenceConfig  `json:"persistence"`
	Resources      ResourceConfig     `json:"resources"`
	LivenessProbe  DBProbeConfig      `json:"livenessProbe"`
	ReadinessProbe DBProbeConfig      `json:"readinessProbe"`
	CommonLabels   CommonLabels       `json:"commonLabels"`
}

// CouchDbVaultConfig represents Vault CSI driver configuration for CouchDB
type CouchDbVaultConfig struct {
	RoleName   string `json:"roleName"`
	SecretPath string `json:"secretPath"`
}

// NewCouchDbValues creates CouchDbValues from CouchDbConfig
// This ensures consistency between Helm values and workflow/activity configuration
func NewCouchDbValues(couchDbConfig fabricconfig.CouchDbConfig) *CouchDbValues {
	return &CouchDbValues{
		Org:       couchDbConfig.OwnerID(),
		Namespace: couchDbConfig.KubernetesNamespace(),
		PeerId:    couchDbConfig.PeerID(),
		Image:     DefaultImageConfig("couchdb", ""),
		Replicas:  1, // Single instance per peer
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
