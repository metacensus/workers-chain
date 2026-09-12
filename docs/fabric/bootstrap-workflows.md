# Fabric Bootstrap Workflows

Orchestrating Hyperledger Fabric deployment through Temporal workflows across three isolated worker types.

**Status Legend:** ✅ Implemented | 🟡 Partial | ❌ Not Started

---

## Why Three Workers?

Fabric deployment requires balancing **security isolation** with **shared infrastructure**. Three workers solve this:

| Problem | Solution |
|---------|----------|
| TLS certificates must be issued from a trusted authority | **Fabric Worker** owns the operator's TLS CAs, signs foreign CSRs via Nexus |
| Each organization needs isolated secrets and workflow history | **Shard Workers** run in per-org Temporal namespaces |
| Someone must create Vault policies and Temporal namespaces | **Infra Worker** has elevated privileges, validates requests |

**Key insight:** The Fabric Worker exists because TLS CAs serve all organizations but shouldn't be owned by any single org. Without it, either every shard would need TLS CA access (security risk) or the Infra Worker would need to understand Fabric CA operations (coupling).

---

## Worker Architecture

### Infra Worker (platform orchestration)
- **Temporal Namespace:** `infra`
- **K8s Namespace:** `workers`
- **Vault Access:** God-mode with policy validation
- **Nexus Services:** Exposes `secrets.v1.CreateAccess`
- **Responsibilities:** Temporal namespace/Nexus management, Vault policy creation, worker deployment

### Fabric Worker (shared TLS infrastructure)
- **Temporal Namespace:** `fabric`
- **K8s Namespace:** `fabric`
- **Vault Access:** operator subtree only (`secret/{data,metadata}/fabric/metacensus/*`)
- **Nexus Calls:** `secrets.v1.CreateAccess` → Infra
- **Nexus Services:** Exposes `fabric.certauth.v1.SignCertificate`, `fabric.v1.BootstrapFabric`
- **Responsibilities:** TLS CA chain (Root + Intermediate), Certificate Store, signing foreign CSRs
- **Identity:** Acts AS the operator (`OPERATOR_ID` injected at startup)

### Shard Workers (per-org deployments)
- **Temporal Namespace:** `fabric-shard-{org}` (one per org)
- **K8s Namespace:** `fabric-shard-{org}` (one per org)
- **Vault Access:** own subtree only (`secret/{data,metadata}/fabric/shard/{org}/*`)
- **Nexus Calls:** `secrets.v1.CreateAccess` → Infra, `fabric.certauth.v1.SignCertificate` → Fabric
- **Nexus Services:** Exposes `fabric.v1.BootstrapShard`
- **Responsibilities:** Org CA chain, Peers, Orderers, CouchDB, identity management
- **Identity:** Acts AS its org (`ORG_ID` injected at startup)

**Isolation Strategy:**
- Temporal namespace boundaries provide workflow history isolation
- Single task queue name (`infra`) per namespace (no routing complexity)
- Cross-namespace calls use Temporal Nexus
- Each worker's Vault policy is scoped to its own owner subtree (`WorkerVaultPolicy`, guarded by `policy_test.go`)

---

## Ownership Model

Every fabric/certauth workflow acts **as its worker's owner**, injected at
startup (`OPERATOR_ID` on the fabric worker, `ORG_ID` on shard workers). No
workflow input carries an `org` string; Vault paths, Kubernetes namespaces, and
CA references all derive from the injected owner. The ownership rule: the
**operator owns `tls` CAs**, an **org owns `org` CAs**.

Workflows that act **on behalf of** others (cell.v1 on the infra worker,
cluster.fabric.v1 everywhere) carry an explicit typed `Owner` message
(oneof `operator_id` / `org_id`) in their inputs.

CA deployment is spec-driven: a single `fabric.v1.DeployCertificateAuthority`
workflow takes a `CertificateAuthoritySpec` (`self`, `signing_issuer`,
`tls_issuer`, `validity_duration`) and covers all four variants (tls/root,
tls/int, org/root, org/int). It validates ownership up front:

