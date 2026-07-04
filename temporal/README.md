# Temporal Workflows

**Code placement:** Workflows and activities belong in `temporal/`. See
[core/README.md](../core/README.md#architecture--layer-placement) for the full layer model.

## Pattern: Proto-Framework-Implementation (PFI)

Protobuf definitions generate framework code; `lib/workflows/` provides implementations.

Full guide: [docs/strategy/PFI_PATTERN.md](../docs/strategy/PFI_PATTERN.md)

### Package structure

```text
temporal/
├── proto/<service>/<version>/     # Interface definitions (.proto)
├── gen/workflows/<service>/<version>/   # Generated messages + Temporal stubs
├── lib/workflows/<service>/<version>/   # workflows.go, activities.go, client.go, fx.go
├── lib/workers/                     # infra-worker and fabric-worker DI modules
├── cmd/workers/                     # Worker entrypoints
└── mocks/                           # Generated test mocks
```

**Reference implementation:** `lib/workflows/vault/v1/`

`proto/README.md` is **auto-generated** by `make proto` (buf `docs-out`) — edit `.proto`
files, not the README directly.

### Registration

Each service module must export `RegisterActivities()` and `RegisterWorkflows()` and be
wired through `fx.Module`. Missing registration causes runtime "activity not found" errors.

## Testing

Run from `temporal/`:

```bash
# Unit tests (no Docker)
go test ./lib/...

# With integration tests (requires Docker)
go test -tags=integration ./lib/...
```

Strategy docs: [docs/strategy/UNIT_TESTING.md](../docs/strategy/UNIT_TESTING.md),
[UNIT_TESTING_WORKFLOWS.md](../docs/strategy/UNIT_TESTING_WORKFLOWS.md),
[UNIT_TESTING_ACTIVITIES.md](../docs/strategy/UNIT_TESTING_ACTIVITIES.md),
[UNIT_TESTING_CLIENTS.md](../docs/strategy/UNIT_TESTING_CLIENTS.md)

Explicit test planning: use the `temporal-workflow-tests` agent skill.

## Test pyramid

| Layer | File | Mocks | Build tag |
|-------|------|-------|-----------|
| Workflows | `workflows_test.go` | Activities | — |
| Activities | `activities_test.go` | Client | — |
| Client | `client_integ_test.go` | Real system (testcontainers) | `integration` |

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

Fabric deploy flow is documented in [docs/fabric/bootstrap-workflows.md](../docs/fabric/bootstrap-workflows.md).

## Code generation

```bash
make gen       # proto + mocks (from repo root)
make proto     # protobuf only
make mocks     # mocks only
```

Generated code is committed; CI verifies it is up to date.

## Related docs

- [docs/strategy/PFI_PATTERN.md](../docs/strategy/PFI_PATTERN.md)
- [docs/fabric/bootstrap-workflows.md](../docs/fabric/bootstrap-workflows.md)
- [terraform/README.md](../terraform/README.md) — what Terraform provisions vs what workflows orchestrate
- [AGENTS.md](../AGENTS.md) — decision framework
