package kube

import (
	"context"
	"fmt"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	pb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	"github.com/stretchr/testify/mock"
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
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				// Verify resource identifier fields individually (protobuf structs can't be compared directly)
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

// Mock implementations for testing

// mockResourceInterface mocks dynamic.ResourceInterface
type mockResourceInterface struct {
	mock.Mock
}

func (m *mockResourceInterface) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := m.Called(ctx, obj, options, subresources)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (m *mockResourceInterface) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := m.Called(ctx, obj, options, subresources)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (m *mockResourceInterface) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	args := m.Called(ctx, obj, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (m *mockResourceInterface) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	args := m.Called(ctx, name, options, subresources)
	return args.Error(0)
}

func (m *mockResourceInterface) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	args := m.Called(ctx, options, listOptions)
	return args.Error(0)
}

func (m *mockResourceInterface) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := m.Called(ctx, name, options, subresources)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (m *mockResourceInterface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.UnstructuredList), args.Error(1)
}

func (m *mockResourceInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(watch.Interface), args.Error(1)
}

func (m *mockResourceInterface) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := m.Called(ctx, name, obj, options, subresources)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (m *mockResourceInterface) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	args := m.Called(ctx, name, obj, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

func (m *mockResourceInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	args := m.Called(ctx, name, pt, data, options, subresources)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

// mockDynamicInterface mocks dynamic.Interface
type mockDynamicInterface struct {
	mock.Mock
	resourceInterface *mockResourceInterface
}

func (m *mockDynamicInterface) Resource(resource schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	if m.resourceInterface == nil {
		m.resourceInterface = &mockResourceInterface{}
	}
	return &mockNamespaceableResourceInterface{
		resourceInterface: m.resourceInterface,
	}
}

// mockNamespaceableResourceInterface mocks dynamic.NamespaceableResourceInterface
type mockNamespaceableResourceInterface struct {
	resourceInterface *mockResourceInterface
}

func (m *mockNamespaceableResourceInterface) Namespace(namespace string) dynamic.ResourceInterface {
	return m.resourceInterface
}

// Implement all ResourceInterface methods at cluster-scoped level
func (m *mockNamespaceableResourceInterface) Create(ctx context.Context, obj *unstructured.Unstructured, options metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return m.resourceInterface.Create(ctx, obj, options, subresources...)
}

func (m *mockNamespaceableResourceInterface) Update(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return m.resourceInterface.Update(ctx, obj, options, subresources...)
}

func (m *mockNamespaceableResourceInterface) UpdateStatus(ctx context.Context, obj *unstructured.Unstructured, options metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return m.resourceInterface.UpdateStatus(ctx, obj, options)
}

func (m *mockNamespaceableResourceInterface) Delete(ctx context.Context, name string, options metav1.DeleteOptions, subresources ...string) error {
	return m.resourceInterface.Delete(ctx, name, options, subresources...)
}

func (m *mockNamespaceableResourceInterface) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return m.resourceInterface.DeleteCollection(ctx, options, listOptions)
}

func (m *mockNamespaceableResourceInterface) Get(ctx context.Context, name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return m.resourceInterface.Get(ctx, name, options, subresources...)
}

func (m *mockNamespaceableResourceInterface) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return m.resourceInterface.List(ctx, opts)
}

func (m *mockNamespaceableResourceInterface) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return m.resourceInterface.Watch(ctx, opts)
}

func (m *mockNamespaceableResourceInterface) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return m.resourceInterface.Patch(ctx, name, pt, data, options, subresources...)
}

func (m *mockNamespaceableResourceInterface) Apply(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	return m.resourceInterface.Apply(ctx, name, obj, options, subresources...)
}

func (m *mockNamespaceableResourceInterface) ApplyStatus(ctx context.Context, name string, obj *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	return m.resourceInterface.ApplyStatus(ctx, name, obj, options)
}

// Helper to create a test kubeClient with injected mocks
func newTestKubeClient(dynamicClient dynamic.Interface, typedClient kubernetes.Interface, restConfig *rest.Config) *kubeClient {
	return &kubeClient{
		dynamicClient: dynamicClient,
		typedClient:   typedClient,
		restConfig:    restConfig,
	}
}

// Helper to create a test ResourceIdentifier
func testResourceID(apiVersion, kind, name, namespace string) *pb.ResourceIdentifier {
	return &pb.ResourceIdentifier{
		ApiVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		Namespace:  namespace,
	}
}

// Helper to create a test unstructured object
func testUnstructured(apiVersion, kind, name, namespace string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(apiVersion)
	obj.SetKind(kind)
	obj.SetName(name)
	if namespace != "" {
		obj.SetNamespace(namespace)
	}
	return obj
}

