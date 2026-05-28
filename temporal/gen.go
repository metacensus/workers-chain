package temporal

// Code generation entrypoints. The Makefile invokes these via
// `go generate -run mockery` (see `make mocks`) and `buf generate`
// (see `make proto`). Tool versions are tracked by the `tool`
// directive in go.mod.

//go:generate buf generate
//go:generate go tool mockery
