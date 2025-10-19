package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"

	"github.com/metacensus/infra/temporal/env"
	"github.com/metacensus/infra/temporal/healthcheck"
	"github.com/metacensus/infra/temporal/workflows"
)

func main() {
	address := env.TemporalAddress()
	log.Println("Connecting to Temporal at", address)

	app := fx.New(
		// Include all workflow modules (client, worker, vault, identity, etc.)
		workflows.Module,

		// Start healthcheck server
		fx.Invoke(StartHealthcheck),

		// Run the worker
		fx.Invoke(RunWorker),
	)

	app.Run()
}

// StartHealthcheck starts the healthcheck HTTP server in a goroutine
func StartHealthcheck(lc fx.Lifecycle, c client.Client) {
	s := healthcheck.NewServer(c)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go s.Run(env.HealthCheckAddress())
			log.Println("Healthcheck server started at", env.HealthCheckAddress())
			return nil
		},
	})
}

// RunWorker runs the Temporal worker until interrupted
func RunWorker(lc fx.Lifecycle, w worker.Worker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start worker in background
			go func() {
				log.Println("Starting Temporal worker on task queue:", env.InfraTaskQueue())
				err := w.Run(worker.InterruptCh())
				if err != nil {
					log.Fatalln("Worker error:", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping worker...")
			w.Stop()
			return nil
		},
	})
}
