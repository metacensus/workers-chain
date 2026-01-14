package workers

import (
	"context"
	"log"

	"github.com/metacensus/infra/temporal/lib/healthcheck"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Shared worker infrastructure

// NewWorker creates a Temporal worker from the client and environment configuration
func NewWorker(c client.Client, cfg *config.TemporalConfig) worker.Worker {
	return worker.New(c, cfg.TaskQueue, worker.Options{})
}

// Shared lifecycle functions

// StartHealthcheck starts the healthcheck HTTP server in a goroutine
func StartHealthcheck(lc fx.Lifecycle, c client.Client, cfg *config.WorkerConfig) {
	s := healthcheck.NewServer(c)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go s.Run(cfg.HealthCheckAddress)
			log.Println("Healthcheck server started at", cfg.HealthCheckAddress)
			return nil
		},
	})
}

// RunWorker runs the Temporal worker until interrupted
func RunWorker(lc fx.Lifecycle, w worker.Worker, temporalCfg *config.TemporalConfig) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start worker in background
			go func() {
				log.Println("Starting Temporal worker on task queue:", temporalCfg.TaskQueue)
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
