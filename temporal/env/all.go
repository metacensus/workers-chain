package env

import "github.com/spf13/viper"

const (
	temporalAddressName    = "TEMPORAL_ADDRESS"
	healthCheckAddressName = "HEALTH_CHECK_ADDRESS"
	infraTaskQueueName     = "INFRA_TASK_QUEUE"

	githubRepoName      = "GITHUB_REPO"
	githubDeployKeyName = "GITHUB_DEPLOY_KEY"
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("") // or set a prefix if needed
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
