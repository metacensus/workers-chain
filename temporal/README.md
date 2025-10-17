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
└── workflows/<service>/<version>/
    ├── workflows.go             # Workflow implementations
    ├── activities.go            # Activity implementations  
    ├── client.go                # External system client (may be different or unnecessary)
    ├── fx.go                    # Dependency injection
    ├── workflows_test.go        # Unit (mock activities)
    ├── activities_test.go       # Unit (mock client)
    └── client_test.go           # Integration (real system, build tag)
```

**Reference:** `workflows/vault/v1/` package implements this pattern.

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

