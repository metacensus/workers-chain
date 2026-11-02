package config

// Version is the release tag this binary was built from, set with
// -ldflags "-X github.com/metacensus/workers-chain/lib/workers/config.Version=...".
// The infra worker deploys the fabric and shard workers at it.
var Version string
