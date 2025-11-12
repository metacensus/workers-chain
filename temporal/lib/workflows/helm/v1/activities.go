package helm

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	pb "github.com/metacensus/infra/temporal/gen/workflows/helm/v1"
	"go.temporal.io/sdk/activity"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
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
	logger := activity.GetLogger(ctx)

	// Validate required fields
	if req.GetChartName() == "" {
		return nil, fmt.Errorf("chart_name is required")
	}

	logger.Info("Rendering Helm chart",
		"chart_name", req.GetChartName(),
		"namespace", req.GetNamespace())

	return a.renderChartWithValues(ctx, req.GetChartName(), req.GetValuesYaml(), req.GetNamespace())
}

// RenderChartWithTemplate renders a Helm values template, then renders the chart with the result
func (a *Activities) RenderChartWithTemplate(ctx context.Context, req *pb.RenderChartWithTemplateInput) (*pb.RenderChartOutput, error) {
	logger := activity.GetLogger(ctx)

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

	logger.Info("Rendering Helm chart with template",
		"chart_name", req.GetChartName(),
		"component", req.GetComponent(),
		"namespace", req.GetNamespace())

	// Step 1: Load values template for the component
	templateStr, err := a.repo.GetValuesTemplate(ctx, req.GetComponent())
	if err != nil {
		logger.Error("Failed to get values template", "error", err)
		return nil, fmt.Errorf("failed to get values template for component %s: %w", req.GetComponent(), err)
	}

	// Step 2: Convert proto Struct to map[string]interface{}
	values := req.GetValues().AsMap()

	// Step 3: Render template with values
	renderedYAML, err := renderTemplate(templateStr, values)
	if err != nil {
		logger.Error("Failed to render template", "error", err)
		return nil, fmt.Errorf("failed to render values template: %w", err)
	}

	// Step 4: Use shared chart rendering logic
	return a.renderChartWithValues(ctx, req.GetChartName(), renderedYAML, req.GetNamespace())
}

// renderChartWithValues is the shared logic for rendering a Helm chart with values YAML
func (a *Activities) renderChartWithValues(ctx context.Context, chartName, valuesYAML, namespace string) (*pb.RenderChartOutput, error) {
	logger := activity.GetLogger(ctx)

	// Acquire chart from repository
	ch, err := a.repo.GetChart(ctx, chartName)
	if err != nil {
		logger.Error("Failed to get chart", "error", err)
		return nil, fmt.Errorf("failed to get chart %s: %w", chartName, err)
	}

	// Parse values from YAML string
	vals, err := parseValues(valuesYAML)
	if err != nil {
		logger.Error("Failed to parse values", "error", err)
		return nil, fmt.Errorf("failed to parse values: %w", err)
	}

	// Render the chart
	release, err := renderHelmChart(ch, namespace, chartName, vals)
	if err != nil {
		logger.Error("Failed to render chart", "error", err)
		return nil, fmt.Errorf("failed to render chart: %w", err)
	}

	// Parse rendered manifests into individual resources
	resources := make(map[string]*pb.Resource)
	failures := make(map[string]string)

	docs := strings.Split(release.Manifest, "---")

	for i, doc := range docs {
		// Skip empty documents and comments
		if skipDocument(doc) {
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

	logger.Info("Chart rendered successfully",
		"total_resources", len(resources)+len(failures),
		"parsed_resources", len(resources),
		"failed_resources", len(failures))

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
	install.ReleaseName = releaseName // Use release name for deterministic output

	return install.Run(ch, vals)
}

// skipDocument checks if a document should be skipped (empty or comment-only)
func skipDocument(doc string) bool {
	doc = strings.TrimSpace(doc)
	return doc == "" || strings.HasPrefix(doc, "#")
}

// parseManifestDocument parses a single YAML document and extracts resource metadata
// Returns the resource ID, manifest bytes, and any parse error
func parseManifestDocument(doc string, index int) (string, []byte, error) {
	// Parse the document to extract resource metadata
	obj := &unstructured.Unstructured{}
	_, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).
		Decode([]byte(doc), nil, obj)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse resource: %w", err)
	}

	// Generate resource ID: "{kind}-{name}"
	kind := gvk.Kind
	if kind == "" {
		kind = "Unknown"
	}
	name := obj.GetName()
	if name == "" {
		name = fmt.Sprintf("resource-%d", index)
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

// parseValues parses Helm values from a YAML string
func parseValues(valuesYAML string) (map[string]interface{}, error) {
	if valuesYAML == "" {
		return make(map[string]interface{}), nil
	}

	// Decode YAML to unstructured object
	obj := &unstructured.Unstructured{}
	_, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).
		Decode([]byte(valuesYAML), nil, obj)
	if err != nil {
		return nil, fmt.Errorf("invalid values YAML: %w", err)
	}

	return obj.Object, nil
}
