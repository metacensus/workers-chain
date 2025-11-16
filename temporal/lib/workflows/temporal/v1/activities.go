package temporal

import (
	"context"
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/durationpb"
)

// RegisterNamespaceRequest and DescribeNamespaceResponse are re-exported from workflowservice

// Activities implements the AdminActivities interface from generated code
type Activities struct {
	namespaceClient client.NamespaceClient
}

// NewActivities creates a new Activities instance
func NewActivities(client client.NamespaceClient) *Activities {
	return &Activities{
		namespaceClient: client,
	}
}

// RegisterNamespace creates a new Temporal namespace with specified retention policy
// Idempotent: returns success if namespace already exists
func (a *Activities) RegisterNamespace(ctx context.Context, req *pb.RegisterNamespaceInput) (*pb.RegisterNamespaceOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Registering namespace",
		"name", req.Name)

	// Determine retention period: use provided value or default to 30 days (2592000 seconds)
	var retentionPeriod *durationpb.Duration
	if req.Retention != nil {
		retentionPeriod = req.Retention
	} else {
		// Default to 30 days
		retentionPeriod = durationpb.New(30 * 24 * time.Hour)
	}

	logger.Info("Using retention period",
		"namespace", req.Name,
		"duration", retentionPeriod)

	// Create RegisterNamespaceRequest for Temporal admin API
	registerReq := &workflowservice.RegisterNamespaceRequest{
		Namespace:                        req.Name,
		WorkflowExecutionRetentionPeriod: retentionPeriod,
		IsGlobalNamespace:                false,
		ActiveClusterName:                "active", // Default cluster
	}

	// Call Temporal admin API
	// This is idempotent - if the namespace already exists, it returns success
	if err := a.namespaceClient.Register(ctx, registerReq); err != nil {
		logger.Error("Failed to register namespace",
			"error", err,
			"namespace", req.Name)
		return nil, fmt.Errorf("register namespace: %w", err)
	}

	logger.Info("Namespace registered successfully",
		"namespace", req.Name,
		"retention_duration", retentionPeriod.String())

	return &pb.RegisterNamespaceOutput{}, nil
}

// DescribeNamespace retrieves the current state of a Temporal namespace
// Used by the workflow to poll until namespace is active
func (a *Activities) DescribeNamespace(ctx context.Context, req *pb.DescribeNamespaceInput) (*pb.DescribeNamespaceOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Describing namespace", "name", req.Name)

	// Call Temporal admin API
	resp, err := a.namespaceClient.Describe(ctx, req.Name)
	if err != nil {
		logger.Error("Failed to describe namespace",
			"error", err,
			"namespace", req.Name)
		return nil, fmt.Errorf("describe namespace: %w", err)
	}

	// Determine namespace state from NamespaceInfo
	state := pb.NamespaceState_UNSPECIFIED
	if resp.NamespaceInfo != nil {
		switch resp.NamespaceInfo.State {
		case enums.NAMESPACE_STATE_REGISTERED:
			state = pb.NamespaceState_REGISTERED
		case enums.NAMESPACE_STATE_DEPRECATED:
			state = pb.NamespaceState_DEPRECATED
		}
	}

	logger.Info("Namespace described successfully",
		"name", req.Name,
		"state", state.String())

	return &pb.DescribeNamespaceOutput{
		Name:  req.Name,
		State: state,
	}, nil
}