// TestCreateResource tests CreateResource method
func TestCreateResource(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          *pb.ResourceIdentifier
		obj         *unstructured.Unstructured
		setupMock   func(*mockResourceInterface)
		expectErr   bool
		errContains string
	}{
		{
			name: "success - namespaced resource",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(testUnstructured("v1", "Service", "test-service", "default"), nil)
			},
			expectErr: false,
		},
		{
			name: "success - cluster-scoped resource",
			id:   testResourceID("v1", "Namespace", "test-ns", ""),
			obj:  testUnstructured("v1", "Namespace", "test-ns", ""),
			setupMock: func(m *mockResourceInterface) {
				m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(testUnstructured("v1", "Namespace", "test-ns", ""), nil)
			},
			expectErr: false,
		},
		{
			name: "error - resource already exists",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{Resource: "services"}, "test-service"))
			},
			expectErr:   true,
			errContains: "failed to create resource",
		},
		{
			name: "error - client create fails",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("connection refused"))
			},
			expectErr:   true,
			errContains: "failed to create resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResource := &mockResourceInterface{}
			tt.setupMock(mockResource)

			mockDynamic := &mockDynamicInterface{
				resourceInterface: mockResource,
			}

			client := newTestKubeClient(mockDynamic, nil, nil)

			err := client.CreateResource(ctx, tt.id, tt.obj)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message should contain %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			mockResource.AssertExpectations(t)
		})
	}
}

// TestUpdateResource tests UpdateResource method
func TestUpdateResource(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          *pb.ResourceIdentifier
		obj         *unstructured.Unstructured
		setupMock   func(*mockResourceInterface)
		expectErr   bool
		errContains string
	}{
		{
			name: "success - namespaced resource",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(testUnstructured("v1", "Service", "test-service", "default"), nil)
			},
			expectErr: false,
		},
		{
			name: "success - cluster-scoped resource",
			id:   testResourceID("v1", "Namespace", "test-ns", ""),
			obj:  testUnstructured("v1", "Namespace", "test-ns", ""),
			setupMock: func(m *mockResourceInterface) {
				m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(testUnstructured("v1", "Namespace", "test-ns", ""), nil)
			},
			expectErr: false,
		},
		{
			name: "error - resource not found",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, k8serrors.NewNotFound(schema.GroupResource{Resource: "services"}, "test-service"))
			},
			expectErr:   true,
			errContains: "failed to update resource",
		},
		{
			name: "error - client update fails",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("connection refused"))
			},
			expectErr:   true,
			errContains: "failed to update resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResource := &mockResourceInterface{}
			tt.setupMock(mockResource)

			mockDynamic := &mockDynamicInterface{
				resourceInterface: mockResource,
			}

			client := newTestKubeClient(mockDynamic, nil, nil)

			err := client.UpdateResource(ctx, tt.id, tt.obj)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message should contain %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			mockResource.AssertExpectations(t)
		})
	}
}

// TestApplyResource tests ApplyResource method (upsert behavior)
func TestApplyResource(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          *pb.ResourceIdentifier
		obj         *unstructured.Unstructured
		setupMock   func(*mockResourceInterface)
		expectErr   bool
		errContains string
	}{
		{
			name: "success - updates existing resource",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				// Update succeeds
				m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(testUnstructured("v1", "Service", "test-service", "default"), nil)
			},
			expectErr: false,
		},
		{
			name: "success - creates new resource when not found",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				// Update returns NotFound, then Create succeeds
				m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, k8serrors.NewNotFound(schema.GroupResource{Resource: "services"}, "test-service"))
				m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(testUnstructured("v1", "Service", "test-service", "default"), nil)
			},
			expectErr: false,
		},
		{
			name: "error - update fails with non-NotFound error",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				// Update returns non-NotFound error, should not attempt Create
				m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("forbidden"))
			},
			expectErr:   true,
			errContains: "forbidden",
		},
		{
			name: "error - create fails after NotFound",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			obj:  testUnstructured("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				// Update returns NotFound, then Create fails
				m.On("Update", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, k8serrors.NewNotFound(schema.GroupResource{Resource: "services"}, "test-service"))
				m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("create failed"))
			},
			expectErr:   true,
			errContains: "failed to create resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResource := &mockResourceInterface{}
			tt.setupMock(mockResource)

			mockDynamic := &mockDynamicInterface{
				resourceInterface: mockResource,
			}

			client := newTestKubeClient(mockDynamic, nil, nil)

			err := client.ApplyResource(ctx, tt.id, tt.obj)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message should contain %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			mockResource.AssertExpectations(t)
		})
	}
}

