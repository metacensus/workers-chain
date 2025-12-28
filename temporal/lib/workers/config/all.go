package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

const (
	temporalAddressName    = "TEMPORAL_ADDRESS"
	temporalNamespaceName  = "TEMPORAL_NAMESPACE"
	temporalTaskQueueName  = "TEMPORAL_TASK_QUEUE"
	healthCheckAddressName = "HEALTH_CHECK_ADDRESS"

	vaultAddressName   = "VAULT_ADDR"
	vaultRoleName      = "VAULT_ROLE"
	vaultMountPathName = "VAULT_MOUNT_PATH"
	vaultEngineType    = "VAULT_ENGINE_TYPE"

	kubeconfigPathName = "KUBECONFIG"
	orgIDName          = "ORG_ID"
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	// Set defaults for non-worker-specific config
	viper.SetDefault(vaultMountPathName, "secret")
	viper.SetDefault(vaultEngineType, "kv")
}

// TemporalConfig holds Temporal cluster connection settings
type TemporalConfig struct {
	Address   string
	Namespace string
	TaskQueue string
}

var (
	temporalGlobal *TemporalConfig
	temporalOnce   sync.Once
	temporalMu     sync.RWMutex
)

func InitializeTemporal() {
	temporalOnce.Do(func() {
		temporalGlobal = &TemporalConfig{
			Address:   viper.GetString(temporalAddressName),
			Namespace: viper.GetString(temporalNamespaceName),
			TaskQueue: viper.GetString(temporalTaskQueueName),
		}
	})
}

func GetTemporal() *TemporalConfig {
	temporalMu.RLock()
	defer temporalMu.RUnlock()
	if temporalGlobal == nil {
		panic("TemporalConfig not initialized - call InitializeTemporal() at startup")
	}
	return temporalGlobal
}

// VaultConfig holds Vault client settings
type VaultConfig struct {
	Address    string
	Role       string
	MountPath  string
	EngineType string
}

var (
	vaultGlobal *VaultConfig
	vaultOnce   sync.Once
	vaultMu     sync.RWMutex
)

func InitializeVault() {
	vaultOnce.Do(func() {
		vaultGlobal = &VaultConfig{
			Address:    viper.GetString(vaultAddressName),
			Role:       viper.GetString(vaultRoleName),
			MountPath:  viper.GetString(vaultMountPathName),
			EngineType: viper.GetString(vaultEngineType),
		}
	})
}

func GetVault() *VaultConfig {
	vaultMu.RLock()
	defer vaultMu.RUnlock()
	if vaultGlobal == nil {
		panic("VaultConfig not initialized - call InitializeVault() at startup")
	}
	return vaultGlobal
}

// WorkerConfig holds worker runtime settings
type WorkerConfig struct {
	HealthCheckAddress string
}

var (
	workerGlobal *WorkerConfig
	workerOnce   sync.Once
	workerMu     sync.RWMutex
)

func InitializeWorker() {
	workerOnce.Do(func() {
		workerGlobal = &WorkerConfig{
			HealthCheckAddress: viper.GetString(healthCheckAddressName),
		}
	})
}

func GetWorker() *WorkerConfig {
	workerMu.RLock()
	defer workerMu.RUnlock()
	if workerGlobal == nil {
		panic("WorkerConfig not initialized - call InitializeWorker() at startup")
	}
	return workerGlobal
}

// KubernetesConfig holds Kubernetes client settings
type KubernetesConfig struct {
	KubeconfigPath string
}

var (
	kubernetesGlobal *KubernetesConfig
	kubernetesOnce   sync.Once
	kubernetesMu     sync.RWMutex
)

func InitializeKubernetes() {
	kubernetesOnce.Do(func() {
		kubernetesGlobal = &KubernetesConfig{
			KubeconfigPath: viper.GetString(kubeconfigPathName),
		}
	})
}

func GetKubernetes() *KubernetesConfig {
	kubernetesMu.RLock()
	defer kubernetesMu.RUnlock()
	if kubernetesGlobal == nil {
		panic("KubernetesConfig not initialized - call InitializeKubernetes() at startup")
	}
	return kubernetesGlobal
}

// OrgConfig holds fabric organization identity and provides derived values
type OrgConfig struct {
	OrgID string
}

var (
	orgGlobal *OrgConfig
	orgOnce   sync.Once
	orgMu     sync.RWMutex
)

func InitializeOrg() error {
	var err error
	orgOnce.Do(func() {
		orgID := viper.GetString(orgIDName)
		if orgID == "" {
			err = fmt.Errorf("ORG_ID environment variable is required")
			return
		}
		orgGlobal = &OrgConfig{OrgID: orgID}
	})
	return err
}

func GetOrg() *OrgConfig {
	orgMu.RLock()
	defer orgMu.RUnlock()
	if orgGlobal == nil {
		panic("OrgConfig not initialized - call InitializeOrg() at startup")
	}
	return orgGlobal
}

// OrgConfig methods for deriving per-org values

func (c OrgConfig) KubernetesNamespace() string {
	return "fabric-" + c.OrgID
}

func (c OrgConfig) TemporalNamespace() string {
	return "fabric-" + c.OrgID
}

func (c OrgConfig) VaultBasePath() string {
	return "fabric/" + c.OrgID
}

func (c OrgConfig) ServiceAccountName() string {
	return "fabric-" + c.OrgID
}

func (c OrgConfig) WorkerDeploymentName() string {
	return "fabric-" + c.OrgID + "-worker"
}

// Module exports config providers for infra worker (no OrgConfig)
var Module = fx.Module("config",
	fx.Provide(
		GetTemporal,
		GetVault,
		GetWorker,
		GetKubernetes,
	),
)

// FabricModule exports config providers for fabric worker (includes OrgConfig)
var FabricModule = fx.Module("config",
	fx.Provide(
		GetTemporal,
		GetVault,
		GetWorker,
		GetKubernetes,
		GetOrg,
	),
)
