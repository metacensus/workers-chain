# `temporal-server`

Custom Temporal server with compile-time `Authorizer` / `ClaimMapper` for Nexus
endpoint allowlisting. Canonical design (including plain-language overview):
[docs/fabric/bootstrap-workflows.md — Nexus Authorization](../docs/fabric/bootstrap-workflows.md#nexus-authorization).

| Path | Role |
|------|------|
| `authz/` | Allowlist + Authorizer + mTLS ClaimMapper, and the Nexus guards `serverfx` wires in |
| `serverfx/` | Upstream's service providers, re-declared to decorate history's endpoint registry and the frontends' matching client |
| `cmd/server/` | Upstream `main.go` with our plugins injected |
| `Dockerfile` | Binary layered onto `temporalio/server:${TEMPORAL_VERSION}` |

Client certificate CN = Temporal namespace (`infra`, `fabric`, `fabric-shard-{org}`).
`temporal-system-admin` and `temporal-internode` get cluster system admin.

Its own Go module: the server's dependency graph stays out of the workers'. `make check` at the root vets and tests it.

Terraform: `server_image = metacensus/temporal-server`, tag = the `go.temporal.io/server` version in `go.mod`. Local: `make docker-build-temporal-server`. Each workers-chain release publishes it for linux/amd64 and linux/arm64.
