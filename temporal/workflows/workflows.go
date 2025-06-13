package workflows

import (
	"github.com/metacensus/infra/temporal/workflows/component"
	"github.com/metacensus/infra/temporal/workflows/founder"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func NewWorker(c client.Client, taskQueue string, opts worker.Options) worker.Worker {
	w := worker.New(c, taskQueue, opts)

	component.RegisterActivities(w)

	founder.RegisterWorkflows(w)

	return w
}
