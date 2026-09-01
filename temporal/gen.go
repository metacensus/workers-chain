package temporal

// Code generation entrypoints. The Makefile invokes these via
// `go generate -run mockery` (see `make mocks`) and `buf generate`
// (see `make proto`). Both run the pinned binaries that `make
// install-tools-gen` places on PATH; their versions are tracked by the
// `tool` directives in go.mod.

//go:generate buf generate
//go:generate mockery
