# Temporal Workflows

**Canonical home** for Temporal package layout, workflow modules, and registration rules.

**Temporal code placement:** this README.

## Pattern: Proto-Framework-Implementation (PFI)

Protobuf definitions generate framework code; `lib/` provides implementations.

### Package structure

```text
workers-chain/
├── proto/<service>/<version>/           # Interface definitions (.proto)
├── gen/<service>/<version>/             # Generated messages + Temporal stubs
├── lib/<service>/<version>/             # workflows.go, activities.go, client.go, fx.go
├── lib/workers/                         # infra / fabric / shard worker DI modules
├── cmd/workers/                         # Worker entrypoints
└── mocks/                               # Generated test mocks
```

**Reference implementation:** `lib/vault/v1/`

`proto/README.md` is **auto-generated** by `make proto` (buf `docs-out`) — edit `.proto` files,
not that README.

### Registration

Each service module must export `RegisterActivities()` and `RegisterWorkflows()` and be
wired through `fx.Module`. Missing registration causes runtime "activity not found" errors.

## Testing

Run from the repository root:

```bash
go test ./lib/...
go test -tags=integration ./lib/...
```

## Key workflow packages

| Package | Responsibility |
|---------|----------------|
| `lib/cell/v1/` | Platform bootstrap (`BootstrapInfra`, `InitFabric`, `InitFabricShard`) |
| `lib/fabric/v1/` | Shared fabric + per-org shard orchestration (CAs, peers, orderers) |
| `lib/fabric/certauth/v1/` | Certificate authority identity and CSR signing |
| `lib/cluster/fabric/v1/` | K8s namespace/deployment for Fabric components |
| `lib/vault/v1/` | Vault secret and policy operations |
| `lib/postgres/v1/` | Postgres schema and queries |
| `lib/secrets/v1/` | Secret generation and storage |
| `lib/temporal/v1/` | Temporal namespace and Nexus management |

Generic Kubernetes and Helm activities come from `github.com/metacensus/library-temporal-kube`.

## Pattern: Ownership Model

Workflows execute in exactly one security context — their worker's owner. Owner configs, typed
`Owner` inputs, Nexus boundaries, and Vault policy generation are governed by
[docs/strategy/OWNERSHIP.md](../docs/strategy/OWNERSHIP.md). Read it before adding workflows that
touch Vault paths or Kubernetes namespaces.

**Fabric deploy flow:** [docs/fabric/bootstrap-workflows.md](../docs/fabric/bootstrap-workflows.md)

## Code generation

`make gen`, `make proto`, `make mocks` — see [README.md — Development](../README.md#development).

## Related docs

- [docs/strategy/OWNERSHIP.md](../docs/strategy/OWNERSHIP.md)
- [docs/fabric/bootstrap-workflows.md](../docs/fabric/bootstrap-workflows.md)
- [AGENTS.md](../AGENTS.md) — decision framework
