package main

import (
	"log"

	"go.uber.org/fx"

	"github.com/metacensus/infra/temporal/lib/workers"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/metacensus/infra/temporal/lib/workers/infra"
)

func main() {
	// Initialize configuration from environment
	config.InitializeTemporal()
	config.InitializeVault()
	config.InitializeKubernetes()
	config.InitializeWorker()
	if err := config.InitializeOperator(); err != nil {
		log.Fatalf("Failed to initialize operator config: %v", err)
	}

	cfg := config.GetTemporal()
	log.Println("Connecting to Temporal at", cfg.Address)

	app := fx.New(
		// Include infra worker module
		infra.Module,

		// Start healthcheck server (shared)
		fx.Invoke(workers.StartHealthcheck),

		// Run the worker (shared)
		fx.Invoke(workers.RunWorker),
	)

	app.Run()
}
