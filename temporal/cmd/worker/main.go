package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/metacensus/infra/temporal/env"
	"github.com/metacensus/infra/temporal/healthcheck"
	"github.com/metacensus/infra/temporal/workflows"
)

func main() {
	address := env.TemporalAddress()
	log.Println("Connecting to Temporal at", address)

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: address,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	s := healthcheck.NewServer(c)
	go s.Run(env.HealthCheckAddress())

	w := workflows.NewWorker(c, env.InfraTaskQueue(), worker.Options{})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