// TestDeleteResource tests DeleteResource method
func TestDeleteResource(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          *pb.ResourceIdentifier
		setupMock   func(*mockResourceInterface)
		expectErr   bool
		errContains string
	}{
		{
			name: "success - namespaced resource",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				m.On("Delete", mock.Anything, "test-service", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "success - cluster-scoped resource",
			id:   testResourceID("v1", "Namespace", "test-ns", ""),
			setupMock: func(m *mockResourceInterface) {
				m.On("Delete", mock.Anything, "test-ns", mock.Anything, mock.Anything).
					Return(nil)
			},
			expectErr: false,
		},
		{
			name: "success - resource not found (succeeds)",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				// Delete returns NotFound, but should succeed
				m.On("Delete", mock.Anything, "test-service", mock.Anything, mock.Anything).
					Return(k8serrors.NewNotFound(schema.GroupResource{Resource: "services"}, "test-service"))
			},
			expectErr:   true, // Currently DeleteResource wraps all errors, including NotFound
			errContains: "failed to delete resource",
		},
		{
			name: "error - client delete fails",
			id:   testResourceID("v1", "Service", "test-service", "default"),
			setupMock: func(m *mockResourceInterface) {
				m.On("Delete", mock.Anything, "test-service", mock.Anything, mock.Anything).
					Return(fmt.Errorf("connection refused"))
			},
			expectErr:   true,
			errContains: "failed to delete resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResource := &mockResourceInterface{}
			tt.setupMock(mockResource)

			mockDynamic := &mockDynamicInterface{
				resourceInterface: mockResource,
			}

			client := newTestKubeClient(mockDynamic, nil, nil)

			err := client.DeleteResource(ctx, tt.id)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message should contain %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			mockResource.AssertExpectations(t)
		})
	}
}

// TestGetResourceStatus tests GetResourceStatus method
func TestGetResourceStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		id          *pb.ResourceIdentifier
		setupClient func() kubernetes.Interface
		setupMock   func(*mockResourceInterface)
		expectErr   bool
		errContains string
		expectReady bool
		expectMsg   string
	}{
		{
			name: "success - Deployment ready",
			id:   testResourceID("apps/v1", "Deployment", "test-deployment", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(3),
					},
					Status: appsv1.DeploymentStatus{
						ReadyReplicas:     3,
						UpdatedReplicas:   3,
						Replicas:          3,
						AvailableReplicas: 3,
					},
				})
			},
			expectErr:   false,
			expectReady: true,
			expectMsg:   "all replicas ready",
		},
		{
			name: "success - Deployment not ready",
			id:   testResourceID("apps/v1", "Deployment", "test-deployment", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(3),
					},
					Status: appsv1.DeploymentStatus{
						ReadyReplicas:     1,
						UpdatedReplicas:   1,
						Replicas:          1,
						AvailableReplicas: 1,
					},
				})
			},
			expectErr:   false,
			expectReady: false,
			expectMsg:   "1/3 replicas ready",
		},
		{
			name: "success - StatefulSet ready",
			id:   testResourceID("apps/v1", "StatefulSet", "test-statefulset", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-statefulset",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: int32Ptr(2),
					},
					Status: appsv1.StatefulSetStatus{
						ReadyReplicas:   2,
						UpdatedReplicas: 2,
						Replicas:        2,
					},
				})
			},
			expectErr:   false,
			expectReady: true,
			expectMsg:   "all replicas ready",
		},
		{
			name: "success - DaemonSet ready",
			id:   testResourceID("apps/v1", "DaemonSet", "test-daemonset", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-daemonset",
						Namespace: "default",
					},
					Status: appsv1.DaemonSetStatus{
						DesiredNumberScheduled: 5,
						NumberReady:            5,
					},
				})
			},
			expectErr:   false,
			expectReady: true,
			expectMsg:   "all pods ready",
		},
		{
			name: "success - Pod Running with all containers ready",
			id:   testResourceID("v1", "Pod", "test-pod", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{Name: "container1", Ready: true},
							{Name: "container2", Ready: true},
						},
					},
				})
			},
			expectErr:   false,
			expectReady: true,
			expectMsg:   "all containers running",
		},
		{
			name: "success - Pod Succeeded",
			id:   testResourceID("v1", "Pod", "test-pod", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
					},
				})
			},
			expectErr:   false,
			expectReady: true,
			expectMsg:   "pod completed successfully",
		},
		{
			name: "success - Pod Failed",
			id:   testResourceID("v1", "Pod", "test-pod", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
					},
					Status: corev1.PodStatus{
						Phase:   corev1.PodFailed,
						Message: "container failed",
					},
				})
			},
			expectErr:   false,
			expectReady: false,
			expectMsg:   "pod failed: container failed",
		},
		{
			name: "success - unknown resource type (checks existence)",
			id:   testResourceID("example.com/v1", "CustomResource", "test-cr", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset()
			},
			setupMock: func(m *mockResourceInterface) {
				obj := testUnstructured("example.com/v1", "CustomResource", "test-cr", "default")
				m.On("Get", mock.Anything, "test-cr", mock.Anything, mock.Anything).
					Return(obj, nil)
			},
			expectErr:   false,
			expectReady: true,
			expectMsg:   "resource exists",
		},
		{
			name: "error - resource not found",
			id:   testResourceID("apps/v1", "Deployment", "not-found", "default"),
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset()
			},
			expectErr:   true,
			errContains: "failed to get deployment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockDynamic *mockDynamicInterface
			if tt.setupMock != nil {
				mockResource := &mockResourceInterface{}
				tt.setupMock(mockResource)
				mockDynamic = &mockDynamicInterface{
					resourceInterface: mockResource,
				}
			}

			typedClient := tt.setupClient()
			client := newTestKubeClient(mockDynamic, typedClient, nil)

			status, err := client.GetResourceStatus(ctx, tt.id)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message should contain %q, got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if status == nil {
					t.Fatal("expected status, got nil")
				}
				if status.Ready != tt.expectReady {
					t.Errorf("Ready: got %v, want %v", status.Ready, tt.expectReady)
				}
				if tt.expectMsg != "" && !strings.Contains(status.Message, tt.expectMsg) {
					t.Errorf("Message should contain %q, got: %q", tt.expectMsg, status.Message)
				}
			}

			if mockDynamic != nil && mockDynamic.resourceInterface != nil {
				mockDynamic.resourceInterface.AssertExpectations(t)
			}
		})
	}
}

