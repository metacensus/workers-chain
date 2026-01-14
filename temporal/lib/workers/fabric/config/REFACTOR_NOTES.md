# FabricConfig Refactor Notes

**Purpose**: Characterization documentation for the fabric config system. Documents what the code currently does (not what it should do) to serve as a baseline for safe refactoring.

**Created**: 2026-01-13  
**Status**: Temporary (delete after refactoring complete)

---

## Certificate Authority Config Matrix

The `CertificateAuthorityConfig` is the most complex config due to the dual-chain PKI architecture. This matrix documents the return values of all methods for each of the 4 CA types.

**Assumptions for examples**: `orgID = "org1"`

---

### 1. Issuer Resolution Methods

These methods determine who signs certificates for each CA. **Critical for correctness.**

| Method | TLS Root CA | TLS Int CA | Org Root CA | Org Int CA |
|--------|-------------|------------|-------------|------------|
| `IsRoot()` | `true` | `false` | `true` | `false` |
| `IsTLS()` | `true` | `true` | `false` | `false` |
| `SigningIssuer()` | `nil` (self-signed) | TLS Root CA | `nil` (self-signed) | Org Root CA |
| `TLSIssuer()` | `nil` (self-signed) | TLS Root CA | TLS Int CA | TLS Int CA |
| `AdminSigningIssuer()` | TLS Root CA (self) | TLS Int CA (self) | Org Root CA (self) | Org Int CA (self) |
| `AdminTLSIssuer()` | TLS Root CA (self) | TLS Int CA (self) | TLS Int CA | TLS Int CA |

**Key Observations:**
1. `AdminSigningIssuer()` always returns `self` - admins get identity from the CA they administer
2. `AdminTLSIssuer()` returns `self` for TLS chain, returns TLS Int CA for Org chain
3. `SigningIssuer()` returns `nil` for root CAs (self-signed), parent root for intermediates
4. `TLSIssuer()` returns `nil` only for TLS Root CA; all others get TLS from TLS chain

**Issuer Resolution Logic (from code):**

```go
// SigningIssuer: who signs this CA's signing certificate
func (ca) SigningIssuer() *CertificateAuthorityConfig {
    if ca.IsRoot() { return nil }  // Self-signed
    return ca.withID("root")       // Same chain's root
}

// TLSIssuer: who signs this CA's TLS certificate  
func (ca) TLSIssuer() *CertificateAuthorityConfig {
    if ca.IsTLS() && ca.IsRoot() { return nil }      // TLS Root self-signs
    if ca.IsTLS() { return ca.withID("root") }       // TLS Int from TLS Root
    return ca.withTypeAndID("tls", "v1")             // Org CAs from TLS Int (cross-chain!)
}

// AdminSigningIssuer: who signs admin's identity cert
func (ca) AdminSigningIssuer() CertificateAuthorityConfig {
    return ca  // Always self - Fabric CA database constraint
}

// AdminTLSIssuer: who signs admin's TLS client cert
func (ca) AdminTLSIssuer() CertificateAuthorityConfig {
    if ca.IsTLS() && ca.IsRoot() { return ca }       // TLS Root admin from TLS Root
    return ca.withTypeAndID("tls", "v1")             // Everyone else from TLS Int
}
```

---

### 2. Common Name Methods (Collision Avoidance)

When the same CA issues both signing and TLS certs for an identity, names must differ to avoid registration collision in Fabric CA's database.

| Method | TLS Root CA | TLS Int CA | Org Root CA | Org Int CA |
|--------|-------------|------------|-------------|------------|
| `ComponentName()` | `certificate-authority-tls-root` | `certificate-authority-tls-v1` | `certificate-authority-org-root` | `certificate-authority-org-v1` |
| `TLSCommonName()` | `certificate-authority-tls-root` | `certificate-authority-tls-v1-tls` | `certificate-authority-org-root` | `certificate-authority-org-v1` |
| `AdminCommonName()` | `admin@certificate-authority-tls-root` | `admin@certificate-authority-tls-v1` | `admin@certificate-authority-org-root` | `admin@certificate-authority-org-v1` |
| `AdminTLSCommonName()` | `admin-tls@certificate-authority-tls-root` | `admin-tls@certificate-authority-tls-v1` | `admin@certificate-authority-org-root` | `admin@certificate-authority-org-v1` |
| `BootstrapAdminCommonName()` | `bootstrap-admin@certificate-authority-tls-root` | `bootstrap-admin@certificate-authority-tls-v1` | `bootstrap-admin@certificate-authority-org-root` | `bootstrap-admin@certificate-authority-org-v1` |
| `BootstrapAdminTLSCommonName()` | `bootstrap-admin-tls@certificate-authority-tls-root` | `bootstrap-admin-tls@certificate-authority-tls-v1` | `bootstrap-admin@certificate-authority-org-root` | `bootstrap-admin@certificate-authority-org-v1` |

**Collision Detection Logic:**

```go
// TLSCommonName: appends "-tls" if signing and TLS issuers are the SAME CA
func (ca) TLSCommonName() string {
    signingIssuer := ca.SigningIssuer()
    tlsIssuer := ca.TLSIssuer()
    
    // Self-signed cases: no collision (not registering with a CA)
    if signingIssuer == nil || tlsIssuer == nil {
        return ca.ComponentName()
    }
    
    // Same CA issues both → need different name
    if signingIssuer.CertType() == tlsIssuer.CertType() && 
       signingIssuer.CAID() == tlsIssuer.CAID() {
        return ca.ComponentName() + "-tls"
    }
    
    return ca.ComponentName()
}
```

**Why TLS Int CA gets `-tls` suffix:**
- `SigningIssuer()` = TLS Root CA (tls/root)
- `TLSIssuer()` = TLS Root CA (tls/root)
- Same CA → collision → appends `-tls`

