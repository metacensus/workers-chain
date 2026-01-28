# MetaCensus Fabric Certificate Architecture

## Overview

MetaCensus uses a **dual-chain PKI architecture** for Hyperledger Fabric deployments:
- **TLS Chain**: Issues certificates for secure communications (HTTPS, mTLS)
- **Org Chain**: Issues certificates for Fabric identities (signing, endorsement)

Both chains use a **Root CA → Intermediate CA** hierarchy for security and operational flexibility.

---

## Certificate Authority Hierarchy

```
TLS Chain                              Org Chain
─────────                              ─────────
TLS Root CA (self-signed)              Org Root CA (self-signed)
    │                                      │
    └── TLS Int CA                         └── Org Int CA
            │                                      │
            ├── CA Server TLS certs                ├── Peer identities
            ├── Admin client TLS certs             ├── Orderer identities  
            ├── Peer/Orderer TLS certs             ├── Admin identities
            └── Chaincode TLS certs                └── Client identities
```

**Key Design Decision**: Org CAs get their TLS certificates from the TLS chain, not their own chain. This provides clean separation between identity signing and secure communications.

---

## Certificate Types Per Entity

Each CA requires **two** certificates with different purposes:
1. **Signing Certificate** (leaf only): For issuing other certificates (no SANs required)
2. **TLS Certificate** (with chain): For HTTPS endpoint (hostname verification)

Each identity (admin, peer, orderer) requires **two** certificates:
1. **Identity Certificate** (with chain): For Fabric operations (signing, authentication)
2. **TLS Certificate** (with chain): For mTLS communications

---

## Certificate Matrix

### TLS Infrastructure

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| TLS Root CA | CA Signing | Self-signed | Offline after bootstrap |
| TLS Root CA | Server TLS | Self-signed | For CA's HTTPS endpoint |
| TLS Int CA | CA Signing | TLS Root CA | Signs all TLS certs |
| TLS Int CA | Server TLS | TLS Root CA | For CA's HTTPS endpoint |