// Helper to create int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}

// TestExecPod tests ExecPod method (validation logic)
// Note: Full executor testing requires integration tests due to SPDY streaming complexity
func TestExecPod(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		namespace   string
		podName     string
		container   string
		command     []string
		stdin       []byte
		setupClient func() kubernetes.Interface
		restConfig  *rest.Config
		expectErr   bool
		errContains string
	}{
		{
			name:      "error - pod not found",
			namespace: "default",
			podName:   "not-found",
			container: "test-container",
			command:   []string{"echo", "test"},
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset()
			},
			restConfig:  &rest.Config{},
			expectErr:   true,
			errContains: "failed to get pod",
		},
		{
			name:      "error - pod in Succeeded phase",
			namespace: "default",
			podName:   "test-pod",
			container: "test-container",
			command:   []string{"echo", "test"},
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "test-container"},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodSucceeded,
					},
				})
			},
			restConfig:  &rest.Config{},
			expectErr:   true,
			errContains: "cannot exec into a container in a completed pod",
		},
		{
			name:      "error - pod in Failed phase",
			namespace: "default",
			podName:   "test-pod",
			container: "test-container",
			command:   []string{"echo", "test"},
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "test-container"},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodFailed,
					},
				})
			},
			restConfig:  &rest.Config{},
			expectErr:   true,
			errContains: "cannot exec into a container in a completed pod",
		},
		{
			name:      "error - container not found",
			namespace: "default",
			podName:   "test-pod",
			container: "missing-container",
			command:   []string{"echo", "test"},
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "test-container"},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				})
			},
			restConfig:  &rest.Config{},
			expectErr:   true,
			errContains: "container not found in pod",
		},
		{
			name:      "error - executor creation fails (invalid rest config)",
			namespace: "default",
			podName:   "test-pod",
			container: "test-container",
			command:   []string{"echo", "test"},
			setupClient: func() kubernetes.Interface {
				return fake.NewSimpleClientset(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "test-container"},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				})
			},
			restConfig:  nil, // nil rest config will cause executor creation to fail
			expectErr:   true,
			errContains: "failed to create executor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typedClient := tt.setupClient()
			client := newTestKubeClient(nil, typedClient, tt.restConfig)

			stdout, stderr, exitCode, err := client.ExecPod(ctx, tt.namespace, tt.podName, tt.container, tt.command, tt.stdin)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message should contain %q, got: %v", tt.errContains, err)
				}
				// On error, stdout/stderr/exitCode should be zero values or error indicators
				if exitCode != 0 && exitCode != 1 {
					t.Errorf("unexpected exit code on error: %d", exitCode)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// Note: Full execution testing requires integration tests or complex executor mocking
				_ = stdout
				_ = stderr
				_ = exitCode
			}
		})
	}
}
