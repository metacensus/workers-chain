package kube_test

import (
	"errors"
	"fmt"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	kube "github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	kubemocks "github.com/metacensus/infra/temporal/mocks/workflows/kube/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

var (
	// Test error sentinels for client errors
	errClientNotFound      = errors.New("resource not found")
	errClientAlreadyExists = errors.New("resource already exists")
	errClientForbidden     = errors.New("forbidden")
)

// Helper to create a valid manifest YAML for a simple Service
func validServiceManifestYAML() []byte {
	return []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: default
spec:
  ports:
  - port: 80
    targetPort: 8080
`)
}

// Helper to create a valid manifest YAML for a Deployment
func validDeploymentManifestYAML() []byte {
	return []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: default
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
        image: nginx:latest
`)
}

// Helper to create an invalid manifest YAML
func invalidManifestYAML() []byte {
	return []byte("not: valid: yaml: syntax:")
}

func TestCreateActivity(t *testing.T) {
	tests := []struct {
		name        string
		resources   map[string]*pb.Resource
		clientErr   error
		expectErr   bool
		expectCount int32
	}{
		{
			name: "success - single service",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			expectErr:   false,
			expectCount: 1,
		},
		{
			name: "success - multiple manifests",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
				"deployment-1": {
					ManifestYaml: validDeploymentManifestYAML(),
				},
			},
			expectErr:   false,
			expectCount: 2,
		},
		{
			name: "error - invalid manifest YAML",
			resources: map[string]*pb.Resource{
				"invalid": {
					ManifestYaml: invalidManifestYAML(),
				},
			},
			expectErr: true,
		},
		{
			name: "error - client returns error",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			clientErr: errClientAlreadyExists,
			expectErr: true,
		},
		{
			name: "success - status check fails after create",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			expectErr:   false,
			expectCount: 1,
		},
		{
			name:      "success - empty resource list",
			resources: map[string]*pb.Resource{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			if !tt.expectErr {
				if len(tt.resources) > 0 {
					// Success cases: CreateResource and GetResourceStatus called
					mockClient.On("CreateResource", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.clientErr).
						Times(len(tt.resources))
					// For "status check fails after create" test case, GetResourceStatus fails
					if tt.name == "success - status check fails after create" {
						mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
							Return(nil, errClientForbidden).
							Times(len(tt.resources))
					} else {
						mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
							Return(&kube.ResourceStatus{
								Ready:           true,
								Message:         "created",
								DesiredReplicas: 1,
								ReadyReplicas:   1,
							}, nil).
							Times(len(tt.resources))
					}
				}
				// For empty resource list, no client calls
			} else if tt.clientErr != nil {
				// Error with client error: CreateResource called once and fails
				mockClient.On("CreateResource", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Once()
			}
			// For YAML parse errors, no client methods called

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.Create)

			req := &pb.CreateInput{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.Create, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.CreateOutput
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if tt.name != "success - empty resource list" {
					if result.Summary.Created != tt.expectCount {
						t.Errorf("created count: got %d, want %d", result.Summary.Created, tt.expectCount)
					}
				} else {
					if result.Summary.Created != 0 {
						t.Errorf("created count: got %d, want 0", result.Summary.Created)
					}
				}
				if result.Summary.Total != int32(len(tt.resources)) {
					t.Errorf("total count: got %d, want %d", result.Summary.Total, int32(len(tt.resources)))
				}
				// Verify StatusUnknown is set when status check fails
				if tt.name == "success - status check fails after create" {
					if len(result.Results) != 1 {
						t.Fatalf("expected 1 result, got %d", len(result.Results))
					}
					for _, res := range result.Results {
						if res.ResourceStatus == nil {
							t.Fatal("expected ResourceStatus to be set")
						}
						if res.ResourceStatus.Status != kube.StatusUnknown {
							t.Errorf("expected StatusUnknown, got %s", res.ResourceStatus.Status)
						}
					}
				}
			}
		})
	}
}

