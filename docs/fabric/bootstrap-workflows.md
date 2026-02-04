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
- **Nexus Services:** Exposes `fabric.certauth.v1.CreateCertifitcate`
- **Responsibilities:** TLS CA chain (Root + Intermediate), Certificate Store, TLS certificate issuance

### Shard Workers (per-org deployments)
- **Temporal Namespace:** `fabric-shard-{org}` (one per org)
- **K8s Namespace:** `fabric-shard-{org}` (one per org)
- **Vault Access:** `secret/fabric/shard/{org}/*`
- **Nexus Calls:** `secrets.v1.CreateAccess` → Infra, `fabric.certauth.v1.CreateCertifitcate` → Fabric
- **Responsibilities:** Org CA chain, Peers, Orderers, CouchDB, identity management

**Isolation Strategy:**
- Temporal namespace boundaries provide workflow history isolation
- Single task queue name (`infra`) per namespace (no routing complexity)
- Cross-namespace calls use Temporal Nexus
- Each worker type has scoped Vault access

---

## Nexus Authorization

Cross-namespace calls require authorization. Since Temporal Nexus ACL management requires server plugins (not available in standard SDK), authorization is handled at the application layer.

### `secrets.v1.CreateAccess` (Infra Worker)

Validates caller namespace maps to requested Vault path:
- `fabric` → `secret/fabric/tls/*` only
- `fabric-shard-{org}` → `secret/fabric/shard/{org}/*` only

### `fabric.certauth.v1.CreateCertifitcate` (Fabric Worker)

Validates:
- Caller namespace matches `fabric-shard-*` pattern
- Certificate CN/SANs match caller's org

---

## Bootstrap Flows

Fabric bootstrap has two distinct flows with different lifecycles:

### One-Time: Fabric Initialization

Run **once per cluster** to establish shared TLS infrastructure. Must complete before any shards can be created.

```
cell.v1.InitFabric (Infra Worker, infra namespace)
│
├─ temporal.v1.CreateNamespace (fabric)
├─ cluster.fabric.v1.CreateNamespace (fabric)
├─ cluster.fabric.v1.CreateServiceAccount
├─ secrets.v1.CreateAccess
├─ cluster.fabric.v1.CreateDeployment (fabric worker)
├─ temporal.v1.CreateNexusEndpoint (CreateCertifitcate)
│
└─ Nexus → fabric.v1.BootstrapFabric (Fabric Worker, fabric namespace)
           │
           └─ fabric.v1.BootstrapTlsCertificateAuthorities
              ├─ fabric.v1.DeployCertificateStore
              ├─ fabric.v1.DeployTlsRootCertificateAuthority
              └─ fabric.v1.DeployTlsIntermediateCertificateAuthority
```

**Result:** TLS CA infrastructure ready. Fabric worker accepting `CreateCertifitcate` requests.

### On-Demand: Shard Creation

Run **per organization** when a new org joins the network. Can be invoked multiple times.

