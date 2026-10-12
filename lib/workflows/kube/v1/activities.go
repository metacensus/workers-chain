package kube

import (
	"context"
	"fmt"
	"time"

	pb "github.com/metacensus/workers-chain/gen/workflows/kube/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// Status constants for ResourceStatus.Status field
const (
	StatusReady    = "ready"
	StatusPending  = "pending"
	StatusDegraded = "degraded"
	StatusError    = "error"
	StatusUnknown  = "unknown"
)

// Activities implements the KubeProviderServiceActivities interface from generated code
type Activities struct {
	client  KubeClient
	decoder runtime.Decoder
}

// NewActivities creates a new Activities instance
func NewActivities(client KubeClient) *Activities {
	return &Activities{
		client:  client,
		decoder: yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
	}
}

// Create creates new resources in the cluster
func (a *Activities) Create(ctx context.Context, req *pb.CreateInput) (*pb.CreateOutput, error) {
	response := &pb.CreateOutput{
		Results: make(map[string]*pb.CreateResult),
		Summary: &pb.CreateSummary{
			Total: int32(len(req.GetResources())),
		},
	}

	// Sequential creation - fail fast on first error
	for resourceID, resource := range req.GetResources() {
		result, err := a.createSingleResource(ctx, resourceID, resource)
		if err != nil {
			result = &pb.CreateResult{
				ResourceId:   resourceID,
				ErrorMessage: err.Error(),
			}
			response.Results[resourceID] = result
			response.Summary.Failed++
			return response, err
		}

		response.Results[resourceID] = result
		response.Summary.Created++
	}

	return response, nil
}

// createSingleResource creates a single resource in the cluster
func (a *Activities) createSingleResource(ctx context.Context, resourceID string, resource *pb.Resource) (*pb.CreateResult, error) {

	// Parse resource to extract resource identifier and object
	id, obj, err := a.parseResource(resourceID, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource %s: %w", resourceID, err)
	}

	// Create the resource
	err = a.client.CreateResource(ctx, id, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource %s: %w", resourceID, err)
	}

	// Get resource status after create
	status, err := a.getResourceStatus(ctx, id)
	if err != nil {
		// Don't fail if status check fails - resource was created

		status = &pb.ResourceStatus{
			Kind:      id.Kind,
			Name:      id.Name,
			Namespace: id.Namespace,
			Status:    StatusUnknown,
			Message:   fmt.Sprintf("created but status check failed: %v", err),
		}
	}

	return &pb.CreateResult{
		ResourceId:     resourceID,
		ResourceStatus: status,
	}, nil
}

// Update updates existing resources in the cluster
func (a *Activities) Update(ctx context.Context, req *pb.UpdateInput) (*pb.UpdateOutput, error) {
	response := &pb.UpdateOutput{
		Results: make(map[string]*pb.UpdateResult),
		Summary: &pb.UpdateSummary{
			Total: int32(len(req.GetResources())),
		},
	}

	// Sequential update - fail fast on first error
	for resourceID, resource := range req.GetResources() {
		result, err := a.updateSingleResource(ctx, resourceID, resource)
		if err != nil {
			result = &pb.UpdateResult{
				ResourceId:   resourceID,
				ErrorMessage: err.Error(),
			}
			response.Results[resourceID] = result
			response.Summary.Failed++
			return response, err
		}

		response.Results[resourceID] = result
		response.Summary.Updated++
	}

	return response, nil
}

// updateSingleResource updates a single resource in the cluster
func (a *Activities) updateSingleResource(ctx context.Context, resourceID string, resource *pb.Resource) (*pb.UpdateResult, error) {

	// Parse resource to extract resource identifier and object
	id, obj, err := a.parseResource(resourceID, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource %s: %w", resourceID, err)
	}

	// Update the resource
	err = a.client.UpdateResource(ctx, id, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource %s: %w", resourceID, err)
	}

	// Get resource status after update
	status, err := a.getResourceStatus(ctx, id)
	if err != nil {
		// Don't fail if status check fails - resource was updated
		status = &pb.ResourceStatus{
			Kind:      id.Kind,
			Name:      id.Name,
			Namespace: id.Namespace,
			Status:    StatusUnknown,
			Message:   fmt.Sprintf("updated but status check failed: %v", err),
		}
	}

	return &pb.UpdateResult{
		ResourceId:     resourceID,
		ResourceStatus: status,
	}, nil
}

// Apply deploys resources to the cluster (create or update)
func (a *Activities) Apply(ctx context.Context, req *pb.ApplyInput) (*pb.ApplyOutput, error) {
	response := &pb.ApplyOutput{
		Results: make(map[string]*pb.ApplyResult),
		Summary: &pb.ApplySummary{
			Total: int32(len(req.GetResources())),
		},
	}

	// Sequential application - fail fast on first error
	for resourceID, resource := range req.GetResources() {
		result, err := a.applySingleResource(ctx, resourceID, resource)
		if err != nil {
			result = &pb.ApplyResult{
				ResourceId:   resourceID,
				ErrorMessage: err.Error(),
			}
			response.Results[resourceID] = result
			response.Summary.Failed++
			return response, err
		}

		response.Results[resourceID] = result
		response.Summary.Applied++
	}

	return response, nil
}

// applySingleResource applies a single resource to the cluster
func (a *Activities) applySingleResource(ctx context.Context, resourceID string, resource *pb.Resource) (*pb.ApplyResult, error) {

	// Parse resource to extract resource identifier and object
	id, obj, err := a.parseResource(resourceID, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource %s: %w", resourceID, err)
	}

	// Apply the resource
	err = a.client.ApplyResource(ctx, id, obj)
	if err != nil {
		return nil, fmt.Errorf("failed to apply resource %s: %w", resourceID, err)
	}

	// Get resource status after apply
	status, err := a.getResourceStatus(ctx, id)
	if err != nil {
		// Don't fail if status check fails - resource was applied

		status = &pb.ResourceStatus{
			Kind:      id.Kind,
			Name:      id.Name,
			Namespace: id.Namespace,
			Status:    StatusUnknown,
			Message:   fmt.Sprintf("applied but status check failed: %v", err),
		}
	}

	return &pb.ApplyResult{
		ResourceId:     resourceID,
		ResourceStatus: status,
	}, nil
}

// Delete removes resources from the cluster
func (a *Activities) Delete(ctx context.Context, req *pb.DeleteInput) (*pb.DeleteOutput, error) {
	response := &pb.DeleteOutput{
		Results: make(map[string]*pb.DeleteResult),
		Summary: &pb.DeleteSummary{
			Total: int32(len(req.GetResources())),
		},
	}

	// Sequential deletion - fail fast on first error
	for resourceID, resource := range req.GetResources() {
		result, err := a.deleteSingleResource(ctx, resourceID, resource)
		if err != nil {
			result = &pb.DeleteResult{
				ResourceId:   resourceID,
				ErrorMessage: err.Error(),
			}
			response.Results[resourceID] = result
			response.Summary.Failed++
			return response, err
		}

		response.Results[resourceID] = result
		response.Summary.Deleted++
	}

	return response, nil
}

// deleteSingleResource deletes a single resource from the cluster
func (a *Activities) deleteSingleResource(ctx context.Context, resourceID string, resource *pb.Resource) (*pb.DeleteResult, error) {
	// Parse resource to extract resource identifier
	// Note: We only need the resource identifier (id), not the full object
	id, _, err := a.parseResource(resourceID, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource %s: %w", resourceID, err)
	}

	// Delete the resource
	err = a.client.DeleteResource(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete resource %s: %w", resourceID, err)
	}

	return &pb.DeleteResult{
		ResourceId: resourceID,
	}, nil
}

// GetStatus retrieves the current status of deployed resources
// Returns error if any resources are not ready (for use with activity retry policies)
func (a *Activities) GetStatus(ctx context.Context, req *pb.GetStatusInput) (*pb.GetStatusOutput, error) {

	response := &pb.GetStatusOutput{
		ResourcesReady: []string{},
	}

	if len(req.GetResources()) == 0 {

		return response, nil
	}

	var notReadyResources []string

	for resourceID, resourceIdentifier := range req.GetResources() {
		status, err := a.getResourceStatus(ctx, resourceIdentifier)
		if err != nil {

			notReadyResources = append(notReadyResources,
				fmt.Sprintf("%s/%s (error: %v)", resourceIdentifier.Kind, resourceIdentifier.Name, err))
			continue
		}

		if status.Status == StatusReady {
			response.ResourcesReady = append(response.ResourcesReady, resourceID)
		} else {
			notReadyResources = append(notReadyResources,
				fmt.Sprintf("%s/%s (status: %s)", resourceIdentifier.Kind, resourceIdentifier.Name, status.Status))
		}
	}

	// Return error if not all resources are ready (triggers activity retry)
	if len(notReadyResources) > 0 {

		return response, fmt.Errorf("resources not ready: %v", notReadyResources)
	}

	return response, nil
}

// parseResource parses YAML and extracts resource identifier and object
func (a *Activities) parseResource(resourceID string, resource *pb.Resource) (*pb.ResourceIdentifier, *unstructured.Unstructured, error) {
	// Parse YAML
	obj := &unstructured.Unstructured{}
	_, gvk, err := a.decoder.Decode(resource.GetManifestYaml(), nil, obj)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode resource %s: %w", resourceID, err)
	}

	// Extract resource metadata from YAML
	id := &pb.ResourceIdentifier{
		ApiVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(), // Empty for cluster-scoped resources
	}

	return id, obj, nil
}