### TLS Certificate Authority Admins

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| TLS Root CA Admin | Identity | TLS Root CA | Admin privs on TLS Root CA |
| TLS Root CA Admin | Client TLS | TLS Root CA | mTLS to TLS Root CA |
| TLS Int CA Admin | Identity | TLS Int CA | Admin privs on TLS Int CA (must be in CA's database) |
| TLS Int CA Admin | Client TLS | TLS Int CA | mTLS to TLS Int CA |

### Organization Infrastructure

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| Org Root CA | CA Signing | Self-signed | Offline after bootstrap |
| Org Root CA | Server TLS | TLS Int CA | For CA's HTTPS endpoint |
| Org Int CA | CA Signing | Org Root CA | Signs all identity certs |
| Org Int CA | Server TLS | TLS Int CA | For CA's HTTPS endpoint |

### Organization Certificate Authority Admins

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| Org Root CA Admin | Identity | Org Root CA | Admin privs on Org Root CA (must be in CA's database) |
| Org Root CA Admin | Client TLS | TLS Int CA | mTLS to Org Root CA |
| Org Int CA Admin | Identity | Org Int CA | Admin privs on Org Int CA (must be in CA's database) |
| Org Int CA Admin | Client TLS | TLS Int CA | mTLS to Org Int CA |

### Organization Entities

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| Orderer | Identity | Org Int CA | MSP signing |
| Orderer | Server/Client TLS | TLS Int CA | mTLS with peers and orderers |
| Peer | Identity | Org Int CA | Endorsement signing |
| Peer | Server/Client TLS | TLS Int CA | mTLS with orderers and clients |
| Org Admin | Identity | Org Int CA | Channel/chaincode ops |
| Org Admin | Client TLS | TLS Int CA | mTLS to peers and orderers |
| Client App | Identity | Org Int CA | Transaction signing |
| Client App | Client TLS | TLS Int CA | mTLS to peers and orderers |
| Chaincode | Server TLS | TLS Int CA | Peer-to-chaincode mTLS |

---

## Certificate Chain Construction

When enrolling an identity, certificate chains are built and stored for later use in mTLS and MSP verification.

### Chain Assembly at Enrollment

1. **Generate keypair** → stored in `keys/{entity-path}/signing` or `keys/{entity-path}/tls`
2. **Submit CSR to issuing CA** → CA signs and returns leaf certificate
3. **Build full chain**:
   - Fetch issuing CA's chain from `certs/ca/{chain}/{id}/signing` (`chain` key)
   - Concatenate: `[leaf cert] + [issuing CA's chain]`
4. **Store results** at single path with two keys:
   - `certificate` key → leaf certificate only
   - `chain` key → full chain (leaf + intermediate(s) + root)

### Chain Content by Certificate Type

| Certificate Type | `chain` Key Contents | Usage |
|-----------------|----------------------|-------|
| TLS (entity) | leaf + TLS Int CA + TLS Root CA | Sent during mTLS handshake |
| Identity (entity) | leaf + Org Int CA + Org Root CA | MSP verification |
| CA Signing (int) | leaf + parent chain | CA operations, client cert verification |
| CA Signing (root) | self-signed leaf only | CA operations |

### CA Chain Files

Intermediate CAs require a `chainfile` for x509 verification of client certificates:

| CA Type | certfile (from `certificate` key) | chainfile (from `chain` key) |
|---------|-----------------------------------|------------------------------|
| TLS Root CA | TLS Root CA (self-signed) | N/A (same as certfile) |
| TLS Int CA | TLS Int CA (leaf only) | TLS Int CA + TLS Root CA |
| Org Root CA | Org Root CA (self-signed) | N/A (same as certfile) |
| Org Int CA | Org Int CA (leaf only) | Org Int CA + Org Root CA |

**Note**: Intermediate CAs include their own cert in the chainfile to verify certificates they issue.

---

## mTLS Certificate Exchange

### Trust Store Configuration

Trust stores contain **only root CA certificates**:
- TLS trust store: TLS Root CA certificate only
- MSP cacerts: Org Root CA certificate only

Intermediate CA certificates are **never** placed in trust stores.

### Handshake Flow

During mutual TLS authentication:

1. **Server sends**: leaf certificate + intermediate CA cert(s)
2. **Client sends**: leaf certificate + intermediate CA cert(s)
3. **Both validate**: received chain terminates at a trusted root in their trust store

```
┌────────┐                              ┌────────┐
│ Client │                              │ Server │
└───┬────┘                              └───┬────┘
    │                                       │
    │  ←── Server cert chain ───────────────┤  (leaf + intermediate)
    │                                       │
    │  ─── Client cert chain ──────────────→│  (leaf + intermediate)
    │                                       │
    ▼                                       ▼
Validate against                    Validate against
TLS Root CA in                      TLS Root CA in
trust store                         trust store
```

### Why Root-Only Trust Stores

- **Simplified management**: Only root CAs in trust stores, which rarely change
- **Intermediate rotation**: Can rotate intermediate CAs without updating trust stores
- **Standard behavior**: Expected by all compliant TLS implementations

---

## Vault Storage Structure

Certificates and keys are stored in two parallel hierarchies:

| Hierarchy | Contents | Access |
|-----------|----------|--------|
| `certs/` | Public certificates | All identities can read |
| `keys/` | Private keys and credentials | Per-identity only |

### Certificate Paths (`certs/`)

Each path contains `certificate` (leaf) and `chain` (leaf + issuer chain to root).

```
secret/fabric/{org}/certs/
├── ca/{chain}/{id}/
│   ├── signing                         # {certificate, chain}
│   ├── tls                             # {certificate, chain}
│   ├── bootstrap-admin/
│   │   ├── signing                     # {certificate, chain}
│   │   └── tls                         # {certificate, chain}
│   └── admin/
│       ├── signing                     # {certificate, chain}
│       └── tls                         # {certificate, chain}
│
├── peer/{id}/
│   ├── signing                         # {certificate, chain}
│   └── tls                             # {certificate, chain}
│
├── orderer/{id}/
│   ├── signing                         # {certificate, chain}
│   └── tls                             # {certificate, chain}
│
├── admin/{id}/
│   ├── signing                         # {certificate, chain}
│   └── tls                             # {certificate, chain}
│
├── client/{id}/
│   ├── signing                         # {certificate, chain}
│   └── tls                             # {certificate, chain}
│
└── chaincode/{id}/
    └── tls                             # {certificate, chain}
```

Where `{chain}` is `tls` or `org`, and `{id}` is the entity identifier.

### Key Paths (`keys/`)

Mirrors `certs/` structure. Each path contains `private_key`. Additional credential paths:

| Path | Contents |
|------|----------|
| `keys/ca/{chain}/{id}/certstore` | `{username, password}` for CA database |
| `keys/ca/{chain}/{id}/bootstrap-admin/credentials` | `{username, password}` for bootstrap |
| `keys/peer/{id}/couchdb` | `{username, password}` for CouchDB |

### Path Examples

| Description | Path | Keys |
|-------------|------|------|
| TLS Root CA signing key | `secret/fabric/org1/keys/ca/tls/root/signing` | `{private_key}` |
| TLS Root CA signing cert | `secret/fabric/org1/certs/ca/tls/root/signing` | `{certificate, chain}` |
| Org Int CA TLS cert | `secret/fabric/org1/certs/ca/org/int/tls` | `{certificate, chain}` |
| Peer0 TLS | `secret/fabric/org1/certs/peer/peer0/tls` | `{certificate, chain}` |
| Peer0 TLS key | `secret/fabric/org1/keys/peer/peer0/tls` | `{private_key}` |
| Bootstrap admin credentials | `secret/fabric/org1/keys/ca/tls/root/bootstrap-admin/credentials` | `{username, password}` |
| CouchDB credentials | `secret/fabric/org1/keys/peer/peer0/couchdb` | `{username, password}` |

---

## Security Boundaries

### Two-Path Access Model

Each identity has access to two conceptual areas:
- **Public** (`certs/*`): All certificates - open read access
- **Private** (`keys/{entity-path}/*`): Own keys and credentials only

| Entity | Public Access | Private Access |
|--------|---------------|----------------|
| CA | `certs/*` (all) | `keys/ca/{chain}/{id}/*` |
| CA Bootstrap Admin | `certs/*` (all) | `keys/ca/{chain}/{id}/bootstrap-admin/*` |
| CA Admin | `certs/*` (all) | `keys/ca/{chain}/{id}/admin/*` |
| Peer | `certs/*` (all) | `keys/peer/{id}/*` |
| Orderer | `certs/*` (all) | `keys/orderer/{id}/*` |

**Critical**: Certificates are public (transmitted during TLS handshakes). Private keys are strictly per-identity.

### Per-Identity Least-Privilege

Each identity gets its own scoped Vault policy:
- Policy grants read to all `certs/*` (certificates are public)
- Policy grants read to only its own `keys/{entity-path}/*`
- One compromised identity cannot access others' private keys

### Worker Isolation

| Worker | Temporal Namespace | Vault Access |
|--------|-------------------|--------------|
| Central Worker | `infra` | Full access (creates policies) |
| Fabric Worker (Org1) | `fabric-org1` | `secret/fabric/org1/*` only |
| Fabric Worker (Org2) | `fabric-org2` | `secret/fabric/org2/*` only |

---

## Bootstrap Order

The TLS chain must complete before the Org chain because Org CAs need TLS Int CA for their server TLS certificates.

```
Phase 1: TLS Chain
├── Deploy TLS Root CA (self-signed signing + self-signed TLS)
│   └── Create Bootstrap Admin → Create Permanent Admin
├── Deploy TLS Int CA (signing from TLS Root, TLS from TLS Root)
│   └── Create Bootstrap Admin → Create Permanent Admin

Phase 2: Org Chain (requires TLS Int CA)
├── Deploy Org Root CA (self-signed signing, TLS from TLS Int CA)
│   └── Create Bootstrap Admin → Create Permanent Admin
├── Deploy Org Int CA (signing from Org Root, TLS from TLS Int CA)
│   └── Create Bootstrap Admin → Create Permanent Admin

Phase 3: Fabric Components
├── Deploy Peers (identity from Org Int, TLS from TLS Int)
├── Deploy Orderers (identity from Org Int, TLS from TLS Int)
└── Deploy Chaincode (TLS from TLS Int)
```

---

## Helm Chart Configuration

### Certificate Authority Values

```yaml
vault:
  # Keys (private, restricted access)
  signingKeyPath: "secret/data/fabric/{org}/keys/ca/{chain}/{id}/signing"
  tlsKeyPath: "secret/data/fabric/{org}/keys/ca/{chain}/{id}/tls"
  
  # Certificates (public) - each path contains {certificate, chain} keys
  signingCertPath: "secret/data/fabric/{org}/certs/ca/{chain}/{id}/signing"
  tlsCertPath: "secret/data/fabric/{org}/certs/ca/{chain}/{id}/tls"
  
  # Credentials (private)
  bootstrapAdminCredentialsPath: "secret/data/fabric/{org}/keys/ca/{chain}/{id}/bootstrap-admin/credentials"
  certstorePath: "secret/data/fabric/{org}/keys/ca/{chain}/{id}/certstore"
```

### Peer/Orderer Values

```yaml
vault:
  # Own keys (private)
  signingKeyPath: "secret/data/fabric/{org}/keys/peer/{id}/signing"
  tlsKeyPath: "secret/data/fabric/{org}/keys/peer/{id}/tls"
  couchdbPath: "secret/data/fabric/{org}/keys/peer/{id}/couchdb"
  
  # Own certificates (public) - each contains {certificate, chain} keys
  signingCertPath: "secret/data/fabric/{org}/certs/peer/{id}/signing"
  tlsCertPath: "secret/data/fabric/{org}/certs/peer/{id}/tls"
  
  # CA root certs for trust stores (public)
  mspRootCaCertPath: "secret/data/fabric/{org}/certs/ca/org/{root-id}/signing"
  tlsRootCaCertPath: "secret/data/fabric/{org}/certs/ca/tls/{root-id}/signing"
```

### Fabric CA Server Config

```yaml
ca:
  certfile: /mnt/secrets-store/signing-cert      # From signing path, certificate key
  keyfile: /mnt/secrets-store/signing-key        # From signing path, private_key key
  chainfile: /mnt/secrets-store/signing-chain    # From signing path, chain key (Int CAs only)

tls:
  enabled: true
  certfile: /mnt/secrets-store/tls-chain         # From tls path, chain key (full chain for transmission)
  keyfile: /mnt/secrets-store/tls-key            # From tls path, private_key key
```

---

## Admin Certificate Requirements

**Fabric CA Database Constraint**: Admin identity certificates must be issued by the CA they administer. Fabric CA validates admin privileges by checking its database.

| Admin | Identity Issued By | TLS Issued By |
|-------|-------------------|---------------|
| TLS Root CA Admin | TLS Root CA | TLS Root CA |
| TLS Int CA Admin | TLS Int CA | TLS Int CA |
| Org Root CA Admin | Org Root CA | TLS Int CA |
| Org Int CA Admin | Org Int CA | TLS Int CA |

### Avoiding Identity Registration Collisions

When an admin gets both identity and TLS certificates from the same CA (e.g., TLS Root CA Admin), append `-tls` suffix to TLS certificate common names:

| Admin | Identity CN | TLS CN |
|-------|-------------|--------|
| TLS Root CA Admin | admin@ca-tls-root | admin-tls@ca-tls-root |
| Org Root CA Admin | admin@ca-org-root | admin@ca-org-root |

---

## Fabric CA Database (Certificate Store)

Each organization has a **shared PostgreSQL database** (`certificate-store`) that stores:
- All certificate enrollment records
- Identity attributes and permissions
- Revocation status

The database is shared across all four CAs in an organization, but each CA has its own PostgreSQL user with scoped permissions.

---

## CSI Secret Delivery

Certificates are delivered via Vault CSI Provider:

1. **Pod Startup**: Kubelet's CSI plugin intercepts volume mount
2. **Authentication**: CSI driver authenticates pod's service account to Vault
3. **Secret Fetch**: Vault CSI Provider fetches secrets from Vault
4. **Mount**: Secrets are mounted as tmpfs into pod filesystem
5. **Application Start**: Container starts with certificates in place

---

## Implementation Notes

### Signing Profiles

The CA ConfigMap defines signing profiles:
- **Default Profile**: Client/identity certificates (1 year)
- **CA Profile**: Intermediate CA certificates (5 years, maxpathlen: 1)
- **TLS Profile**: Server/client TLS certificates (1 year)

### Permanent Admin Attributes

```
hf.Registrar.Roles: "*"
hf.Registrar.DelegateRoles: "*"
hf.Registrar.Attributes: "*"
hf.Revoker: true
hf.IntermediateCA: true
hf.GenCRL: true
hf.AffiliationMgr: true
```

### Cross-Chain Certificate Issuance

The TLS Int CA issues all non-root TLS certificates:
- Org CAs get server TLS from TLS Int CA
- Admin TLS certs come from TLS chain (except TLS Root Admin)
- Peer/Orderer TLS certs come from TLS Int CA

This allows rotating Org CA signing capabilities without affecting mTLS infrastructure.

---

## Related Documentation

- [cert-design-issues.md](cert-design-issues.md) - Known implementation issues
- [Fabric Certs.csv](Fabric%20Certs.csv) - Complete certificate matrix
- [temporal/README.md](../../temporal/README.md) - Workflow patterns

---

**Architecture**: Two-path model (`certs/` public, `keys/` private) with `certificate` and `chain` as KV keys at each path.
