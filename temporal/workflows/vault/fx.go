package vault

import (
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides vault workflows and activities for dependency injection
var Module = fx.Options(
	// Provide Vault configuration
	fx.Provide(GetVaultConfig),

	// Provide Vault client
	fx.Provide(NewVaultClient),

	// Provide Activities
	fx.Provide(NewActivities),

	// Provide Workflows
	fx.Provide(NewWorkflows),
)

// RegisterAll registers both workflows and activities with the worker
func RegisterAll(w worker.Worker, workflows *Workflows, activities *Activities) {
	RegisterWorkflows(w, workflows)
	RegisterActivities(w, activities)
}

// RegisterWorkflows registers all vault workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows *Workflows) {
	w.RegisterWorkflow(workflows.StoreSecretWorkflow)
	w.RegisterWorkflow(workflows.FetchSecretWorkflow)
}

// RegisterActivities registers all vault activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) {
	w.RegisterActivity(activities.StoreSecretActivity)
	w.RegisterActivity(activities.FetchSecretActivity)
}
