package config

import (
	"fmt"

	vaultv1 "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
)

// WorkerVaultPolicy creates the Vault policy document for an owner's fabric
// worker (the shared fabric worker when owner is the operator, a shard worker
// otherwise). It is derived from the same VaultBasePath() the component
// configs (CertificateAuthorityConfig, CertificateStoreConfig, ...) build
// their paths from, so worker permissions and component paths cannot drift.
//
// Every worker gets full access to its owner's own Vault subtree:
//
//	secret/data/{owner.VaultBasePath()}/*
//
// The operator's worker additionally gets cross-shard create/update for:
//   - org certs/*: issued TLS certificates written back to requesting shards
//   - org keys/*: private key material generated during Fabric CA enrollment
//
// The keys/* carve-out intentionally omits read/list/delete. It lets the
// operator complete foreign TLS issuance without allowing it to fetch shard
// private keys back from Vault.
func WorkerVaultPolicy(owner OwnerConfig) *vaultv1.PolicyDocument {
	rules := []*vaultv1.PolicyRule{
		// Owner's subtree: full access
		{
			Path: fmt.Sprintf("%s/%s/*", vaultSecretDataPrefix, owner.VaultBasePath()),
			Capabilities: []vaultv1.Capability{
				vaultv1.Capability_CAPABILITY_CREATE,
				vaultv1.Capability_CAPABILITY_READ,
				vaultv1.Capability_CAPABILITY_UPDATE,
				vaultv1.Capability_CAPABILITY_LIST,
			},
		},
		{
			Path: fmt.Sprintf("%s/%s/*", vaultSecretMetadataPrefix, owner.VaultBasePath()),
			Capabilities: []vaultv1.Capability{
				vaultv1.Capability_CAPABILITY_LIST,
				vaultv1.Capability_CAPABILITY_READ,
			},
		},
	}

	if owner.IsOperator() {
		// Cross-shard carve-out: write issued TLS material into any org's
		// certs/* and keys/* subtrees. Vault uses "+" as a single path-segment
		// wildcard, so this matches "fabric/shard/{org}/..." without matching
		// unrelated layouts.
		allOrgsCerts := NewOrgConfig("+").VaultBasePath() + "/certs/*"
		allOrgsKeys := NewOrgConfig("+").VaultBasePath() + "/keys/*"
		rules = append(rules,
			&vaultv1.PolicyRule{
				Path: fmt.Sprintf("%s/%s", vaultSecretDataPrefix, allOrgsCerts),
				Capabilities: []vaultv1.Capability{
					vaultv1.Capability_CAPABILITY_CREATE,
					vaultv1.Capability_CAPABILITY_UPDATE,
				},
			},
			&vaultv1.PolicyRule{
				Path: fmt.Sprintf("%s/%s", vaultSecretDataPrefix, allOrgsKeys),
				Capabilities: []vaultv1.Capability{
					vaultv1.Capability_CAPABILITY_CREATE,
					vaultv1.Capability_CAPABILITY_UPDATE,
				},
			},
			&vaultv1.PolicyRule{
				Path: fmt.Sprintf("%s/%s", vaultSecretMetadataPrefix, allOrgsCerts),
				Capabilities: []vaultv1.Capability{
					vaultv1.Capability_CAPABILITY_READ,
				},
			},
		)
	}

	return &vaultv1.PolicyDocument{Rules: rules}
}
