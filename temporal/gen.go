package temporal

// Code generation directives for protobuf and mocks.
// Run via: make gen (or make proto / make mocks for individual targets)

//go:generate sh -c "rm -rf gen/workflows && buf dep update && buf generate"
//go:generate mockery
