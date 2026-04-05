# Fabric Channel Workflows

**Status Legend:**
- ✅ Implemented
- 🟡 Partial
- ❌ Not Started

---

## Overview

This document outlines workflows for creating Fabric channels and joining nodes. It extends the bootstrap process in [`bootstrap-workflows.md`](bootstrap-workflows.md) to include channel operations after org infrastructure is deployed.

**Scope:** Channel creation, orderer/peer joining, anchor peer configuration, and adding nodes to existing channels.

---

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Consensus** | Raft (etcdraft) | CFT sufficient for trusted consortium; 3 nodes minimum |
| **Fabric Version** | v2.5+ | Channel participation API required; system channel deprecated |
| **Channel Timing** | Post-bootstrap workflow | Separate from `BootstrapShard`; channels span multiple orgs |
| **Genesis Storage** | Vault | Consistent with other secrets; enables workflow access |

### Open Questions

- **Multi-org orderer topology:** Single orderer org or orderers distributed across peer orgs?
- **Channel naming convention:** `<network>-<purpose>` or flat names?
- **Anchor peer selection:** First peer per org, or configurable?

---

## Fabric Version Context

| Version | Key Change |
|---------|------------|
| v2.3+ | Channel participation API; system channel deprecated |
| v3.0 | System channel removed; `OrdererEndpoints` must be per-org |
| v3.0 | SmartBFT consensus added (requires 4+ orderers) |

**Target:** v2.5+ with Raft consensus. SmartBFT deferred until multi-org trust boundaries require BFT.

---

## Worker Architecture Integration

Channel operations run on **Fabric Workers** (not Central Worker) because they require:
- Fabric CA client access (for admin identity)
- Fabric CLI tools (`configtxgen`, `osnadmin`, `peer`)
- Access to org-specific secrets in Vault

| Operation | Worker | Vault Access | K8s Access |
|-----------|--------|--------------|------------|
| Generate genesis block | Fabric Worker | Read org MSP certs | None |
| Join orderer to channel | Fabric Worker | Read TLS certs, admin creds | None (calls orderer API) |
| Join peer to channel | Fabric Worker | Read admin MSP | None (calls peer API) |
| Set anchor peer | Fabric Worker | Read admin MSP | None (calls peer API) |

**Note:** Channel creation spans multiple orgs. The initiating org's Fabric Worker generates the genesis block; other orgs join their nodes using their own workers.

---

## Implementation Status

### Root Workflows

These are the entry points for channel operations.

```
Workflow: fabric.v1.CreateChannel ❌ (Runs in fabric-<org> namespace)
├─ Activity: fabric.v1.GenerateConfigTx ❌
│   └─ Generates configtx.yaml from org definitions
├─ Activity: fabric.v1.GenerateGenesisBlock ❌
│   └─ Runs: configtxgen -profile <Profile> -outputBlock genesis.block -channelID <channel>
├─ Activity: secrets.v1.StoreGenesisBlock ❌
│   └─ Stores genesis block in Vault for other orgs to fetch
│
├─ for each Orderer (SEQUENTIAL)
│   └─ Workflow: fabric.v1.JoinOrdererToChannel ❌ [→ see Component Workflows]
│
├─ for each Peer (PARALLEL after orderers ready)
│   └─ Workflow: fabric.v1.JoinPeerToChannel ❌ [→ see Component Workflows]
│
└─ for each Org
    └─ Workflow: fabric.v1.SetAnchorPeer ❌ [→ see Component Workflows]
```

```
Workflow: fabric.v1.AddOrdererToChannel ❌
├─ Activity: fabric.v1.PeerChannelFetch ❌
│   └─ Fetches LATEST config block (NOT genesis) for follower sync
├─ Workflow: fabric.v1.JoinOrdererToChannel ❌ [→ see Component Workflows]
│   └─ Joins as FOLLOWER; polls until synced
└─ Workflow: fabric.v1.AddOrdererToConsenterSet ❌ [→ see Component Workflows]
    └─ Promotes follower to consenter via config update
```

**Critical:** Never add orderers when existing consenters are offline. Adding a 4th node to a 3-node cluster changes quorum from 2 to 3.

```
Workflow: fabric.v1.AddPeerToChannel ❌
├─ (If org NOT in channel)
│   └─ Workflow: fabric.v1.AddOrgToChannel ❌ [→ see Component Workflows]
│
└─ Workflow: fabric.v1.JoinPeerToChannel ❌ [→ see Component Workflows]
```

