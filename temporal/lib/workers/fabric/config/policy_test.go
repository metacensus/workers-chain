package config

import (
	"fmt"
	"strings"
	"testing"

	fabricv1 "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	vaultv1 "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
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

// TestVaultPolicies_AcceptedByCreateAccess is the other half of the CreateAccess
// scope check (workerconfig.VaultPathAllowedForKubernetesNamespace): every policy
// bootstrap actually asks for must be accepted for the owner's namespace. If this
// fails, either a policy reaches outside its owner's subtree or the scope check is
// too strict and would break a deploy.
func TestVaultPolicies_AcceptedByCreateAccess(t *testing.T) {
	for _, owner := range []OwnerConfig{NewOperatorConfig("metacensus"), NewOrgConfig("org1")} {
		caRef := &fabricv1.CertificateAuthorityRef{CertType: "org", Id: "int"}
		policies := map[string]*vaultv1.PolicyDocument{
			"worker":                WorkerVaultPolicy(owner),
			"certificate-store":     NewCertificateStoreConfig(owner).VaultPolicy(),
			"certificate-authority": NewCertificateAuthorityConfig(owner, caRef).VaultPolicy(),
			"couchdb":               NewCouchDbConfig(owner, "peer0").VaultPolicy(),
			"peer":                  NewPeerConfig(owner, "peer0").VaultPolicy(),
			"orderer":               NewOrdererConfig(owner, "orderer0").VaultPolicy(),
		}

		namespace := owner.KubernetesNamespace()
		for component, policy := range policies {
			if len(policy.GetRules()) == 0 {
				t.Errorf("%s/%s: policy has no rules, so the scope check proves nothing", namespace, component)
			}
			for _, rule := range policy.GetRules() {
				if !workerconfig.VaultPathAllowedForKubernetesNamespace(namespace, rule.GetPath()) {
					t.Errorf("%s/%s: CreateAccess would reject %q", namespace, component, rule.GetPath())
				}
			}
		}
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
