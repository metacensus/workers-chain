package config

import (
	"fmt"
	"strings"
	"testing"

	fabricv1 "github.com/metacensus/workers-chain/gen/fabric/v1"
	vaultv1 "github.com/metacensus/workers-chain/gen/vault/v1"
	workerconfig "github.com/metacensus/workers-chain/lib/workers/config"
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

// Fail if a real bootstrap policy would be rejected by CreateAccess, or if a
// policy has no rules (the check would prove nothing).
func TestVaultPolicies_AcceptedByCreateAccess(t *testing.T) {
	operator := NewOperatorConfig("metacensus")
	for _, owner := range []OwnerConfig{NewOperatorConfig("metacensus"), NewOrgConfig("org1")} {
		caRef := &fabricv1.CertificateAuthorityRef{CertType: "org", Id: "int"}
		policies := map[string]*vaultv1.PolicyDocument{
			"worker":                WorkerVaultPolicy(owner),
			"certificate-store":     NewCertificateStoreConfig(owner).VaultPolicy(),
			"certificate-authority": NewCertificateAuthorityConfig(owner, caRef).VaultPolicy(),
			"couchdb":               NewCouchDbConfig(owner, "peer0").VaultPolicy(),
			"peer":                  NewPeerConfig(owner, operator, "peer0").VaultPolicy(),
			"orderer":               NewOrdererConfig(owner, operator, "orderer0").VaultPolicy(),
		}

		namespace := owner.KubernetesNamespace()
		for component, policy := range policies {
			if len(policy.GetRules()) == 0 {
				t.Errorf("%s/%s: policy has no rules, so the scope check proves nothing", namespace, component)
			}
			for _, rule := range policy.GetRules() {
				if !workerconfig.VaultRuleAllowedForKubernetesNamespace(namespace, rule) {
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

// TestComponentVaultPolicies_ReadOperatorTLSRootCA asserts that an org's peers
// and orderers may read the operator's TLS root CA certificate their charts
// mount (#37), and nothing else outside their owner's subtree.
func TestComponentVaultPolicies_ReadOperatorTLSRootCA(t *testing.T) {
	operator := NewOperatorConfig("metacensus")
	org := NewOrgConfig("org1")
	want := "secret/data/fabric/metacensus/certs/ca/tls/root/signing"
	policies := map[string]struct {
		tlsRootCA CertificateAuthorityConfig
		policy    *vaultv1.PolicyDocument
	}{
		"peer":    {NewPeerConfig(org, operator, "peer0").TLSRootCA(), NewPeerConfig(org, operator, "peer0").VaultPolicy()},
		"orderer": {NewOrdererConfig(org, operator, "orderer0").TLSRootCA(), NewOrdererConfig(org, operator, "orderer0").VaultPolicy()},
	}
	ownPrefix := fmt.Sprintf("%s/%s/", vaultSecretDataPrefix, org.VaultBasePath())
	ownMetadataPrefix := fmt.Sprintf("%s/%s/", vaultSecretMetadataPrefix, org.VaultBasePath())

	for component, tc := range policies {
		if got := tc.tlsRootCA.VaultSigningCertDataPath(); got != want {
			t.Errorf("%s: chart mounts %q, want %q", component, got, want)
		}
		var granted bool
		for _, rule := range tc.policy.GetRules() {
			path := rule.GetPath()
			switch {
			case strings.HasPrefix(path, ownPrefix), strings.HasPrefix(path, ownMetadataPrefix):
			case path == want:
				granted = true
				if caps := rule.GetCapabilities(); len(caps) != 1 || caps[0] != vaultv1.Capability_CAPABILITY_READ {
					t.Errorf("%s: %q grants %v, want read only", component, path, caps)
				}
			default:
				t.Errorf("%s: rule %q is outside the org's subtree", component, path)
			}
		}
		if !granted {
			t.Errorf("%s: policy does not grant read on %q", component, want)
		}
	}
}
