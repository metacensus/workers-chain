package workers

import (
	"context"
	"log"

	"github.com/metacensus/infra/temporal/lib/env"
	"github.com/metacensus/infra/temporal/lib/healthcheck"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

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
				log.Println("Starting Temporal worker on task queue:", env.TaskQueue())
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
