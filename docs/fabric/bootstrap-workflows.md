# Fabric Bootstrap Workflows

**Status Legend:**
- ✅ Implemented
- 🟡 Partial  
- ❌ Not Started

---

## Overview

This document outlines the complete execution flow for deploying a Fabric organization. It shows the hierarchy of workflows and activities across two worker types, the call sequence via task queue routing, dependency chains, architecture, security boundaries, and implementation status.

**Visual Workflow Tree:** See [`Fabric Shard Bootstrap Workflows.txt`](Fabric%20Shard%20Bootstrap%20Workflows.txt) for the complete hierarchical structure.

---

## Worker Architecture

### Central Worker (infra Temporal namespace, infra task queue)
- **Temporal Namespace:** `infra` (or `default`)
- **Task Queue:** `infra` (single queue)
- **K8s Location:** `workers` namespace
- **K8s RBAC:** Manages only `workers` namespace
- **Vault Token:** God-mode with policy validation
- **Nexus Services:** Exposes `CreateScopedIdentityAccess` in "infra" namespace
- **Responsibilities:** Platform bootstrap, Temporal namespace/Nexus management, Vault policy/role creation

### Per-Org Fabric Workers (dedicated Temporal namespace, infra task queue)
- **Temporal Namespace:** `fabric-{org}` (e.g., `fabric-org1`) - one per org
- **Task Queue:** `infra` (same name, different namespace)
- **K8s Location:** `fabric-{org}` namespace (one pod per org, runs in its own namespace)
- **K8s RBAC:** Role bound to `fabric-{org}` namespace (manages resources in same namespace)
- **Vault Token:** Scoped to `secret/fabric/{org}/*`
- **Nexus Calls:** Calls `CreateScopedIdentityAccess` in "infra" namespace via Nexus
- **Responsibilities:** Fabric deployment, identity provisioning, certificate management

**Isolation Strategy:** 
- Temporal namespace boundaries provide workflow history isolation
- Single task queue per namespace (no routing complexity)
- Cross-namespace calls use Temporal Nexus (org → central for policy creation)
- Central → org uses standard Temporal signals/StartWorkflow

---

## Nexus Authorization Strategy

### Authorization for `CreateScopedIdentityAccess`

**Note:** Nexus endpoint ACL management requires Temporal server plugins (not available in standard SDK). Authorization is handled at the application layer.

**Workflow Validation (Application-level)**
- **Setup:** Code in `CreateScopedIdentityAccess` workflow implementation
- **Mechanism:** Workflow receives caller namespace context and validates request
- **Validation:** Confirms caller can only create policies for their own organization
  - Caller namespace `fabric-org1` → can only create policies under `secret/fabric/org1/*`
  - Caller namespace `fabric-org2` → can only create policies under `secret/fabric/org2/*`
  - Requests for paths outside org scope → rejected
- **Purpose:** Fine-grained authorization; prevents one org from accessing another's secrets

---

## Execution Phases

### Phase 1: Organization Infrastructure Bootstrap

Central Worker orchestrates infrastructure setup:

1. **CreateTemporalNamespace** workflow - async namespace creation with polling
2. **CreateKubernetesNamespace** workflow - create org K8s namespace
3. **CreateServiceAccount** workflow - create fabric worker service account in `workers` namespace
4. **CreateVaultRole** workflow - create org-scoped Vault policy and role for fabric worker
5. **DeployFabricWorker** workflow - render and deploy fabric worker pod with `TEMPORAL_NAMESPACE=fabric-org1` env
6. **CreateNexusEndpoint** activity - register Nexus endpoint for `DeployOrgWorkflow` (enables central→org calls)
7. **Nexus Call** - invoke `DeployOrgWorkflow` in org's Temporal namespace (triggers Phase 2)

**Note:** `CreateScopedIdentityAccess` Nexus endpoint is registered once at central worker init time (not per-org).

### Phase 2: Fabric Worker Deployment

Fabric Worker orchestrates organization deployment:

1. **DeployCA** workflow - deploy Fabric CA pod
2. **CreateAffiliations** activity - create identity affiliations in CA
3. **BootstrapAdmin** workflow - create and enroll bootstrap admin identity
4. **For each peer (SEQUENTIAL):**
   - **DeployCouchDB** workflow - deploy peer-specific CouchDB instance
   - **DeployPeer** workflow - orchestrates peer provisioning and deployment
