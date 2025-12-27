package env

import "github.com/spf13/viper"

const (
	temporalAddressName    = "TEMPORAL_ADDRESS"
	temporalNamespaceName  = "TEMPORAL_NAMESPACE"
	healthCheckAddressName = "HEALTH_CHECK_ADDRESS"
	taskQueueName          = "TASK_QUEUE"

	githubRepoName      = "GITHUB_REPO"
	githubDeployKeyName = "GITHUB_DEPLOY_KEY"

	// Vault configuration
	vaultAddressName   = "VAULT_ADDR"
	vaultRoleName      = "VAULT_ROLE"
	vaultMountPathName = "VAULT_MOUNT_PATH"
	vaultEngineType    = "VAULT_ENGINE_TYPE"

	// Fabric CA configuration
	fabricCAURLName     = "FABRIC_CA_URL"
	fabricCANameName    = "FABRIC_CA_NAME"
	fabricCAHomeDirName = "FABRIC_CA_HOME_DIR"

	// Kubernetes configuration
	kubeconfigPathName = "KUBECONFIG"

	// Fabric organization configuration
	orgIDName = "ORG_ID"
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("") // or set a prefix if needed

	// Set defaults
	viper.SetDefault(temporalNamespaceName, "infra")
	viper.SetDefault(vaultRoleName, "infra-worker")
	viper.SetDefault(vaultMountPathName, "secret")
	viper.SetDefault(vaultEngineType, "kv")
	viper.SetDefault(fabricCAHomeDirName, "/tmp/fabric-ca-client")
}

func TemporalAddress() string {
	return viper.GetString(temporalAddressName)
}

func TemporalNamespace() string {
	return viper.GetString(temporalNamespaceName)
}

func HealthCheckAddress() string {
	return viper.GetString(healthCheckAddressName)
}

func TaskQueue() string {
	return viper.GetString(taskQueueName)
}

func GithubRepo() string {
	return viper.GetString(githubRepoName)
}

func GithubDeployKey() string {
	return viper.GetString(githubDeployKeyName)
}

// Vault configuration
func VaultAddress() string {
	return viper.GetString(vaultAddressName)
}

func VaultRole() string {
	return viper.GetString(vaultRoleName)
}

func VaultMountPath() string {
	return viper.GetString(vaultMountPathName)
}

func VaultEngineType() string {
	return viper.GetString(vaultEngineType)
}

// Fabric CA configuration
func FabricCAURL() string {
	return viper.GetString(fabricCAURLName)
}

func FabricCAName() string {
	return viper.GetString(fabricCANameName)
}

func FabricCAHomeDir() string {
	return viper.GetString(fabricCAHomeDirName)
}

// Kubernetes configuration
func KubeconfigPath() string {
	return viper.GetString(kubeconfigPathName)
}

// Fabric organization configuration
func OrgID() string {
	return viper.GetString(orgIDName)
}
