# Fabric Bootstrap Workflows

Orchestrating Hyperledger Fabric deployment through Temporal workflows across three isolated worker types.

**Status Legend:** ✅ Implemented | 🟡 Partial | ❌ Not Started

---

## Why Three Workers?

Fabric deployment requires balancing **security isolation** with **shared infrastructure**. Three workers solve this:

| Problem | Solution |
|---------|----------|
| TLS certificates must be issued from a trusted authority | **Fabric Worker** owns TLS CAs, issues certs via Nexus |
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
- **Vault Access:** `secret/fabric/tls/*` (TLS CA secrets only)
- **Nexus Calls:** `secrets.v1.CreateAccess` → Infra
- **Nexus Services:** Exposes `fabric.certauth.v1.CreateCertificate`, `fabric.v1.BootstrapFabric`
- **Responsibilities:** TLS CA chain (Root + Intermediate), Certificate Store, TLS certificate issuance
- **Identity:** Acts AS the operator (`OPERATOR_ID` injected at startup)

### Shard Workers (per-org deployments)
- **Temporal Namespace:** `fabric-shard-{org}` (one per org)
- **K8s Namespace:** `fabric-shard-{org}` (one per org)
- **Vault Access:** `secret/fabric/shard/{org}/*`
- **Nexus Calls:** `secrets.v1.CreateAccess` → Infra, `fabric.certauth.v1.CreateCertificate` → Fabric
- **Nexus Services:** Exposes `fabric.v1.BootstrapShard`
- **Responsibilities:** Org CA chain, Peers, Orderers, CouchDB, identity management
- **Identity:** Acts AS its org (`ORG_ID` injected at startup)

**Isolation Strategy:**
- Temporal namespace boundaries provide workflow history isolation
- Single task queue name (`infra`) per namespace (no routing complexity)
- Cross-namespace calls use Temporal Nexus
- Each worker type has scoped Vault access

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

## Nexus Authorization (Future Work)

**Status: ❌ Not implemented.** Cross-namespace calls currently have no application-layer authorization; the design below is planned but deferred.

Cross-namespace calls require authorization. Since Temporal Nexus ACL management requires server plugins (not available in standard SDK), authorization will be handled at the application layer.

### `secrets.v1.CreateAccess` (Infra Worker)

Validates caller namespace maps to requested Vault path:
- `fabric` → `secret/fabric/tls/*` only
- `fabric-shard-{org}` → `secret/fabric/shard/{org}/*` only

### `fabric.certauth.v1.CreateCertificate` (Fabric Worker)

Validates:
- Caller namespace matches `fabric-shard-*` pattern
- Certificate CN/SANs match caller's org

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
├─ temporal.v1.CreateNexusEndpoint (CreateCertificate)
├─ temporal.v1.CreateNexusEndpoint (BootstrapFabric)
│
└─ Nexus → fabric.v1.BootstrapFabric (Fabric Worker, fabric namespace, acts AS operator)
           │   input: ordered repeated CertificateAuthoritySpec (tls/root, tls/int)
           │
           ├─ fabric.v1.DeployCertificateStore
           ├─ fabric.v1.DeployCertificateAuthority (spec: tls/root, self-signed)
           └─ fabric.v1.DeployCertificateAuthority (spec: tls/int, issued by tls/root)
```

**Result:** TLS CA infrastructure ready. Fabric worker accepting `CreateCertificate` requests.

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
           │  └─ Nexus → fabric.certauth.v1.CreateCertificate (Fabric Worker)
           ├─ fabric.v1.DeployCertificateAuthority (spec: org/int, TLS via Nexus)
           │  └─ Nexus → fabric.certauth.v1.CreateCertificate (Fabric Worker)
           │
           ├─ for each peer:
           │  ├─ fabric.v1.DeployCouchDb
           │  └─ fabric.v1.DeployPeer
           │     └─ fabric.certauth.v1.CreateIdentity (org + TLS certs)
           │
           └─ for each orderer:
              └─ fabric.v1.DeployOrderer
                 └─ fabric.certauth.v1.CreateIdentity (org + TLS certs)
```

**Result:** Organization deployed with isolated CAs, peers, and orderers.

---

## Security Boundaries

### Per-Worker Isolation

| Worker | Temporal NS | K8s NS | Vault Access | Nexus Access |
|--------|-------------|--------|--------------|--------------|
| Infra | `infra` | `workers` | All paths (god-mode) | Exposes `CreateAccess` |
| Fabric | `fabric` | `fabric` | `secret/fabric/tls/*` | Exposes `CreateCertificate` |
| Shard (org1) | `fabric-shard-org1` | `fabric-shard-org1` | `secret/fabric/shard/org1/*` | Calls Infra + Fabric |
| Shard (org2) | `fabric-shard-org2` | `fabric-shard-org2` | `secret/fabric/shard/org2/*` | Calls Infra + Fabric |

### Per-Identity Isolation (within shards)

| Pod | Service Account | Vault Access |
|-----|-----------------|--------------|
| peer0 (org1) | `fabric-shard-org1-peer0` | `secret/fabric/shard/org1/msp/peer-peer0/*` |
| peer1 (org1) | `fabric-shard-org1-peer1` | `secret/fabric/shard/org1/msp/peer-peer1/*` |
| orderer0 (org1) | `fabric-shard-org1-orderer0` | `secret/fabric/shard/org1/msp/orderer-orderer0/*` |

