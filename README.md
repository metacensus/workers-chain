# workers-chain

Temporal workers and Helm charts that bootstrap and operate MetaCensus Fabric infrastructure. Deploy and orchestration that *starts* these workers lives in the MetaCensus `infra` repository; this repository is where the workers build, test, and release.

Images:

- `metacensus/worker-infra`
- `metacensus/worker-fabric`
- `metacensus/worker-fabric-shard`

Design decisions are in [AGENTS.md](AGENTS.md). Ownership boundaries: [`docs/strategy/OWNERSHIP.md`](docs/strategy/OWNERSHIP.md). Fabric bootstrap flows: [`docs/fabric/bootstrap-workflows.md`](docs/fabric/bootstrap-workflows.md).

## Layout

| Path | |
|---|---|
| `cmd/`, `lib/`, `proto/`, `gen/` | Workers, workflows, activities, protos |
| `helm/` | Charts the workers render/deploy |
| `docs/fabric/` | Fabric bootstrap and certificate design |

## Tests

`make help` lists targets; `make check` is the CI gate ([#2](https://github.com/metacensus/workers-chain/issues/2)).

## Releasing

Semver tags publish the three worker images ([#3](https://github.com/metacensus/workers-chain/issues/3)).

## License

[AGPL-3.0-only](LICENSE). See [NOTICE](NOTICE) for BSD-3-Clause material retained from the infra carve.