5. **For each orderer (SEQUENTIAL):**
   - **DeployOrderer** workflow - orchestrates orderer provisioning and deployment
6. Return `DeployOrgOutput` with references to all deployed resources

---

## Security Boundaries

### Per-Worker Isolation

| Worker | Temporal Namespace | K8s Namespace | K8s RBAC | Vault Access | Can Create Policies? |
|--------|-------------------|---------------|----------|--------------|---------------------|
| Central Worker | `infra` | `workers` | `workers` namespace only | God-mode token (all paths) | ✅ Yes (with validation) |
| Fabric Worker (Org1) | `fabric-org1` | `fabric-org1` | Role in `fabric-org1` namespace | `secret/fabric/org1/*` only | ❌ No (calls via Nexus) |
| Fabric Worker (Org2) | `fabric-org2` | `fabric-org2` | Role in `fabric-org2` namespace | `secret/fabric/org2/*` only | ❌ No (calls via Nexus) |

**Temporal Namespace Isolation:**
- Complete workflow history separation per org
- Org1 worker cannot see Org2's workflow executions
- Temporal RBAC enforces namespace boundaries
- Nexus provides controlled cross-namespace RPC

### Per-Identity Isolation

| Pod | Service Account | Vault Policy | Vault Access |
|-----|-----------------|--------------|-------------|
| peer0 (org1) | `fabric-org1-peer0` | `fabric-org1-peer-peer0` | `secret/fabric/org1/msp/peer-peer0/*` + CA certs |
| peer1 (org1) | `fabric-org1-peer1` | `fabric-org1-peer-peer1` | `secret/fabric/org1/msp/peer-peer1/*` + CA certs |
| orderer0 (org1) | `fabric-org1-orderer0` | `fabric-org1-orderer-orderer0` | `secret/fabric/org1/msp/orderer-orderer0/*` + CA certs |

**Security Result:**
- Compromised peer0 pod cannot read peer1's secrets (Vault policy isolation)
- Compromised org1 worker cannot access org2's K8s namespace (RBAC isolation)
- Compromised org1 worker cannot read org2's Vault secrets (Vault token scoping)
- Compromised org1 worker cannot see org2's workflow history (Temporal namespace isolation)
- Compromised org1 worker cannot create arbitrary Vault policies (Nexus validation at central worker)

---

## Implementation Status: Detailed Workflows

---

## Root Bootstrap Workflow - Infra Worker

```
Workflow: cell.v1.BootstrapInfra ✅ (Runs in infra namespace)
└─ Activity: temporal.v1.CreateNexusEndpoint ✅ (creates secrets-CreateAccess endpoint)
```

```
Workflow: cell.v1.InitFabricShard ✅ (Runs in infra namespace)
├─ Workflow: temporal.v1.CreateNamespace ✅
├─ Workflow: cluster.fabric.v1.CreateNamespace ✅
├─ Workflow: cluster.fabric.v1.CreateServiceAccount ✅
├─ Workflow: secrets.v1.CreateAccess ✅
├─ Workflow: cluster.fabric.v1.CreateDeployment ✅
├─ Activity: temporal.v1.CreateNexusEndpoint ✅
│
└─ Nexus Call: fabric.v1.BootstrapShard [→ Fabric Worker in fabric-<org> namespace]
```


## Root Bootstrap Workflow - Fabric Worker

```
Workflow: fabric.v1.BootstrapShard 🟡 (Runs in fabric-<org> namespace)
├─ Workflow: fabric.v1.BootstrapCertificateAuthorities ✅
│
├─ for each Peer (at least 3)
│ ├─ Workflow: fabric.v1.DeployCouchDb ✅
│ └─ Workflow: fabric.v1.DeployPeer ✅
│
├─ for each Orderer (at least 1)
│ └─ Workflow: fabric.v1.DeployOrderer ✅
│
├─ for each ChainCode (at least 2)
│ └─ Workflow: fabric.v1.DeployChainCode 🟡 (skeleton implemented, deployment TODO)
│
└─ TODO: Hook into existing chains/channels (osnadmin channel join)
```


## Fabric Certificate Authority Bootstrap