// getResourceStatus converts client status to proto status
func (a *Activities) getResourceStatus(ctx context.Context, id *pb.ResourceIdentifier) (*pb.ResourceStatus, error) {
	status, err := a.client.GetResourceStatus(ctx, id)
	if err != nil {
		return nil, err
	}

	pbStatus := &pb.ResourceStatus{
		Kind:            id.Kind,
		Name:            id.Name,
		Namespace:       id.Namespace,
		Message:         status.Message,
		DesiredReplicas: status.DesiredReplicas,
		ReadyReplicas:   status.ReadyReplicas,
	}

	if status.Ready {
		pbStatus.Status = StatusReady
	} else if status.ReadyReplicas > 0 && status.DesiredReplicas > 0 {
		pbStatus.Status = StatusDegraded
	} else {
		pbStatus.Status = StatusPending
	}

	return pbStatus, nil
}

// pollResourcesReady waits for resources to reach ready state
// Polls status up to timeout duration
func (a *Activities) pollResourcesReady(ctx context.Context, resources []*pb.ResourceIdentifier, timeout time.Duration) error {

	pollingInterval := 1 * time.Second
	deadline := time.Now().Add(timeout)

	for {
		// Check if all resources are ready
		allReady := true
		for _, id := range resources {
			status, err := a.client.GetResourceStatus(ctx, id)
			if err != nil {

				allReady = false
				break
			}
			if !status.Ready {
				allReady = false
				break
			}
		}

		if allReady {

			return nil
		}

		// Check timeout
		if time.Now().Add(pollingInterval).After(deadline) {
			return fmt.Errorf("timeout waiting for resources to be ready")
		}

		// Sleep before next poll
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollingInterval):
			// Continue polling
		}
	}
}