```
cell.v1.InitFabricShard (Infra Worker, infra namespace)
│
├─ temporal.v1.CreateNamespace (fabric-shard-{org})
├─ cluster.fabric.v1.CreateNamespace (fabric-shard-{org})
├─ cluster.fabric.v1.CreateServiceAccount
├─ secrets.v1.CreateAccess
├─ cluster.fabric.v1.CreateDeployment (shard worker)
├─ temporal.v1.CreateNexusEndpoint (BootstrapShard)
│
└─ Nexus → fabric.v1.BootstrapShard (Shard Worker, fabric-shard-{org} namespace)
           │
           ├─ fabric.v1.BootstrapOrgCertificateAuthorities
           │  ├─ fabric.v1.DeployCertificateStore
           │  ├─ fabric.v1.DeployOrgRootCertificateAuthority
           │  │  └─ Nexus → fabric.certauth.v1.CreateCertifitcate (Fabric Worker)
           │  └─ fabric.v1.DeployOrgIntermediateCertificateAuthority
           │     └─ Nexus → fabric.certauth.v1.CreateCertifitcate (Fabric Worker)
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
| Fabric | `fabric` | `fabric` | `secret/fabric/tls/*` | Exposes `CreateCertifitcate` |
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
└─ cell.v1.InitFabric 🟡
```

```
cell.v1.InitFabric 🟡
├─ temporal.v1.CreateNamespace ❌ (fabric)
├─ cluster.fabric.v1.CreateNamespace ✅
├─ cluster.fabric.v1.CreateServiceAccount ✅
├─ secrets.v1.CreateAccess ✅
├─ cluster.fabric.v1.CreateDeployment ❌ (fabric worker)
├─ temporal.v1.CreateNexusEndpoint ✅ (CreateCertifitcate)
└─ Nexus → fabric.v1.BootstrapFabric ❌
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

```
fabric.v1.BootstrapFabric ❌
└─ fabric.v1.BootstrapTlsCertificateAuthorities ✅
   ├─ fabric.v1.DeployCertificateStore ✅
   ├─ fabric.v1.DeployTlsRootCertificateAuthority ✅
   └─ fabric.v1.DeployTlsIntermediateCertificateAuthority ✅
```

**TLS CA Deployment Details:**

```
fabric.v1.DeployTlsRootCertificateAuthority ✅
├─ Nexus → secrets.v1.CreateAccess ✅
├─ fabric.v1.CreateCertificateDatabase ✅
├─ secrets.v1.CreateKeyPair ✅ (signing)
├─ secrets.v1.CreateSelfSignedCertificate ✅ (signing)
├─ secrets.v1.CreateKeyPair ✅ (TLS)
├─ secrets.v1.CreateSelfSignedCertificate ✅ (TLS)
├─ secrets.v1.CreatePassword ✅ (bootstrap)
├─ cluster.fabric.v1.CreateDeployment ✅
├─ fabric.certauth.v1.CreateBootstrapAdmin ✅
└─ fabric.certauth.v1.CreateAdmin ✅
```

```
fabric.v1.DeployTlsIntermediateCertificateAuthority ✅
├─ Nexus → secrets.v1.CreateAccess ✅
├─ fabric.v1.CreateCertificateDatabase ✅
├─ fabric.certauth.v1.CreateCertificate ✅ (signing from TLS Root)
├─ fabric.certauth.v1.CreateCertificate ✅ (TLS from TLS Root)
├─ secrets.v1.CreatePassword ✅ (bootstrap)
├─ cluster.fabric.v1.CreateDeployment ✅
├─ fabric.certauth.v1.CreateBootstrapAdmin ✅
└─ fabric.certauth.v1.CreateAdmin ✅
```

---

### Shard Worker Workflows (fabric.v1 package)

```
fabric.v1.BootstrapShard 🟡
├─ fabric.v1.BootstrapOrgCertificateAuthorities ✅
│  ├─ fabric.v1.DeployCertificateStore ✅
│  ├─ fabric.v1.DeployOrgRootCertificateAuthority ✅
│  └─ fabric.v1.DeployOrgIntermediateCertificateAuthority ✅
├─ for each peer:
│  ├─ fabric.v1.DeployCouchDb ✅
│  └─ fabric.v1.DeployPeer ✅
├─ for each orderer:
│  └─ fabric.v1.DeployOrderer ✅
└─ TODO: chaincode, channel join
```

**Org CA Deployment Details:**

```
fabric.v1.DeployOrgRootCertificateAuthority ✅
├─ Nexus → secrets.v1.CreateAccess ✅
├─ fabric.v1.CreateCertificateDatabase ✅
├─ secrets.v1.CreateKeyPair ✅ (signing)
├─ secrets.v1.CreateSelfSignedCertificate ✅ (signing)
├─ Nexus → fabric.certauth.v1.CreateCertifitcate ✅
├─ secrets.v1.CreatePassword ✅ (bootstrap)
├─ cluster.fabric.v1.CreateDeployment ✅
├─ fabric.certauth.v1.CreateBootstrapAdmin ✅
└─ fabric.certauth.v1.CreateAdmin ✅
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
├─ secrets.v1.CreateKeyPair ✅
├─ fabric.certauth.v1.RegisterIdentity ✅ (org cert)
├─ fabric.certauth.v1.EnrollIdentity ✅ (org cert)
├─ fabric.certauth.v1.RegisterIdentity ✅ (TLS cert)
└─ fabric.certauth.v1.EnrollIdentity ✅ (TLS cert)
```

```
fabric.certauth.v1.CreateBootstrapAdmin ✅
├─ secrets.v1.CreateKeyPair ✅
├─ fabric.certauth.v1.BootstrapEnrollIdentity ✅
└─ secrets.v1.StoreIdentityCredentials ✅
```

```
fabric.certauth.v1.CreateAdmin ✅
├─ fabric.certauth.v1.CreateCertificate ✅ (signing)
└─ fabric.certauth.v1.CreateCertificate ✅ (TLS)
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