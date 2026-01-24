# Certificate Architecture - Implementation Issues

Current implementation deviations from [cert-design.md](cert-design.md). This document serves as the guide for fixing these issues.

---

## Core Problem: Certificate Access Implies Key Access

The fundamental issue is that **accessing any certificate also grants access to the corresponding private key**. Certificates and private keys are stored at the same Vault path.

### Why This Matters

All services need to trust root CA certificates from both chains (TLS and Org). To have those certificates available for trust stores:

- Current: Services get read access to CA paths → exposes CA private keys
- Required: Services read certificates without any access to private keys

This isn't just a peer/orderer issue. It affects **every component** that needs CA certificates for trust validation.

### The Solution

Complete separation into two parallel hierarchies:

```
secret/fabric/{org}/
├── certs/                   # PUBLIC - All identities can read
│   └── ...                  # Certificates with {certificate, chain} keys
│
└── keys/                    # PRIVATE - Per-identity access only
    └── ...                  # Private keys and credentials
```

**Access pattern**:
- `certs/*` - All Fabric roles can read (certificates are public)
- `keys/{entity-path}/*` - Each identity can only read its own private data

**Storage pattern**:
- Each certificate path contains `{certificate, chain}` keys at the same Vault KV path
- Each key path contains `{private_key}` or credentials as appropriate

---

## Current Vault Path Structure

```
secret/fabric/{org}/
├── ca/{chain}/{id}/
│   ├── signing/
│   │   ├── private_key      # CA signing key
│   │   ├── public_key
│   │   └── certificate      # CA signing cert
│   ├── signing-chain/       # Int CAs only
│   │   └── chain
│   ├── tls/
│   │   ├── private_key      # CA TLS key
│   │   ├── public_key
│   │   └── certificate      # CA TLS cert
│   ├── bootstrap-admin/
│   │   ├── credentials/
│   │   ├── signing/         # Contains private_key + certificate
│   │   └── tls/             # Contains private_key + certificate
│   ├── admin/
│   │   ├── signing/         # Contains private_key + certificate
│   │   └── tls/             # Contains private_key + certificate
│   └── certstore/
│       └── user             # DB credentials
│
├── peer/{id}/
│   ├── signing/             # Contains private_key + certificate
│   └── tls/                 # Contains private_key + certificate
│
├── orderer/{id}/
│   ├── signing/             # Contains private_key + certificate
│   └── tls/                 # Contains private_key + certificate
│
├── couchdb/{id}/
│   └── admin                # Credentials only
│
└── certstore/
    └── admin                # Credentials only
```

**Problem**: Every path containing a `certificate` also contains the `private_key`.

---

## Target Vault Path Structure

Two parallel hierarchies with certificate and chain stored as keys at the same path:

### Certificate Paths (`certs/`) - PUBLIC

All identities can read. Each path contains `{certificate, chain}` keys.

```
secret/fabric/{org}/certs/
├── ca/{chain}/{id}/
│   ├── signing                              # {certificate, chain}
│   ├── tls                                  # {certificate, chain}
│   ├── bootstrap-admin/
│   │   ├── signing                          # {certificate, chain}
│   │   └── tls                              # {certificate, chain}
│   └── admin/
│       ├── signing                          # {certificate, chain}
│       └── tls                              # {certificate, chain}
│
├── peer/{id}/
│   ├── signing                              # {certificate, chain}
│   └── tls                                  # {certificate, chain}
│
├── orderer/{id}/
│   ├── signing                              # {certificate, chain}
│   └── tls                                  # {certificate, chain}
│
├── admin/{id}/
│   ├── signing                              # {certificate, chain}
│   └── tls                                  # {certificate, chain}
│
├── client/{id}/
│   ├── signing                              # {certificate, chain}
│   └── tls                                  # {certificate, chain}
│
└── chaincode/{id}/
    └── tls                                  # {certificate, chain}
```

Where `{chain}` is `tls` or `org`, and `{id}` is the entity identifier.

### Key Paths (`keys/`) - PRIVATE

