# Temporal Workflows

## Pattern: Proto-Framework-Implementation (PFI)

Protobuf definitions generate framework, implementation provides logic.

**Note:** This pattern has a name so you can reference it in conversation with AI assistants: "Follow the PFI pattern" or "Use the Test Pyramid approach."

### Package Structure

```
temporal/
├── proto/<service>/<version>/
│   └── <service>.proto          # Interface definition
│
├── gen/workflows/<service>/<version>/
│   ├── *.pb.go                  # Generated messages
│   └── *_temporal.pb.go         # Generated framework
│
└── lib/workflows/<service>/<version>/
    ├── workflows.go             # Workflow implementations
    ├── activities.go            # Activity implementations  
    ├── client.go                # External system client (may be different or unnecessary)
    ├── fx.go                    # Dependency injection
    ├── workflows_test.go        # Unit (mock activities)
    ├── activities_test.go       # Unit (mock client)
    └── client_test.go           # Integration (real system, build tag)
```

**Reference:** `lib/workflows/vault/v1/` package implements this pattern.

## Testing Commands

```bash
# Fast unit tests (default - no Docker needed)
go test ./workflows/...

# With integration tests (requires Docker)
go test -tags=integration ./workflows/...
```

## Pattern: Test Pyramid

| Layer | File | Mocks | Build Tag |
|-------|------|-------|-----------|
| Workflows | `workflows_test.go` | Activities | - |
| Activities | `activities_test.go` | Client | - |
| Client | `client_test.go` | Real system | `integration` |

## Terraform Workflows Architecture

The `workflows/terraform/v1/` package provides **generic, composable workflows** for declarative infrastructure deployment via Terraform.

### Concept

Rather than running Terraform ad-hoc, all infrastructure changes flow through Temporal workflows. This provides:
- **Repeatability**: Identical deployments across environments
- **Observability**: Full audit trail in Temporal UI
- **Composability**: Entity-specific workflows orchestrate generic Terraform workflows with locking

### Workflow Types

**Generic Terraform Workflows** (reusable, stateless):
- `CreateTerraform` - Initialize and apply fresh Terraform state
- `UpdateTerraform` - Update existing Terraform state
- `PlanTerraform` - Generate Terraform plan (dry-run)
- `ApplyTerraform` - Apply existing Terraform plan
- `DestroyTerraform` - Destroy infrastructure managed by Terraform

**Entity Workflows** (specific components, hold locks):
- Entity-specific workflows (e.g., `DeployVault`, `DeployFabric`) call generic workflows
- Entity workflows manage **Temporal-level locking** to prevent concurrent operations
- Callers invoke entity workflows, not generic workflows directly

### State Management

Terraform state is stored in S3 with full audit trail:
```
s3://bucket/
├── {module}/
│   ├── {org}/                    # org is optional (many modules have no org)
│   │   ├── 2025-10-19T14:30:22Z.tfplan      # Plan output
│   │   ├── 2025-10-19T14:30:22Z.tfstate     # Terraform state
│   │   └── 2025-10-19T14:31:00Z.destroyed   # Tombstone marker
```

**Design Pattern:**
- Single timestamp groups related artifacts (plan + state from same action)
- ISO 8601 format: `YYYY-MM-DDTHH:MM:SSZ` (lexicographically sortable)
- Tombstone file `{timestamp}.destroyed` marks successful destruction
- Full history visible when browsing S3 bucket

### Locking Strategy

**Locking happens at the workflow level** (not activities):
1. Entity workflow acquires Temporal lock (via semaphore or dedicated activity)
2. Entity workflow invokes generic Terraform workflow
3. Entity workflow releases lock on completion

This design:
- Keeps generic workflows simple and stateless
- Prevents concurrent Terraform applies on same resources
- Leverages Temporal's workflow determinism for lock safety

### Error Handling

**Current State (v1):**
- If `ApplyTerraform` fails, cluster may be left in partial state
- Operator must investigate logs and manually recover
- Plan and state files in S3 aid debugging

**Future Work:**
- State version restoration (manual rollback)
- Automatic rollback to last known-good state
- Destroy + redeploy for immutable workloads

See reference implementation: `workflows/terraform/v1/`

