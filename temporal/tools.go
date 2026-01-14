//go:build tools
// +build tools

package temporal

// This file declares dependencies on tools used for code generation.
// These dependencies will be tracked in go.mod but won't be included in production builds.
// The imports below keep these tools from being removed by 'go mod tidy'.
//
// To install tools: make install-tools
// To generate code: make generate
//
// See: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

//go:generate buf generate
//go:generate mockery

import (
	// Generates Temporal workflow/activity interfaces from protobuf
	_ "github.com/cludden/protoc-gen-go-temporal/pkg/scheme"

	// Generates Nexus service interfaces from protobuf (cmd packages for tools)
	_ "github.com/bergundy/protoc-gen-go-nexus-temporal/cmd/protoc-gen-go-nexus-temporal"
	_ "github.com/bergundy/protoc-gen-go-nexus/cmd/protoc-gen-go-nexus"

	// Generates mocks for testing
	_ "github.com/vektra/mockery/v2/pkg/config"
)

// Note: For tools that are pure binaries (no library packages), import the cmd package
// directly with a blank import. This keeps the module in go.mod so 'go install' works.
