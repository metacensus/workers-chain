# Working on workers-chain

Decisions with no single declaration to sit beside. Per-declaration detail is a comment beside the code; the [README](README.md) orients an operator.

Temporal workflow and activity authoring follows the Temporal skill in [`metacensus/skills`](https://github.com/metacensus/skills) (`skills/temporal`). Ownership of Vault paths, Kubernetes namespaces, and Nexus boundaries: [docs/strategy/OWNERSHIP.md](docs/strategy/OWNERSHIP.md).

## Rules below the seam

- **Cross-owner work goes through Nexus**, not a child workflow on the caller's worker ([OWNERSHIP.md P2](docs/strategy/OWNERSHIP.md)).
- **Private keys never cross Nexus** — only public CSRs and signed cert material.
- **A change to a workflow's calls, a worker's registrations or a Nexus boundary changes its `.twf` in the same PR.**
- **This repo builds and releases workers; it does not deploy them or start workflows.**

## Exemplars

Read the file beside a practice before writing a sibling.

[`dev-testing-go`](https://github.com/metacensus/skills/tree/main/skills/dev-testing-go):

- **Every test is a table:** [lib/postgres/v1/activities_test.go](lib/postgres/v1/activities_test.go); named builders in [lib/cluster/fabric/v1/workflows_test.go](lib/cluster/fabric/v1/workflows_test.go).
- **Validate apart from execution:** [lib/cluster/fabric/v1/workflows_test.go](lib/cluster/fabric/v1/workflows_test.go), [lib/fabric/v1/workflows_test.go](lib/fabric/v1/workflows_test.go).
- **Generated mocks standardize the double:** [.mockery.yaml](.mockery.yaml); the matcher in [lib/temporal/v1/workflows_test.go](lib/temporal/v1/workflows_test.go); each layer mocks only its direct dependency in [lib/cell/v1/workflows_test.go](lib/cell/v1/workflows_test.go) and [lib/postgres/v1/activities_test.go](lib/postgres/v1/activities_test.go).
- **Integration tests own their dependencies:** [lib/postgres/v1/client_integ_test.go](lib/postgres/v1/client_integ_test.go).
- **Conventions are tests, not prose:** [lib/workers/fabric/config/policy_test.go](lib/workers/fabric/config/policy_test.go), a security property for every principal.

[`dev-ci-github`](https://github.com/metacensus/skills/tree/main/skills/dev-ci-github):

- **One toolchain, read from a committed file:** the [Makefile](Makefile)'s `GEN_TOOLCHAIN`.
- **Committed output matches its source:** [`make check`](Makefile).
- **Workflow shape:** [pull_request.yml](.github/workflows/pull_request.yml) and [main.yml](.github/workflows/main.yml), one orchestrator per trigger, each calling [build.yml](.github/workflows/build.yml).

## Not yet

- The deploy repository becomes deploy-only after the first publish.