**Why Org CAs don't need `-tls` suffix:**
- `SigningIssuer()` = Org Root CA (org/root) or nil
- `TLSIssuer()` = TLS Int CA (tls/v1)
- Different CAs → no collision → use base name

---

### 3. Vault Path Methods

All paths are relative to `secret/` (KV v2 engine).

| Method | TLS Root CA | TLS Int CA | Org Root CA | Org Int CA |
|--------|-------------|------------|-------------|------------|
| `VaultBasePath()` | `fabric/org1/ca/tls/root` | `fabric/org1/ca/tls/v1` | `fabric/org1/ca/org/root` | `fabric/org1/ca/org/v1` |
| `VaultSigningRef()` | `.../tls/root/signing` | `.../tls/v1/signing` | `.../org/root/signing` | `.../org/v1/signing` |
| `VaultSigningChainRef()` | `.../tls/root/signing-chain` | `.../tls/v1/signing-chain` | `.../org/root/signing-chain` | `.../org/v1/signing-chain` |
| `VaultTLSRef()` | `.../tls/root/tls` | `.../tls/v1/tls` | `.../org/root/tls` | `.../org/v1/tls` |
| `VaultDatabaseUserRef()` | `.../tls/root/certstore/user` | `.../tls/v1/certstore/user` | `.../org/root/certstore/user` | `.../org/v1/certstore/user` |
| `VaultBootstrapAdminCredentialsRef()` | `.../tls/root/bootstrap-admin/credentials` | `.../tls/v1/bootstrap-admin/credentials` | `.../org/root/bootstrap-admin/credentials` | `.../org/v1/bootstrap-admin/credentials` |
| `VaultBootstrapAdminSigningRef()` | `.../tls/root/bootstrap-admin/signing` | `.../tls/v1/bootstrap-admin/signing` | `.../org/root/bootstrap-admin/signing` | `.../org/v1/bootstrap-admin/signing` |
| `VaultBootstrapAdminTLSRef()` | `.../tls/root/bootstrap-admin/tls` | `.../tls/v1/bootstrap-admin/tls` | `.../org/root/bootstrap-admin/tls` | `.../org/v1/bootstrap-admin/tls` |
| `VaultAdminSigningRef()` | `.../tls/root/admin/signing` | `.../tls/v1/admin/signing` | `.../org/root/admin/signing` | `.../org/v1/admin/signing` |
| `VaultAdminTLSRef()` | `.../tls/root/admin/tls` | `.../tls/v1/admin/tls` | `.../org/root/admin/tls` | `.../org/v1/admin/tls` |

**Path Pattern**: `fabric/{orgID}/ca/{certType}/{caID}/{subpath}`

**Note**: All Vault refs use the same `VaultBasePath()` + suffix pattern. Very consistent.

---

### 4. Kubernetes Properties

| Method | TLS Root CA | TLS Int CA | Org Root CA | Org Int CA |
|--------|-------------|------------|-------------|------------|
| `ComponentName()` | `certificate-authority-tls-root` | `certificate-authority-tls-v1` | `certificate-authority-org-root` | `certificate-authority-org-v1` |
| `ServiceAccountName()` | `certificate-authority-tls-root` | `certificate-authority-tls-v1` | `certificate-authority-org-root` | `certificate-authority-org-v1` |
| `KubernetesNamespace()` | `fabric-org1` | `fabric-org1` | `fabric-org1` | `fabric-org1` |
| `URL()` | `https://certificate-authority-tls-root.fabric-org1.svc.cluster.local:7054` | `https://certificate-authority-tls-v1.fabric-org1.svc.cluster.local:7054` | `https://certificate-authority-org-root.fabric-org1.svc.cluster.local:7054` | `https://certificate-authority-org-v1.fabric-org1.svc.cluster.local:7054` |
| `ResourceIdentifier().Kind` | `Deployment` | `Deployment` | `Deployment` | `Deployment` |

**TLSDNSNames()** returns 4 entries for each CA:
```
[
  "certificate-authority-{type}-{id}",
  "certificate-authority-{type}-{id}.fabric-org1",
  "certificate-authority-{type}-{id}.fabric-org1.svc",
  "certificate-authority-{type}-{id}.fabric-org1.svc.cluster.local"
]
```

---

### 5. Database Properties

| Method | TLS Root CA | TLS Int CA | Org Root CA | Org Int CA |
|--------|-------------|------------|-------------|------------|
| `DatabaseName()` | `fabric_org1_ca_tls_root` | `fabric_org1_ca_tls_v1` | `fabric_org1_ca_org_root` | `fabric_org1_ca_org_v1` |
| `DatabaseUsername()` | `fabric_org1_ca_tls_root_user` | `fabric_org1_ca_tls_v1_user` | `fabric_org1_ca_org_root_user` | `fabric_org1_ca_org_v1_user` |

**Note**: Database names are sanitized for PostgreSQL (hyphens → underscores, lowercase).

---

### 6. Certificate Subject Properties

| Method | TLS Root CA | TLS Int CA | Org Root CA | Org Int CA |
|--------|-------------|------------|-------------|------------|
| `CertType()` | `tls` | `tls` | `org` | `org` |
| `CAID()` | `root` | `v1` | `root` | `v1` |
| `OrganizationalUnit()` | `root-ca` | `v1` | `root-ca` | `v1` |
| `Description()` | `TLS Root CA for org1` | `TLS Intermediate CA for org1` | `Org Root CA for org1` | `Org Intermediate CA for org1` |

---

### 7. Vault Policy Properties

