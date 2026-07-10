package config

import (
	"fmt"
	"testing"

	vaultv1 "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
)

func TestWorkerVaultPolicy_OperatorCrossShardKeysWriteOnly(t *testing.T) {
	policy := WorkerVaultPolicy(NewOperatorConfig("metacensus"))

	keysPath := fmt.Sprintf("%s/%s", vaultSecretDataPrefix, NewOrgConfig("+").VaultBasePath()+"/keys/*")
	rule := findPolicyRule(policy, keysPath)
	if rule == nil {
		t.Fatalf("expected operator policy to include cross-shard keys rule for %q", keysPath)
	}

	assertCapabilities(t, rule.Capabilities,
		vaultv1.Capability_CAPABILITY_CREATE,
		vaultv1.Capability_CAPABILITY_UPDATE,
	)
	assertNoCapabilities(t, rule.Capabilities,
		vaultv1.Capability_CAPABILITY_READ,
		vaultv1.Capability_CAPABILITY_LIST,
		vaultv1.Capability_CAPABILITY_DELETE,
	)
}

func TestWorkerVaultPolicy_ShardWorkerDoesNotGetCrossShardKeys(t *testing.T) {
	policy := WorkerVaultPolicy(NewOrgConfig("org1"))

	keysPath := fmt.Sprintf("%s/%s", vaultSecretDataPrefix, NewOrgConfig("+").VaultBasePath()+"/keys/*")
	if rule := findPolicyRule(policy, keysPath); rule != nil {
		t.Fatalf("did not expect shard policy to include cross-shard keys rule for %q", keysPath)
	}
}

func findPolicyRule(policy *vaultv1.PolicyDocument, path string) *vaultv1.PolicyRule {
	for _, rule := range policy.GetRules() {
		if rule.GetPath() == path {
			return rule
		}
	}
	return nil
}

func assertCapabilities(t *testing.T, actual []vaultv1.Capability, expected ...vaultv1.Capability) {
	t.Helper()
	actualSet := capabilitySet(actual)
	for _, capability := range expected {
		if !actualSet[capability] {
			t.Fatalf("expected capability %s in %v", capability, actual)
		}
	}
}

func assertNoCapabilities(t *testing.T, actual []vaultv1.Capability, forbidden ...vaultv1.Capability) {
	t.Helper()
	actualSet := capabilitySet(actual)
	for _, capability := range forbidden {
		if actualSet[capability] {
			t.Fatalf("did not expect capability %s in %v", capability, actual)
		}
	}
}

func capabilitySet(capabilities []vaultv1.Capability) map[vaultv1.Capability]bool {
	set := make(map[vaultv1.Capability]bool, len(capabilities))
	for _, capability := range capabilities {
		set[capability] = true
	}
	return set
}
