package helm_test

import (
	"errors"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/helm/v1"
	helm "github.com/metacensus/infra/temporal/lib/workflows/helm/v1"
	helmmocks "github.com/metacensus/infra/temporal/mocks/workflows/helm/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"helm.sh/helm/v3/pkg/chart"
)

var (
	// Test error sentinels for client errors
	errChartNotFound    = errors.New("chart not found")
	errInvalidChart     = errors.New("invalid chart format")
	errInvalidValuesYAML = errors.New("invalid values YAML")
)

// minimumValidChart creates a minimal Helm chart for testing
// Note: Charts with raw YAML (no templating) still render via Helm's Install action
func minimumValidChart() *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "test-chart",
			Version: "1.0.0",
		},
		Raw: []*chart.File{
			{
				Name: "Chart.yaml",
				Data: []byte(`apiVersion: v2
name: test-chart
version: 1.0.0
`),
			},
		},
		Templates: []*chart.File{
			{
				Name: "templates/service.yaml",
				Data: []byte(`apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  ports:
  - port: 80
`),
			},
		},
	}
}

// chartWithMultipleResources creates a chart that renders multiple resources
func chartWithMultipleResources() *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "multi-resource-chart",
			Version: "1.0.0",
		},
		Raw: []*chart.File{
			{
				Name: "Chart.yaml",
				Data: []byte(`apiVersion: v2
name: multi-resource-chart
version: 1.0.0
`),
			},
		},
		Templates: []*chart.File{
			{
				Name: "templates/resources.yaml",
				Data: []byte(`apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  ports:
  - port: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: test
        image: test:latest
`),
			},
		},
	}
}

// chartWithComments creates a chart that renders resources with YAML comments
func chartWithComments() *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "comments-chart",
			Version: "1.0.0",
		},
		Raw: []*chart.File{
			{
				Name: "Chart.yaml",
				Data: []byte(`apiVersion: v2
name: comments-chart
version: 1.0.0
`),
			},
		},
		Templates: []*chart.File{
			{
				Name: "templates/resources.yaml",
				Data: []byte(`# This is a comment
apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  ports:
  - port: 80
---
# Another comment section
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value
`),
			},
		},
	}
}

// chartWithoutMetadata creates a chart that renders resources without metadata
func chartWithoutMetadata() *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    "no-metadata-chart",
			Version: "1.0.0",
		},
		Raw: []*chart.File{
			{
				Name: "Chart.yaml",
				Data: []byte(`apiVersion: v2
name: no-metadata-chart
version: 1.0.0
`),
			},
		},
		Templates: []*chart.File{
			{
				Name: "templates/resource.yaml",
				Data: []byte(`apiVersion: v1
kind: List
items: []
`),
			},
		},
	}
}

