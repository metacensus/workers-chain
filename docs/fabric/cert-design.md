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
2. **TLS Certificate** (with SANs): For HTTPS endpoint (hostname verification)

Each identity (admin, peer, orderer) requires **two** certificates:
1. **Identity Certificate**: For Fabric operations (signing, authentication) - includes full chain for verification
2. **TLS Certificate**: For mTLS communications - includes full chain for verification

**Key Implementation Detail**: CA signing certificates are stored as **leaf only** (no chain) because Fabric CA expects only the leaf cert in its signing config. Client certificates are stored with **full chains** for verification purposes.

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
| TLS Root CA Admin | Identity | TLS Root CA | TLS Root; admin privs |
| TLS Root CA Admin | Client TLS | TLS Root CA | TLS Root; admin privs |
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
| Orderer _ | Identity | Org Int CA | MSP signing |
| Orderer _ | Server/Client TLS | TLS Int CA | mTLS with peers and orderers |
| Peer _ | Identity | Org Int CA | Endorsement signing |
| Peer _ | Server/Client TLS | TLS Int CA | mTLS with orderers and clients |
| Org Admin | Identity | Org Int CA | Channel/chaincode ops |
| Org Admin | Client TLS | TLS Int CA | mTLS to peers and orderers |
| Client App | Identity | Org Int CA | Transaction signing |
| Client App | Client TLS | TLS Int CA | mTLS to peers and orderers |

### Other Identities

| Service/Entity | Purpose | Issued By | Notes |
|---|---|---|---|
| Chaincode | Server TLS | TLS Int CA | Peer-to-chaincode mTLS |

---

## Chain Storage Requirements

### For CAs (signing-chain)

Fabric CA uses a `chainfile` for x509 verification of client certificates. The chainfile content determines what certificates the CA trusts.

| CA Type | signing (certfile) | signing-chain (chainfile) |
|---------|-------------------|---------------------------|
| TLS Root CA | TLS Root CA (self-signed) | N/A (not used) |
| TLS Int CA | TLS Int CA (leaf only) | **TLS Int CA + TLS Root CA** |
| Org Root CA | Org Root CA (self-signed) | N/A (not used) |
| Org Int CA | Org Int CA (leaf only) | **Org Int CA + Org Root CA** |

**Critical**: Intermediate CAs must include their own cert in the chainfile to verify certificates they issue.

### For Client Certificates

Client certificates include the full chain for verification:

```
[Client cert]       ← Leaf (signed by issuing CA)
[Issuing CA cert]   ← For verification
[Root CA cert]      ← Trust anchor
```

---

## Admin Certificate Requirements

**Fabric CA Database Constraint**: Admin identity certificates must be issued by the CA they administer. Fabric CA validates admin privileges by checking its database, and the admin's certificate must exist in the CA's database where it was registered.

**Actual Implementation** (as of 2026-01-04): All admin identity certificates come from their own CA (the one they administer), ensuring they exist in that CA's database for privilege verification.

| Admin | Identity Issued By | TLS Issued By |
|-------|-------------------|---------------|
| TLS Root CA Admin | TLS Root CA | TLS Root CA |
| TLS Int CA Admin | TLS Int CA | TLS Int CA |
| Org Root CA Admin | Org Root CA | TLS Int CA |
| Org Int CA Admin | Org Int CA | TLS Int CA |

**Explanation**: 
- Identity certs must be issued by the CA itself (database requirement)
- TLS certs for non-root CAs come from TLS Int CA for security/operational separation
- This provides **certificate separation**: identity signing is distinct from communications security

---

## Vault Storage Structure

Certificates and credentials are organized by CA and identity type:

```
secret/fabric/{org}/ca/{chain}/{id}/
├── signing/
│   ├── private_key                 # CA's private key for signing
│   ├── public_key                  # CA's public key
│   └── certificate                 # CA's signing cert (leaf only, no chain)
│
├── signing-chain/                  # Intermediate CAs only
│   └── chain                       # Full chain: this CA's cert + parent chain
│                                   # Used for x509 verification
├── tls/
│   ├── private_key                 # CA's private key for TLS
│   ├── public_key                  # CA's public key
│   └── certificate                 # CA's TLS cert with SANs (with chain)
│
├── bootstrap-admin/
│   ├── credentials
│   │   ├── username
│   │   └── password                # For initial CA access
│   ├── signing/
│   │   ├── private_key
│   │   ├── public_key
│   │   └── certificate             # Identity cert + chain
│   └── tls/
│       ├── private_key
│       ├── public_key
│       └── certificate             # TLS client cert + chain
│
├── admin/                          # Permanent admin
│   ├── signing/
│   │   ├── private_key
│   │   ├── public_key
│   │   └── certificate             # Identity cert + chain
│   └── tls/
│       ├── private_key
│       ├── public_key
│       └── certificate             # TLS client cert + chain
│
└── certstore/                      # PostgreSQL database credentials
    ├── username
    └── password
```