| Method | TLS Root CA | TLS Int CA | Org Root CA | Org Int CA |
|--------|-------------|------------|-------------|------------|
| `PolicyName()` | `fabric-org1-certificate-authority-tls-root` | `fabric-org1-certificate-authority-tls-v1` | `fabric-org1-certificate-authority-org-root` | `fabric-org1-certificate-authority-org-v1` |
| `AuthRoleName()` | `fabric-org1-certificate-authority-tls-root` | `fabric-org1-certificate-authority-tls-v1` | `fabric-org1-certificate-authority-org-root` | `fabric-org1-certificate-authority-org-v1` |
| `VaultDataPath()` | `secret/data/fabric/org1/ca/tls/root/*` | `secret/data/fabric/org1/ca/tls/v1/*` | `secret/data/fabric/org1/ca/org/root/*` | `secret/data/fabric/org1/ca/org/v1/*` |
| `VaultMetadataPath()` | `secret/metadata/fabric/org1/ca/tls/root/*` | `secret/metadata/fabric/org1/ca/tls/v1/*` | `secret/metadata/fabric/org1/ca/org/root/*` | `secret/metadata/fabric/org1/ca/org/v1/*` |

---

## Summary: Critical Decision Points

### When to use `-tls` suffix (collision avoidance)

The `-tls` suffix is added when:
1. The signing issuer and TLS issuer are the **same CA**
2. Both issuers are **not nil** (not self-signed)

| Entity | Has `-tls` suffix? | Reason |
|--------|-------------------|--------|
| TLS Root CA's TLS cert | No | Self-signed (TLSIssuer=nil) |
| TLS Int CA's TLS cert | **Yes** | Both issuers = TLS Root CA |
| Org Root CA's TLS cert | No | SigningIssuer=nil (self-signed) |
| Org Int CA's TLS cert | No | Different issuers (Org Root vs TLS Int) |
| TLS Root Admin TLS cert | **Yes** | Both issuers = TLS Root CA |
| TLS Int Admin TLS cert | **Yes** | Both issuers = TLS Int CA |
| Org Root Admin TLS cert | No | Different issuers (Org Root vs TLS Int) |
| Org Int Admin TLS cert | No | Different issuers (Org Int vs TLS Int) |

### Cross-Chain Dependencies

The Org chain depends on the TLS chain being bootstrapped first:

```
Bootstrap Order:
1. TLS Root CA (self-signed signing + self-signed TLS)
2. TLS Int CA (signing from TLS Root, TLS from TLS Root)
3. Org Root CA (self-signed signing, TLS from TLS Int ← DEPENDS ON TLS Int)
4. Org Int CA (signing from Org Root, TLS from TLS Int ← DEPENDS ON TLS Int)
```

---

## Other Component Configs

### CertificateStoreConfig

The shared PostgreSQL database for all CAs in an organization.

**Assumptions**: `orgID = "org1"`

| Method | Return Value |
|--------|--------------|
| `ComponentName()` | `certificate-store` |
| `ServiceAccountName()` | `certificate-store` |
| `KubernetesNamespace()` | `fabric-org1` |
| `ServiceHost()` | `certificate-store.fabric-org1.svc.cluster.local` |
| `Port()` | `5432` |
| `AdminDatabase()` | `postgres` |
| `ResourceIdentifier().Kind` | `StatefulSet` |
| `PolicyName()` | `fabric-org1-certificate-store` |
| `AuthRoleName()` | `fabric-org1-certificate-store` |
| `VaultAdminRef()` | `fabric/org1/certstore/admin` |
| `VaultDataPath()` | `secret/data/fabric/org1/certstore/*` |
| `VaultMetadataPath()` | `secret/metadata/fabric/org1/certstore/*` |

**Pattern**: Single component per org, no parameterization needed.

---

### CouchDbConfig

Per-peer state database. Each peer gets its own CouchDB instance.

**Assumptions**: `orgID = "org1"`, `peerID = "peer0"`

| Method | Return Value |
|--------|--------------|
| `ComponentName()` | `couchdb-peer0` |
| `ServiceAccountName()` | `couchdb-peer0` |
| `KubernetesNamespace()` | `fabric-org1` |
| `Port()` | `5984` |
| `PeerID()` | `peer0` |
| `ResourceIdentifier().Kind` | `StatefulSet` |
| `PolicyName()` | `fabric-org1-couchdb-peer0` |
| `AuthRoleName()` | `fabric-org1-couchdb-peer0` |
| `VaultBasePath()` | `fabric/org1/couchdb/peer0` |
| `VaultAdminRef()` | `fabric/org1/couchdb/peer0/admin` |
| `VaultDataPath()` | `secret/data/fabric/org1/couchdb/peer0/*` |
| `VaultMetadataPath()` | `secret/metadata/fabric/org1/couchdb/peer0/*` |

**Pattern**: Parameterized by `peerID`. Path pattern: `fabric/{orgID}/couchdb/{peerID}/...`

---

### PeerConfig

Fabric peer node configuration.

**Assumptions**: `orgID = "org1"`, `peerID = "peer0"`

| Method | Return Value |
|--------|--------------|
| `ComponentName()` | `peer0` |
| `ServiceAccountName()` | `peer0` |
| `KubernetesNamespace()` | `fabric-org1` |
| `Port()` | `7051` |
| `ChaincodePort()` | `7052` |
| `OperationsPort()` | `9443` |
| `PeerID()` | `peer0` |
| `MSPID()` | `org1MSP` |
| `ResourceIdentifier().Kind` | `StatefulSet` |
| `PolicyName()` | `fabric-org1-peer0` |
| `AuthRoleName()` | `fabric-org1-peer0` |
| `VaultBasePath()` | `fabric/org1/peer/peer0` |
| `VaultSigningRef()` | `fabric/org1/peer/peer0/signing` |
| `VaultTLSRef()` | `fabric/org1/peer/peer0/tls` |
| `VaultDataPath()` | `secret/data/fabric/org1/peer/peer0/*` |
| `VaultMetadataPath()` | `secret/metadata/fabric/org1/peer/peer0/*` |
| `IdentityName()` | `peer0` |
| `CommonName()` | `peer0@org1` |
| `OrganizationalUnit()` | `peer` |
| `OrgIssuerCertType()` | `org` |
| `OrgIssuerID()` | `v1` |
| `TLSIssuerCertType()` | `tls` |
| `TLSIssuerID()` | `v1` |

