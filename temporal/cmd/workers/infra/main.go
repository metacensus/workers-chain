package main

import (
	"log"

	"go.uber.org/fx"

	"github.com/metacensus/infra/temporal/lib/env"
	"github.com/metacensus/infra/temporal/lib/workers"
)

func main() {
	address := env.TemporalAddress()
	log.Println("Connecting to Temporal at", address)

	app := fx.New(
		// Include all workflow modules (client, worker, vault, identity, etc.)
		workers.InfraModule,

		// Start healthcheck server
		fx.Invoke(workers.StartHealthcheck),

		// Run the worker
		fx.Invoke(workers.RunWorker),
	)

	app.Run()
}
