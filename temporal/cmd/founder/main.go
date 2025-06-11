package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/metacensus/infra/temporal/env"
	"github.com/metacensus/infra/temporal/workflows/founder"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: env.TemporalAddress(),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "founder",
		TaskQueue: env.InfraTaskQueue(),
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, founder.Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
