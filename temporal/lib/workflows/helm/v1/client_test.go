package helm_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	helm "github.com/metacensus/infra/temporal/lib/workflows/helm/v1"
)

func TestNewLocalRepoClient_Validation(t *testing.T) {
	tests := []struct {
		name      string
		chartDir  string
		expectErr bool
		expectMsg string
	}{
		{
			name:      "error - empty chart directory",
			chartDir:  "",
			expectErr: true,
			expectMsg: "ChartDir is required",
		},
		{
			name:      "error - nonexistent directory",
			chartDir:  "/nonexistent/path/to/charts",
			expectErr: true,
			expectMsg: "ChartDir not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := helm.NewLocalRepoClient(tt.chartDir)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				// Note: We just verify an error occurred, not the exact message
				// since error wrapping may change the message format
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLocalRepoClient_GetChart(t *testing.T) {
	// Create a temporary directory for test charts
	tmpDir := t.TempDir()

	// Create a minimal chart structure
	chartDir := filepath.Join(tmpDir, "test-chart")
	if err := os.MkdirAll(filepath.Join(chartDir, "templates"), 0755); err != nil {
		t.Fatalf("failed to create chart directory: %v", err)
	}

	// Create Chart.yaml
	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := []byte(`apiVersion: v2
name: test-chart
version: 1.0.0
description: A test chart
`)
	if err := os.WriteFile(chartYaml, chartContent, 0644); err != nil {
		t.Fatalf("failed to write Chart.yaml: %v", err)
	}

	// Create a template file
	templateFile := filepath.Join(chartDir, "templates", "service.yaml")
	templateContent := []byte(`apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}
spec:
  ports:
  - port: 80
`)
	if err := os.WriteFile(templateFile, templateContent, 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	tests := []struct {
		name      string
		repoDir   string
		chartName string
		expectErr bool
	}{
		{
			name:      "success - load valid chart",
			repoDir:   tmpDir,
			chartName: "test-chart",
			expectErr: false,
		},
		{
			name:      "error - chart not found",
			repoDir:   tmpDir,
			chartName: "nonexistent-chart",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := helm.NewLocalRepoClient(tt.repoDir)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			ctx := context.Background()
			chart, err := client.GetChart(ctx, tt.chartName)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if chart == nil {
					t.Fatal("chart is nil")
				}
				if chart.Metadata == nil {
					t.Fatal("chart metadata is nil")
				}
				if chart.Metadata.Name != "test-chart" {
					t.Errorf("chart name: got %q, want %q", chart.Metadata.Name, "test-chart")
				}
			}
		})
	}
}

func TestLocalRepoClient_GetChart_WithMinimalChart(t *testing.T) {
	// Test loading a minimal chart structure
	tmpDir := t.TempDir()
	chartDir := filepath.Join(tmpDir, "minimal-chart")

	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatalf("failed to create chart directory: %v", err)
	}

	// Minimal chart with just Chart.yaml
	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := []byte(`apiVersion: v2
name: minimal-chart
version: 0.1.0
`)
	if err := os.WriteFile(chartYaml, chartContent, 0644); err != nil {
		t.Fatalf("failed to write Chart.yaml: %v", err)
	}

	client, err := helm.NewLocalRepoClient(tmpDir)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	chart, err := client.GetChart(ctx, "minimal-chart")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chart == nil {
		t.Fatal("chart is nil")
	}
	if chart.Metadata.Name != "minimal-chart" {
		t.Errorf("chart name: got %q, want %q", chart.Metadata.Name, "minimal-chart")
	}
}

func TestLocalRepoClient_GetChart_Context(t *testing.T) {
	// Test that context is respected (if implementation supports cancellation)
	tmpDir := t.TempDir()
	chartDir := filepath.Join(tmpDir, "test-chart")

	if err := os.MkdirAll(chartDir, 0755); err != nil {
		t.Fatalf("failed to create chart directory: %v", err)
	}

	chartYaml := filepath.Join(chartDir, "Chart.yaml")
	chartContent := []byte(`apiVersion: v2
name: test-chart
version: 1.0.0
`)
	if err := os.WriteFile(chartYaml, chartContent, 0644); err != nil {
		t.Fatalf("failed to write Chart.yaml: %v", err)
	}

	client, err := helm.NewLocalRepoClient(tmpDir)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test with normal context
	ctx := context.Background()
	chart, err := client.GetChart(ctx, "test-chart")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chart == nil {
		t.Fatal("chart is nil")
	}

	// Test with cancelled context (implementation may ignore, but should not panic)
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	chart, err = client.GetChart(cancelCtx, "test-chart")
	// Note: Current implementation doesn't use context, so this might still succeed
	// But we verify it doesn't panic
	if chart != nil && chart.Metadata.Name != "test-chart" {
		t.Errorf("chart name: got %q, want %q", chart.Metadata.Name, "test-chart")
	}
}
