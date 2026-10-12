# Working on workers-chain

Decisions with no single declaration to sit beside. Per-declaration detail is a comment beside the code; the [README](README.md) orients an operator.

Temporal workflow and activity authoring follows the Temporal skill in [`metacensus/skills`](https://github.com/metacensus/skills) (Temporal / workflow-authoring). Ownership of Vault paths, Kubernetes namespaces, and Nexus boundaries: [docs/strategy/OWNERSHIP.md](docs/strategy/OWNERSHIP.md).

## Rules below the seam

- **Cross-owner work goes through Nexus**, not a child workflow on the caller's worker ([OWNERSHIP.md P2](docs/strategy/OWNERSHIP.md)).
- **Private keys never cross Nexus** — only public CSRs and signed cert material.
- **This repo builds and releases workers; it does not deploy them or start workflows.** Deploy stays in `metacensus/infra`.

## Not yet

- Module path rename to `github.com/metacensus/workers-chain` and root `make check` ([#2](https://github.com/metacensus/workers-chain/issues/2)).
- CI/release for the three images ([#3](https://github.com/metacensus/workers-chain/issues/3)).
- Infra deploy-only cleanup after publish ([metacensus/infra#72](https://github.com/metacensus/infra/issues/72)).
