package helm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	pb "github.com/metacensus/workers-chain/gen/workflows/helm/v1"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"sigs.k8s.io/yaml"
)

// Sentinel errors for helm activities
var (
	ErrGetValuesTemplate = errors.New("get values template")
	ErrRenderTemplate    = errors.New("render template")
	ErrGetChart          = errors.New("get chart")
	ErrParseValues       = errors.New("parse values")
	ErrRenderChart       = errors.New("render chart")
)

// Activities implements the HelmServiceActivities interface from generated code
type Activities struct {
	repo RepoClient
}

// NewActivities creates a new Activities instance
func NewActivities(repo RepoClient) *Activities {
	return &Activities{
		repo: repo,
	}
}

// RenderChart renders a Helm chart and returns parsed Kubernetes resources
func (a *Activities) RenderChart(ctx context.Context, req *pb.RenderChartInput) (*pb.RenderChartOutput, error) {
	// Validate required fields
	if req.GetChartName() == "" {
		return nil, fmt.Errorf("chart_name is required")
	}

	return a.renderChartWithValues(ctx, req.GetChartName(), req.GetValuesYaml(), req.GetNamespace())
}

// RenderChartWithTemplate renders a Helm values template, then renders the chart with the result
func (a *Activities) RenderChartWithTemplate(ctx context.Context, req *pb.RenderChartWithTemplateInput) (*pb.RenderChartOutput, error) {
	// Validate required fields
	if req.GetChartName() == "" {
		return nil, fmt.Errorf("chart_name is required")
	}
	if req.GetComponent() == "" {
		return nil, fmt.Errorf("component is required")
	}
	if req.GetValues() == nil {
		return nil, fmt.Errorf("values is required")
	}

	// Step 1: Load values template from the chart directory
	templateStr, err := a.repo.GetValuesTemplate(ctx, req.GetChartName())
	if err != nil {
		return nil, fmt.Errorf("%w for chart %s: %w", ErrGetValuesTemplate, req.GetChartName(), err)
	}

	// Step 2: Convert proto Struct to map[string]interface{}
	values := req.GetValues().AsMap()

	// Step 3: Render template with values
	renderedYAML, err := renderTemplate(templateStr, values)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderTemplate, err)
	}

	// Step 4: Use shared chart rendering logic
	return a.renderChartWithValues(ctx, req.GetChartName(), renderedYAML, req.GetNamespace())
}

// renderChartWithValues is the shared logic for rendering a Helm chart with values YAML
func (a *Activities) renderChartWithValues(ctx context.Context, chartName, valuesYAML, namespace string) (*pb.RenderChartOutput, error) {
	// Acquire chart from repository
	ch, err := a.repo.GetChart(ctx, chartName)
	if err != nil {
		return nil, fmt.Errorf("%w %s: %w", ErrGetChart, chartName, err)
	}

	// Parse values from YAML string
	vals, err := parseValues(valuesYAML)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParseValues, err)
	}

	// Convert chart name to valid Helm release name (e.g., "fabric/namespace" → "fabric-namespace")
	releaseName := strings.ReplaceAll(chartName, "/", "-")

	// Render the chart
	release, err := renderHelmChart(ch, namespace, releaseName, vals)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderChart, err)
	}

	// Parse rendered manifests into individual resources
	resources := make(map[string]*pb.Resource)
	failures := make(map[string]string)

	docs := strings.Split(release.Manifest, "---")

	for i, doc := range docs {
		// Skip empty documents
		if strings.TrimSpace(doc) == "" {
			continue
		}

		// Parse the document and extract resource metadata
		resourceID, manifest, err := parseManifestDocument(doc, i)
		if err != nil {
			resourceID := fmt.Sprintf("unknown-resource-%d", i)
			failures[resourceID] = err.Error()
			continue
		}

		resources[resourceID] = &pb.Resource{
			ManifestYaml: manifest,
		}
	}

	return &pb.RenderChartOutput{
		Resources: resources,
		Failures:  failures,
	}, nil
}

// renderHelmChart renders a Helm chart using the action API
// Uses deterministic settings: dry-run mode, client-only, includes CRDs
// Returns the rendered release or an error
func renderHelmChart(ch *chart.Chart, namespace, releaseName string, vals map[string]interface{}) (*release.Release, error) {
	// Create Helm action configuration with suppressed logging
	actionConfig := &action.Configuration{}
	actionConfig.Log = func(format string, v ...interface{}) {
		// Suppress Helm's logging output
	}

	// Create install action configured for rendering (not actual installation)
	install := action.NewInstall(actionConfig)
	install.DryRun = true      // Don't actually install
	install.ClientOnly = true  // No Kubernetes API calls
	install.IncludeCRDs = true // Include CRD manifests
	install.Namespace = namespace
	install.ReleaseName = releaseName

	return install.Run(ch, vals)
}

// parseManifestDocument parses a single YAML document and extracts resource metadata
// Returns the resource ID, manifest bytes, and any parse error
func parseManifestDocument(doc string, index int) (string, []byte, error) {
	// Parse the document to extract resource metadata
	obj := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(doc), &obj)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse resource: %w", err)
	}

	// Extract Kind and Name from the parsed object
	kind := "Unknown"
	if k, ok := obj["kind"]; ok {
		if kindStr, ok := k.(string); ok {
			kind = kindStr
		}
	}

	name := fmt.Sprintf("resource-%d", index)
	if metadata, ok := obj["metadata"].(map[string]interface{}); ok {
		if n, ok := metadata["name"].(string); ok && n != "" {
			name = n
		}
	}

	resourceID := fmt.Sprintf("%s-%s", strings.ToLower(kind), name)

	return resourceID, []byte(doc), nil
}

// renderTemplate renders a Go template with provided values
func renderTemplate(templateStr string, values interface{}) (string, error) {
	tmpl, err := template.New("values").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, values); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// parseValues parses Helm values from a YAML string into a map
func parseValues(valuesYAML string) (map[string]interface{}, error) {
	if valuesYAML == "" {
		return make(map[string]interface{}), nil
	}

	// Decode YAML to a plain map (not a Kubernetes resource)
	var values map[string]interface{}
	err := yaml.Unmarshal([]byte(valuesYAML), &values)
	if err != nil {
		return nil, fmt.Errorf("invalid values YAML: %w", err)
	}

	// Handle case where YAML is empty or null
	if values == nil {
		values = make(map[string]interface{})
	}

	return values, nil
}