func TestUpdateActivity(t *testing.T) {
	tests := []struct {
		name        string
		resources   map[string]*pb.Resource
		clientErr   error
		expectErr   bool
		expectCount int32
	}{
		{
			name: "success - single service",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			expectErr:   false,
			expectCount: 1,
		},
		{
			name: "success - multiple manifests",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
				"deployment-1": {
					ManifestYaml: validDeploymentManifestYAML(),
				},
			},
			expectErr:   false,
			expectCount: 2,
		},
		{
			name: "error - invalid manifest YAML",
			resources: map[string]*pb.Resource{
				"invalid": {
					ManifestYaml: invalidManifestYAML(),
				},
			},
			expectErr: true,
		},
		{
			name: "error - resource not found",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			clientErr: errClientNotFound,
			expectErr: true,
		},
		{
			name: "success - status check fails after update",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			expectErr:   false,
			expectCount: 1,
		},
		{
			name:      "success - empty resource list",
			resources: map[string]*pb.Resource{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			if !tt.expectErr {
				if len(tt.resources) > 0 {
					// Success cases: UpdateResource and GetResourceStatus called
					mockClient.On("UpdateResource", mock.Anything, mock.Anything, mock.Anything).
						Return(tt.clientErr).
						Times(len(tt.resources))
					// For "status check fails after update" test case, GetResourceStatus fails
					if tt.name == "success - status check fails after update" {
						mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
							Return(nil, errClientForbidden).
							Times(len(tt.resources))
					} else {
						mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
							Return(&kube.ResourceStatus{
								Ready:           true,
								Message:         "updated",
								DesiredReplicas: 1,
								ReadyReplicas:   1,
							}, nil).
							Times(len(tt.resources))
					}
				}
				// For empty resource list, no client calls
			} else if tt.clientErr != nil {
				// Error with client error: UpdateResource called once and fails
				mockClient.On("UpdateResource", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Once()
			}
			// For YAML parse errors, no client methods called

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.Update)

			req := &pb.UpdateInput{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.Update, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.UpdateOutput
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if tt.name != "success - empty resource list" {
					if result.Summary.Updated != tt.expectCount {
						t.Errorf("updated count: got %d, want %d", result.Summary.Updated, tt.expectCount)
					}
				} else {
					if result.Summary.Updated != 0 {
						t.Errorf("updated count: got %d, want 0", result.Summary.Updated)
					}
				}
				if result.Summary.Total != int32(len(tt.resources)) {
					t.Errorf("total count: got %d, want %d", result.Summary.Total, int32(len(tt.resources)))
				}
				// Verify StatusUnknown is set when status check fails
				if tt.name == "success - status check fails after update" {
					if len(result.Results) != 1 {
						t.Fatalf("expected 1 result, got %d", len(result.Results))
					}
					for _, res := range result.Results {
						if res.ResourceStatus == nil {
							t.Fatal("expected ResourceStatus to be set")
						}
						if res.ResourceStatus.Status != kube.StatusUnknown {
							t.Errorf("expected StatusUnknown, got %s", res.ResourceStatus.Status)
						}
					}
				}
			}
		})
	}
}

