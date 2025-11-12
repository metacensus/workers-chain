package helm

import (
	"fmt"
	"os"

	pb "github.com/metacensus/infra/temporal/gen/workflows/helm/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides RepoClient, activities, and workflow registration
var Module = fx.Options(
	// Provide repository client (local filesystem-based for now)
	fx.Provide(provideRepoClient),

	// Provide activities
	fx.Provide(NewActivities),
)

// provideRepoClient creates a RepoClient from environment configuration
func provideRepoClient() (RepoClient, error) {
	chartDir := os.Getenv("HELM_CHART_DIR")
	if chartDir == "" {
		chartDir = "/helm" // Default to /helm (from Docker COPY)
	}

	// For now, use local filesystem repo. Future: read from env to support Helm repos, OCI registries, etc.
	return NewLocalRepoClient(chartDir)
}

// RegisterActivities registers Helm activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) error {
	if activities == nil {
		return fmt.Errorf("helm activities must not be nil")
	}

	pb.RegisterHelmActivities(w, activities)
	return nil
}
