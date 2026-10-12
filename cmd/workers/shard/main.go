package main

import (
	"log"

	"go.uber.org/fx"

	"github.com/metacensus/workers-chain/lib/workers"
	"github.com/metacensus/workers-chain/lib/workers/config"
	"github.com/metacensus/workers-chain/lib/workers/shard"
)

func main() {
	// Initialize configuration from environment.
	// ORG_ID (worker identity) and OPERATOR_ID (platform context) are loaded
	// via fx in the shard module.
	config.InitializeTemporal()
	config.InitializeVault()
	config.InitializeKubernetes()
	config.InitializeWorker()

	cfg := config.GetTemporal()
	log.Println("Connecting to Temporal at", cfg.Address)

	app := fx.New(
		// Include shard worker module
		shard.Module,

		// Start healthcheck server (shared)
		fx.Invoke(workers.StartHealthcheck),

		// Run the worker (shared)
		fx.Invoke(workers.RunWorker),
	)

	app.Run()
}