Per-identity access only. Contains private keys and credentials.

```
secret/fabric/{org}/keys/
├── ca/{chain}/{id}/
│   ├── signing                              # {private_key}
│   ├── tls                                  # {private_key}
│   ├── certstore                            # {username, password}
│   ├── bootstrap-admin/
│   │   ├── credentials                      # {username, password}
│   │   ├── signing                          # {private_key}
│   │   └── tls                              # {private_key}
│   └── admin/
│       ├── signing                          # {private_key}
│       └── tls                              # {private_key}
│
├── peer/{id}/
│   ├── signing                              # {private_key}
│   ├── tls                                  # {private_key}
│   └── couchdb                              # {username, password}
│
├── orderer/{id}/
│   ├── signing                              # {private_key}
│   └── tls                                  # {private_key}
│
├── admin/{id}/
│   ├── signing                              # {private_key}
│   └── tls                                  # {private_key}
│
├── client/{id}/
│   ├── signing                              # {private_key}
│   └── tls                                  # {private_key}
│
└── chaincode/{id}/
    └── tls                                  # {private_key}
```

---

## Access Policy Changes

### Current Policy Pattern

Each component gets access to:
- Its own secret paths
- CA signing paths (to read certificates) → **inadvertently includes CA private keys**

Example from `PeerConfig.VaultPolicy()`:
```go
// Grants read to secret/data/fabric/{org}/ca/tls/root/signing
// This path contains BOTH certificate AND private_key
tlsRootCASigningPath := fmt.Sprintf("%s/%s/ca/tls/root/signing", vaultSecretDataPrefix, p.VaultBasePath())
```

### Target Policy Pattern

Each identity gets two access paths:
- **Public**: Read all `certs/*` (certificates are public)
- **Private**: Read only own `keys/{entity-path}/*`

```go
func (p PeerConfig) VaultPolicy() *vaultpb.PolicyDocument {
    return &vaultpb.PolicyDocument{
        Rules: []*vaultpb.PolicyRule{
            // Public: read all certificates
            {
                Path: fmt.Sprintf("%s/%s/certs/*", vaultSecretDataPrefix, p.VaultBasePath()),
                Capabilities: []vaultpb.Capability{
                    vaultpb.Capability_CAPABILITY_READ,
                    vaultpb.Capability_CAPABILITY_LIST,
                },
            },
            // Private: read only own keys
            {
                Path: fmt.Sprintf("%s/%s/keys/peer/%s/*", vaultSecretDataPrefix, p.VaultBasePath(), p.peerID),
                Capabilities: []vaultpb.Capability{
                    vaultpb.Capability_CAPABILITY_READ,
                },
            },
        },
    }
}
```

**Result**: 
- Any Fabric role can read any certificate (including all CA certs for trust stores)
- Each role can only read its own private keys and credentials
- CA private keys are only accessible to the CA itself

---

## Issue 1: Trust Stores Contain Full Certificate Chains

### Current Behavior

Trust stores include both intermediate and root CA certificates.

### Affected Components

| Component | File | Line |
|-----------|------|------|
| Peer | `helm/fabric/peer/templates/statefulset.yaml` | ~205 |
| Orderer | `helm/fabric/orderer/templates/statefulset.yaml` | ~184 |

### Current Code

```bash
cat /secrets/tls-int-ca-cert /secrets/tls-root-ca-cert > /tls/ca.crt
```

### Problem

- Trust stores should contain **only root CA certificates**
- Intermediate certificates should be sent during mTLS handshake
- Current approach masks broken certificate chain transmission

### Fix

```bash
# Trust store = root CA only
cp /secrets/tls-root-ca-cert /tls/ca.crt
```

TLS must be configured to send the full chain during handshake (depends on Issue 2).

---

## Issue 2: Certificate Chains Not Built at Enrollment

### Current Behavior

Only leaf certificates are stored. Chains are assembled at deployment time in init containers.

### Affected Enrollment Workflows

