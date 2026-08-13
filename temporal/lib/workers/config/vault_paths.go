package config

import "strings"

// Kubernetes namespaces owners are deployed into. Vault authority is derived
// from these (see VaultPathAllowedForKubernetesNamespace), and they double as
// Temporal namespace names.
const (
	OperatorKubernetesNamespace    = "fabric"
	ShardKubernetesNamespacePrefix = "fabric-shard-"
)

// Vault KV v2 mount prefixes. Policies address the data and metadata planes
// separately, so both appear in every owner policy.
const (
	VaultKVDataPrefix     = "secret/data"
	VaultKVMetadataPrefix = "secret/metadata"
)

// operatorVaultPrefix is the root every owner's subtree lives under, and
// shardVaultPrefix is the part of it reserved for shards. Both mirror
// OperatorConfig.VaultBasePath / OrgConfig.VaultBasePath.
const (
	operatorVaultPrefix = OperatorKubernetesNamespace + "/"
	shardVaultPrefix    = operatorVaultPrefix + "shard/"
)

// VaultPathAllowedForKubernetesNamespace reports whether a Vault policy path may
// be granted to a service account in kubernetesNamespace.
//
// A Nexus handler cannot identify its caller (Temporal supplies no caller
// identity to handlers), so CreateAccess binds authority to the Kubernetes
// namespace the policy is being attached to instead: whoever asks, a policy
// mounted in a shard's namespace can only reach that shard's Vault subtree. The
// caller namespace itself is gated separately by the Temporal server Authorizer
// in temporal-server/authz.
//
// The mapping is the inverse of OwnerConfig.KubernetesNamespace() and
// OwnerConfig.VaultBasePath():
//
//	fabric              → fabric/{operatorID}, i.e. anything under fabric/ that is not a shard's
//	fabric-shard-{org}  → fabric/shard/{org}
//
// Any other namespace is denied. The operator rule is a subtree exclusion rather
// than an exact base path because the operator ID is not knowable here, and the
// operator owns the whole non-shard remainder of the tree anyway.
func VaultPathAllowedForKubernetesNamespace(kubernetesNamespace, path string) bool {
	// Vault treats these literally rather than resolving them, so a policy
	// containing one is malformed at best and an escape attempt at worst.
	if strings.Contains(path, "..") {
		return false
	}

	rel, ok := kvRelativePath(path)
	if !ok {
		return false
	}

	if kubernetesNamespace == OperatorKubernetesNamespace {
		return strings.HasPrefix(rel, operatorVaultPrefix) && !strings.HasPrefix(rel, shardVaultPrefix)
	}
	org, ok := shardOrg(kubernetesNamespace)
	if !ok {
		return false
	}
	return strings.HasPrefix(rel, shardVaultPrefix+org+"/")
}

// IsOwnerKubernetesNamespace reports whether a namespace belongs to a fabric
// owner and therefore has a Vault subtree at all. Namespaces outside this set
// (platform namespaces such as workers or observe) get their Vault access from
// Terraform, never from CreateAccess.
func IsOwnerKubernetesNamespace(kubernetesNamespace string) bool {
	if kubernetesNamespace == OperatorKubernetesNamespace {
		return true
	}
	_, ok := shardOrg(kubernetesNamespace)
	return ok
}

func shardOrg(kubernetesNamespace string) (string, bool) {
	org, ok := strings.CutPrefix(kubernetesNamespace, ShardKubernetesNamespacePrefix)
	if !ok || org == "" {
		return "", false
	}
	return org, true
}

// kvRelativePath strips the KV v2 data/metadata prefix, returning the owner
// portion of the path. Paths outside the KV mount are rejected.
func kvRelativePath(path string) (string, bool) {
	for _, prefix := range []string{VaultKVDataPrefix + "/", VaultKVMetadataPrefix + "/"} {
		if rel, ok := strings.CutPrefix(path, prefix); ok {
			return rel, true
		}
	}
	return "", false
}
