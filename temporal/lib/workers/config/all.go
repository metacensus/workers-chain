package config

import (
	"sync"

	"github.com/spf13/viper"
)

const (
	temporalAddressName      = "TEMPORAL_ADDRESS"
	temporalNamespaceName    = "TEMPORAL_NAMESPACE"
	temporalTaskQueueName    = "TEMPORAL_TASK_QUEUE"
	temporalTLSCAPathName    = "TEMPORAL_TLS_CA_PATH"
	temporalTLSCertPathName  = "TEMPORAL_TLS_CERT_PATH"
	temporalTLSKeyPathName   = "TEMPORAL_TLS_KEY_PATH"
	temporalTLSServerNameKey = "TEMPORAL_TLS_SERVER_NAME"
	healthCheckAddressName   = "HEALTH_CHECK_ADDRESS"

	vaultAddressName   = "VAULT_ADDR"
	vaultRoleName      = "VAULT_ROLE"
	vaultMountPathName = "VAULT_MOUNT_PATH"
	vaultEngineType    = "VAULT_ENGINE_TYPE"

	kubeconfigPathName = "KUBECONFIG"
	operatorIDName     = "OPERATOR_ID"
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
	// TLS client identity (required when frontend requireClientAuth is enabled).
	// Empty paths keep plaintext dial for local unit tests.
	TLSCAPath     string
	TLSCertPath   string
	TLSKeyPath    string
	TLSServerName string
}

var (
	temporalGlobal *TemporalConfig
	temporalOnce   sync.Once
	temporalMu     sync.RWMutex
)

func InitializeTemporal() {
	temporalOnce.Do(func() {
		temporalGlobal = &TemporalConfig{
			Address:       viper.GetString(temporalAddressName),
			Namespace:     viper.GetString(temporalNamespaceName),
			TaskQueue:     viper.GetString(temporalTaskQueueName),
			TLSCAPath:     viper.GetString(temporalTLSCAPathName),
			TLSCertPath:   viper.GetString(temporalTLSCertPathName),
			TLSKeyPath:    viper.GetString(temporalTLSKeyPathName),
			TLSServerName: viper.GetString(temporalTLSServerNameKey),
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

// OrgConfig holds fabric organization identity and provides derived values.
// Constructed via LoadOrgConfig (from ORG_ID at worker startup) or NewOrgConfig
// (from a typed Owner input); there is no global singleton.
type OrgConfig struct {
	OrgID string
}

// ============= Kubernetes Infrastructure =============

func (c OrgConfig) KubernetesNamespace() string {
	return ShardKubernetesNamespacePrefix + c.OrgID
}

func (c OrgConfig) WorkerServiceAccountName() string {
	return "fabric-worker-" + c.OrgID
}

func (c OrgConfig) WorkerDeploymentName() string {
	return "fabric-worker-" + c.OrgID
}

// ============= Temporal Orchestration =============

func (c OrgConfig) TemporalNamespace() string {
	return ShardKubernetesNamespacePrefix + c.OrgID
}

func (c OrgConfig) NexusEndpointBootstrapShard() string {
	return "fabric-shard-" + c.OrgID + "-BootstrapShard"
}

// ============= Vault Security/RBAC =============

func (c OrgConfig) VaultBasePath() string {
	return shardVaultPrefix + c.OrgID
}

func (c OrgConfig) WorkerVaultPolicyName() string {
	return "fabric-shard-" + c.OrgID + "-worker"
}

func (c OrgConfig) WorkerVaultAuthRole() string {
	return "fabric-shard-" + c.OrgID + "-worker"
}

// VaultSecretPath constructs a full Vault KV v2 secret path.
// Example: VaultSecretPath("database/cert-store/credentials") → "secret/fabric/shard/org1/database/cert-store/credentials"
func (c OrgConfig) VaultSecretPath(subpath string) string {
	if subpath == "" {
		return "secret/" + c.VaultBasePath()
	}
	return "secret/" + c.VaultBasePath() + "/" + subpath
}