**Path Examples**:
- TLS Root CA signing cert: `secret/fabric/org1/ca/tls/root/signing`
- Org Int CA signing chain: `secret/fabric/org1/ca/org/v1/signing-chain`
- Admin identity: `secret/fabric/org1/ca/tls/root/admin/signing`

---

## Fabric CA Database (Certificate Store)

Each organization has a **shared PostgreSQL database** (`certificate-store`) that stores:
- All certificate enrollment records
- Identity attributes and permissions
- Revocation status
- CA configuration

**Critical Design Point**: Fabric CA uses database lookups to verify admin privileges. When an admin enrolls an identity, the CA checks if the admin's certificate exists in its database with the required permissions. This is why admin identity certificates must be issued by and registered with their own CA.

The database is **shared across all four CAs** in an organization (TLS Root, TLS Int, Org Root, Org Int), but each CA has its own PostgreSQL user with scoped permissions.

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

The `certificate-authority` Helm chart supports all four CA types with conditional configuration.

### Values Template

The values must provide:

```yaml
vault:
  signingPath: "secret/data/fabric/{org}/ca/{type}/{id}/signing"
                # Path to CA's signing keypair + certificate
  
  signingChainPath: "secret/data/fabric/{org}/ca/{type}/{id}/signing-chain"
                   # Path to chain (intermediate CAs only, omit for root CAs)
                   # Enables intermediate CAs to verify client certificate chains
  
  tlsPath: "secret/data/fabric/{org}/ca/{type}/{id}/tls"
           # Path to CA's TLS keypair + certificate
  
  adminCredentialsPath: "secret/data/fabric/{org}/ca/{type}/{id}/bootstrap-admin/credentials"
                       # Bootstrap admin username/password
  
  dbCredentialsPath: "secret/data/fabric/{org}/ca/{type}/{id}/certstore/user"
                    # Database user credentials
```

### Fabric CA Server Config

The ConfigMap generates `fabric-ca-server-config.yaml` with:

```yaml
ca:
  certfile: /mnt/secrets-store/signing-cert
           # CA signing certificate (leaf only, no chain)
  keyfile: /mnt/secrets-store/signing-private-key
  {{- if certificate_is_intermediate }}
  chainfile: /mnt/secrets-store/signing-chain
            # For intermediate CAs: self cert + parent chain for x509 verification
  {{- end }}

tls:
  enabled: true
  certfile: /mnt/secrets-store/tls-cert
           # Separate TLS certificate with SANs for hostname verification
  keyfile: /mnt/secrets-store/tls-private-key
```

**Key Implementation Detail**: Fabric CA stores signing certificates as **leaf only** (no chain), but intermediate CAs require a `chainfile` containing their own cert plus the parent chain for x509 verification of client certificates.

---

## Certificate Lifecycle & Admin Creation

### Bootstrap Admin → Permanent Admin Transition

Each CA follows a two-phase admin creation process:

**Phase 1: Bootstrap Admin** (immediate, username/password only)
- Created during CA deployment
- Used only to create the permanent admin
- Credentials stored in Vault

**Phase 2: Permanent Admin** (certificate-based)
- Created using bootstrap admin credentials
- Gets two certificates:
  1. **Identity Certificate** - issued by the CA itself for registration/privileges
  2. **TLS Certificate** - issued by TLS Int CA (or TLS Root for root CAs) for mTLS
- Both certificates stored in Vault
- Can be used for all admin operations going forward

### Certificate Generation Workflow

For each identity certificate:

1. **Key Generation** → Pre-generate keypair, store in Vault
2. **Registration** → Register identity name with issuing CA via bootstrap admin
3. **Enrollment** → Enroll with pre-generated key to get signed certificate
4. **Storage** → Certificate stored with keypair (creates complete identity)