// pollResourcesDeleted waits for resources to be deleted
// Polls until resources no longer exist or timeout
func (a *Activities) pollResourcesDeleted(ctx context.Context, resources []*pb.ResourceIdentifier, timeout time.Duration) error {
	pollingInterval := 1 * time.Second
	deadline := time.Now().Add(timeout)

	for {
		// Check if all resources are deleted
		allDeleted := true
		for _, id := range resources {
			status, err := a.client.GetResourceStatus(ctx, id)
			if err != nil {
				// Resource not found or can't be accessed - consider it deleted
				continue
			}
			if status != nil {
				// Resource still exists
				allDeleted = false
				break
			}
		}

		if allDeleted {

			return nil
		}

		// Check timeout
		if time.Now().Add(pollingInterval).After(deadline) {
			return fmt.Errorf("timeout waiting for resources to be deleted")
		}

		// Sleep before next poll
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollingInterval):
			// Continue polling
		}
	}
}

// pollResourceReady adapts batch polling to single-resource interface
// Used when polling for individual resource readiness
func (a *Activities) pollResourceReady(ctx context.Context, resourceID *pb.ResourceIdentifier, timeout time.Duration) error {
	return a.pollResourcesReady(ctx, []*pb.ResourceIdentifier{resourceID}, timeout)
}

// pollResourceDeleted adapts batch polling to single-resource interface
// Used when polling for individual resource deletion
func (a *Activities) pollResourceDeleted(ctx context.Context, resourceID *pb.ResourceIdentifier, timeout time.Duration) error {
	return a.pollResourcesDeleted(ctx, []*pb.ResourceIdentifier{resourceID}, timeout)
}

// ExecPod executes a command in a pod container
func (a *Activities) ExecPod(ctx context.Context, req *pb.ExecPodInput) (*pb.ExecPodOutput, error) {
	// Validate input
	id := req.GetContainerId()
	if id == nil || id.GetNamespace() == "" || id.GetName() == "" {
		return nil, fmt.Errorf("container_id with namespace and name is required")
	}
	if req.GetContainer() == "" {
		return nil, fmt.Errorf("container is required")
	}
	if len(req.GetCommand()) == 0 {
		return nil, fmt.Errorf("command is required")
	}

	// Execute command
	stdout, stderr, exitCode, err := a.client.ExecPod(
		ctx,
		id,
		req.GetContainer(),
		req.GetCommand(),
		req.GetStdin(),
	)

	output := &pb.ExecPodOutput{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}

	if err != nil {
		output.ErrorMessage = err.Error()

		// Return the output even on error so caller can see stdout/stderr
		return output, err
	}

	return output, nil
}
