package authz

import (
	"strings"
)

// Endpoint names registered by cell bootstrap workflows.
// Keep in sync with config.OrgConfig's NexusEndpoint* methods (lib/workers/config)
// and lib/cell/v1/workflows.go in this repository.
const (
	EndpointSecretsCreateAccess   = "secrets-CreateAccess" // pragma: allowlist secret
	EndpointFabricBootstrapFabric = "fabric-BootstrapFabric"

	suffixBootstrapShard    = "-BootstrapShard"
	suffixCreateCertificate = "-CreateCertificate"
)

const (
	NamespaceInfra       = "infra"
	NamespaceFabric      = "fabric"
	namespaceShardPrefix = "fabric-shard-"
)

const (
	apiRegisterNamespace   = "/temporal.api.workflowservice.v1.WorkflowService/RegisterNamespace"
	apiDescribeNamespace   = "/temporal.api.workflowservice.v1.WorkflowService/DescribeNamespace"
	apiCreateNexusEndpoint = "/temporal.api.operatorservice.v1.OperatorService/CreateNexusEndpoint"
)

// Allowed reports whether callerNS may invoke nexusEndpoint.
//
// Enforced on the in-workflow path by the endpoint registry decorator (see
// nexus_registry.go), which is the only place the caller's namespace is known.
func Allowed(callerNS, nexusEndpoint string) bool {
	if nexusEndpoint == "" {
		return false
	}
	switch {
	case callerNS == NamespaceInfra:
		return infraMayCall(nexusEndpoint)
	case callerNS == NamespaceFabric:
		return nexusEndpoint == EndpointSecretsCreateAccess
	default:
		org, ok := OrgForShardNamespace(callerNS)
		if !ok {
			return false
		}
		return shardMayCall(org, nexusEndpoint)
	}
}

// infraMayCall covers the bootstrap fan-out: the infra worker starts the
// operator's fabric bootstrap and each shard's bootstrap.
func infraMayCall(endpoint string) bool {
	if endpoint == EndpointFabricBootstrapFabric {
		return true
	}
	_, ok := orgForShardEndpoint(endpoint, suffixBootstrapShard)
	return ok
}

// shardMayCall restricts a shard to its own certificate-issuance endpoint. The
// per-org endpoint is what makes cross-org issuance refusable: the endpoint
// name identifies the org, and only that org's namespace may reach it.
func shardMayCall(org, endpoint string) bool {
	if endpoint == EndpointSecretsCreateAccess {
		return true
	}
	certOrg, ok := orgForShardEndpoint(endpoint, suffixCreateCertificate)
	return ok && certOrg == org
}

// infraMayOperate is the infra worker's non-Nexus exception: create/describe
// fabric and shard Temporal namespaces, and CreateNexusEndpoint (that RPC has
// no GetNamespace(), so it cannot be scoped the same way).
func infraMayOperate(apiName, targetNS string) bool {
	switch apiName {
	case apiRegisterNamespace, apiDescribeNamespace:
		if targetNS == NamespaceFabric {
			return true
		}
		_, ok := OrgForShardNamespace(targetNS)
		return ok
	case apiCreateNexusEndpoint:
		return true
	default:
		return false
	}
}

// OrgForCreateCertificateEndpoint returns the org a per-org certificate
// issuance endpoint belongs to. The Authorizer uses it to bind the endpoint to
// the requester declared in the payload.
func OrgForCreateCertificateEndpoint(endpoint string) (string, bool) {
	return orgForShardEndpoint(endpoint, suffixCreateCertificate)
}

// OrgForShardNamespace returns the org owning a shard Temporal namespace.
func OrgForShardNamespace(namespace string) (string, bool) {
	if !strings.HasPrefix(namespace, namespaceShardPrefix) {
		return "", false
	}
	return validOrg(strings.TrimPrefix(namespace, namespaceShardPrefix))
}

func orgForShardEndpoint(endpoint, suffix string) (string, bool) {
	if !strings.HasPrefix(endpoint, namespaceShardPrefix) || !strings.HasSuffix(endpoint, suffix) {
		return "", false
	}
	return validOrg(strings.TrimSuffix(strings.TrimPrefix(endpoint, namespaceShardPrefix), suffix))
}

// validOrg rejects org IDs that could smuggle a separator into a derived name.
func validOrg(org string) (string, bool) {
	if org == "" || strings.ContainsAny(org, "/@.") {
		return "", false
	}
	return org, true
}