**Additional Methods**:
- `TLSDNSNames()`: `["peer0", "peer0.fabric-org1", "peer0.fabric-org1.svc", "peer0.fabric-org1.svc.cluster.local"]`
- `CouchDbAddress()`: `couchdb-peer0.fabric-org1.svc.cluster.local:5984`
- `GossipBootstrapAddress()`: `""` for peer0, `peer0.fabric-org1.svc.cluster.local:7051` for others
- `GossipExternalEndpoint()`: `peer0.fabric-org1.svc.cluster.local:7051`

**Pattern**: 
- Parameterized by `peerID`
- Path pattern: `fabric/{orgID}/peer/{peerID}/...`
- Always issued by Org Int CA (signing) and TLS Int CA (TLS) - hardcoded "v1"

---

### OrdererConfig

Fabric orderer node configuration.

**Assumptions**: `orgID = "org1"`, `ordererID = "orderer0"`

| Method | Return Value |
|--------|--------------|
| `ComponentName()` | `orderer0` |
| `ServiceAccountName()` | `orderer0` |
| `KubernetesNamespace()` | `fabric-org1` |
| `Port()` | `7050` |
| `AdminPort()` | `7053` |
| `OperationsPort()` | `9443` |
| `OrdererID()` | `orderer0` |
| `MSPID()` | `org1MSP` |
| `ResourceIdentifier().Kind` | `StatefulSet` |
| `PolicyName()` | `fabric-org1-orderer0` |
| `AuthRoleName()` | `fabric-org1-orderer0` |
| `VaultBasePath()` | `fabric/org1/orderer/orderer0` |
| `VaultSigningRef()` | `fabric/org1/orderer/orderer0/signing` |
| `VaultTLSRef()` | `fabric/org1/orderer/orderer0/tls` |
| `VaultDataPath()` | `secret/data/fabric/org1/orderer/orderer0/*` |
| `VaultMetadataPath()` | `secret/metadata/fabric/org1/orderer/orderer0/*` |
| `IdentityName()` | `orderer0` |
| `CommonName()` | `orderer0@org1` |
| `OrganizationalUnit()` | `orderer` |
| `OrgIssuerCertType()` | `org` |
| `OrgIssuerID()` | `v1` |
| `TLSIssuerCertType()` | `tls` |
| `TLSIssuerID()` | `v1` |

**Additional Methods**:
- `TLSDNSNames()`: `["orderer0", "orderer0.fabric-org1", "orderer0.fabric-org1.svc", "orderer0.fabric-org1.svc.cluster.local"]`
- `ClusterAddress()`: `orderer0.fabric-org1.svc.cluster.local:7050`

**Pattern**: 
- Nearly identical to PeerConfig
- Parameterized by `ordererID`
- Path pattern: `fabric/{orgID}/orderer/{ordererID}/...`
- Always issued by Org Int CA (signing) and TLS Int CA (TLS) - hardcoded "v1"

---

## Cross-Config Pattern Comparison

| Aspect | CertificateStoreConfig | CouchDbConfig | PeerConfig | OrdererConfig | CertificateAuthorityConfig |
|--------|------------------------|---------------|------------|---------------|---------------------------|
| **Parameterized by** | None | peerID | peerID | ordererID | certType + caID |
| **Path base** | `certstore` | `couchdb/{peerID}` | `peer/{peerID}` | `orderer/{ordererID}` | `ca/{certType}/{caID}` |
| **Has VaultBasePath()** | No (inline) | Yes | Yes | Yes | Yes |
| **Has VaultSigningRef()** | No | No | Yes | Yes | Yes |
| **Has VaultTLSRef()** | No | No | Yes | Yes | Yes |
| **Has Issuer methods** | No | No | Hardcoded | Hardcoded | Dynamic resolution |
| **ResourceIdentifier Kind** | StatefulSet | StatefulSet | StatefulSet | StatefulSet | Deployment |
| **VaultPolicy capabilities** | READ,LIST | READ,LIST | READ,LIST + CA certs | READ,LIST + CA certs | CREATE,READ,UPDATE,LIST |

**Key Observations**:
1. **PeerConfig and OrdererConfig are nearly identical** - differ only in ports, OU, and component prefix
2. **CertificateStoreConfig lacks VaultBasePath()** - inline path construction differs from others
3. **Only CertificateAuthorityConfig has dynamic issuer resolution** - Peer/Orderer hardcode "v1"
4. **Only CertificateAuthorityConfig uses Deployment** - all others use StatefulSet

---

## Caller Usage Patterns

### workflows.go - Config Method Usage by Workflow

#### DeployCertificateStoreWorkflow
```
fabricConfig.CertificateStore() → certStore
  - certStore.PolicyName()
  - certStore.VaultPolicy()
  - certStore.AuthRoleName()
  - certStore.ServiceAccountName()
  - certStore.KubernetesNamespace()
  - certStore.VaultAdminRef()
  - certStore.ComponentName()
```

#### DeployCertificateAuthorityWorkflow
```
fabricConfig.CertificateAuthority(certType, caID) → caConfig
  - caConfig.Description()
  - caConfig.PolicyName()
  - caConfig.VaultPolicy()
  - caConfig.AuthRoleName()
  - caConfig.ServiceAccountName()
  - caConfig.KubernetesNamespace()
  - caConfig.SigningIssuer() → signingIssuer
    - signingIssuer.Description()
    - signingIssuer.CertType()
    - signingIssuer.CAID()
    - signingIssuer.VaultAdminSigningRef()
  - caConfig.VaultSigningRef()
  - caConfig.ComponentName()
  - caConfig.OrganizationalUnit()
  - caConfig.VaultSigningChainRef()
  - caConfig.TLSIssuer() → tlsIssuer
    - tlsIssuer.Description()
    - tlsIssuer.CertType()
    - tlsIssuer.CAID()
    - tlsIssuer.VaultAdminSigningRef()
  - caConfig.VaultTLSRef()
  - caConfig.TLSCommonName()
  - caConfig.TLSDNSNames()
  - caConfig.VaultBootstrapAdminCredentialsRef()
```