---

### Component Workflows

Reusable workflows called by root workflows above.

```
Workflow: fabric.v1.JoinOrdererToChannel ❌
├─ Activity: fabric.v1.OsnadminChannelJoin ❌
│   └─ Runs: osnadmin channel join --config-block <block>
│   └─ Uses genesis block → joins as CONSENTER
│   └─ Uses config block → joins as FOLLOWER
└─ Activity: fabric.v1.WaitForOrdererActive ❌
    └─ Polls osnadmin channel list until status: "active"
    └─ For consenter: consensusRelation = "consenter"
    └─ For follower: height matches cluster before promotion
```

```
Workflow: fabric.v1.JoinPeerToChannel ❌
├─ Activity: fabric.v1.PeerChannelFetch ❌
│   └─ Runs: peer channel fetch oldest <channel>.block -c <channel>
└─ Activity: fabric.v1.PeerChannelJoin ❌
    └─ Runs: peer channel join -b <channel>.block
```

```
Workflow: fabric.v1.SetAnchorPeer ❌
├─ Workflow: fabric.v1.UpdateChannelConfig ❌ [→ see below]
│   └─ Modification: Add AnchorPeers to org's Application group
└─ (No additional signatures required - org admin can set own anchors)
```

```
Workflow: fabric.v1.AddOrdererToConsenterSet ❌
├─ Workflow: fabric.v1.UpdateChannelConfig ❌ [→ see below]
│   └─ Modification: Add orderer to ConsensusType.metadata.consenters[]
└─ Activity: fabric.v1.WaitForOrdererActive ❌
    └─ Polls until consensusRelation changes: follower → consenter
```

```
Workflow: fabric.v1.AddOrgToChannel ❌
├─ Activity: fabric.v1.GenerateOrgDefinition ❌
│   └─ Creates org MSP definition for channel config
├─ Workflow: fabric.v1.UpdateChannelConfig ❌ [→ see below]
│   └─ Modification: Add org to Application.groups
└─ Activity: fabric.v1.CollectSignatures ❌
    └─ Requires signatures from MAJORITY of existing channel admins
```

---

### Config Update Workflow

Common pattern for all channel configuration changes.

```
Workflow: fabric.v1.UpdateChannelConfig ❌
├─ Activity: fabric.v1.FetchChannelConfig ❌
│   └─ Runs: peer channel fetch config
│   └─ Decodes: configtxlator proto_decode --type common.Block
│   └─ Extracts: .data.data[0].payload.data.config
│
├─ Activity: fabric.v1.ComputeConfigUpdate ❌
│   └─ Input: original config JSON, modification function
│   └─ Encodes both: configtxlator proto_encode --type common.Config
│   └─ Computes delta: configtxlator compute_update
│   └─ Output: ConfigUpdate protobuf
│
├─ Activity: fabric.v1.SignConfigUpdate ❌ (optional, for multi-org updates)
│   └─ Runs: peer channel signconfigtx -f <update.pb>
│   └─ Called by each org that needs to sign
│
└─ Activity: fabric.v1.SubmitChannelUpdate ❌
    └─ Wraps in envelope with channel header
    └─ Runs: peer channel update -f <envelope.pb>
```

**Modification functions** (passed to ComputeConfigUpdate):
- `AddAnchorPeer(org, host, port)` - for SetAnchorPeer
- `AddConsenter(host, port, tlsCert)` - for AddOrdererToConsenterSet
- `AddOrg(orgDefinition)` - for AddOrgToChannel
- `RemoveConsenter(host, port)` - for removing orderers (future)

---

## Activity Details

### All Activities