```
Workflow: fabric.v1.BootstrapCertificateAuthorities ✅
├─ Workflow: fabric.v1.DeployCertificateStore ✅
│
└─ for each (tls, org)
  ├─ Workflow: fabric.v1.DeployRootCertificateAuthority ✅
  └─ Workflow: fabric.v1.DeployCertificateAuthority 🟡 (DB+access ready, intermediate cert TODO)
```


## Fabric Component Deployment Workflows

```
Workflow: fabric.v1.DeployCouchDb ✅
├─ Nexus Call: secrets.v1.CreateAccess ✅ [→ Infra Worker in infra namespace]
├─ Activity: secrets.v1.CreatePassword ✅
└─ Workflow: cluster.fabric.v1.CreateDeployment ✅
```

```
Workflow: fabric.v1.DeployPeer ✅
├─ Nexus Call: secrets.v1.CreateAccess ✅ [→ Infra Worker in infra namespace]
├─ Workflow: fabric.v1.CreateIdentity ✅ (creates both org + TLS certs with IdentityType: "peer")
│
└─ Workflow: cluster.fabric.v1.CreateDeployment ✅
```

```
Workflow: fabric.v1.DeployOrderer ✅
├─ Nexus Call: secrets.v1.CreateAccess ✅ [→ Infra Worker in infra namespace]
├─ Workflow: fabric.v1.CreateIdentity ✅ (creates both org + TLS certs with IdentityType: "orderer")
│
└─ Workflow: cluster.fabric.v1.CreateDeployment ✅
```

```
Workflow: fabric.v1.DeployChainCode 🟡
├─ Nexus Call: secrets.v1.CreateAccess ✅ [→ Infra Worker in infra namespace]
├─ for only ("org")
│ └─ Workflow: fabric.v1.CreateIdentity ✅
│
└─ Workflow: cluster.fabric.v1.CreateDeployment ✅
```


## Fabric Certificate Authority Setup Workflows

```
Workflow: fabric.v1.DeployCertificateStore ✅
├─ Nexus Call: secrets.v1.CreateAccess ✅ [→ Infra Worker in infra namespace]
├─ Activity: secrets.v1.CreatePassword ✅
├─ Workflow: cluster.fabric.v1.CreateDeployment ✅
└─ Activity: kube.v1.GetStatus ✅ (wait for ready)
```

```
Workflow: fabric.v1.CreateCertificateDatabase ✅
├─ Activity: postgres.v1.CreateDatabase ✅
├─ Activity: secrets.v1.CreatePassword ✅
├─ Activity: postgres.v1.CreateUser ✅
└─ Activity: postgres.v1.GrantPrivileges ✅
```

```
Workflow: fabric.v1.DeployRootCertificateAuthority ✅
├─ Nexus Call: secrets.v1.CreateAccess ✅ [→ Infra Worker in infra namespace]
├─ Workflow: fabric.v1.CreateCertificateDatabase ✅
├─ Activity: secrets.v1.CreateKeyPair ✅ (for CA's signing key)
├─ Activity: secrets.v1.CreateSelfSignedCertificate ✅ (CA's certificate)
├─ Activity: secrets.v1.CreatePassword ✅ (bootstrap admin credentials)
├─ Workflow: cluster.fabric.v1.CreateDeployment ✅
└─ Workflow: fabric.v1.BootstrapCertificateAuthorityAdmin ✅
```

```
Workflow: fabric.v1.DeployCertificateAuthority ✅
├─ Nexus Call: secrets.v1.CreateAccess ✅ [→ Infra Worker in infra namespace]
├─ Workflow: fabric.v1.CreateCertificateDatabase ✅
├─ Workflow: cluster.fabric.v1.ScaleDeployment ❌ (optional, not implemented)
├─ Activity: secrets.v1.CreateKeyPair ✅ (intermediate CA key)
├─ Activity: fabric.v1.RegisterIdentity ✅ (register with root CA)
├─ Activity: fabric.v1.EnrollIdentity ✅ (get cert signed by root CA)
├─ Workflow: cluster.fabric.v1.ScaleDeployment ❌ (optional, not implemented)
├─ Activity: secrets.v1.CreatePassword ✅ (bootstrap admin credentials)
├─ Workflow: cluster.fabric.v1.CreateDeployment ✅
└─ Workflow: fabric.v1.BootstrapCertificateAuthorityAdmin ✅
```


## Fabric Certificate Authority Admin Bootstrap

