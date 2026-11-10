package authz

import (
	"context"

	"go.temporal.io/api/serviceerror"
	persistencespb "go.temporal.io/server/api/persistence/v1"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/namespace"
	commonnexus "go.temporal.io/server/common/nexus"
)

// guardedEndpointRegistry applies the Nexus allowlist to endpoint resolution in
// the history service.
//
// This is the only seam where a caller namespace is available for an
// in-workflow Nexus call: `ScheduleNexusOperation` resolves the endpoint
// through EndpointRegistry.GetByName with the calling workflow's namespace ID,
// and Temporal's command handler turns a PermissionDenied from that lookup into
// a workflow task failure. The frontend Authorizer cannot do this job — the
// history service dispatches to its own frontend over HTTP as the internode
// identity, so the caller's namespace is not part of that request.
type guardedEndpointRegistry struct {
	commonnexus.EndpointRegistry

	namespaces namespace.Registry
	logger     log.Logger
}

// NexusEndpointRegistry wraps an endpoint registry with the caller-namespace
// allowlist. Wired as an fx decoration of the history service's registry.
func NexusEndpointRegistry(
	inner commonnexus.EndpointRegistry,
	namespaces namespace.Registry,
	logger log.Logger,
) commonnexus.EndpointRegistry {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	return &guardedEndpointRegistry{
		EndpointRegistry: inner,
		namespaces:       namespaces,
		logger:           logger,
	}
}

func (r *guardedEndpointRegistry) GetByName(
	ctx context.Context,
	callerNamespaceID namespace.ID,
	endpointName string,
) (*persistencespb.NexusEndpointEntry, error) {
	callerNS, err := r.namespaces.GetNamespaceName(callerNamespaceID)
	if err != nil {
		return nil, err
	}
	caller := callerNS.String()
	if !Allowed(caller, endpointName) {
		r.logger.Warn("nexus endpoint denied for caller namespace",
			tag.WorkflowNamespace(caller),
			tag.NewStringTag("nexus-endpoint", endpointName),
			tag.NewStringTag("reason", "caller namespace may not invoke nexus endpoint"),
		)
		return nil, serviceerror.NewPermissionDeniedf(
			"nexus endpoint not allowed for caller namespace",
			"caller namespace %q may not invoke nexus endpoint %q", caller, endpointName,
		)
	}
	return r.EndpointRegistry.GetByName(ctx, callerNamespaceID, endpointName)
}
