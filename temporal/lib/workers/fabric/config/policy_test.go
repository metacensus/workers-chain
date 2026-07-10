package config

import (
	"fmt"
	"strings"
	"testing"
)

// TestWorkerVaultPolicy_OnlyOwnSubtree asserts that every worker - operator or
// shard - gets ONLY its own-subtree rules and NO cross-owner rules. Under the
// CSR-based issuance flow the operator never reads or writes shard-owned Vault
// paths, so Vault access must align exactly with ownership.
func TestWorkerVaultPolicy_OnlyOwnSubtree(t *testing.T) {
	cases := []struct {
		name  string
		owner OwnerConfig
	}{
		{name: "operator", owner: NewOperatorConfig("metacensus")},
		{name: "shard", owner: NewOrgConfig("org1")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			policy := WorkerVaultPolicy(tc.owner)

			ownDataPrefix := fmt.Sprintf("%s/%s/", vaultSecretDataPrefix, tc.owner.VaultBasePath())
			ownMetadataPrefix := fmt.Sprintf("%s/%s/", vaultSecretMetadataPrefix, tc.owner.VaultBasePath())

			for _, rule := range policy.GetRules() {
				path := rule.GetPath()
				if !strings.HasPrefix(path, ownDataPrefix) && !strings.HasPrefix(path, ownMetadataPrefix) {
					t.Errorf("rule path %q is outside owner subtree (data=%q metadata=%q)",
						path, ownDataPrefix, ownMetadataPrefix)
				}
			}
		})
	}
}

// TestWorkerVaultPolicy_NoCrossOwnerWildcard asserts symmetrically that no
// worker (operator or shard) is granted the single-segment "+" wildcard used to
// reach other owners' subtrees.
func TestWorkerVaultPolicy_NoCrossOwnerWildcard(t *testing.T) {
	crossOwnerBase := NewOrgConfig("+").VaultBasePath()

	for _, owner := range []OwnerConfig{NewOperatorConfig("metacensus"), NewOrgConfig("org1")} {
		policy := WorkerVaultPolicy(owner)
		for _, rule := range policy.GetRules() {
			if strings.Contains(rule.GetPath(), crossOwnerBase) {
				t.Errorf("owner %q has cross-owner rule %q (contains %q)",
					owner.OwnerID(), rule.GetPath(), crossOwnerBase)
			}
		}
	}
}