```
Workflow: fabric.v1.BootstrapCertificateAuthorityAdmin ✅
├─ Activity: secrets.v1.CreateKeyPair ✅
├─ Activity: fabric.v1.BootstrapEnrollIdentity ✅ (fetches key, enrolls with CA, stores keypair+cert)
├─ Activity: secrets.v1.StoreIdentityCredentials ✅ (stores keypair+cert together for client auth)
├─ Workflow: fabric.v1.CreateIdentity ✅ (creates permanent admin)
└─ Activity: fabric.v1.RevokeIdentity 🟡 (TODO: revoke bootstrap credentials after admin is enrolled)
```


## Fabric Identity Workflows

```
Workflow: fabric.v1.CreateIdentity ✅
├─ Activity: secrets.v1.CreateKeyPair ✅
├─ Activity: fabric.v1.RegisterIdentity ✅ (registers identity with CA, returns enrollment secret)
└─ Activity: fabric.v1.EnrollIdentity ✅ (enrolls using secret, fetches & stores certificate with keypair)
```

**Fabric CA Activities:**
- `fabric.v1.RegisterIdentity` ✅ - Register new identity with CA (admin auth required)
- `fabric.v1.EnrollIdentity` ✅ - Enroll identity using enrollment secret, store keypair+cert
- `fabric.v1.BootstrapEnrollIdentity` ✅ - Bootstrap enrollment using username/password
- `fabric.v1.RevokeIdentity` ✅ - Revoke identity from CA (admin auth required)


## Secrets Workflows (Secret Management)

```
Workflow: secrets.v1.CreateAccess ✅
├─ Activity: vault.v1.CreatePolicy ✅
└─ Activity: vault.v1.CreateAuthRole ✅
```

**Secrets Activities:**
- `secrets.v1.CreatePassword` ✅ - Generate and store password credentials
- `secrets.v1.CreateKeyPair` ✅ - Generate and store ECDSA keypair
- `secrets.v1.CreateSelfSignedCertificate` ✅ - Generate self-signed CA certificate
- `secrets.v1.StoreCertificate` ✅ - Store certificate data
- `secrets.v1.StoreIdentityCredentials` ✅ - Store keypair+cert together for client auth


## Cluster/Infrastructure Workflows

```
Workflow: cluster.fabric.v1.CreateNamespace ✅
├─ Activity: helm.v1.RenderChartWithTemplate ✅
└─ Activity: kube.v1.Apply ✅
```

```
Workflow: cluster.fabric.v1.CreateServiceAccount ✅
├─ Activity: helm.v1.RenderChartWithTemplate ✅
└─ Activity: kube.v1.Apply ✅
```

```
Workflow: cluster.fabric.v1.CreateDeployment ✅
├─ Activity: helm.v1.RenderChartWithTemplate ✅
└─ Activity: kube.v1.Apply ✅
```

```
Workflow: cluster.fabric.v1.ScaleDeployment ❌
└─ Activity: kube.v1.Apply ✅
```

```
Workflow: temporal.v1.CreateNamespace ✅
├─ Activity: temporal.v1.RegisterNamespace ✅
└─ Activity: temporal.v1.DescribeNamespace ✅
```

---

## Error Handling

**Failure Strategy:** Fail immediately on critical sections
- If any step fails, the workflow stops immediately
- No rollback or cleanup is performed
- Partial deployments may exist (e.g., namespace created but peers not)

**Retry Policy:** Each activity/workflow has retry logic configured:
- Initial interval: 1 second
- Backoff coefficient: 2.0
- Maximum interval: 30 seconds (activities) / 1 minute (workflows)
- Maximum attempts: 3

---

## See Also

- [`Fabric Shard Bootstrap Workflows.txt`](Fabric%20Shard%20Bootstrap%20Workflows.txt) - Complete workflow tree visualization
- [`FABRIC_CERTIFICATE_ARCHITECTURE.md`](FABRIC_CERTIFICATE_ARCHITECTURE.md) - Architecture and design decisions
- [`DEPLOYMENT_WORKFLOW_EXECUTION_TREE.md`](DEPLOYMENT_WORKFLOW_EXECUTION_TREE.md) - Detailed execution tree (now merged into this document)
- [`Fabric Certs.csv`](Fabric%20Certs.csv) - Certificate matrix

---

**Last Updated:** 2026-01-05