#### CreateCertificateDatabaseWorkflow
```
fabricConfig.CertificateAuthority(certType, caID) → caConfig
fabricConfig.CertificateStore() → certStore
  - certStore.ResourceIdentifier()
  - caConfig.DatabaseName()
  - caConfig.VaultDatabaseUserRef()
  - caConfig.DatabaseUsername()
```

#### CreateCertificateWorkflow
```
fabricConfig.CertificateAuthority(certType, caID) → caConfig
  - caConfig.VaultTLSRef()
```

#### CreateBootstrapAdminWorkflow
```
fabricConfig.CertificateAuthority(certType, caID) → caConfig
  - caConfig.Description()
  - caConfig.ComponentName()
  - caConfig.ResourceIdentifier()
  - caConfig.VaultBootstrapAdminSigningRef()
  - caConfig.URL()
  - caConfig.VaultBootstrapAdminCredentialsRef()
  - caConfig.BootstrapAdminCommonName()
  - caConfig.OrganizationalUnit()
  - caConfig.VaultTLSRef()
  - caConfig.VaultBootstrapAdminTLSRef()
  - caConfig.AdminTLSIssuer() → tlsIssuer
    - tlsIssuer.URL()
    - tlsIssuer.VaultBootstrapAdminCredentialsRef()
    - tlsIssuer.VaultTLSRef()
  - caConfig.BootstrapAdminTLSCommonName()
```

#### CreateAdminWorkflow
```
fabricConfig.CertificateAuthority(certType, caID) → caConfig
  - caConfig.Description()
  - caConfig.AdminSigningIssuer() → signingIssuer
    - signingIssuer.CertType()
    - signingIssuer.CAID()
    - signingIssuer.VaultAdminSigningRef()
  - caConfig.CertType()
  - caConfig.CAID()
  - caConfig.VaultBootstrapAdminSigningRef()
  - caConfig.AdminCommonName()
  - caConfig.OrganizationalUnit()
  - caConfig.VaultAdminSigningRef()
  - caConfig.AdminTLSIssuer() → tlsIssuer
    - tlsIssuer.CertType()
    - tlsIssuer.CAID()
    - tlsIssuer.VaultAdminSigningRef()
  - caConfig.AdminTLSCommonName()
  - caConfig.VaultAdminTLSRef()
```

#### CreateIdentityWorkflow
```
fabricConfig.CertificateAuthority(orgIssuer.CertType, orgIssuer.Id) → orgIssuerConfig
fabricConfig.CertificateAuthority(tlsIssuer.CertType, tlsIssuer.Id) → tlsIssuerConfig
  - orgIssuerConfig.VaultAdminSigningRef()
  - tlsIssuerConfig.VaultAdminSigningRef()
```

#### DeployCouchDbWorkflow
```
fabricConfig.CouchDb(peerID) → couchDbConfig
  - couchDbConfig.PolicyName()
  - couchDbConfig.VaultPolicy()
  - couchDbConfig.AuthRoleName()
  - couchDbConfig.ServiceAccountName()
  - couchDbConfig.KubernetesNamespace()
  - couchDbConfig.VaultAdminRef()
  - couchDbConfig.ComponentName()
  - couchDbConfig.ResourceIdentifier()
```

#### DeployPeerWorkflow
```
fabricConfig.Peer(peerID) → peerConfig
  - peerConfig.PolicyName()
  - peerConfig.VaultPolicy()
  - peerConfig.AuthRoleName()
  - peerConfig.ServiceAccountName()
  - peerConfig.KubernetesNamespace()
  - peerConfig.IdentityName()
  - peerConfig.CommonName()
  - peerConfig.OrganizationalUnit()
  - peerConfig.TLSDNSNames()
  - peerConfig.OrgIssuerCertType()
  - peerConfig.OrgIssuerID()
  - peerConfig.TLSIssuerCertType()
  - peerConfig.TLSIssuerID()
  - peerConfig.VaultSigningRef()
  - peerConfig.VaultTLSRef()
  - peerConfig.ComponentName()
```

#### DeployOrdererWorkflow
```
fabricConfig.Orderer(ordererID) → ordererConfig
  - ordererConfig.PolicyName()
  - ordererConfig.VaultPolicy()
  - ordererConfig.AuthRoleName()
  - ordererConfig.ServiceAccountName()
  - ordererConfig.KubernetesNamespace()
  - ordererConfig.IdentityName()
  - ordererConfig.CommonName()
  - ordererConfig.OrganizationalUnit()
  - ordererConfig.TLSDNSNames()
  - ordererConfig.OrgIssuerCertType()
  - ordererConfig.OrgIssuerID()
  - ordererConfig.TLSIssuerCertType()
  - ordererConfig.TLSIssuerID()
  - ordererConfig.VaultSigningRef()
  - ordererConfig.VaultTLSRef()
  - ordererConfig.ComponentName()
```

---

### activities.go - Config Method Usage

#### RegisterIdentity
```
fabricConfig.CertificateAuthority(certType, caID) → caConfig
  - Passed to certauthority.NewConfig(caConfig, adminCredentialsRef)
```

#### EnrollIdentity
```
fabricConfig.CertificateAuthority(certType, caID) → caConfig
  - caConfig.URL()
  - caConfig.VaultTLSRef()
```

#### RevokeIdentity
```
fabricConfig.CertificateAuthority(certType, caID) → caConfig
  - Passed to certauthority.NewConfig(caConfig, adminCredentialsRef)
```

---

### certauthority/provider.go - Config Method Usage

