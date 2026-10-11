# Ownership Model

**Every workflow executes in exactly one security context: its worker's owner. Vault paths, Kubernetes namespaces, and Vault policies all derive from that single identity, so a workflow that writes a path its policy doesn't grant becomes unrepresentable — not merely unauthorized.**

This document defines four principles (P1-P4) that govern how ownership flows through workflows, activities, and Vault policies. They apply to all fabric, shard, cell, cluster, and secrets workflows.

## Vocabulary

- **Owner**: the identity that owns a set of resources — either the *operator* (shared infrastructure: TLS CAs, certificate store, fabric worker) or an *org* (per-shard infrastructure: org CAs, peers, orderers).
- **Owner config**: `OperatorConfig` or `OrgConfig` (`temporal/lib/workers/config/owner.go`), implementing `OwnerConfig`. All resource names, Kubernetes namespaces, and Vault paths derive from it.
- **Acting AS an owner**: fabric and shard worker workflows run with their worker's identity — the operator for the fabric worker (`OPERATOR_ID`), an org for shard workers (`ORG_ID`).
- **Acting ON BEHALF OF an owner**: infra worker workflows (cell, cluster.fabric, secrets, temporal packages) provision resources *for* an owner they receive as an explicit, typed input.

## Case Study: The Certificate-Store 403

The bug that motivates this model:

The shared fabric worker runs with `OPERATOR_ID=metacensus` (no `ORG_ID`). Its Vault role `fabric-metacensus-worker` grants `secret/data/fabric/metacensus/*` plus create/update on `secret/data/fabric/shard/*/certs/*` (the Nexus TLS-issuance carve-out).

During operator bootstrap, `cell.v1.InitFabric` passed the operator ID as a plain `Org` string into `BootstrapFabric` over Nexus. `fabric.v1.DeployCertificateStore` then unconditionally called `config.NewOrgConfig(org)`, deriving the **org-shaped** Vault path `fabric/shard/metacensus/keys/certstore/admin`. The worker's policy — built for the **operator-shaped** path `fabric/metacensus/*` — doesn't cover it. Vault returned 403 in `secrets.v1.CreatePassword`.

The same string-typed ownership produced sibling latent bugs: `CreateCertificateDatabase` polled certificate-store readiness in the wrong Kubernetes namespace, and the certificate-store branch of `cluster.fabric.v1.CreateDeployment` always used org config regardless of the actual owner.

The root cause is not a wrong string value. It is that ownership was representable as a bare string at all, and that paths and policies were derived from it in different places. Each principle below closes one way this class of bug can be written.

## P1 — Worker Identity Is the Security Context

**Fabric and shard worker workflows act AS their worker's owner. Provisioning workflows act ON BEHALF OF an explicit, typed Owner. Neither ever derives ownership from a bare string in workflow input.**

The worker's Vault token comes from its Kubernetes service account, bound to a policy generated for one owner. Any config the workflow derives must come from that same owner, or the token and the paths disagree — that disagreement *is* the 403.

- On fabric/shard workers: the owner config (`OperatorConfig` from `OPERATOR_ID`, or `OrgConfig` from `ORG_ID`) is injected at worker startup and held by `Workflows`. Workflow inputs do not carry an owner at all — the worker already knows who it is.
- On the infra worker: provisioning workflows (cell, cluster.fabric, secrets, temporal) receive an explicit, typed `Owner` (proto `oneof operator_id / org_id`) and resolve it to an owner config once, at the top of the workflow.

**Don't** — reconstruct ownership from an input string (the bug):

```go
// fabric worker workflow — assumes every caller is an org
org := w.input.Req.GetOrg()
orgConfig := config.NewOrgConfig(org) // operator ID in, org-shaped paths out
certStore := fabricconfig.NewCertificateStoreConfig(orgConfig)
```

**Do** — use the injected worker identity:

```go
// fabric worker workflow — owner injected via fx at startup
certStore := fabricconfig.NewCertificateStoreConfig(w.owner)
```

**Don't** — model ownership as `org string` + `is_operator bool` in proto: the two fields can contradict each other, and every consumer re-implements the switch.

**Do** — use the typed `Owner` message (`oneof operator_id / org_id`): the variant carries the shape, and resolution happens in exactly one place.

## P2 — Local Call = My Resources; Nexus = Someone Else's

**A direct child workflow stays inside the caller's security context. Crossing an ownership boundary happens only via Nexus.**

Nexus is the authorization and audit boundary between Temporal namespaces. A child workflow executes on the same worker with the same Vault token, so a direct child that touches another owner's Vault paths or Kubernetes namespace silently smuggles a cross-owner operation past that boundary — it either fails with a 403 or, worse, succeeds because a policy was widened to accommodate it.

- A workflow never touches another owner's Vault paths or Kubernetes namespace through a direct child workflow.
- Under the CSR-based issuance flow there are no cross-owner Vault carve-outs at all: the key owner generates its private key and CSR locally and stores the signed certificate back into its own subtree; the operator's `certauth.SignCertificate` Nexus handler only signs the public CSR and returns public material, so it never reads or writes a shard-owned `keys/*` or `certs/*` path.

Concretely, in CA deployment: a signing parent is always same-owner, so signing certificates are produced inline with the local `SignCsr` activity. A TLS issuer owned by the operator, called from a shard worker, is someone else's CA — only the public CSR crosses via the `SignCertificate` Nexus operation, and the signed cert + chain come back to be stored locally.

