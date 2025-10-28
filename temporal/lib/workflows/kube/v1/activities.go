package kube

import (
	"context"
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	"go.temporal.io/sdk/activity"
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
func (a *Activities) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating resources", "resource_count", len(req.GetResources()))

	response := &pb.CreateResponse{
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

	logger.Info("Resources created successfully",
		"total", response.Summary.Total,
		"created", response.Summary.Created,
		"failed", response.Summary.Failed)

	return response, nil
}

// createSingleResource creates a single resource in the cluster
func (a *Activities) createSingleResource(ctx context.Context, resourceID string, resource *pb.Resource) (*pb.CreateResult, error) {
	logger := activity.GetLogger(ctx)

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
		logger.Warn("Failed to get resource status after create", "resource_id", resourceID, "error", err)
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
func (a *Activities) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Updating resources", "resource_count", len(req.GetResources()))

	response := &pb.UpdateResponse{
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

	logger.Info("Resources updated successfully",
		"total", response.Summary.Total,
		"updated", response.Summary.Updated,
		"failed", response.Summary.Failed)

	return response, nil
}

// updateSingleResource updates a single resource in the cluster
func (a *Activities) updateSingleResource(ctx context.Context, resourceID string, resource *pb.Resource) (*pb.UpdateResult, error) {
	logger := activity.GetLogger(ctx)

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
		logger.Warn("Failed to get resource status after update", "resource_id", resourceID, "error", err)
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
func (a *Activities) Apply(ctx context.Context, req *pb.ApplyRequest) (*pb.ApplyResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Applying resources", "resource_count", len(req.GetResources()))

	response := &pb.ApplyResponse{
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

	logger.Info("Resources applied successfully",
		"total", response.Summary.Total,
		"applied", response.Summary.Applied,
		"failed", response.Summary.Failed)

	return response, nil
}

// applySingleResource applies a single resource to the cluster
func (a *Activities) applySingleResource(ctx context.Context, resourceID string, resource *pb.Resource) (*pb.ApplyResult, error) {
	logger := activity.GetLogger(ctx)

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
		logger.Warn("Failed to get resource status after apply", "resource_id", resourceID, "error", err)
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
func (a *Activities) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Deleting resources", "resource_count", len(req.GetResources()))

	response := &pb.DeleteResponse{
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

	logger.Info("Resources deleted successfully",
		"total", response.Summary.Total,
		"deleted", response.Summary.Deleted,
		"failed", response.Summary.Failed)

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
func (a *Activities) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Getting resource status", "resource_count", len(req.GetResources()))

	response := &pb.GetStatusResponse{
		Results:       make(map[string]*pb.ResourceStatus),
		OverallStatus: "unknown",
	}

	if len(req.GetResources()) == 0 {
		logger.Warn("No resources provided for status check")
		return response, nil
	}

	allReady := true
	anyDegraded := false
	anyError := false

	for resourceID, resource := range req.GetResources() {
		status := a.getStatusForResource(ctx, resourceID, resource)
		response.Results[resourceID] = status

		// Track overall status
		switch status.Status {
		case StatusReady:
			// Good
		case StatusDegraded, StatusPending:
			allReady = false
			anyDegraded = true
		default:
			allReady = false
			anyError = true
		}
	}

	// Determine overall status
	if anyError {
		response.OverallStatus = StatusError
	} else if allReady {
		response.OverallStatus = StatusReady
	} else if anyDegraded {
		response.OverallStatus = StatusDegraded
	} else {
		response.OverallStatus = StatusPending
	}

	logger.Info("Status retrieved",
		"overall_status", response.OverallStatus,
		"resource_count", len(response.Results))

	return response, nil
}

// getStatusForResource retrieves the status for a single resource
func (a *Activities) getStatusForResource(ctx context.Context, resourceID string, resource *pb.Resource) *pb.ResourceStatus {
	logger := activity.GetLogger(ctx)

	// Parse resource to extract resource identifier
	id, _, err := a.parseResource(resourceID, resource)
	if err != nil {
		logger.Error("Failed to parse resource for status check", "resource_id", resourceID, "error", err)
		return &pb.ResourceStatus{
			Status:  StatusError,
			Message: err.Error(),
		}
	}

	status, err := a.getResourceStatus(ctx, id)
	if err != nil {
		logger.Error("Failed to get resource status", "resource_id", resourceID, "error", err)
		return &pb.ResourceStatus{
			Kind:      id.Kind,
			Name:      id.Name,
			Namespace: id.Namespace,
			Status:    StatusError,
			Message:   err.Error(),
		}
	}

	return status
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
	logger := activity.GetLogger(ctx)
	pollingInterval := 1 * time.Second
	deadline := time.Now().Add(timeout)

	for {
		// Check if all resources are ready
		allReady := true
		for _, id := range resources {
			status, err := a.client.GetResourceStatus(ctx, id)
			if err != nil {
				logger.Warn("Failed to get resource status during polling", "resource", id.Name, "error", err)
				allReady = false
				break
			}
			if !status.Ready {
				allReady = false
				break
			}
		}

		if allReady {
			logger.Info("All resources are ready")
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
	logger := activity.GetLogger(ctx)
	pollingInterval := 1 * time.Second
	deadline := time.Now().Add(timeout)

	for {
		// Check if all resources are deleted
		allDeleted := true
		for _, id := range resources {
			status, err := a.client.GetResourceStatus(ctx, id)
			if err != nil {
				// Resource not found or can't be accessed - consider it deleted
				logger.Debug("Resource not found during deletion poll", "resource", id.Name)
				continue
			}
			if status != nil {
				// Resource still exists
				allDeleted = false
				break
			}
		}

		if allDeleted {
			logger.Info("All resources are deleted")
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