#### BootstrapConfig (embeds CertificateAuthorityConfig)
```
GetSecretRefs():
  - c.VaultBootstrapAdminCredentialsRef()
  - c.VaultTLSRef()

Hydrate():
  - c.ComponentName()
  - c.KubernetesNamespace()
```

#### Config (embeds CertificateAuthorityConfig)
```
GetSecretRefs():
  - c.VaultAdminSigningRef() (or AdminCredentialsRef if provided)
  - c.VaultTLSRef()

Hydrate():
  - c.ComponentName()
  - c.KubernetesNamespace()
```

---

### Common Method Bundles (Repeated Patterns)

The same groups of methods are called together repeatedly:

#### "Vault Access Bundle" (used for Nexus CreateAccess calls)
```
PolicyName()
VaultPolicy()
AuthRoleName()
ServiceAccountName()
KubernetesNamespace()
```
**Used in**: DeployCertificateStore, DeployCertificateAuthority, DeployCouchDb, DeployPeer, DeployOrderer

#### "Deployment Bundle" (used for cluster.CreateDeployment calls)
```
ComponentName()
(implicit: OrgID from fabricConfig)
```
**Used in**: All deployment workflows

#### "Identity Creation Bundle" (used for CreateIdentity calls)
```
IdentityName()
CommonName()
OrganizationalUnit()
TLSDNSNames()
OrgIssuerCertType()
OrgIssuerID()
TLSIssuerCertType()
TLSIssuerID()
VaultSigningRef()
VaultTLSRef()
```
**Used in**: DeployPeer, DeployOrderer (nearly identical code)

---

## Pattern Analysis

### 1. Repeated Logic

#### Issuer Resolution Pattern
All four issuer methods follow similar decision trees:

```go
// Pattern: Check self-signed → Check TLS chain → Default to TLS Int
SigningIssuer():     IsRoot() ? nil : sameChainRoot
TLSIssuer():         (IsTLS() && IsRoot()) ? nil : (IsTLS() ? TLSRoot : TLSInt)
AdminSigningIssuer(): self (always)
AdminTLSIssuer():    (IsTLS() && IsRoot()) ? self : TLSInt
```

**Observation**: The logic could be extracted into a decision table or rule engine.

#### Collision Avoidance Pattern
Three methods repeat the same collision detection logic:

```go
// Pattern: If both issuers are same CA, append "-tls" suffix
func needsTLSSuffix(signingIssuer, tlsIssuer *CA) bool {
    if signingIssuer == nil || tlsIssuer == nil { return false }
    return signingIssuer.CertType() == tlsIssuer.CertType() && 
           signingIssuer.CAID() == tlsIssuer.CAID()
}
```

**Used in**: `TLSCommonName()`, `AdminTLSCommonName()`, `BootstrapAdminTLSCommonName()`

**Opportunity**: Extract to shared helper function.

#### VaultRef Pattern
All Vault refs follow: `VaultBasePath() + "/subpath"`

```go
VaultSigningRef()                    → basePath + "/signing"
VaultSigningChainRef()               → basePath + "/signing-chain"
VaultTLSRef()                        → basePath + "/tls"
VaultDatabaseUserRef()               → basePath + "/certstore/user"
VaultBootstrapAdminCredentialsRef()  → basePath + "/bootstrap-admin/credentials"
VaultBootstrapAdminSigningRef()      → basePath + "/bootstrap-admin/signing"
VaultBootstrapAdminTLSRef()          → basePath + "/bootstrap-admin/tls"
VaultAdminSigningRef()               → basePath + "/admin/signing"
VaultAdminTLSRef()                   → basePath + "/admin/tls"
```

**Observation**: Could be simplified to `VaultRef(subpath string)` or use constants.

---

### 2. String-Based Differentiation

Current code uses magic strings throughout:

```go
certType = "tls" | "org"           // Could be enum: CertType{TLS, Org}
caID = "root" | "v1" | ...         // Could be typed: CAID{Root, V1}
profile = "ca" | "tls" | ""        // Could be enum: SigningProfile
identityType = "peer" | "orderer"  // Could be enum: IdentityType
```

**Risk**: Typos cause runtime errors, not compile errors.

**Hard-coded constant**:
```go
const tlsIntermediateCAID = "v1"  // Used in TLSIssuer(), AdminTLSIssuer()
```

**Observation**: This assumes there's only one intermediate CA version. If multiple versions exist, this becomes problematic.

---

### 3. Method Explosion

CertificateAuthorityConfig has **34 methods**:

| Category | Count | Methods |
|----------|-------|---------|
| Type checks | 2 | IsRoot, IsTLS |
| Issuer resolution | 4 | SigningIssuer, TLSIssuer, AdminSigningIssuer, AdminTLSIssuer |
| Common names | 6 | ComponentName, TLSCommonName, AdminCommonName, AdminTLSCommonName, BootstrapAdminCommonName, BootstrapAdminTLSCommonName |
| Vault refs | 10 | VaultBasePath, VaultSigningRef, VaultSigningChainRef, VaultTLSRef, VaultDatabaseUserRef, VaultBootstrapAdmin*, VaultAdmin* |
| Kubernetes | 5 | ServiceAccountName, KubernetesNamespace, URL, ResourceIdentifier, TLSDNSNames |
| Database | 2 | DatabaseName, DatabaseUsername |
| Subject | 4 | CertType, CAID, OrganizationalUnit, Description |
| Vault policy | 4 | PolicyName, AuthRoleName, VaultDataPath, VaultMetadataPath, VaultPolicy |

**Compare to other configs**:
- CertificateStoreConfig: 12 methods
- CouchDbConfig: 14 methods
- PeerConfig: 22 methods
- OrdererConfig: 22 methods

---

### 4. Separation of Concerns

The `CertificateAuthorityConfig` mixes:

1. **Data storage** (certType, caID, OrgConfig)
2. **Path generation** (Vault paths, Kubernetes names)
3. **PKI hierarchy knowledge** (issuer resolution)
4. **Collision avoidance logic** (common name suffixes)
5. **Policy generation** (VaultPolicy())