func TestApplyActivity(t *testing.T) {
	tests := []struct {
		name          string
		resources     map[string]*pb.Resource
		updateErr     error
		createErr     error
		expectErr     bool
		expectApplied int32
	}{
		{
			name: "success - create new resource",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			updateErr:     errClientNotFound, // Triggers create fallback
			createErr:     nil,
			expectErr:     false,
			expectApplied: 1,
		},
		{
			name: "success - update existing resource",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			updateErr:     nil, // Update succeeds
			expectErr:     false,
			expectApplied: 1,
		},
		{
			name: "error - invalid manifest YAML",
			resources: map[string]*pb.Resource{
				"invalid": {
					ManifestYaml: invalidManifestYAML(),
				},
			},
			expectErr: true,
		},
		{
			name: "error - both update and create fail",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			updateErr: errClientForbidden, // Not NotFound, so create won't be tried
			expectErr: true,
		},
		{
			name: "success - status check fails after apply",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			expectErr:     false,
			expectApplied: 1,
		},
		{
			name:      "success - empty resource list",
			resources: map[string]*pb.Resource{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			// Note: ApplyActivity uses the ApplyResource method, not UpdateResource/CreateResource
			if !tt.expectErr {
				if len(tt.resources) > 0 {
					// Success cases: ApplyResource and GetResourceStatus called
					mockClient.On("ApplyResource", mock.Anything, mock.Anything, mock.Anything).
						Return(nil).
						Times(len(tt.resources))
					// For "status check fails after apply" test case, GetResourceStatus fails
					if tt.name == "success - status check fails after apply" {
						mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
							Return(nil, errClientForbidden).
							Times(len(tt.resources))
					} else {
						mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
							Return(&kube.ResourceStatus{
								Ready:           true,
								Message:         "applied",
								DesiredReplicas: 1,
								ReadyReplicas:   1,
							}, nil).
							Times(len(tt.resources))
					}
				}
				// For empty resource list, no client calls
			} else if tt.createErr == nil && tt.updateErr != nil {
				// Error case: ApplyResource fails
				mockClient.On("ApplyResource", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.updateErr).
					Once()
			}
			// For YAML parse errors, no client methods called

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.Apply)

			req := &pb.ApplyInput{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.Apply, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.ApplyOutput
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if tt.name != "success - empty resource list" {
					if result.Summary.Applied != tt.expectApplied {
						t.Errorf("applied count: got %d, want %d", result.Summary.Applied, tt.expectApplied)
					}
				} else {
					if result.Summary.Applied != 0 {
						t.Errorf("applied count: got %d, want 0", result.Summary.Applied)
					}
				}
				// Verify StatusUnknown is set when status check fails
				if tt.name == "success - status check fails after apply" {
					if len(result.Results) != 1 {
						t.Fatalf("expected 1 result, got %d", len(result.Results))
					}
					for _, res := range result.Results {
						if res.ResourceStatus == nil {
							t.Fatal("expected ResourceStatus to be set")
						}
						if res.ResourceStatus.Status != kube.StatusUnknown {
							t.Errorf("expected StatusUnknown, got %s", res.ResourceStatus.Status)
						}
					}
				}
			}
		})
	}
}

func TestDeleteActivity(t *testing.T) {
	tests := []struct {
		name        string
		resources   map[string]*pb.Resource
		clientErr   error
		expectErr   bool
		expectCount int32
	}{
		{
			name: "success - single service",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			expectErr:   false,
			expectCount: 1,
		},
		{
			name: "success - multiple manifests",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
				"deployment-1": {
					ManifestYaml: validDeploymentManifestYAML(),
				},
			},
			expectErr:   false,
			expectCount: 2,
		},
		{
			name: "error - invalid manifest YAML",
			resources: map[string]*pb.Resource{
				"invalid": {
					ManifestYaml: invalidManifestYAML(),
				},
			},
			expectErr: true,
		},
		{
			name: "error - client returns error",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
			},
			clientErr: errClientForbidden,
			expectErr: true,
		},
		{
			name:      "success - empty resource list",
			resources: map[string]*pb.Resource{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			if !tt.expectErr {
				if len(tt.resources) > 0 {
					// Success cases: DeleteResource called
					mockClient.On("DeleteResource", mock.Anything, mock.Anything).
						Return(tt.clientErr).
						Times(len(tt.resources))
				}
				// For empty resource list, no client calls
			} else if tt.clientErr != nil {
				// Error with client error: DeleteResource called once and fails
				mockClient.On("DeleteResource", mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Once()
			}
			// For YAML parse errors, no client methods called

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.Delete)

			req := &pb.DeleteInput{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.Delete, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.DeleteOutput
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if tt.name != "success - empty resource list" {
					if result.Summary.Deleted != tt.expectCount {
						t.Errorf("deleted count: got %d, want %d", result.Summary.Deleted, tt.expectCount)
					}
				} else {
					if result.Summary.Deleted != 0 {
						t.Errorf("deleted count: got %d, want 0", result.Summary.Deleted)
					}
				}
				if result.Summary.Total != int32(len(tt.resources)) {
					t.Errorf("total count: got %d, want %d", result.Summary.Total, int32(len(tt.resources)))
				}
			}
		})
	}
}

