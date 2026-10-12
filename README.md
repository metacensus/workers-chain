# workers-chain

Temporal workers and Helm charts that bootstrap and operate MetaCensus Fabric infrastructure. Deploy and orchestration that *starts* these workers lives in [`metacensus/infra`](https://github.com/metacensus/infra); this repository is where the workers build, test, and release.

Three images (names unchanged for deploy compatibility):

- `metacensus/infra-worker`
- `metacensus/fabric-worker`
- `metacensus/shard-worker`

Design decisions are in [AGENTS.md](AGENTS.md). Ownership boundaries: [`docs/strategy/OWNERSHIP.md`](docs/strategy/OWNERSHIP.md). Fabric bootstrap flows: [`docs/fabric/bootstrap-workflows.md`](docs/fabric/bootstrap-workflows.md).

## Layout

| Path | |
|---|---|
| `temporal/` | Workers, workflows, activities, protos |
| `helm/` | Charts the workers render/deploy |
| `docs/fabric/` | Fabric bootstrap and certificate design |

## Tests

`make help` lists targets once the module rename lands ([#2](https://github.com/metacensus/workers-chain/issues/2)); `make check` is the CI gate.

## Releasing

Semver tags publish the three worker images ([#3](https://github.com/metacensus/workers-chain/issues/3)).

## License

[AGPL-3.0-only](LICENSE). See [NOTICE](NOTICE) for BSD-3-Clause material retained from the infra carve.
