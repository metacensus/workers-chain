package config

import "strings"

const (
	OperatorKubernetesNamespace    = "fabric"
	ShardKubernetesNamespacePrefix = "fabric-shard-"
)

const (
	VaultKVDataPrefix     = "secret/data"
	VaultKVMetadataPrefix = "secret/metadata"
)

const (
	operatorVaultPrefix = OperatorKubernetesNamespace + "/"
	shardVaultPrefix    = operatorVaultPrefix + "shard/"
)

// VaultPathAllowedForKubernetesNamespace is the inverse of OwnerConfig
// KubernetesNamespace / VaultBasePath. CreateAccess cannot identify the Nexus
// caller, so it scopes the grant to the namespace the policy is bound to.
//
//	fabric              → fabric/{operatorID} (anything under fabric/ that is not a shard)
//	fabric-shard-{org}  → fabric/shard/{org}
func VaultPathAllowedForKubernetesNamespace(kubernetesNamespace, path string) bool {
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

func kvRelativePath(path string) (string, bool) {
	for _, prefix := range []string{VaultKVDataPrefix + "/", VaultKVMetadataPrefix + "/"} {
		if rel, ok := strings.CutPrefix(path, prefix); ok {
			return rel, true
		}
	}
	return "", false
}
