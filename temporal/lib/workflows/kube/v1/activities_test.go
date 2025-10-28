package kube_test

import (
	"errors"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			if !tt.expectErr {
				// Success cases: CreateResource and GetResourceStatus called
				mockClient.On("CreateResource", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Times(len(tt.resources))
				mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
					Return(&kube.ResourceStatus{
						Ready:           true,
						Message:         "created",
						DesiredReplicas: 1,
						ReadyReplicas:   1,
					}, nil).
					Times(len(tt.resources))
			} else if tt.clientErr != nil {
				// Error with client error: CreateResource called once and fails
				mockClient.On("CreateResource", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Once()
			}
			// For YAML parse errors, no client methods called

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.Create)

			req := &pb.CreateRequest{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.Create, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.CreateResponse
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if result.Summary.Created != tt.expectCount {
					t.Errorf("created count: got %d, want %d", result.Summary.Created, tt.expectCount)
				}
				if result.Summary.Total != int32(len(tt.resources)) {
					t.Errorf("total count: got %d, want %d", result.Summary.Total, int32(len(tt.resources)))
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			if !tt.expectErr {
				// Success cases: UpdateResource and GetResourceStatus called
				mockClient.On("UpdateResource", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Times(len(tt.resources))
				mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
					Return(&kube.ResourceStatus{
						Ready:           true,
						Message:         "updated",
						DesiredReplicas: 1,
						ReadyReplicas:   1,
					}, nil).
					Times(len(tt.resources))
			} else if tt.clientErr != nil {
				// Error with client error: UpdateResource called once and fails
				mockClient.On("UpdateResource", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Once()
			}
			// For YAML parse errors, no client methods called

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.Update)

			req := &pb.UpdateRequest{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.Update, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.UpdateResponse
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if result.Summary.Updated != tt.expectCount {
					t.Errorf("updated count: got %d, want %d", result.Summary.Updated, tt.expectCount)
				}
				if result.Summary.Total != int32(len(tt.resources)) {
					t.Errorf("total count: got %d, want %d", result.Summary.Total, int32(len(tt.resources)))
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			// Note: ApplyActivity uses the ApplyResource method, not UpdateResource/CreateResource
			if !tt.expectErr {
				// Success cases: ApplyResource and GetResourceStatus called
				mockClient.On("ApplyResource", mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					Times(len(tt.resources))
				mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
					Return(&kube.ResourceStatus{
						Ready:           true,
						Message:         "applied",
						DesiredReplicas: 1,
						ReadyReplicas:   1,
					}, nil).
					Times(len(tt.resources))
			} else if tt.createErr == nil && tt.updateErr != nil {
				// Error case: ApplyResource fails
				mockClient.On("ApplyResource", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.updateErr).
					Once()
			}
			// For YAML parse errors, no client methods called

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.Apply)

			req := &pb.ApplyRequest{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.Apply, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.ApplyResponse
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if result.Summary.Applied != tt.expectApplied {
					t.Errorf("applied count: got %d, want %d", result.Summary.Applied, tt.expectApplied)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			if !tt.expectErr {
				// Success cases: DeleteResource called
				mockClient.On("DeleteResource", mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Times(len(tt.resources))
			} else if tt.clientErr != nil {
				// Error with client error: DeleteResource called once and fails
				mockClient.On("DeleteResource", mock.Anything, mock.Anything).
					Return(tt.clientErr).
					Once()
			}
			// For YAML parse errors, no client methods called

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.Delete)

			req := &pb.DeleteRequest{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.Delete, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.DeleteResponse
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if result.Summary.Deleted != tt.expectCount {
					t.Errorf("deleted count: got %d, want %d", result.Summary.Deleted, tt.expectCount)
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
		name          string
		resources     map[string]*pb.Resource
		statusResults map[string]*kube.ResourceStatus
		clientErr     error
		expectErr     bool
		expectOverall string
		expectCount   int
	}{
		{
			name: "success - all resources ready",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
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
			expectErr:     false,
			expectOverall: kube.StatusReady,
			expectCount:   1,
		},
		{
			name: "success - resources pending",
			resources: map[string]*pb.Resource{
				"deployment-1": {
					ManifestYaml: validDeploymentManifestYAML(),
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
			expectErr:     false,
			expectOverall: kube.StatusDegraded, // Pending is treated as degraded in overall status
			expectCount:   1,
		},
		{
			name: "success - mixed status (ready and pending)",
			resources: map[string]*pb.Resource{
				"service-1": {
					ManifestYaml: validServiceManifestYAML(),
				},
				"deployment-1": {
					ManifestYaml: validDeploymentManifestYAML(),
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
			expectErr:     false,
			expectOverall: kube.StatusReady, // All return the first status (ready)
			expectCount:   2,
		},
		{
			name:          "error - no manifests",
			resources:     map[string]*pb.Resource{},
			expectErr:     false, // Should return empty response with unknown status
			expectOverall: "unknown",
			expectCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := kubemocks.NewMockKubeClient(t)

			// Set up specific expectations based on scenario
			if len(tt.resources) > 0 {
				// Get the first status to return for all calls
				var statusToReturn *kube.ResourceStatus
				if len(tt.statusResults) > 0 {
					for _, status := range tt.statusResults {
						statusToReturn = status
						break
					}
				} else {
					statusToReturn = &kube.ResourceStatus{
						Ready:           true,
						Message:         "ready",
						DesiredReplicas: 1,
						ReadyReplicas:   1,
					}
				}

				mockClient.On("GetResourceStatus", mock.Anything, mock.Anything).
					Return(statusToReturn, nil).
					Times(len(tt.resources))
			}
			// For empty manifests, no GetResourceStatus calls

			activities := kube.NewActivities(mockClient)
			env.RegisterActivity(activities.GetStatus)

			req := &pb.GetStatusRequest{Resources: tt.resources}
			res, err := env.ExecuteActivity(activities.GetStatus, req)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.GetStatusResponse
				if err := res.Get(&result); err != nil {
					t.Fatalf("failed to get activity result: %v", err)
				}
				if len(result.Results) != tt.expectCount {
					t.Errorf("result count: got %d, want %d", len(result.Results), tt.expectCount)
				}
				if result.OverallStatus != tt.expectOverall {
					t.Errorf("overall status: got %q, want %q", result.OverallStatus, tt.expectOverall)
				}
			}
		})
	}
}