### Avoiding Identity Registration Collisions

**Problem**: When an admin gets both identity and TLS certificates from the same CA (e.g., TLS Root CA Admin), registering both with the same name would cause a collision.

**Solution**: Append `-tls` suffix to TLS certificate common names when issued by the same CA:

| Admin | Identity CN | TLS CN | Issued By |
|-------|-------------|--------|-----------|
| TLS Root CA Admin | admin@ca-tls-root | admin-tls@ca-tls-root | Both from TLS Root CA |
| Org Root CA Admin | admin@ca-org-root | admin@ca-org-root | From different CAs (Org Root, TLS Int) |

This ensures each identity name is unique within its issuing CA.

---

## CSI Secret Delivery

Certificates are delivered via Vault CSI Provider:

1. **Pod Startup**: Kubelet's CSI plugin intercepts volume mount
2. **Authentication**: CSI driver authenticates pod's service account to Vault
3. **Secret Fetch**: Vault CSI Provider fetches secrets from Vault
4. **Mount**: Secrets are mounted as tmpfs into pod filesystem
5. **Application Start**: Container starts with certificates in place

**Benefits**: No Kubernetes Secrets, no bootstrap volumes, atomic delivery at pod startup.

---

## Security Boundaries

### Per-Identity Least-Privilege

Each identity (peer, orderer, admin) gets its own scoped Vault policy:
- Peer can only read its own certificates
- Orderer can only read its own certificates
- One compromised identity cannot access others' secrets

### Worker Isolation

| Worker | Temporal Namespace | Vault Access |
|--------|-------------------|--------------|
| Central Worker | `infra` | God-mode (creates policies) |
| Fabric Worker (Org1) | `fabric-org1` | `secret/fabric/org1/*` only |
| Fabric Worker (Org2) | `fabric-org2` | `secret/fabric/org2/*` only |

---

## Related Documentation

- [`Fabric Certs.csv`](Fabric%20Certs.csv) - Complete certificate matrix
- [`DEPLOYMENT_WORKFLOW_EXECUTION_TREE.md`](DEPLOYMENT_WORKFLOW_EXECUTION_TREE.md) - Workflow execution details
- [`Fabric Bootstrap Workflows.md`](Fabric%20Bootstrap%20Workflows.md) - Implementation status
- [temporal/README.md](../temporal/README.md) - Workflow patterns

---

## Implementation Notes

### Credential Injection in Helm Chart

The certificate-authority Helm chart uses **sed-based placeholder substitution** for bootstrap admin credentials in `fabric-ca-server-config.yaml` rather than Kubernetes environment variable overrides. This is because Fabric CA's Viper configuration doesn't support environment variable syntax for array elements like `registry.identities[0].name`.

### Signing Profiles

The CA ConfigMap defines signing profiles for certificate issuance:

- **Default Profile**: Used for client/identity certificates (1 year)
- **CA Profile**: Used for intermediate CA certificates (5 years, maxpathlen: 1)
- **TLS Profile**: Used for server/client TLS certificates (1 year, with server auth)

### Permanent Admin Attributes

When creating the permanent admin certificate, the following Fabric CA attributes are granted:

```
hf.Registrar.Roles: "*"             # Can register any identity type
hf.Registrar.DelegateRoles: "*"     # Can delegate any role
hf.Registrar.Attributes: "*"        # Can assign any attribute (critical for hf.IntermediateCA)
hf.Revoker: true                    # Can revoke certificates
hf.IntermediateCA: true             # Can be used to create intermediate CA certificates
hf.GenCRL: true                     # Can generate Certificate Revocation Lists
hf.AffiliationMgr: true             # Can manage affiliations
```

These attributes enable the admin to bootstrap peers and orderers.

### Cross-Chain Certificate Issuance

The architecture uses the TLS Int CA for all non-root TLS certificates to provide operational separation:

- **Org CAs don't issue their own TLS certs** - they get server TLS from TLS Int CA
- **Admin TLS certs come from TLS chain** - except TLS Root Admin (self-signed)
- **Peer/Orderer TLS certs come from TLS Int CA** - consistent with CAs

This separation allows rotating Org CA signing capabilities without affecting mTLS infrastructure.

---

**Last Updated:** 2026-01-06  
**Architecture Version:** 3.1 (Updated with implementation details and Fabric CA database constraints)
