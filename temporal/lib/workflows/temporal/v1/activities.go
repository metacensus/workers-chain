package temporal

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/temporal/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/nexus/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Sentinel errors for temporal activities
var (
	ErrRegisterNamespace     = errors.New("register namespace")
	ErrDescribeNamespace     = errors.New("describe namespace")
	ErrCreateNexusEndpoint   = errors.New("create nexus endpoint")
	ErrDescribeNexusEndpoint = errors.New("describe nexus endpoint")
	ErrUpdateNexusEndpoint   = errors.New("update nexus endpoint")
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
	// Determine retention period: use provided value or default to 30 days (2592000 seconds)
	var retentionPeriod *durationpb.Duration
	if req.Retention != nil {
		retentionPeriod = req.Retention
	} else {
		// Default to 30 days
		retentionPeriod = durationpb.New(30 * 24 * time.Hour)
	}

	// Create RegisterNamespaceRequest for Temporal admin API
	registerReq := &workflowservice.RegisterNamespaceRequest{
		Namespace:                        req.Name,
		WorkflowExecutionRetentionPeriod: retentionPeriod,
		IsGlobalNamespace:                false,
		ActiveClusterName:                "active", // Default cluster
	}

	// Call Temporal admin API.
	// Idempotent: a namespace that already exists is treated as success. The server
	// returns a typed *serviceerror.NamespaceAlreadyExists; match the type rather than
	// the message text (which varies by server version, e.g. "already exists").
	if err := a.namespaceClient.Register(ctx, registerReq); err != nil {
		var alreadyExists *serviceerror.NamespaceAlreadyExists
		if errors.As(err, &alreadyExists) {
			return &pb.RegisterNamespaceOutput{}, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrRegisterNamespace, err)
	}

	return &pb.RegisterNamespaceOutput{}, nil
}

// DescribeNamespace retrieves the current state of a Temporal namespace
// Used by the workflow to poll until namespace is active
func (a *Activities) DescribeNamespace(ctx context.Context, req *pb.DescribeNamespaceInput) (*pb.DescribeNamespaceOutput, error) {
	// Call Temporal admin API
	resp, err := a.namespaceClient.Describe(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDescribeNamespace, err)
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

	return &pb.DescribeNamespaceOutput{
		Name:  req.Name,
		State: state,
	}, nil
}

// CreateNexusEndpoint creates a global Nexus endpoint for cross-namespace workflow calls
// Idempotent: returns success if endpoint already exists
func (a *Activities) CreateNexusEndpoint(ctx context.Context, req *pb.CreateNexusEndpointInput) (*pb.CreateNexusEndpointOutput, error) {
	logger := activity.GetLogger(ctx)

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
		// Idempotent: an endpoint that already exists is treated as success. The server
		// returns a typed *serviceerror.AlreadyExists (message: "... already registered");
		// match the type rather than the message text, which varies by server version.
		var alreadyExists *serviceerror.AlreadyExists
		if errors.As(err, &alreadyExists) {
			logger.Info("Nexus endpoint already exists, treating as success",
				"name", req.Name)
			return &pb.CreateNexusEndpointOutput{}, nil
		}
		return nil, fmt.Errorf("%w: %w", ErrCreateNexusEndpoint, err)
	}

	endpointID := ""
	if resp.Endpoint != nil {
		endpointID = resp.Endpoint.Id
	}

	return &pb.CreateNexusEndpointOutput{
		Id: endpointID,
	}, nil
}

// DescribeNexusEndpoint retrieves the current configuration of a Nexus endpoint by ID
// Used to fetch existing endpoint state before updating
func (a *Activities) DescribeNexusEndpoint(ctx context.Context, req *pb.DescribeNexusEndpointInput) (*pb.DescribeNexusEndpointOutput, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("endpoint id cannot be empty")
	}

	// Call GetNexusEndpoint with ID
	getReq := &operatorservice.GetNexusEndpointRequest{
		Id: req.Id,
	}

	resp, err := a.operatorService.GetNexusEndpoint(ctx, getReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDescribeNexusEndpoint, err)
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

	return &pb.DescribeNexusEndpointOutput{
		Endpoint: endpoint,
	}, nil
}

// UpdateNexusEndpoint updates an existing Nexus endpoint configuration
// REPLACES entire endpoint configuration (caller must provide complete desired state including ID)
func (a *Activities) UpdateNexusEndpoint(ctx context.Context, req *pb.UpdateNexusEndpointInput) (*pb.UpdateNexusEndpointOutput, error) {
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
		return nil, fmt.Errorf("%w: %w", ErrUpdateNexusEndpoint, err)
	}

	return &pb.UpdateNexusEndpointOutput{}, nil
}
