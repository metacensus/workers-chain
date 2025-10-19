package env

import "github.com/spf13/viper"

const (
	temporalAddressName    = "TEMPORAL_ADDRESS"
	healthCheckAddressName = "HEALTH_CHECK_ADDRESS"
	infraTaskQueueName     = "INFRA_TASK_QUEUE"

	githubRepoName      = "GITHUB_REPO"
	githubDeployKeyName = "GITHUB_DEPLOY_KEY"

	// Vault configuration
	vaultAddressName   = "VAULT_ADDR"
	vaultTokenName     = "VAULT_TOKEN"
	vaultMountPathName = "VAULT_MOUNT_PATH"
	vaultEngineType    = "VAULT_ENGINE_TYPE"

	// Fabric CA configuration
	fabricCAURLName     = "FABRIC_CA_URL"
	fabricCANameName    = "FABRIC_CA_NAME"
	fabricCAHomeDirName = "FABRIC_CA_HOME_DIR"
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("") // or set a prefix if needed

	// Set defaults
	viper.SetDefault(vaultMountPathName, "secret")
	viper.SetDefault(vaultEngineType, "kv")
	viper.SetDefault(fabricCAHomeDirName, "/tmp/fabric-ca-client")
}

func TemporalAddress() string {
	return viper.GetString(temporalAddressName)
}

func HealthCheckAddress() string {
	return viper.GetString(healthCheckAddressName)
}

func InfraTaskQueue() string {
	return viper.GetString(infraTaskQueueName)
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

func VaultToken() string {
	return viper.GetString(vaultTokenName)
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