- `spec.self` and `spec.signing_issuer` MUST be owned by the executing worker
  (signing issuance never crosses the ownership boundary)
- `spec.tls_issuer` may be foreign — TLS certs for org CAs are issued by the
  operator's TLS chain via Nexus

`BootstrapFabric` and `BootstrapShard` carry ordered
`repeated CertificateAuthoritySpec` and deploy each in order.

---

## Certificate Issuance: the Unified CSR Pipeline

All certificate issuance flows through one pipeline, so a private key never
leaves the owner that generated it. `IssueCertificate` (and its register-less
sibling `BootstrapIssueCertificate`) run these steps:

1. **`secrets.v1.EnsureKeyPair`** — get-or-create the private key at
   `storage.key_ref`, in the requesting owner's own Vault subtree.
2. **`fabric.certauth.v1.GenerateCsr`** — build a CSR from that key + subject.
   Only the **public** CSR leaves this step; the private key stays put.
3. **Sign the CSR**, chosen structurally by
   `OwnsCertificateAuthority(owner, issuerRef)` (OWNERSHIP.md P3) — not by any
   input flag:
   - **Local issuer** → `fabric.certauth.v1.SignCsr` activity, authenticating as
     the caller-selected registrar (`adminIdentity`).
   - **Foreign issuer** → `fabric.certauth.v1.SignCertificate` workflow via Nexus
     on the *issuer's own* worker. Only the public CSR crosses the boundary; the
     signed cert + chain come back.
   - **Bootstrap enrollment** (a CA's first admin) uses
     `fabric.certauth.v1.BootstrapEnrollCsr` instead, authenticating with the
     CA's preset bootstrap username/password.
4. **`secrets.v1.StoreCertificate`** — store the returned cert + chain at
   `storage.cert_ref`.

**Security boundary (verified):** no private key, key reference, or registrar
identity ever crosses Nexus. `adminIdentity` is honored only on the local
branch; a foreign issuer always signs as its **own** permanent admin, so no
caller-chosen registrar crosses the ownership boundary. The only cross-owner
Nexus operation is `SignCertificate`, and it carries a **public CSR only**.

> **Note (stale wire name):** the Nexus *endpoint* is still registered under the
> legacy string `fabric-CreateCertificate` (`cell/v1/workflows.go`
> `FabricCreateCertificateNexusEndpoint`). The served operation is
> `SignCertificate`; the endpoint rename is deferred to a coordinated redeploy.

---

## Nexus Authorization

**Status: 🟢 Happy path ready** ([PR #51](https://github.com/metacensus/infra/pull/51), issue [#43](https://github.com/metacensus/infra/issues/43)). Server gate, mTLS, handler payload checks, and live happy-path/negative probes are in. Cross-owner CN/SAN namespacing remains deferred.

Temporal Cloud has per-endpoint caller-namespace allowlists. Self-hosted does not — the supported path is a compile-time [Authorizer](https://docs.temporal.io/self-hosted-guide/security#authorizer-plugin) injected into a custom Temporal server binary ([docs](https://docs.temporal.io/nexus/security)). Runtime-loadable auth plugins are not supported.

### Architecture (locked)

```text
Caller worker (mTLS identity = Temporal NS)
        │
        ▼
┌───────────────────────────────────────────────────────────┐
│  Temporal server Authorizer (THE GATE)                    │
│  default-deny: caller NS → Nexus endpoint                 │
│  binary: temporal-server/ (WithAuthorizer/WithClaimMapper)│
└───────────────────────────────────────────────────────────┘
        │ allowed
        ▼
┌───────────────────────────────────────────────────────────┐
│  Handler payload policy (defense in depth)                │
│  CreateAccess: Vault path ⊆ bound k8s namespace subtree   │
│  SignCertificate: reject ca / hf.IntermediateCA /         │
│    hf.Registrar.* (leaf issuance only on this path)       │
└───────────────────────────────────────────────────────────┘
```

**Locked decisions**

| Decision | Choice |
|----------|--------|
| Gate | Custom Temporal server Authorizer (not app-only, not Cloud ACLs) |
| Worker identity | mTLS client certs; ClaimMapper maps cert CN → Temporal namespace |
| Allowlist source | Go code in `temporal-server/` (versioned with the image) |
| Nexus surface | Authorizer endpoint allowlist **and** narrow worker Nexus registration to intended ops |
| Cheap SignCertificate harden as standalone PR | **No** — lands with handler policy |

### Endpoint allowlist (default deny)

| Caller Temporal NS | Endpoint | Allowed operation(s) |
|--------------------|----------|----------------------|
| `infra` | `fabric-BootstrapFabric` | `BootstrapFabric` only |
| `infra` | `fabric-shard-{org}-BootstrapShard` | `BootstrapShard` only (matching `{org}`) |
| `fabric` | `secrets-CreateAccess` | `CreateAccess` iff policy ⊆ operator Vault subtree |
| `fabric-shard-{org}` | `secrets-CreateAccess` | `CreateAccess` iff policy ⊆ that org's Vault subtree |
| `fabric-shard-{org}` | `fabric-shard-{org}-CreateCertificate` | `SignCertificate` only; requester must match `{org}`; privileged profiles/attrs rejected |

All other caller × endpoint × operation pairs are denied. In-workflow calls are gated by the history-service endpoint registry; HTTP Nexus dispatch is gated by the frontend Authorizer (caller when present, plus endpoint→payload binding for certificate endpoints).

`temporal-system-admin` and `temporal-internode` are Terraform-issued CNs with a cluster system role (admintools/UI/CLI, and Temporal's own frontend calls). They are never given to workers.

The infra worker is not a system admin. It may `RegisterNamespace` / `DescribeNamespace` for Temporal namespaces `fabric` and `fabric-shard-*`, and `CreateNexusEndpoint` (that RPC has no `GetNamespace()`). Shard and fabric workers cannot.

Nexus handlers do not receive a trustworthy caller identity. Handler policy therefore binds to what the handler can verify: CreateAccess to the Kubernetes namespace the policy is attached to; SignCertificate to request shape (no CA/registrar attributes) plus CN/SAN scoped to the declared requester. Certificate issuance uses per-org endpoints (`fabric-shard-{org}-CreateCertificate`), not a shared `fabric-CreateCertificate`.

### Implementation checklist

Done:

1. **Design lock** — this section ✅
2. **Worker↔Temporal mTLS** — frontend `requireClientAuth`; TF-issued CA + leafs; workers dial with TLS; runtime leaf issuance for shards via `EnsureTemporalMTLSClientCert` ✅
3. **Custom server image** — Authorizer + ClaimMapper in `metacensus/temporal-server`; Terraform `server_image` defaults there; built locally by `make tf-apply` and published by the release workflow ✅
4. **Handler policy (partial)** — `CreateAccess` Vault-subtree scoping and `SignCertificate` privileged-request refusal ✅
5. **Happy-path verify** — live `BootstrapInfra` + `InitFabricShard` ✅
6. **In-workflow Nexus allowlist** — history-service `EndpointRegistry` decorator (`temporal-server/authz/nexus_registry.go` via `serverfx`) returns `PermissionDenied` for disallowed caller→endpoint pairs; Temporal turns that into `caller namespace %q unauthorized for %q` ✅
7. **Per-org CreateCertificate endpoints** — `fabric-shard-{org}-CreateCertificate`; frontend Authorizer binds endpoint→payload requester; shared `fabric-CreateCertificate` retired ✅
8. **Structured deny logs** — history registry logs `nexus endpoint denied for caller namespace` (caller + endpoint); frontend Authorizer logs `authorization denied` with a full reason (not `"..."`) ✅

Live negatives (re-run 2026-08-26 after deploying the registry gate):

| Probe | Result |
|-------|--------|
| Shard `acme` StartWorkflow on `infra` / `fabric` | Denied (`Request unauthorized`) ✅ |
| Shard `acme` `RegisterNamespace` | Denied ✅ |
| Fabric StartWorkflow on `infra` | Denied ✅ |
| `CreateAccess` for `fabric-shard-acme` with org-B / operator Vault paths | Denied by handler (`policy path outside … Vault subtree`) ✅ |
| `SignCertificate` with `Profile: "ca"` | Denied by handler (`privileged issuance…`) ✅ |
| Shard → Nexus `fabric-BootstrapFabric` (in-workflow) | Denied at schedule (`caller namespace "fabric-shard-acme" unauthorized for "fabric-BootstrapFabric"`) ✅ |
| Infra → Nexus `secrets-CreateAccess` (in-workflow) | Denied at schedule (`caller namespace "infra" unauthorized for "secrets-CreateAccess"`) ✅ |

The two seams: history binds **caller→endpoint** (and logs `nexus endpoint denied for caller namespace` with caller + endpoint tags); frontend Authorizer binds **endpoint→payload** and logs `authorization denied` with a full reason string. Handler policy remains defense in depth.

Remaining (issue [#43](https://github.com/metacensus/infra/issues/43)):

- [x] **Narrow Nexus registration** — fabric worker: `BootstrapFabric` + `SignCertificate` only; shard worker: `BootstrapShard` only; infra already `CreateAccess` only ✅
- [x] **Open PR** — [#51](https://github.com/metacensus/infra/pull/51) (`erikforsman/nexus-authorization`)
- [ ] **Land PR** — merge #51 to `main`
- [ ] **CN/SAN namespacing** — scope leaf subject (CN/SAN) to the declared requester on the CreateCertificate path

Code lives in `temporal-server/` (gate) and `temporal/lib/workflows/` (handler policy). Ownership boundary: [OWNERSHIP.md P2](../strategy/OWNERSHIP.md#p2--local-call--my-resources-nexus--someone-elses).

---

## Bootstrap Flows

Fabric bootstrap has two distinct flows with different lifecycles:

### One-Time: Fabric Initialization

Run **once per cluster** to establish shared TLS infrastructure. Must complete before any shards can be created.

```
cell.v1.InitFabric (Infra Worker, infra namespace, on behalf of the operator)
│
├─ temporal.v1.CreateNamespace (fabric)
├─ cluster.fabric.v1.CreateNamespace (Owner: operator)
├─ cluster.fabric.v1.CreateServiceAccount (Owner: operator)
├─ secrets.v1.CreateAccess (WorkerVaultPolicy(operator))
├─ cluster.fabric.v1.CreateDeployment (Owner: operator, fabric worker)
├─ temporal.v1.CreateNexusEndpoint (SignCertificate)
├─ temporal.v1.CreateNexusEndpoint (BootstrapFabric)
│
└─ Nexus → fabric.v1.BootstrapFabric (Fabric Worker, fabric namespace, acts AS operator)
           │   input: ordered repeated CertificateAuthoritySpec (tls/root, tls/int)
           │
           ├─ fabric.v1.DeployCertificateStore
           ├─ fabric.v1.DeployCertificateAuthority (spec: tls/root, self-signed)
           └─ fabric.v1.DeployCertificateAuthority (spec: tls/int, issued by tls/root)
```

**Result:** TLS CA infrastructure ready. Fabric worker accepting `SignCertificate` requests.

### On-Demand: Shard Creation

Run **per organization** when a new org joins the network. Can be invoked multiple times.

```
cell.v1.InitFabricShard (Infra Worker, infra namespace, on behalf of {org})
│
├─ temporal.v1.CreateNamespace (fabric-shard-{org})
├─ cluster.fabric.v1.CreateNamespace (Owner: org)
├─ cluster.fabric.v1.CreateServiceAccount (Owner: org)
├─ secrets.v1.CreateAccess (WorkerVaultPolicy(org))
├─ cluster.fabric.v1.CreateDeployment (Owner: org, shard worker)
├─ temporal.v1.CreateNexusEndpoint (BootstrapShard)
│
└─ Nexus → fabric.v1.BootstrapShard (Shard Worker, fabric-shard-{org} namespace, acts AS org)
           │   input: DeploymentSpec with ordered CertificateAuthoritySpec (org/root, org/int)
           │
           ├─ fabric.v1.DeployCertificateStore
           ├─ fabric.v1.DeployCertificateAuthority (spec: org/root, TLS via Nexus)
           │  └─ Nexus → fabric.certauth.v1.SignCertificate (Fabric Worker)
           ├─ fabric.v1.DeployCertificateAuthority (spec: org/int, TLS via Nexus)
           │  └─ Nexus → fabric.certauth.v1.SignCertificate (Fabric Worker)
           │
           └─ 🟡 peers/orderers: OUT OF SCOPE for now (see note below)
```

**Result:** Organization's CAs deployed. Peer/orderer/CouchDB startup is
intentionally deferred (see below).

> **Scope note:** `BootstrapShard` currently stops after deploying the org CAs —
> `deployWorkloads := false` in `bootstrapShardWorkflow.Execute` short-circuits
> before CouchDB/peer/orderer deployment. This is deliberate: peer startup (and
> the Vault-policy + network wiring it needs) is tracked as separate follow-up
> work, not a debug leftover. The peer/orderer branches below are implemented but
> gated off until that lands.

---

## Security Boundaries

### Per-Worker Isolation

| Worker | Temporal NS | K8s NS | Vault Access | Nexus Access |
|--------|-------------|--------|--------------|--------------|
| Infra | `infra` | `workers` | All paths (god-mode) | Exposes `CreateAccess` |
| Fabric | `fabric` | `fabric` | operator subtree only | Exposes `SignCertificate` |
| Shard (org1) | `fabric-shard-org1` | `fabric-shard-org1` | `secret/fabric/shard/org1/*` | Calls Infra + Fabric |
| Shard (org2) | `fabric-shard-org2` | `fabric-shard-org2` | `secret/fabric/shard/org2/*` | Calls Infra + Fabric |

### Compromise Scenarios

| If compromised... | Cannot access... | Because... |
|-------------------|------------------|------------|
| shard-org1 worker | org2 K8s resources | K8s RBAC namespace binding |
| shard-org1 worker | org2 Vault secrets | Vault token scoped to org1 subtree |
| shard-org1 worker | org2 workflow history | Temporal namespace isolation |
| any shard worker | arbitrary Vault policies | Nexus Authorizer + CreateAccess handler checks (see [Nexus Authorization](#nexus-authorization)) |
| any shard worker | another org's private keys | keys never cross Nexus — only public CSRs do |

---

## Implementation Status

---

### Infra Worker Workflows (cell package)

```
cell.v1.BootstrapInfra ✅
├─ temporal.v1.CreateNexusEndpoint ✅ (secrets-CreateAccess)
└─ cell.v1.InitFabric ✅
```

```
cell.v1.InitFabric ✅
├─ temporal.v1.CreateNamespace ✅ (fabric)
├─ cluster.fabric.v1.CreateNamespace ✅
├─ cluster.fabric.v1.CreateServiceAccount ✅
├─ secrets.v1.CreateAccess ✅
├─ cluster.fabric.v1.CreateDeployment ✅ (fabric worker)
├─ temporal.v1.CreateNexusEndpoint ✅ (SignCertificate — endpoint still named fabric-CreateCertificate)
├─ temporal.v1.CreateNexusEndpoint ✅ (BootstrapFabric)
└─ Nexus → fabric.v1.BootstrapFabric ✅
```

```
cell.v1.InitFabricShard ✅
├─ temporal.v1.CreateNamespace ✅
├─ cluster.fabric.v1.CreateNamespace ✅
├─ cluster.fabric.v1.CreateServiceAccount ✅
├─ secrets.v1.CreateAccess ✅
├─ cluster.fabric.v1.CreateDeployment ✅
├─ temporal.v1.CreateNexusEndpoint ✅
└─ Nexus → fabric.v1.BootstrapShard 🟡 (CA deployment only; peers gated off)
```

---

### Fabric Worker Workflows (fabric.v1 package)

The fabric worker acts AS the operator (injected `OPERATOR_ID`); shard workers
act AS their org (injected `ORG_ID`). The same workflow implementations serve
both — behavior derives from the injected owner and the CA specs.

```
fabric.v1.BootstrapFabric ✅ (on the fabric worker; specs: tls/root, tls/int)
├─ fabric.v1.DeployCertificateStore ✅
└─ for each CertificateAuthoritySpec (in order):
   └─ fabric.v1.DeployCertificateAuthority ✅
```

**CA Deployment (single spec-driven workflow, all four variants):**

```
fabric.v1.DeployCertificateAuthority ✅
│  validates: spec.self and spec.signing_issuer owned by this worker (fail fast)
│
├─ Nexus → secrets.v1.CreateAccess ✅
├─ fabric.v1.CreateCertificateDatabase ✅
│
├─ signing identity (CSR pipeline):
│  ├─ signing_issuer == nil (root CA):
│  │  ├─ secrets.v1.EnsureKeyPair ✅
│  │  └─ secrets.v1.CreateSelfSignedCertificate ✅
│  └─ signing_issuer set (intermediate CA, always local):
│     └─ IssueCertificate ✅ (local SignCsr branch, hf.IntermediateCA)
│
├─ TLS identity (CSR pipeline):
│  ├─ tls_issuer == nil (TLS root only):
│  │  ├─ secrets.v1.EnsureKeyPair ✅
│  │  └─ secrets.v1.CreateSelfSignedCertificate ✅
│  ├─ tls_issuer owned locally (operator's TLS CAs):
│  │  └─ IssueCertificate ✅ (local SignCsr branch)
│  └─ tls_issuer foreign (org CAs):
│     └─ IssueCertificate ✅ → Nexus fabric.certauth.v1.SignCertificate (Fabric Worker)
│
├─ secrets.v1.CreatePassword ✅ (bootstrap admin)
├─ cluster.fabric.v1.CreateDeployment ✅ (Owner from injected identity)
├─ fabric.certauth.v1.CreateBootstrapAdmin ✅
└─ fabric.certauth.v1.CreateAdmin ✅
```

Child workflow IDs are distinguished per spec: `{owner}/deploy-ca/{certType}-{caID}`.

---

### Shard Worker Workflows (fabric.v1 package)

```
fabric.v1.BootstrapShard 🟡 (on a shard worker; specs: org/root, org/int)
├─ fabric.v1.DeployCertificateStore ✅
├─ for each CertificateAuthoritySpec (in order):
│  └─ fabric.v1.DeployCertificateAuthority ✅ (TLS via Nexus → Fabric Worker)
│
└─ ⛔ deployWorkloads == false — stops here (peer startup is separate follow-up work)
   ├─ for each peer:      fabric.v1.DeployCouchDb + DeployPeer   🟡 (implemented, gated off)
   ├─ for each orderer:   fabric.v1.DeployOrderer                🟡 (implemented, gated off)
   └─ TODO: chaincode, channel join
```

**Component Deployment (implemented, gated off until peer startup lands):**

```
fabric.v1.DeployPeer 🟡
├─ Nexus → secrets.v1.CreateAccess ✅
├─ fabric.certauth.v1.CreateIdentity ✅ (org + TLS certs)
└─ cluster.fabric.v1.CreateDeployment ✅
```

```
fabric.v1.DeployOrderer 🟡
├─ Nexus → secrets.v1.CreateAccess ✅
├─ fabric.certauth.v1.CreateIdentity ✅ (org + TLS certs)
└─ cluster.fabric.v1.CreateDeployment ✅
```

---

### Certificate Authority Workflows (fabric.certauth.v1 package)

Identity and certificate management workflows used by both Fabric and Shard
workers. All build on the unified CSR pipeline (`IssueCertificate` /
`BootstrapIssueCertificate`); there is no longer a `CreateCertificate` workflow
or `RegisterIdentity`/`EnrollIdentity` split.

```
fabric.certauth.v1.SignCertificate ✅  (the Nexus-served issuer-side op)
│  input: public CSR + subject + attributes (NO private key, NO registrar identity)
├─ fabric.certauth.v1.SignCsr ✅ (signs as this CA's own permanent admin)
└─ returns cert + chain
```

```
fabric.certauth.v1.CreateIdentity ✅ (org + TLS certs for peers/orderers)
├─ IssueCertificate ✅ (org cert — local SignCsr)
└─ TLS identity:
   ├─ tls_issuer owned locally → IssueCertificate ✅ (local SignCsr)
   └─ tls_issuer foreign        → IssueCertificate ✅ (Nexus SignCertificate)
```

```
fabric.certauth.v1.CreateBootstrapAdmin ✅
├─ kube.v1.GetStatus ✅ (wait for CA readiness)
├─ signing identity: BootstrapIssueCertificate ✅
│  (EnsureKeyPair → GenerateCsr → BootstrapEnrollCsr → StoreCertificate)
└─ TLS identity:
   ├─ tls_issuer owned locally → BootstrapIssueCertificate ✅ (register-less)
   └─ tls_issuer foreign        → IssueCertificate ✅ (Nexus SignCertificate)
```

```
fabric.certauth.v1.CreateAdmin ✅
│  registers the org's permanent admin with the full registrar grant
│  (hf.Registrar.Roles=*, …), signed by the CA's bootstrap admin
├─ signing identity: IssueCertificate ✅ (local SignCsr, adminIdentity = bootstrap admin)
└─ TLS identity:
   ├─ tls_issuer owned locally → IssueCertificate ✅ (local SignCsr)
   └─ tls_issuer foreign        → IssueCertificate ✅ (Nexus SignCertificate)
```

**Activities (fabric.certauth.v1):**
- `GenerateCsr` ✅ — build a CSR from an existing key + subject
- `SignCsr` ✅ — sign a CSR at a local CA as a chosen registrar
- `BootstrapEnrollCsr` ✅ — register-less enrollment with the CA's bootstrap username/password
- `RevokeIdentity` ✅ — defined + registered, currently **no caller** (intended bootstrap-admin revocation hook)


---

### Shared Infrastructure Workflows

**secrets.v1 (Infra Worker only):**

```
secrets.v1.CreateAccess ✅
├─ vault.v1.CreatePolicy ✅
└─ vault.v1.CreateAuthRole ✅
```

Activities: `CreatePassword` ✅, `EnsureKeyPair` ✅ (get-or-create), `CreateSelfSignedCertificate` ✅, `StoreCertificate` ✅

**cluster.fabric.v1 (All Workers):**

```
cluster.fabric.v1.CreateNamespace ✅
├─ helm.v1.RenderChartWithTemplate ✅
└─ kube.v1.Apply ✅
```

```
cluster.fabric.v1.CreateDeployment ✅
├─ helm.v1.RenderChartWithTemplate ✅
└─ kube.v1.Apply ✅
```

**temporal.v1 (Infra Worker only):**

```
temporal.v1.CreateNamespace ✅
├─ temporal.v1.RegisterNamespace ✅
└─ temporal.v1.DescribeNamespace ✅
```
