package component

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const ApplyActivityName = "component.Apply"

// ApplyInput enhanced input with more options
type ApplyInput struct {
	Component    string
	Version      string
	Namespace    string // default namespace if not specified in manifest
	FieldManager string // defaults to "temporal-deployment"
}

// ApplyOutput provides detailed results
type ApplyOutput struct {
	Results        []ApplyResult
	TotalResources int
	Created        int
	Updated        int
	Unchanged      int
	Failed         int
}

// ApplyResult tracks what happened to each resource
type ApplyResult struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
	Action     ApplyAction
	Error      string
}

type ApplyAction string

const (
	ApplyActionCreated   ApplyAction = "Created"
	ApplyActionUpdated   ApplyAction = "Updated"
	ApplyActionUnchanged ApplyAction = "Unchanged"
	ApplyActionFailed    ApplyAction = "Failed"
)

// Apply applies a series resource and returns the result
func (a *Activities) Apply(ctx context.Context, input ApplyInput) (*ApplyOutput, error) {
	logger := activity.GetLogger(ctx)
	output := &ApplyOutput{
		Results: []ApplyResult{},
	}

	// Set defaults
	if input.FieldManager == "" {
		input.FieldManager = "temporal-deployment"
	}

	resources, err := a.store.Get(ctx, input.Component, input.Version)
	if err != nil {
		return nil, temporal.NewApplicationError("failed to read manifest content", "READ_ERROR", err)
	}

	for i, obj := range resources {
		// Heartbeat for long-running operations
		if i%10 == 0 {
			activity.RecordHeartbeat(ctx, fmt.Sprintf("Processing resource %d/%d", i+1, len(resources)))
		}

		result := a.applyResource(ctx, obj, input.FieldManager)
		output.Results = append(output.Results, result)

		switch result.Action {
		case ApplyActionCreated:
			output.Created++
		case ApplyActionUpdated:
			output.Updated++
		case ApplyActionUnchanged:
			output.Unchanged++
		case ApplyActionFailed:
			output.Failed++
			if isCriticalResource(obj) {
				return output, temporal.NewApplicationError(
					fmt.Sprintf("failed to apply critical resource %s/%s@%s", result.Kind, obj.GetName(), result.APIVersion),
					"CRITICAL_RESOURCE_FAILED",
					err,
				)
			}
		}
	}

	logger.Info("Apply completed",
		"total", output.TotalResources,
		"created", output.Created,
		"updated", output.Updated,
		"unchanged", output.Unchanged,
		"failed", output.Failed,
	)

	if output.Failed > 0 {
		return output, temporal.NewApplicationError(
			fmt.Sprintf("%d resources failed to apply", output.Failed),
			"PARTIAL_FAILURE",
			nil,
		)
	}

	return output, nil
}

// applyResource applies a single resource and returns the result
func (a *Activities) applyResource(ctx context.Context, obj *unstructured.Unstructured, fieldManager string) ApplyResult {
	gvk := obj.GroupVersionKind()
	result := ApplyResult{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
	}

	resourceInterface, err := a.kubernetes.GetResourceInterface(obj)
	if err != nil {
		result.Action = ApplyActionFailed
		result.Error = err.Error()
		return result
	}

	// Apply options
	applyOptions := v1.ApplyOptions{
		FieldManager: fieldManager,
	}

	// Check if resource exists
	existing, err := resourceInterface.Get(ctx, obj.GetName(), v1.GetOptions{})
	resourceExists := err == nil

	// Server-side apply
	applied, err := resourceInterface.Apply(ctx, obj.GetName(), obj, applyOptions)
	if err != nil {
		if errors.IsConflict(err) {
			err = fmt.Errorf("Conflict error: %w", err)
		}
		result.Action = ApplyActionFailed
		result.Error = err.Error()
		return result
	}

	// Determine what happened
	if !resourceExists {
		result.Action = ApplyActionCreated
		return result
	} else if existing.GetResourceVersion() != applied.GetResourceVersion() {
		result.Action = ApplyActionUpdated
		return result
	}
	result.Action = ApplyActionUnchanged
	return result
}

// isCriticalResource determines if a resource failure should stop the entire apply
func isCriticalResource(obj *unstructured.Unstructured) bool {
	criticalKinds := []string{
		"Namespace",
		"CustomResourceDefinition",
		"StorageClass",
		"PersistentVolume",
	}

	kind := obj.GetKind()
	for _, critical := range criticalKinds {
		if kind == critical {
			return true
		}
	}
	return false
}
