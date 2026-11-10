package authz

import (
	"context"
	"fmt"
	"strings"

	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

type Authorizer struct {
	logger log.Logger
}

func NewAuthorizer(logger log.Logger) authorization.Authorizer {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	return &Authorizer{logger: logger}
}

func (a *Authorizer) Authorize(_ context.Context, claims *authorization.Claims, target *authorization.CallTarget) (authorization.Result, error) {
	if target == nil {
		return a.denied("missing call target"), nil
	}

	// Nexus dispatch is checked before the system-role shortcut below. An
	// in-workflow Nexus call arrives here from the history service under the
	// internode identity, so the caller's namespace is unavailable (the
	// caller→endpoint allowlist runs in the history service instead — see
	// nexus_registry.go) but the endpoint name and request payload are not.
	if ep := target.NexusEndpointName; ep != "" {
		if err := checkCertificateIssuance(ep, target.Request); err != nil {
			return a.denied("nexus endpoint %q payload rejected: %v", ep, err), nil
		}
	}

	if target.Namespace == "temporal-system" {
		return allow(), nil
	}

	if claims == nil {
		return a.denied("unauthenticated caller for api %s namespace %q", target.APIName, target.Namespace), nil
	}

	if claims.System&(authorization.RoleAdmin|authorization.RoleWriter) != 0 {
		return allow(), nil
	}

	callerNS := claims.Subject
	if callerNS == "" {
		return a.denied("caller identity missing subject for api %s", target.APIName), nil
	}

	if ep := target.NexusEndpointName; ep != "" {
		if Allowed(callerNS, ep) {
			return allow(), nil
		}
		return a.denied("caller namespace %q may not invoke nexus endpoint %q", callerNS, ep), nil
	}

	if callerNS == NamespaceInfra && infraMayOperate(target.APIName, target.Namespace) {
		return allow(), nil
	}

	if target.Namespace == "" {
		// SDK CheckHealth is grpc.health.v1.Health/Check (no namespace). Worker probes use it.
		if strings.Contains(target.APIName, "GetClusterInfo") ||
			strings.Contains(target.APIName, "GetSystemInfo") ||
			strings.Contains(target.APIName, "ListNamespaces") ||
			strings.Contains(target.APIName, "grpc.health.v1.Health") {
			return allow(), nil
		}
		return a.denied("caller namespace %q denied namespace-less api %s", callerNS, target.APIName), nil
	}

	if role, ok := claims.Namespaces[target.Namespace]; ok && role != authorization.RoleUndefined {
		return allow(), nil
	}

	return a.denied("caller namespace %q has no role on target namespace %q (api %s)", callerNS, target.Namespace, target.APIName), nil
}

func allow() authorization.Result {
	return authorization.Result{Decision: authorization.DecisionAllow}
}

func (a *Authorizer) denied(format string, args ...any) authorization.Result {
	reason := fmt.Sprintf(format, args...)
	a.logger.Warn("authorization denied", tag.NewStringTag("reason", reason))
	return authorization.Result{Decision: authorization.DecisionDeny, Reason: reason}
}