Corollary for reviewers: if a change makes a local child workflow need another owner's paths, the fix is a Nexus operation (or questioning the design), never a policy widening.

Nexus routing is not the authorization gate. Caller allowlist and handler checks: [bootstrap-workflows — Nexus Authorization](../fabric/bootstrap-workflows.md#nexus-authorization).

## P3 — Reuse via Data and Plain Functions, Never Shared Workflow Types

**Variant behavior branches on structure, not labels. Workflows are never shared across security contexts to save code.**

Label-based branching (`if certType == "tls"`, `if caID == "root"`) encodes ownership assumptions as string comparisons scattered across call sites — the same failure mode as P1, one level down. Structural branching (`issuer == nil?`, `is this issuer owned by me?`) makes each variant's behavior follow from its data.

Case study: the four CA deployment workflows (TLS root, TLS intermediate, org root, org intermediate) share an identical 8-step spine — create Vault access, create database, obtain signing identity, obtain TLS identity, bootstrap password, deploy, bootstrap admin, permanent admin — and differ only in how the signing and TLS identities are obtained. That is expressible as one `DeployCertificateAuthority(CertificateAuthoritySpec)` workflow plus two helper functions:

```go
// ensureSigningIdentity: issuer == nil → key pair + self-signed cert;
// otherwise → IssueCertificate (local SignCsr; signing parents are always same-owner).

// ensureTLSIdentity: issuer == nil → self-sign;
// otherwise → IssueCertificate, which branches on ownership:
//   issuer owned by this worker → local SignCsr;
//   otherwise → Nexus SignCertificate (only the public CSR crosses).
```

**Don't** — four 200-line workflow bodies whose differences hide in label checks, or an owner-guessing switch inside a shared workflow:

```go
// certauth — guesses the owner from a cert-type label
if certType == "tls" {
    ownerConfig = config.GetOperator()
} else {
    ownerConfig = config.NewOrgConfig(req.GetOrg())
}
```

**Do** — one workflow body; the spec data determines the variant; helpers branch on structure:

```go
if spec.SigningIssuer == nil {
    // self-signed root
} else {
    // intermediate: local child — signing parent is mine (validated)
}
```

The workflow validates that `Self` and any `SigningIssuer` are owned by the executing worker (P1/P2 enforcement), and workflow IDs distinguish variants (e.g., `{owner}/deploy-ca/tls-root`).

Sharing a *workflow type* across security contexts is different from sharing *functions*: a plain function is called with the caller's already-resolved config; a shared workflow with an owner switch re-decides identity per invocation, and gets it wrong exactly when the caller isn't who it assumes.

## P4 — Vault Policies Are Generated from the Same Config That Generates Paths

**A worker or component policy is composed from the component configs' path methods. Policies are never hand-assembled in a distant package.**

If the code that writes a secret computes its path from `CertificateStoreConfig.VaultKeysBasePath()` but the policy granting access is assembled by string concatenation in another package, the two drift independently — and drift surfaces as a runtime 403 during bootstrap, the most expensive possible place to find it. The old `buildOperatorVaultPolicy` lived in `temporal/lib/workflows/cell/v1/workflows.go` while the paths it needed to cover lived in `temporal/lib/workers/fabric/config`; that distance is exactly what let the policy and the paths disagree.

- Component configs (`temporal/lib/workers/fabric/config`) own both the path methods (`VaultKeysBasePath()`, `VaultDataPath()`, …) and the `VaultPolicy()` documents built from them.
- Worker policies are produced by a `WorkerVaultPolicy(owner)` function in the same package, composed from those same path methods.
- Every worker gets access to only its own owner's subtree — there are no cross-owner rules. The CSR-based issuance flow keeps each private key and signed cert inside its owner's subtree (P2), so no cross-shard `certs/*` carve-out is needed.

**Don't** — assemble a policy by hand, far from the path definitions:

```go
// cell/v1/workflows.go — nowhere near the configs whose paths it must cover
dataPathPattern := path.Join(vaultSecretDataPrefix, operatorCfg.VaultBasePath(), "*")
allOrgCertsPattern := path.Join(vaultSecretDataPrefix, fabricSecretBase, "shard", "*", "certs/*")
```

**Do** — compose the policy from the component configs' own methods, in the config package:

```go
// workers/fabric/config — same package, same path methods the components use
func WorkerVaultPolicy(owner OwnerConfig) *vaultv1.PolicyDocument {
    certStore := NewCertificateStoreConfig(owner)
    // rules built from certStore.VaultDataPath(), ca.VaultKeysDataPath(), ...
}
```

Test for compliance: change a single path method and confirm both the secret write and the policy rule move together. If a path can change without the policy changing, P4 is violated.

## Quick Reference

| Situation | Rule |
|---|---|
| Fabric/shard workflow needs an owner config | Use the injected worker identity; never `NewOrgConfig(input.Org)` (P1) |
| Provisioning workflow needs an owner | Typed `Owner` oneof in proto; resolve once at the top (P1) |
| Workflow needs another owner's CA/paths/namespace | Nexus operation, never a direct child (P2) |
| Two workflows look almost identical | Share spec data + plain functions; branch on structure, not labels (P3) |
| Writing or reviewing a Vault policy | It must be composed from config path methods in the same package (P4) |

---

**Philosophy:** Ownership is an identity, not a string. Derive everything — paths, namespaces, policies — from one typed owner in one place, and the "worker writes a path its policy doesn't grant" bug class disappears at the type level.
