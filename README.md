# workers-chain

Temporal workers and Helm charts that bootstrap and operate MetaCensus Fabric infrastructure. Deploy and orchestration that *starts* these workers lives in the deploy repository; this repository is where the workers build, test, and release.

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

`make help` lists targets; `make check` is the CI gate.

`make hooks` opts in to git hooks that run the same checks before code leaves your machine: pre-commit checks formatting and, when a `.proto` or mocked interface is staged, proto lint and generated-code freshness; pre-push runs `make check`.

## Releasing

`make release TYPE=patch|minor|major` tags the next semver, and the tag publishes the three worker images.

## License

[AGPL-3.0-only](LICENSE). See [NOTICE](NOTICE) for BSD-3-Clause material retained from the infra carve.
