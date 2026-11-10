package authz

import (
	"context"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/server/api/matchingservice/v1"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/resource"
	"google.golang.org/grpc"
)

// guardedMatchingClient applies the endpoint→payload binding to Nexus
// dispatch in the frontend.
//
// This is the first seam where the payload exists. nexusHandler.StartOperation
// runs the Authorizer before it reads the request body into the dispatch, so a
// payload check there only ever sees nil; the body is in place by the time the
// handler hands the dispatch to matching, and the frontend has stamped the
// endpoint name on it (Request.Endpoint).
type guardedMatchingClient struct {
	resource.MatchingClient

	logger log.Logger
}

// NexusMatchingClient wraps the frontend's matching client with
// checkCertificateIssuance. Wired as an fx decoration of the frontend and
// internal-frontend services.
func NexusMatchingClient(inner resource.MatchingClient, logger log.Logger) resource.MatchingClient {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	return &guardedMatchingClient{MatchingClient: inner, logger: logger}
}

func (c *guardedMatchingClient) DispatchNexusTask(
	ctx context.Context,
	request *matchingservice.DispatchNexusTaskRequest,
	opts ...grpc.CallOption,
) (*matchingservice.DispatchNexusTaskResponse, error) {
	endpoint := request.GetRequest().GetEndpoint()
	if err := checkCertificateIssuance(endpoint, request); err != nil {
		c.logger.Warn("nexus dispatch denied",
			tag.NewStringTag("nexus-endpoint", endpoint),
			tag.NewStringTag("reason", err.Error()),
		)
		return nil, serviceerror.NewPermissionDeniedf("nexus dispatch denied", "%v", err)
	}
	return c.MatchingClient.DispatchNexusTask(ctx, request, opts...)
}
