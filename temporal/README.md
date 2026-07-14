# Temporal Workflows

**Canonical home** for Temporal package layout, workflow modules, and registration rules.

**Temporal code placement:** this README. **Core Go placement** (API, chaincode, fabric client):
[core/README.md](../core/README.md#architecture--layer-placement)

## Pattern: Proto-Framework-Implementation (PFI)

Protobuf definitions generate framework code; `lib/workflows/` provides implementations.

**PFI guide (canonical):** [docs/strategy/PFI_PATTERN.md](../docs/strategy/PFI_PATTERN.md)

### Package structure

```text
temporal/
├── proto/<service>/<version>/           # Interface definitions (.proto)
├── gen/workflows/<service>/<version>/   # Generated messages + Temporal stubs
├── lib/workflows/<service>/<version>/   # workflows.go, activities.go, client.go, fx.go
├── lib/workers/                         # infra-worker and fabric-worker DI modules
├── cmd/workers/                         # Worker entrypoints
└── mocks/                               # Generated test mocks
```

**Reference implementation:** `lib/workflows/vault/v1/`

`proto/README.md` is **auto-generated** by `make proto` (buf `docs-out`) — edit `.proto` files,
not that README.

### Registration

Each service module must export `RegisterActivities()` and `RegisterWorkflows()` and be
wired through `fx.Module`. Missing registration causes runtime "activity not found" errors.

## Testing

Run from `temporal/`:

```bash
go test ./lib/...
go test -tags=integration ./lib/...
```

**Testing strategy (canonical):** [UNIT_TESTING.md](../.agents/skills/temporal-workflow-tests/references/UNIT_TESTING.md)
and linked workflow/activity/client guides. Explicit test planning: `temporal-workflow-tests`
skill.

## Key workflow packages

| Package | Responsibility |
|---------|----------------|
| `lib/workflows/cell/v1/` | Platform bootstrap (`BootstrapInfra`, `InitFabricShard`) |
| `lib/workflows/fabric/v1/` | Per-org Fabric shard operations (CAs, peers, channels) |
| `lib/workflows/cluster/fabric/v1/` | K8s namespace/deployment for Fabric components |
| `lib/workflows/kube/v1/` | Generic Kubernetes create/update |
| `lib/workflows/vault/v1/` | Vault secret and policy operations |
| `lib/workflows/postgres/v1/` | Postgres schema and queries |
| `lib/workflows/secrets/v1/` | Secret generation and storage |
| `lib/workflows/helm/v1/` | Helm chart rendering |
| `lib/workflows/temporal/v1/` | Temporal namespace and Nexus management |

**Fabric deploy flow:** [docs/fabric/bootstrap-workflows.md](../docs/fabric/bootstrap-workflows.md)

## Code generation

`make gen`, `make proto`, `make mocks` — see [README.md — Development](../README.md#development).

## Related docs

- [docs/strategy/PFI_PATTERN.md](../docs/strategy/PFI_PATTERN.md)
- [docs/fabric/bootstrap-workflows.md](../docs/fabric/bootstrap-workflows.md)
- [terraform/README.md](../terraform/README.md) — what Terraform provisions vs what workflows orchestrate
- [AGENTS.md](../AGENTS.md) — decision framework
