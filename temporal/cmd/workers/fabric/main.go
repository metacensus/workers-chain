package main

import (
	"log"

	"go.uber.org/fx"

	"github.com/metacensus/infra/temporal/lib/workers"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/metacensus/infra/temporal/lib/workers/fabric"
)

func main() {
	// Initialize configuration from environment.
	// The shared fabric worker is operator-owned: OPERATOR_ID is required
	// (loaded via fx in the fabric module), no ORG_ID.
	config.InitializeTemporal()
	config.InitializeVault()
	config.InitializeKubernetes()
	config.InitializeWorker()

	cfg := config.GetTemporal()
	log.Println("Connecting to Temporal at", cfg.Address)

	app := fx.New(
		// Include fabric worker module
		fabric.Module,

		// Start healthcheck server (shared)
		fx.Invoke(workers.StartHealthcheck),

		// Run the worker (shared)
		fx.Invoke(workers.RunWorker),
	)

	app.Run()
}
