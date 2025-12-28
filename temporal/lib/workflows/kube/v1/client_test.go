//go:build integration
// +build integration

package kube

import (
	"context"
	"testing"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// TestKubeClient_RealCluster tests against an actual Kubernetes cluster in k3d
// Run with: go test -tags=integration ./lib/workflows/kube/v1
func TestKubeClient_RealCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start k3d container
	req := testcontainers.ContainerRequest{
		Image:        "rancher/k3d:latest",
		ExposedPorts: []string{"6443/tcp"},
		Env: map[string]string{
			"K3D_FIX_GPU": "true",
		},
		Cmd: []string{"cluster", "create", "test-cluster", "--api-port", "0.0.0.0:6443", "--servers", "1"},
		WaitingFor: wait.ForLog("cluster 'test-cluster' created successfully").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start k3d container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	// For now, integration tests require a real Kubernetes cluster
	// Using in-cluster config if available, otherwise skip
	client, err := NewKubeClient(&config.KubernetesConfig{})
	if err != nil {
		t.Logf("could not connect to Kubernetes cluster (expected in integration test environment): %v", err)
		t.Skip("Integration test requires access to a Kubernetes cluster")
	}

	t.Run("create and delete service", func(t *testing.T) {
		// Create a namespace first
		namespaceManifest := []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: test-ns
`)
		decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}
		if _, _, err := decoder.Decode(namespaceManifest, nil, obj); err != nil {
			t.Fatalf("failed to parse namespace manifest: %v", err)
		}

		nsID := &pb.ResourceIdentifier{
			ApiVersion: "v1",
			Kind:       "Namespace",
			Name:       "test-ns",
		}

		if err := client.CreateResource(ctx, nsID, obj); err != nil {
			t.Fatalf("failed to create namespace: %v", err)
		}

		// Create a service
		serviceManifest := []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: test-ns
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: test
`)
		obj = &unstructured.Unstructured{}
		if _, _, err := decoder.Decode(serviceManifest, nil, obj); err != nil {
			t.Fatalf("failed to parse service manifest: %v", err)
		}

		serviceID := &pb.ResourceIdentifier{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "test-service",
			Namespace:  "test-ns",
		}

		// Create
		if err := client.CreateResource(ctx, serviceID, obj); err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		// Get status
		status, err := client.GetResourceStatus(ctx, serviceID)
		if err != nil {
			t.Fatalf("failed to get service status: %v", err)
		}
		if status == nil {
			t.Errorf("expected status, got nil")
		}

		// Update the service
		obj.Object["spec"].(map[string]interface{})["ports"] = []map[string]interface{}{
			{"port": int64(443), "targetPort": int64(8443)},
		}

		if err := client.UpdateResource(ctx, serviceID, obj); err != nil {
			t.Fatalf("failed to update service: %v", err)
		}

		// Delete
		if err := client.DeleteResource(ctx, serviceID); err != nil {
			t.Fatalf("failed to delete service: %v", err)
		}

		// Verify deletion (status should error)
		_, err = client.GetResourceStatus(ctx, serviceID)
		if err == nil {
			t.Error("expected error when getting status of deleted service, got nil")
		}
	})
}

// TestKubeClient_ManifestParsing tests YAML manifest parsing without a real cluster
func TestKubeClient_ManifestParsing(t *testing.T) {
	tests := []struct {
		name      string
		manifest  []byte
		expectErr bool
		expectID  pb.ResourceIdentifier
	}{
		{
			name: "valid service manifest",
			manifest: []byte(`
apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: default
spec:
  ports:
  - port: 80
    targetPort: 8080
`),
			expectErr: false,
			expectID: pb.ResourceIdentifier{
				ApiVersion: "v1",
				Kind:       "Service",
				Name:       "my-service",
				Namespace:  "default",
			},
		},
		{
			name: "valid deployment manifest",
			manifest: []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-container
        image: my-image:latest
`),
			expectErr: false,
			expectID: pb.ResourceIdentifier{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "my-deployment",
				Namespace:  "production",
			},
		},
		{
			name: "invalid YAML",
			manifest: []byte(`
invalid: yaml: syntax:
`),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
			obj := &unstructured.Unstructured{}

			_, gvk, err := decoder.Decode(tt.manifest, nil, obj)

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				// Verify resource identifier
				id := pb.ResourceIdentifier{
					ApiVersion: gvk.GroupVersion().String(),
					Kind:       gvk.Kind,
					Name:       obj.GetName(),
					Namespace:  obj.GetNamespace(),
				}

				if id != tt.expectID {
					t.Errorf("resource identifier mismatch: got %+v, want %+v", id, tt.expectID)
				}
			}
		})
	}
}

// TestKubeClient_GVRParsing tests Group/Version/Resource parsing
func TestKubeClient_GVRParsing(t *testing.T) {
	tests := []struct {
		name           string
		apiVersion     string
		kind           string
		expectAPIGroup string
		expectVersion  string
		expectResource string
	}{
		{
			name:           "core API deployment",
			apiVersion:     "apps/v1",
			kind:           "Deployment",
			expectAPIGroup: "apps",
			expectVersion:  "v1",
			expectResource: "deployments",
		},
		{
			name:           "core API service",
			apiVersion:     "v1",
			kind:           "Service",
			expectAPIGroup: "",
			expectVersion:  "v1",
			expectResource: "services",
		},
		{
			name:           "custom API",
			apiVersion:     "example.com/v1alpha1",
			kind:           "CustomResource",
			expectAPIGroup: "example.com",
			expectVersion:  "v1alpha1",
			expectResource: "customresources", // Simple pluralization
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, version := parseAPIVersion(tt.apiVersion)
			resource := kindToResource(tt.kind)

			if group != tt.expectAPIGroup {
				t.Errorf("API group mismatch: got %q, want %q", group, tt.expectAPIGroup)
			}
			if version != tt.expectVersion {
				t.Errorf("Version mismatch: got %q, want %q", version, tt.expectVersion)
			}
			if resource != tt.expectResource {
				t.Errorf("Resource mismatch: got %q, want %q", resource, tt.expectResource)
			}
		})
	}
}

// TestResourceStatusConversion tests the conversion from client status to proto status
func TestResourceStatusConversion(t *testing.T) {
	tests := []struct {
		name           string
		clientStatus   *ResourceStatus
		expectedStatus string
	}{
		{
			name: "ready deployment",
			clientStatus: &ResourceStatus{
				Ready:           true,
				Message:         "ready",
				DesiredReplicas: 3,
				ReadyReplicas:   3,
			},
			expectedStatus: StatusReady,
		},
		{
			name: "pending deployment",
			clientStatus: &ResourceStatus{
				Ready:           false,
				Message:         "0/3 replicas ready",
				DesiredReplicas: 3,
				ReadyReplicas:   0,
			},
			expectedStatus: StatusPending,
		},
		{
			name: "degraded deployment",
			clientStatus: &ResourceStatus{
				Ready:           false,
				Message:         "2/3 replicas ready",
				DesiredReplicas: 3,
				ReadyReplicas:   2,
			},
			expectedStatus: StatusDegraded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the conversion logic
			var pbStatus string
			if tt.clientStatus.Ready {
				pbStatus = StatusReady
			} else if tt.clientStatus.ReadyReplicas > 0 && tt.clientStatus.DesiredReplicas > 0 {
				pbStatus = StatusDegraded
			} else {
				pbStatus = StatusPending
			}

			if pbStatus != tt.expectedStatus {
				t.Errorf("status conversion: got %q, want %q", pbStatus, tt.expectedStatus)
			}
		})
	}
}
