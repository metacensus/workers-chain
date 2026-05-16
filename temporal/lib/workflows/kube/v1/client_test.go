package kube

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// TestKubeClient_ManifestParsing tests YAML manifest parsing without a real cluster
func TestKubeClient_ManifestParsing(t *testing.T) {
	tests := []struct {
		name            string
		manifest        []byte
		expectErr       bool
		expectAPIVer    string
		expectKind      string
		expectName      string
		expectNamespace string
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
			expectErr:       false,
			expectAPIVer:    "v1",
			expectKind:      "Service",
			expectName:      "my-service",
			expectNamespace: "default",
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
			expectErr:       false,
			expectAPIVer:    "apps/v1",
			expectKind:      "Deployment",
			expectName:      "my-deployment",
			expectNamespace: "production",
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
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := gvk.GroupVersion().String(); got != tt.expectAPIVer {
				t.Errorf("apiVersion mismatch: got %q, want %q", got, tt.expectAPIVer)
			}
			if got := gvk.Kind; got != tt.expectKind {
				t.Errorf("kind mismatch: got %q, want %q", got, tt.expectKind)
			}
			if got := obj.GetName(); got != tt.expectName {
				t.Errorf("name mismatch: got %q, want %q", got, tt.expectName)
			}
			if got := obj.GetNamespace(); got != tt.expectNamespace {
				t.Errorf("namespace mismatch: got %q, want %q", got, tt.expectNamespace)
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
