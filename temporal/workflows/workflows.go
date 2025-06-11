package workflows

import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/metacensus/infra/temporal/workflows/founder"
)

func NewWorker(c client.Client, taskQueue string, opts worker.Options) worker.Worker {
	w := worker.New(c, taskQueue, opts)

	RegisterWorkflow(w, founder.Workflow, "founder.Workflow")
	RegisterActivity(w, founder.Activity, "founder.Activity")

	return w
}

func RegisterWorkflow(w worker.Worker, wf interface{}, name string) {
	w.RegisterWorkflowWithOptions(wf, workflow.RegisterOptions{
		Name: name,
	})
}

func RegisterActivity(w worker.Worker, a interface{}, name string) {
	w.RegisterActivityWithOptions(a, activity.RegisterOptions{
		Name: name,
	})
}