### Compromise Scenarios

| If compromised... | Cannot access... | Because... |
|-------------------|------------------|------------|
| peer0 pod | peer1 secrets | Vault policy isolation |
| shard-org1 worker | org2 K8s resources | K8s RBAC namespace binding |
| shard-org1 worker | org2 Vault secrets | Vault token scoped to org1 |
| shard-org1 worker | org2 workflow history | Temporal namespace isolation |
| any shard worker | arbitrary Vault policies | Nexus validation at Infra worker |
| any shard worker | TLS certs for other orgs | Nexus validation at Fabric worker |

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
├─ temporal.v1.CreateNexusEndpoint ✅ (CreateCertificate)
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
└─ Nexus → fabric.v1.BootstrapShard ✅
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
├─ signing identity:
│  ├─ signing_issuer == nil (root CA):
│  │  ├─ secrets.v1.CreateKeyPair ✅
│  │  └─ secrets.v1.CreateSelfSignedCertificate ✅
│  └─ signing_issuer set (intermediate CA, always local):
│     └─ fabric.certauth.v1.CreateCertificate ✅ (hf.IntermediateCA)
│
├─ TLS identity:
│  ├─ tls_issuer == nil (TLS root only):
│  │  ├─ secrets.v1.CreateKeyPair ✅
│  │  └─ secrets.v1.CreateSelfSignedCertificate ✅
│  ├─ tls_issuer owned locally (operator's TLS CAs):
│  │  └─ fabric.certauth.v1.CreateCertificate ✅ (local child)
│  └─ tls_issuer foreign (org CAs):
│     └─ Nexus → fabric.certauth.v1.CreateCertificate ✅ (Fabric Worker)
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
├─ for each peer:
│  ├─ fabric.v1.DeployCouchDb ✅
│  └─ fabric.v1.DeployPeer ✅
├─ for each orderer:
│  └─ fabric.v1.DeployOrderer ✅
└─ TODO: chaincode, channel join
```

**Component Deployment:**

```
fabric.v1.DeployPeer ✅
├─ Nexus → secrets.v1.CreateAccess ✅
├─ fabric.certauth.v1.CreateIdentity ✅ (org + TLS certs)
└─ cluster.fabric.v1.CreateDeployment ✅
```

```
fabric.v1.DeployOrderer ✅
├─ Nexus → secrets.v1.CreateAccess ✅
├─ fabric.certauth.v1.CreateIdentity ✅ (org + TLS certs)
└─ cluster.fabric.v1.CreateDeployment ✅
```

---

### Certificate Authority Workflows (fabric.certauth.v1 package)

Identity and certificate management workflows used by both Fabric and Shard workers.

```
fabric.certauth.v1.CreateCertificate ✅
├─ secrets.v1.CreateKeyPair ✅
├─ fabric.certauth.v1.RegisterIdentity ✅
└─ fabric.certauth.v1.EnrollIdentity ✅
```

```
fabric.certauth.v1.CreateIdentity ✅
├─ fabric.certauth.v1.CreateCertificate ✅ (org cert, local child)
└─ TLS identity:
   ├─ tls_issuer owned locally:
   │  └─ fabric.certauth.v1.CreateCertificate ✅ (local child)
   └─ tls_issuer foreign:
      └─ Nexus → fabric.certauth.v1.CreateCertificate ✅ (issuer's worker)
```

```
fabric.certauth.v1.CreateBootstrapAdmin ✅
├─ kube.v1.GetStatus ✅ (wait for CA readiness)
├─ secrets.v1.CreateKeyPair ✅ (signing)
├─ fabric.certauth.v1.BootstrapEnrollIdentity ✅ (signing)
└─ TLS identity:
   ├─ tls_issuer owned locally:
   │  ├─ secrets.v1.CreateKeyPair ✅
   │  └─ fabric.certauth.v1.BootstrapEnrollIdentity ✅
   └─ tls_issuer foreign:
      └─ Nexus → fabric.certauth.v1.CreateCertificate ✅ (issuer's worker)
```

```
fabric.certauth.v1.CreateAdmin ✅
├─ fabric.certauth.v1.CreateCertificate ✅ (signing, always local)
└─ TLS identity:
   ├─ tls_issuer owned locally:
   │  └─ fabric.certauth.v1.CreateCertificate ✅ (local child)
   └─ tls_issuer foreign:
      └─ Nexus → fabric.certauth.v1.CreateCertificate ✅ (issuer's worker)
```

**Activities (fabric.certauth.v1):**
- `RegisterIdentity` ✅ - Register identity with CA
- `EnrollIdentity` ✅ - Enroll and store certificate
- `BootstrapEnrollIdentity` ✅ - Bootstrap enrollment with username/password
- `RevokeIdentity` ✅ - Revoke identity from CA


---

### Shared Infrastructure Workflows

**secrets.v1 (Infra Worker only):**

```
secrets.v1.CreateAccess ✅
├─ vault.v1.CreatePolicy ✅
└─ vault.v1.CreateAuthRole ✅
```

Activities: `CreatePassword` ✅, `CreateKeyPair` ✅, `CreateSelfSignedCertificate` ✅, `StoreCertificate` ✅

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