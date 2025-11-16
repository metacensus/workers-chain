package temporal

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Activities implements the AdminActivities interface from generated code
type Activities struct {
	namespaceClient client.NamespaceClient
	operatorService operatorservice.OperatorServiceClient
}

// NewActivities creates a new Activities instance
func NewActivities(namespaceClient client.NamespaceClient, operatorService operatorservice.OperatorServiceClient) *Activities {
	return &Activities{
		namespaceClient: namespaceClient,
		operatorService: operatorService,
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

// CreateNexusEndpoint creates a global Nexus endpoint for cross-namespace workflow calls
// Idempotent: returns success if endpoint already exists
func (a *Activities) CreateNexusEndpoint(ctx context.Context, req *pb.CreateNexusEndpointInput) (*pb.CreateNexusEndpointOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating Nexus endpoint",
		"name", req.Name,
		"namespace", req.Namespace,
		"task_queue", req.TaskQueue)

	if req.Name == "" {
		return nil, fmt.Errorf("endpoint name cannot be empty")
	}

	// Build CreateNexusEndpointRequest with nexus/v1 types
	createReq := &operatorservice.CreateNexusEndpointRequest{
		Spec: &nexus.EndpointSpec{
			Name: req.Name,
			Target: &nexus.EndpointTarget{
				Variant: &nexus.EndpointTarget_Worker_{
					Worker: &nexus.EndpointTarget_Worker{
						Namespace: req.Namespace,
						TaskQueue: req.TaskQueue,
					},
				},
			},
		},
	}

	// Call Temporal operator service
	resp, err := a.operatorService.CreateNexusEndpoint(ctx, createReq)
	if err != nil {
		// Idempotent: treat "already exists" as success
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "AlreadyExists") {
			logger.Info("Nexus endpoint already exists, treating as success",
				"name", req.Name)
			return &pb.CreateNexusEndpointOutput{}, nil
		}
		logger.Error("Failed to create Nexus endpoint",
			"error", err,
			"endpoint_name", req.Name)
		return nil, fmt.Errorf("create nexus endpoint: %w", err)
	}

	endpointID := ""
	if resp.Endpoint != nil {
		endpointID = resp.Endpoint.Id
	}

	logger.Info("Nexus endpoint created successfully",
		"name", req.Name,
		"id", endpointID)

	return &pb.CreateNexusEndpointOutput{
		Id: endpointID,
	}, nil
}

// DescribeNexusEndpoint retrieves the current configuration of a Nexus endpoint by ID
// Used to fetch existing endpoint state before updating
func (a *Activities) DescribeNexusEndpoint(ctx context.Context, req *pb.DescribeNexusEndpointInput) (*pb.DescribeNexusEndpointOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Describing Nexus endpoint", "id", req.Id)

	if req.Id == "" {
		return nil, fmt.Errorf("endpoint id cannot be empty")
	}

	// Call GetNexusEndpoint with ID
	getReq := &operatorservice.GetNexusEndpointRequest{
		Id: req.Id,
	}

	resp, err := a.operatorService.GetNexusEndpoint(ctx, getReq)
	if err != nil {
		logger.Error("Failed to describe Nexus endpoint",
			"error", err,
			"endpoint_id", req.Id)
		return nil, fmt.Errorf("describe nexus endpoint: %w", err)
	}

	endpoint := &pb.Endpoint{}
	if resp.Endpoint != nil {
		endpoint.Id = resp.Endpoint.Id
		if resp.Endpoint.Spec != nil {
			endpoint.Name = resp.Endpoint.Spec.Name
			if resp.Endpoint.Spec.Target != nil && resp.Endpoint.Spec.Target.GetWorker() != nil {
				worker := resp.Endpoint.Spec.Target.GetWorker()
				endpoint.Namespace = worker.Namespace
				endpoint.TaskQueue = worker.TaskQueue
			}
		}
	}

	logger.Info("Nexus endpoint described successfully",
		"id", req.Id,
		"name", endpoint.Name)

	return &pb.DescribeNexusEndpointOutput{
		Endpoint: endpoint,
	}, nil
}

// UpdateNexusEndpoint updates an existing Nexus endpoint configuration
// REPLACES entire endpoint configuration (caller must provide complete desired state including ID)
func (a *Activities) UpdateNexusEndpoint(ctx context.Context, req *pb.UpdateNexusEndpointInput) (*pb.UpdateNexusEndpointOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Updating Nexus endpoint",
		"id", req.Endpoint.Id,
		"name", req.Endpoint.Name)

	if req.Endpoint.Id == "" {
		return nil, fmt.Errorf("endpoint id cannot be empty")
	}
	if req.Endpoint.Name == "" {
		return nil, fmt.Errorf("endpoint name cannot be empty")
	}

	// Build UpdateNexusEndpointRequest with nexus/v1 types
	updateReq := &operatorservice.UpdateNexusEndpointRequest{
		Spec: &nexus.EndpointSpec{
			Name: req.Endpoint.Name,
			Target: &nexus.EndpointTarget{
				Variant: &nexus.EndpointTarget_Worker_{
					Worker: &nexus.EndpointTarget_Worker{
						Namespace: req.Endpoint.Namespace,
						TaskQueue: req.Endpoint.TaskQueue,
					},
				},
			},
		},
	}

	// Call Temporal operator service
	_, err := a.operatorService.UpdateNexusEndpoint(ctx, updateReq)
	if err != nil {
		logger.Error("Failed to update Nexus endpoint",
			"error", err,
			"endpoint_id", req.Endpoint.Id)
		return nil, fmt.Errorf("update nexus endpoint: %w", err)
	}

	logger.Info("Nexus endpoint updated successfully",
		"id", req.Endpoint.Id)

	return &pb.UpdateNexusEndpointOutput{}, nil
}
