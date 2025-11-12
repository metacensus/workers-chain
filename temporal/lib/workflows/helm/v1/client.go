package helm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// RepoClient is an interface for acquiring Helm charts and templates from a repository
// Implementations can read from local filesystem, Helm repos, OCI registries, etc.
type RepoClient interface {
	// GetChart retrieves a Helm chart by name from the repository
	// Returns the loaded chart or an error if not found
	GetChart(ctx context.Context, chartName string) (*chart.Chart, error)

	// GetValuesTemplate retrieves a values template for a component
	// Returns the template string or an error if not found
	GetValuesTemplate(ctx context.Context, component string) (string, error)
}

// NewLocalRepoClient creates a RepoClient that reads charts from the local filesystem
func NewLocalRepoClient(chartDir string) (RepoClient, error) {
	if chartDir == "" {
		return nil, fmt.Errorf("ChartDir is required")
	}

	// Verify chart directory exists
	if _, err := os.Stat(chartDir); err != nil {
		return nil, fmt.Errorf("ChartDir not found: %w", err)
	}

	return &localRepoClient{
		chartDir: chartDir,
	}, nil
}

// localRepoClient implements RepoClient for local filesystem-based chart repositories
type localRepoClient struct {
	chartDir string
}

// GetChart loads a chart from the local filesystem repository
func (lrc *localRepoClient) GetChart(ctx context.Context, chartName string) (*chart.Chart, error) {
	// Construct chart path from chart name
	chartPath := filepath.Join(lrc.chartDir, chartName)

	// Verify chart exists
	if _, err := os.Stat(chartPath); err != nil {
		return nil, fmt.Errorf("chart not found at %s: %w", chartPath, err)
	}

	// Load the chart
	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart from %s: %w", chartPath, err)
	}

	return ch, nil
}

// GetValuesTemplate loads a values template from the local filesystem repository
// Looks for files named "{component}-values.yaml.tpl" in a templates subdirectory
func (lrc *localRepoClient) GetValuesTemplate(ctx context.Context, component string) (string, error) {
	// Construct template path: chartDir/templates/{component}-values.yaml.tpl
	templatePath := filepath.Join(lrc.chartDir, "templates", fmt.Sprintf("%s-values.yaml.tpl", component))

	// Read the template file
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load template for component %s: %w", component, err)
	}

	return string(templateBytes), nil
}