**Principle violation**: Single Responsibility - the config knows too much about the PKI hierarchy.

---

### 5. Caller Duplication

#### DeployPeer vs DeployOrderer
These workflows are 95% identical:

```go
// Both do:
// 1. Create Vault access (same pattern)
// 2. Create identity with CreateIdentity (same pattern)
// 3. Deploy with CreateDeployment (same pattern)
```

Only differences: `peerID` vs `ordererID`, port numbers, OU value.

**Opportunity**: Could share implementation or use common abstraction.

---

### 6. Inconsistencies

| Aspect | CertificateStoreConfig | Others |
|--------|------------------------|--------|
| VaultBasePath() | No (inline in VaultAdminRef) | Yes |
| Path pattern | `certstore/admin` | `{component}/{subpath}` |

| Aspect | CertificateAuthorityConfig | PeerConfig/OrdererConfig |
|--------|---------------------------|--------------------------|
| Issuer resolution | Dynamic methods | Hardcoded "v1" strings |

---

### 7. Identified Opportunities

#### A. Extract Collision Avoidance Helper
```go
// Before: repeated in 3 places
func (ca) AdminTLSCommonName() string {
    signingIssuer := ca.AdminSigningIssuer()
    tlsIssuer := ca.AdminTLSIssuer()
    if signingIssuer.CertType() == tlsIssuer.CertType() && 
       signingIssuer.CAID() == tlsIssuer.CAID() {
        return fmt.Sprintf("admin-tls@%s", ca.ComponentName())
    }
    return ca.AdminCommonName()
}

// After: single helper
func (ca) needsTLSSuffix(signingIssuer, tlsIssuer CertificateAuthorityConfig) bool {
    return signingIssuer.CertType() == tlsIssuer.CertType() && 
           signingIssuer.CAID() == tlsIssuer.CAID()
}
```

#### B. Type Safety for certType/caID
```go
type CertChain string
const (
    CertChainTLS CertChain = "tls"
    CertChainOrg CertChain = "org"
)

type CALevel string
const (
    CALevelRoot         CALevel = "root"
    CALevelIntermediate CALevel = "v1"
)
```

#### C. Extract Vault Access Interface
```go
type VaultAccessConfig interface {
    PolicyName() string
    VaultPolicy() *vaultpb.PolicyDocument
    AuthRoleName() string
    ServiceAccountName() string
    KubernetesNamespace() string
}
```

#### D. Unify VaultRef Generation
```go
func (ca) VaultRef(subpath string) *pb.SecretRef {
    return secrets.NewVaultRef(ca.VaultBasePath() + "/" + subpath)
}

// Usage:
ca.VaultRef("signing")           // replaces VaultSigningRef()
ca.VaultRef("admin/tls")         // replaces VaultAdminTLSRef()
```

---

## Summary: What's Working Well

1. **Consistent path patterns** - All Vault paths follow predictable structure
2. **Clear separation of component configs** - Factory methods on FabricConfig
3. **Self-documenting method names** - Easy to understand what each returns
4. **Comments explain the "why"** - Especially for issuer resolution

## Summary: Areas for Improvement

1. **Collision avoidance logic is duplicated** - 3 places with same pattern
2. **String-based types** - "tls"/"org" could be compile-time checked
3. **PKI hierarchy embedded in config** - Could be extracted
4. **Peer/Orderer nearly identical** - Could share more code
5. **Inconsistent VaultBasePath** - CertificateStoreConfig differs from others

---

## Refactoring Options

Based on the analysis above, here are three options ranging from minimal (low risk) to comprehensive (higher reward).

---

### Option A: Minimal - Extract Collision Avoidance Helper (Low Risk)

**Scope**: CertificateAuthorityConfig only, internal refactor

**Changes**:
1. Extract `needsTLSSuffix()` private helper function
2. Refactor `TLSCommonName()`, `AdminTLSCommonName()`, `BootstrapAdminTLSCommonName()` to use it

**Before** (3 places with same logic):
```go
func (ca CertificateAuthorityConfig) TLSCommonName() string {
    signingIssuer := ca.SigningIssuer()
    tlsIssuer := ca.TLSIssuer()
    if signingIssuer == nil || tlsIssuer == nil {
        return ca.ComponentName()
    }
    if signingIssuer.CertType() == tlsIssuer.CertType() && 
       signingIssuer.CAID() == tlsIssuer.CAID() {
        return fmt.Sprintf("%s-tls", ca.ComponentName())
    }
    return ca.ComponentName()
}
```

**After**:
```go
// sameIssuer returns true if both issuers refer to the same CA.
// Returns false if either issuer is nil (self-signed case).
func sameIssuer(a, b *CertificateAuthorityConfig) bool {
    if a == nil || b == nil {
        return false
    }
    return a.CertType() == b.CertType() && a.CAID() == b.CAID()
}

func (ca CertificateAuthorityConfig) TLSCommonName() string {
    if sameIssuer(ca.SigningIssuer(), ca.TLSIssuer()) {
        return fmt.Sprintf("%s-tls", ca.ComponentName())
    }
    return ca.ComponentName()
}

func (ca CertificateAuthorityConfig) AdminTLSCommonName() string {
    signingIssuer := ca.AdminSigningIssuer()
    tlsIssuer := ca.AdminTLSIssuer()
    if sameIssuer(&signingIssuer, &tlsIssuer) {
        return fmt.Sprintf("admin-tls@%s", ca.ComponentName())
    }
    return ca.AdminCommonName()
}
```

**Pros**:
- Minimal change, easy to verify
- DRY principle - collision logic in one place
- No API changes, no caller impact
- Easy to understand the pattern

**Cons**:
- Doesn't address other issues (type safety, method explosion)
- Small improvement

**Effort**: Low (1-2 hours)

**Verification**: Existing tests should pass unchanged