func TestGetStatusActivity(t *testing.T) {
	tests := []struct {
		name             string
		resources        map[string]*pb.ResourceIdentifier
		statusResults    map[string]*kube.ResourceStatus
		clientErr        error
		expectErr        bool
		expectReadyCount int
	}{
		{
			name: "success - all resources ready",
			resources: map[string]*pb.ResourceIdentifier{
				"service-1": {
					ApiVersion: "v1",
					Kind:       "Service",
					Name:       "test-service",
					Namespace:  "default",
				},
			},
			statusResults: map[string]*kube.ResourceStatus{
				"service-1": {
					Ready:           true,
					Message:         "ready",
					DesiredReplicas: 1,
					ReadyReplicas:   1,
				},
			},
			expectErr:        false,
			expectReadyCount: 1,
		},
		{
			name: "error - resources pending (not ready)",
			resources: map[string]*pb.ResourceIdentifier{
				"deployment-1": {
					ApiVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
					Namespace:  "default",
				},
			},
			statusResults: map[string]*kube.ResourceStatus{
				"deployment-1": {
					Ready:           false,
					Message:         "0/1 replicas ready",
					DesiredReplicas: 1,
					ReadyReplicas:   0,
				},
			},
			expectErr:        true, // Activity fails if resources not ready
			expectReadyCount: 0,
		},
		{
			name: "error - mixed status (some ready, some not)",
			resources: map[string]*pb.ResourceIdentifier{
				"service-1": {
					ApiVersion: "v1",
					Kind:       "Service",
					Name:       "test-service",
					Namespace:  "default",
				},
				"deployment-1": {
					ApiVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
					Namespace:  "default",
				},
			},
			statusResults: map[string]*kube.ResourceStatus{
				"service-1": {
					Ready:   true,
					Message: "ready",
				},
				"deployment-1": {
					Ready:   false,
					Message: "0/1 replicas ready",
				},
			},
			expectErr:        true, // Activity fails if any resource not ready
			expectReadyCount: 1,    // One resource is ready
		},
		{
			name: "success - multiple resources all ready",
			resources: map[string]*pb.ResourceIdentifier{
				"service-1": {
					ApiVersion: "v1",
					Kind:       "Service",
					Name:       "test-service",
					Namespace:  "default",
				},
				"deployment-1": {
					ApiVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
					Namespace:  "default",
				},
			},
			statusResults: map[string]*kube.ResourceStatus{
				"service-1": {
					Ready:   true,
					Message: "ready",
				},
				"deployment-1": {
					Ready:           true,
					Message:         "1/1 replicas ready",
					DesiredReplicas: 1,
					ReadyReplicas:   1,
				},
			},
			expectErr:        false,
			expectReadyCount: 2,
		},
		{
			name:             "success - no resources",
			resources:        map[string]*pb.ResourceIdentifier{},
			expectErr:        false,
			expectReadyCount: 0,
		},
		{
			name: "error - client error getting status",
			resources: map[string]*pb.ResourceIdentifier{
				"service-1": {
					ApiVersion: "v1",
					Kind:       "Service",
					Name:       "test-service",
					Namespace:  "default",
				},
			},
			clientErr:        errClientForbidden,
			expectErr:        true,
			expectReadyCount: 0,
		},
		{
			name: "error - degraded status (partial replicas)",
			resources: map[string]*pb.ResourceIdentifier{
				"deployment-1": {
					ApiVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "test-deployment",
					Namespace:  "default",
				},
			},
			statusResults: map[string]*kube.ResourceStatus{
				"deployment-1": {
					Ready:           false,
					Message:         "1/2 replicas ready",
					DesiredReplicas: 2,
					ReadyReplicas:   1,
				},
			},
			expectErr:        true, // Activity fails if resources not ready
			expectReadyCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up mock expectations based on scenario
			if len(tt.resources) > 0 && tt.clientErr == nil {
				// Set up GetResourceStatus for each resource
				for resourceID := range tt.resources {
					status := tt.statusResults[resourceID]
					if status == nil {
						// Default status if not specified
						status = &kube.ResourceStatus{
							Ready:   true,
							Message: "ready",
						}
					}
					mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
						Return(status, nil).
						Once()
				}
			} else if tt.clientErr != nil {
				// Client error scenario
				mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
					Return(nil, tt.clientErr).
					Once()
			}
			// For empty resources, no GetResourceStatus calls

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.GetStatus)

			req := &pb.GetStatusInput{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.GetStatus, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				// Error case - activity failed, we're done
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.GetStatusOutput
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if len(result.ResourcesReady) != tt.expectReadyCount {
					t.Errorf("ready count: got %d, want %d", len(result.ResourcesReady), tt.expectReadyCount)
				}
			}
		})
	}
}