| Workflow | Creates |
|----------|---------|
| `fabric.v1.DeployCertificateAuthority` | CA signing cert, CA TLS cert |
| `fabric.v1.BootstrapCertificateAuthorityAdmin` | Bootstrap admin certs, permanent admin certs |
| `fabric.v1.CreateIdentity` | Any identity (peer, orderer, admin, etc.) |
| `fabric.v1.DeployPeer` | Peer identity + TLS certs |
| `fabric.v1.DeployOrderer` | Orderer identity + TLS certs |

### Current Storage

```
secret/fabric/{org}/peer/{id}/tls
├── private_key
├── public_key
└── certificate    # Leaf only
```

### Target Storage

Two paths, each with multiple keys:

```
secret/fabric/{org}/keys/peer/{id}/tls      # {private_key}
secret/fabric/{org}/certs/peer/{id}/tls     # {certificate, chain}
```

### Chain Construction

At enrollment time:
1. CA returns leaf certificate
2. Fetch issuing CA's chain from `certs/ca/{chain}/{id}/signing` (`chain` key)
3. Concatenate: `[leaf] + [issuing CA's chain]`
4. Store at single cert path with two keys:
   - `certificate` key → leaf only
   - `chain` key → full chain (leaf + intermediate(s) + root)

---

## Implementation Plan

### Phase 1: Vault Path Restructuring

Create parallel `keys/` and `certs/` structures.

**Files to modify:**

| File | Changes |
|------|---------|
| `temporal/lib/workers/fabric/config/certificate_authority.go` | Add `VaultSigningKeyPath()`, `VaultSigningCertPath()`, etc. |
| `temporal/lib/workers/fabric/config/components.go` | Same for peer, orderer, couchdb configs |
| `temporal/lib/workers/config/all.go` | Update base path construction if needed |

**New methods pattern:**
```go
// Keys (private)
func (ca CertificateAuthorityConfig) VaultSigningKeyPath() string {
    return fmt.Sprintf("%s/keys/ca/%s/%s/signing", ca.VaultBasePath(), ca.chain, ca.id)
}

// Certs (public)
func (ca CertificateAuthorityConfig) VaultSigningCertPath() string {
    return fmt.Sprintf("%s/certs/ca/%s/%s/signing", ca.VaultBasePath(), ca.chain, ca.id)
}

// Credentials (private, under keys/)
func (ca CertificateAuthorityConfig) VaultCertstorePath() string {
    return fmt.Sprintf("%s/keys/ca/%s/%s/certstore", ca.VaultBasePath(), ca.chain, ca.id)
}

func (ca CertificateAuthorityConfig) VaultBootstrapAdminCredentialsPath() string {
    return fmt.Sprintf("%s/keys/ca/%s/%s/bootstrap-admin/credentials", ca.VaultBasePath(), ca.chain, ca.id)
}
```

**Note**: `{chain}` is `tls` or `org`, `{id}` is the CA identifier (e.g., `root`, `int`).

### Phase 2: Policy Updates

Update all `VaultPolicy()` methods for new access patterns.

**Files to modify:**

| File | Methods |
|------|---------|
| `temporal/lib/workers/fabric/config/certificate_authority.go` | `CertificateAuthorityConfig.VaultPolicy()` |
| `temporal/lib/workers/fabric/config/components.go` | `PeerConfig.VaultPolicy()`, `OrdererConfig.VaultPolicy()`, `CouchDbConfig.VaultPolicy()` |

**New policy pattern:**
```go
func (p PeerConfig) VaultPolicy() *vaultpb.PolicyDocument {
    return &vaultpb.PolicyDocument{
        Rules: []*vaultpb.PolicyRule{
            // Public: read all certificates
            {
                Path: fmt.Sprintf("%s/%s/certs/*", vaultSecretDataPrefix, p.VaultBasePath()),
                Capabilities: []vaultpb.Capability{
                    vaultpb.Capability_CAPABILITY_READ,
                    vaultpb.Capability_CAPABILITY_LIST,
                },
            },
            // Private: read only own keys (includes couchdb credentials)
            {
                Path: fmt.Sprintf("%s/%s/keys/peer/%s/*", vaultSecretDataPrefix, p.VaultBasePath(), p.peerID),
                Capabilities: []vaultpb.Capability{
                    vaultpb.Capability_CAPABILITY_READ,
                },
            },
        },
    }
}
```