| Activity | CLI/Tool | Credentials | Description |
|----------|----------|-------------|-------------|
| `GenerateConfigTx` | Template engine | MSP certs (read) | Generates configtx.yaml from org definitions |
| `GenerateGenesisBlock` | `configtxgen` | MSP certs (read) | Creates channel genesis block |
| `StoreGenesisBlock` | Vault client | Vault token | Stores genesis block for cross-org access |
| `OsnadminChannelJoin` | `osnadmin channel join` | Admin TLS cert | Joins orderer to channel |
| `WaitForOrdererActive` | `osnadmin channel list` | Admin TLS cert | Polls until orderer status is "active" |
| `PeerChannelFetch` | `peer channel fetch` | Admin MSP, orderer TLS CA | Fetches block from orderer |
| `PeerChannelJoin` | `peer channel join` | Admin MSP | Joins peer to channel |
| `FetchChannelConfig` | `peer channel fetch config` + `configtxlator` | Admin MSP, orderer TLS CA | Fetches and decodes current config |
| `ComputeConfigUpdate` | `configtxlator` | None (local) | Computes config delta |
| `SignConfigUpdate` | `peer channel signconfigtx` | Admin MSP | Signs config update (multi-org) |
| `SubmitChannelUpdate` | `peer channel update` | Admin MSP, orderer TLS CA | Submits signed config update |
| `GenerateOrgDefinition` | Template engine | MSP certs (read) | Creates org definition for AddOrgToChannel |
| `CollectSignatures` | Coordination | Multiple admin MSPs | Gathers required signatures from other orgs |

### osnadmin Authentication

The `osnadmin` CLI requires mTLS to the orderer's admin port:

```
Admin.ListenAddress: 0.0.0.0:7053
Admin.TLS.ClientAuthRequired: true
Admin.TLS.ClientRootCAs: [<ca-cert>]
```

**Credential source:** Admin TLS client cert must be signed by CA in `ClientRootCAs`. Store in Vault at `secret/fabric/<org>/admin/tls/*`.

### Cross-Org Coordination

`CollectSignatures` requires coordination between orgs. Options:
1. **Nexus call** - Each org's Fabric Worker exposes a signing endpoint
2. **Shared storage** - Store unsigned update in Vault; orgs poll and sign
3. **External coordination** - Human-triggered signing workflow per org

**Open question:** Which coordination pattern to use for multi-org config updates?

---

## Security Boundaries

### Per-Channel Isolation

| Resource | Isolation Mechanism |
|----------|-------------------|
| Channel config | Only orgs in channel can read/modify |
| Block data | Only channel members receive blocks |
| Orderer consensus | Consenters list defines participants |
| Peer gossip | Anchor peers enable cross-org discovery |

### Credential Requirements by Operation

| Operation | Identity Type | Vault Path |
|-----------|--------------|------------|
| Create channel | Org admin | `secret/fabric/<org>/users/admin/*` |
| Join orderer | Orderer admin TLS | `secret/fabric/<org>/admin/tls/*` |
| Join peer | Org admin | `secret/fabric/<org>/users/admin/*` |
| Config updates | Org admin(s) | Multiple orgs may need to sign |

---

## Prerequisites

Before running channel workflows, the following must exist (from [`bootstrap-workflows.md`](bootstrap-workflows.md)):

- ✅ Fabric CA infrastructure deployed
- ✅ Orderer nodes running with channel participation API enabled
- ✅ Peer nodes running
- ✅ Admin identities enrolled and stored in Vault
- ❌ `configtxgen` binary available in Fabric Worker container
- ❌ `osnadmin` binary available in Fabric Worker container
- ❌ `configtxlator` binary available in Fabric Worker container

---

## Integration with BootstrapShard

Channel creation is **separate** from `BootstrapShard`:

```
BootstrapShard (existing)
└─ Deploys: CA, peers, orderers, identities
   └─ Output: Running nodes, admin credentials in Vault

CreateChannel (new, runs after BootstrapShard)
└─ Uses: Running nodes, admin credentials
   └─ Output: Functional channel with all nodes joined
```

**Rationale:** Channels span multiple orgs. Each org runs `BootstrapShard` independently; channel creation requires coordination across orgs.

---

## References

- [Fabric Channel Participation API](https://hyperledger-fabric.readthedocs.io/en/latest/create_channel/create_channel_participation.html) — Official guide for `osnadmin` commands
- [Fabric configtxgen](https://hyperledger-fabric.readthedocs.io/en/latest/commands/configtxgen.html) — Genesis block generation
- [Fabric configtxlator](https://hyperledger-fabric.readthedocs.io/en/latest/commands/configtxlator.html) — Config block encoding/decoding
- [Fabric Channel Update Tutorial](https://hyperledger-fabric.readthedocs.io/en/latest/config_update.html) — Anchor peer and config updates

---

**Last Updated:** 2026-01-21