func TestExecPodActivity(t *testing.T) {
	tests := []struct {
		name           string
		input          *pb.ExecPodInput
		execErr        error
		execStdout     []byte
		execStderr     []byte
		execExitCode   int32
		expectErr      bool
		expectExitCode int32
	}{
		{
			name: "success - command executes",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "test-container",
				Command:   []string{"echo", "hello"},
			},
			execStdout:     []byte("hello\n"),
			execStderr:     []byte(""),
			execExitCode:   0,
			expectErr:      false,
			expectExitCode: 0,
		},
		{
			name: "success - command with stderr",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "test-container",
				Command:   []string{"sh", "-c", "echo hello >&2"},
			},
			execStdout:     []byte(""),
			execStderr:     []byte("hello\n"),
			execExitCode:   0,
			expectErr:      false,
			expectExitCode: 0,
		},
		{
			name: "success - command with stdin",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "test-container",
				Command:   []string{"cat"},
				Stdin:     []byte("input data"),
			},
			execStdout:     []byte("input data"),
			execStderr:     []byte(""),
			execExitCode:   0,
			expectErr:      false,
			expectExitCode: 0,
		},
		{
			name: "error - pod not found",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "nonexistent-pod",
				Container: "test-container",
				Command:   []string{"echo", "hello"},
			},
			execErr:        errClientNotFound,
			expectErr:      true,
			expectExitCode: 0,
		},
		{
			name: "error - container not found",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "nonexistent-container",
				Command:   []string{"echo", "hello"},
			},
			execErr:        fmt.Errorf("container nonexistent-container not found in pod default/test-pod"),
			expectErr:      true,
			expectExitCode: 0,
		},
		{
			name: "error - command fails",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "test-container",
				Command:   []string{"false"},
			},
			execStdout:     []byte(""),
			execStderr:     []byte(""),
			execExitCode:   1,
			execErr:        fmt.Errorf("command execution failed"),
			expectErr:      true,
			expectExitCode: 1,
		},
		{
			name: "error - missing namespace",
			input: &pb.ExecPodInput{
				Namespace: "",
				PodName:   "test-pod",
				Container: "test-container",
				Command:   []string{"echo", "hello"},
			},
			expectErr: true,
		},
		{
			name: "error - missing pod name",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "",
				Container: "test-container",
				Command:   []string{"echo", "hello"},
			},
			expectErr: true,
		},
		{
			name: "error - missing container",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "",
				Command:   []string{"echo", "hello"},
			},
			expectErr: true,
		},
		{
			name: "error - missing command",
			input: &pb.ExecPodInput{
				Namespace: "default",
				PodName:   "test-pod",
				Container: "test-container",
				Command:   []string{},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up mock expectations only when the activity will call the client.
			// Validation errors (missing namespace/pod/container/command) return before calling ExecPod.
			inputValid := tt.input.GetNamespace() != "" && tt.input.GetPodName() != "" && tt.input.GetContainer() != "" && len(tt.input.GetCommand()) > 0
			if inputValid {
				mockClient.On("ExecPod", mock.Anything, tt.input.GetNamespace(), tt.input.GetPodName(), tt.input.GetContainer(), tt.input.GetCommand(), tt.input.GetStdin()).
					Return(tt.execStdout, tt.execStderr, tt.execExitCode, tt.execErr).
					Once()
			}

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.ExecPod)

			res, err := env.ExecuteActivity(activities.ExecPod, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				// For command execution errors, we still get output (if the framework provides it)
				if tt.execErr != nil && tt.execExitCode != 0 && res != nil {
					var result *pb.ExecPodOutput
					if err := res.Get(&result); err == nil && result != nil {
						if result.ExitCode != tt.expectExitCode {
							t.Errorf("exit code: got %d, want %d", result.ExitCode, tt.expectExitCode)
						}
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.ExecPodOutput
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if string(result.Stdout) != string(tt.execStdout) {
					t.Errorf("stdout: got %q, want %q", string(result.Stdout), string(tt.execStdout))
				}
				if string(result.Stderr) != string(tt.execStderr) {
					t.Errorf("stderr: got %q, want %q", string(result.Stderr), string(tt.execStderr))
				}
				if result.ExitCode != tt.expectExitCode {
					t.Errorf("exit code: got %d, want %d", result.ExitCode, tt.expectExitCode)
				}
			}
		})
	}
}