---

### Option B: Moderate - Add Type Safety + Extract Helper (Medium Risk)

**Scope**: CertificateAuthorityConfig primary, minor changes to FabricConfig factory

**Changes**:
1. Include Option A (extract collision helper)
2. Add typed constants for `certType` and `caID`
3. Update factory method signature to use types
4. Keep string-based storage for backward compatibility

**New types**:
```go
// CertChain represents the certificate chain type (TLS or Org)
type CertChain string

const (
    CertChainTLS CertChain = "tls"
    CertChainOrg CertChain = "org"
)

// CALevel represents the CA hierarchy level
type CALevel string

const (
    CALevelRoot         CALevel = "root"
    CALevelIntermediate CALevel = "v1"
)

// IsIntermediate returns true if this is an intermediate CA level
func (l CALevel) IsIntermediate() bool {
    return l != CALevelRoot
}
```

**Updated factory**:
```go
// CertificateAuthority returns the configuration for a specific certificate authority.
func (fc FabricConfig) CertificateAuthority(chain CertChain, level CALevel) CertificateAuthorityConfig {
    return CertificateAuthorityConfig{
        OrgConfig: fc.OrgConfig,
        certType:  string(chain),
        caID:      string(level),
    }
}
```

**Caller changes** (minimal - replace strings with constants):
```go
// Before
caConfig := fabricConfig.CertificateAuthority("tls", "root")

// After
caConfig := fabricConfig.CertificateAuthority(CertChainTLS, CALevelRoot)
```

**Pros**:
- Compile-time checking of cert type and CA level
- Typos become compile errors instead of runtime errors
- Documents valid values in code
- Backward compatible (internal storage unchanged)

**Cons**:
- Requires updating all callers (workflows.go, activities.go)
- More changes to review
- Still doesn't address method explosion

**Effort**: Medium (3-4 hours)

**Verification**: 
- All existing tests should pass
- Compiler catches any missed call sites

---

### Option C: Comprehensive - Separate Issuer Resolution (Higher Risk)

**Scope**: Major restructure of CertificateAuthorityConfig

**Changes**:
1. Include Options A and B
2. Extract issuer resolution to a separate `PKIHierarchy` type
3. Make `CertificateAuthorityConfig` purely data-driven
4. Reduce method count through consolidation

**New structure**:
```go
// PKIHierarchy encodes the certificate authority hierarchy rules.
// This is separate from CertificateAuthorityConfig to separate
// data (config) from business logic (PKI rules).
type PKIHierarchy struct {
    tlsIntermediateID CALevel // configurable, defaults to "v1"
}

// SigningIssuer returns who signs this CA's signing certificate.
func (h PKIHierarchy) SigningIssuer(ca CertificateAuthorityConfig) *CertificateAuthorityConfig {
    if ca.IsRoot() {
        return nil
    }
    issuer := ca.withLevel(CALevelRoot)
    return &issuer
}

// TLSIssuer returns who signs this CA's TLS certificate.
func (h PKIHierarchy) TLSIssuer(ca CertificateAuthorityConfig) *CertificateAuthorityConfig {
    switch {
    case ca.IsTLS() && ca.IsRoot():
        return nil
    case ca.IsTLS():
        issuer := ca.withLevel(CALevelRoot)
        return &issuer
    default:
        issuer := ca.withChainAndLevel(CertChainTLS, h.tlsIntermediateID)
        return &issuer
    }
}
```

**Consolidated Vault methods**:
```go
// VaultRef returns a SecretRef for the given subpath under this CA's base path.
func (ca CertificateAuthorityConfig) VaultRef(subpath string) *pb.SecretRef {
    return secrets.NewVaultRef(ca.VaultBasePath() + "/" + subpath)
}

// Common subpaths as constants for discoverability
const (
    VaultSubpathSigning                  = "signing"
    VaultSubpathSigningChain             = "signing-chain"
    VaultSubpathTLS                      = "tls"
    VaultSubpathDatabaseUser             = "certstore/user"
    VaultSubpathBootstrapAdminCredentials = "bootstrap-admin/credentials"
    VaultSubpathBootstrapAdminSigning    = "bootstrap-admin/signing"
    VaultSubpathBootstrapAdminTLS        = "bootstrap-admin/tls"
    VaultSubpathAdminSigning             = "admin/signing"
    VaultSubpathAdminTLS                 = "admin/tls"
)

// Usage: ca.VaultRef(VaultSubpathSigning) replaces ca.VaultSigningRef()
```

**Pros**:
- Clear separation of data vs business logic
- PKI hierarchy is explicit and testable independently
- Reduced method count (10 VaultRef methods → 1 + constants)
- Easier to understand and modify PKI rules
- Future-proof: can have different PKI hierarchies

**Cons**:
- Significant refactor with many file changes
- Higher risk of introducing bugs
- Callers need updates (either import change or method rename)
- Requires careful review of all issuer resolution paths
- May need to deprecate old methods during transition

**Effort**: High (8-12 hours)

**Verification**:
- Comprehensive unit tests for PKIHierarchy
- Existing workflow tests must pass
- Manual verification of generated certificate chains

---

## Recommendation

**Start with Option A** (extract collision helper). It's low-risk, addresses a real DRY violation, and can be done quickly. This gives us:
1. Immediate improvement
2. Confidence in the refactoring approach
3. A foundation for further changes

**Consider Option B** as a follow-up if Option A goes well. The type safety is valuable for preventing bugs, and the changes are mechanical (compiler-guided).

**Defer Option C** unless there's a specific need (e.g., multiple intermediate CA versions, need to test PKI hierarchy separately). The current structure works, and Option C is significant effort for moderate gain.

---

## Next Steps (after user decision)

1. User selects option (A, B, C, or none)
2. Create characterization tests if needed
3. Execute refactor in small commits
4. Verify with existing tests
5. Delete this REFACTOR_NOTES.md file
