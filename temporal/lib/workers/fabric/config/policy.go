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
// Every worker - operator or shard - gets full access to ONLY its owner's own
// Vault subtree, and nothing else:
//
//	secret/data/{owner.VaultBasePath()}/*
//	secret/metadata/{owner.VaultBasePath()}/*
//
// There are no cross-owner rules. Under the CSR-based cross-owner issuance flow,
// the key owner generates its private key and CSR locally and writes the signed
// certificate back into its own subtree; the operator's TLS CA only signs the
// public CSR and returns public material. The operator therefore never reads or
// writes any shard-owned keys/* or certs/* path, so Vault access aligns exactly
// with ownership.
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

	return &vaultv1.PolicyDocument{Rules: rules}
}
