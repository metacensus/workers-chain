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
| `cmd/`, `lib/`, `proto/`, `gen/`, `mocks/` | Workers, workflows, activities, protos. `proto`, `gen`, `lib` and `mocks` each hold `<application space>/<version>`, as `fabric/certauth/v1`; `lib/workers` and `lib/healthcheck` are not spaces |
| `helm/` | Charts the workers render/deploy; the render and apply activities come from [library-temporal-kube](https://github.com/metacensus/library-temporal-kube) |
| `twf/`, `scripts/check-twf.sh` | The design in `.twf`: one package per application space, `twf/deploy` the workers; the script checks it against `proto/` and `lib/workers` |
| `temporal-server/` | The custom Temporal server whose Authorizer gates Nexus; its own Go module |
| `docs/fabric/` | Fabric bootstrap and certificate design |

## Tests

`make help` lists targets; `make check` is the CI gate.

`make hooks` opts in to git hooks that run the same checks before code leaves your machine: pre-commit checks formatting and, when a `.proto` or mocked interface is staged, proto lint and generated-code freshness; pre-push runs `make check`.

## Releasing

`make release TYPE=patch|minor|major` tags the next semver, and the tag publishes the three worker images.

**Clusters built with v0.2.0 or earlier are rebuilt, not upgraded.** Those workers wrote Kubernetes objects with update calls; later ones apply server-side as `workers-chain` without force, so their first apply to an existing object fails with `Kube.Conflict` against the old manager ([#36](https://github.com/metacensus/workers-chain/issues/36)).

## License

[AGPL-3.0-only](LICENSE). See [NOTICE](NOTICE) for retained BSD-3-Clause material.
