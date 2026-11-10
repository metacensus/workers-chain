package config

import (
	"strings"

	vaultv1 "github.com/metacensus/workers-chain/gen/vault/v1"
)

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

// VaultRuleAllowedForKubernetesNamespace is what CreateAccess may grant a
// policy bound to the namespace: its own subtree, plus read of the operator's
// TLS CA signing certificates, which every org's peers and orderers mount for
// their TLS trust store. Those secrets hold only {certificate, chain}; the
// CAs' keys are under keys/, which this never admits.
func VaultRuleAllowedForKubernetesNamespace(kubernetesNamespace string, rule *vaultv1.PolicyRule) bool {
	if VaultPathAllowedForKubernetesNamespace(kubernetesNamespace, rule.GetPath()) {
		return true
	}
	return isOperatorTLSCACertPath(rule.GetPath()) && isReadOnly(rule.GetCapabilities())
}

// isOperatorTLSCACertPath matches exactly
// secret/data/fabric/{operatorID}/certs/ca/tls/{caID}/signing, with no
// wildcard in any segment.
func isOperatorTLSCACertPath(path string) bool {
	rel, ok := strings.CutPrefix(path, VaultKVDataPrefix+"/")
	if !ok || strings.HasPrefix(rel, shardVaultPrefix) {
		return false
	}
	segments := strings.Split(rel, "/")
	if len(segments) != 7 {
		return false
	}
	for _, segment := range segments {
		if segment == "" || segment == ".." || strings.ContainsAny(segment, "*+") {
			return false
		}
	}
	return segments[0] == OperatorKubernetesNamespace &&
		segments[2] == "certs" && segments[3] == "ca" && segments[4] == "tls" &&
		segments[6] == "signing"
}

func isReadOnly(capabilities []vaultv1.Capability) bool {
	if len(capabilities) == 0 {
		return false
	}
	for _, capability := range capabilities {
		if capability != vaultv1.Capability_CAPABILITY_READ {
			return false
		}
	}
	return true
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