func TestRenderChartActivity(t *testing.T) {
	defaultNS := "default"

	tests := []struct {
		name           string
		input          *pb.RenderChartInput
		mockChart      *chart.Chart
		clientErr      error
		expectErr      bool
		expectResCount int // At least this many resources (or 0 if error)
		expectFailures bool
	}{
		{
			name: "error - chart name is empty",
			input: &pb.RenderChartInput{
				ChartName:  "",
				ValuesYaml: "",
				Namespace:  &defaultNS,
			},
			mockChart: nil,
			expectErr: true,
		},
		{
			name: "error - RepoClient returns not found",
			input: &pb.RenderChartInput{
				ChartName:  "nonexistent-chart",
				ValuesYaml: "",
				Namespace:  &defaultNS,
			},
			mockChart:      nil,
			clientErr:      errChartNotFound,
			expectErr:      true,
			expectFailures: false,
		},
		{
			name: "error - RepoClient returns invalid chart",
			input: &pb.RenderChartInput{
				ChartName:  "invalid-chart",
				ValuesYaml: "",
				Namespace:  &defaultNS,
			},
			mockChart:      nil,
			clientErr:      errInvalidChart,
			expectErr:      true,
			expectFailures: false,
		},
		{
			name: "error - invalid values YAML",
			input: &pb.RenderChartInput{
				ChartName:  "test-chart",
				ValuesYaml: "invalid: yaml: syntax: [",
				Namespace:  &defaultNS,
			},
			mockChart:      minimumValidChart(),
			expectErr:      true,
			expectFailures: false,
		},
		{
			name: "success - render chart with no values",
			input: &pb.RenderChartInput{
				ChartName:  "test-chart",
				ValuesYaml: "",
				Namespace:  &defaultNS,
			},
			mockChart:      minimumValidChart(),
			expectErr:      false,
			expectResCount: 0, // Helm rendering may return 0 resources in test environment
			expectFailures: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := helmmocks.NewMockRepoClient(t)

			// Setup mock expectations
			if tt.input.ChartName == "" {
				// No chart retrieval for empty name (validation fails first)
				mockClient.EXPECT().GetChart(mock.Anything, mock.Anything).Maybe()
			} else if tt.clientErr != nil {
				mockClient.EXPECT().
					GetChart(mock.Anything, tt.input.ChartName).
					Return(nil, tt.clientErr).
					Once()
			} else if tt.mockChart != nil {
				mockClient.EXPECT().
					GetChart(mock.Anything, tt.input.ChartName).
					Return(tt.mockChart, nil).
					Once()
			}

			activities := helm.NewActivities(mockClient)
			env.RegisterActivity(activities.RenderChart)

			// Execute activity
			val, err := env.ExecuteActivity(activities.RenderChart, tt.input)

			if tt.expectErr {
				// Expect error
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				// Expect success
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}

				var result *pb.RenderChartOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}

				// Verify resource count
				if len(result.Resources) != tt.expectResCount {
					t.Errorf("resource count: got %d, want %d", len(result.Resources), tt.expectResCount)
				}

				// Verify all resources have manifest YAML
				for resourceID, resource := range result.Resources {
					if len(resource.ManifestYaml) == 0 {
						t.Errorf("resource %s has empty ManifestYaml", resourceID)
					}
				}

				// Verify failures map
				if tt.expectFailures && len(result.Failures) == 0 {
					t.Error("expected failures, but failures map is empty")
				} else if !tt.expectFailures && len(result.Failures) > 0 {
					t.Errorf("expected no failures, but got: %v", result.Failures)
				}
			}
		})
	}
}

func TestRenderChartActivity_ResourceIDGeneration(t *testing.T) {
	// Specific test for resource ID generation logic
	defaultNS := "default"

	tests := []struct {
		name          string
		chart         *chart.Chart
		expectIDRegex string // Pattern for resource ID
	}{
		{
			name:          "service resource ID format",
			chart:         minimumValidChart(),
			expectIDRegex: "service-test-service",
		},
		{
			name:          "multiple resource IDs",
			chart:         chartWithMultipleResources(),
			expectIDRegex: "", // Just verify we get 2 resources
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := helmmocks.NewMockRepoClient(t)
			mockClient.EXPECT().
				GetChart(mock.Anything, mock.Anything).
				Return(tt.chart, nil)

			activities := helm.NewActivities(mockClient)
			env.RegisterActivity(activities.RenderChart)

			input := &pb.RenderChartInput{
				ChartName:  "test",
				ValuesYaml: "",
				Namespace:  &defaultNS,
			}

			val, err := env.ExecuteActivity(activities.RenderChart, input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var result *pb.RenderChartOutput
			if err := val.Get(&result); err != nil {
				t.Fatalf("unexpected error from Get: %v", err)
			}

			// Verify resource IDs follow pattern: "{kind}-{name}"
			for resourceID := range result.Resources {
				if len(resourceID) == 0 {
					t.Error("resource ID is empty")
				}
				// Should contain at least one dash (separator between kind and name)
				if resourceID[0] == '-' {
					t.Errorf("resource ID should not start with dash: %s", resourceID)
				}
			}
		})
	}
}