**Note**: CouchDB credentials are now under `keys/peer/{id}/couchdb`, covered by the `keys/peer/{id}/*` policy.

### Phase 3: Helm Values and Templates

Update Helm value generation and SecretProviderClass templates.

**Files to modify:**

| Category | Files |
|----------|-------|
| Helm values | `temporal/lib/workflows/cluster/fabric/v1/helm/certificate_authority.go` |
| | `temporal/lib/workflows/cluster/fabric/v1/helm/peer.go` |
| | `temporal/lib/workflows/cluster/fabric/v1/helm/orderer.go` |
| | `temporal/lib/workflows/cluster/fabric/v1/helm/couchdb.go` |
| | `temporal/lib/workflows/cluster/fabric/v1/helm/certificate_store.go` |
| Templates | `helm/fabric/certificate-authority/templates/secret.yaml` |
| | `helm/fabric/peer/templates/secret.yaml` |
| | `helm/fabric/orderer/templates/secret.yaml` |
| | `helm/fabric/couchdb/templates/secret.yaml` |
| | `helm/fabric/certificate-store/templates/secret.yaml` |
| Statefulsets | `helm/fabric/peer/templates/statefulset.yaml` |
| | `helm/fabric/orderer/templates/statefulset.yaml` |

### Phase 4: Enrollment Chain Construction

Build and store certificate chains at enrollment time. Each certificate path stores both `certificate` and `chain` keys.

**Files to modify:**

| File | Changes |
|------|---------|
| `temporal/lib/workflows/fabric/v1/activities.go` | `EnrollIdentity` builds chain after enrollment |
| `temporal/lib/workflows/fabric/v1/workflows.go` | Pass chain storage refs, fetch issuer chains |
| `temporal/proto/workflows/fabric/v1/fabric.proto` | Update enrollment messages for new storage pattern |
| `temporal/proto/workflows/secrets/v1/secrets.proto` | Update activity to store `{certificate, chain}` at single path |

**Chain construction logic:**
```go
// After enrollment, build and store chain
leafCert := enrollmentResult.Certificate
issuerChain := fetchFromVault(issuerCertPath, "chain")  // Get issuer's full chain
fullChain := leafCert + "\n" + issuerChain

// Store both at same path with different keys
storeToVault(certPath, map[string]string{
    "certificate": leafCert,
    "chain":       fullChain,
})
```

### Phase 5: Trust Store Correction

Update init containers to use root CA only for trust stores.

**Files to modify:**

| File | Changes |
|------|---------|
| `helm/fabric/peer/templates/statefulset.yaml` | Trust store = root only, TLS uses chain |
| `helm/fabric/orderer/templates/statefulset.yaml` | Same |

---

## Migration Considerations

### Existing Deployments

Existing deployments have secrets at old paths. Options:

1. **Dual-read period**: Temporarily support both path structures
2. **Migration workflow**: Move secrets to new paths before upgrading
3. **Clean redeploy**: For dev/test environments, redeploy from scratch

### Backward Compatibility

Consider feature flag to enable new path structure:
```go
if config.UseNewVaultPaths {
    return newPathStructure()
}
return legacyPathStructure()
```

---

## Testing Requirements

1. **Unit tests**: Verify new path methods return correct paths
2. **Policy tests**: Verify policies grant correct access
3. **Integration tests**: 
   - Peer cannot read CA private keys
   - Peer can read all CA certificates
   - mTLS works with chain transmission
4. **Security audit**: Confirm no path allows unintended key access

---

## Related

- [cert-design.md](cert-design.md) - Target architecture
- [bootstrap-workflows.md](bootstrap-workflows.md) - Workflow execution details
