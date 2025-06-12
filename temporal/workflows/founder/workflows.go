package founder

import (
	"github.com/metacensus/infra/temporal/workflows/util"
	"go.temporal.io/sdk/worker"
)

func RegisterWorkflows(w worker.Worker) {
	util.RegisterWorkflow(w, Workflow, WorkflowName)
	util.RegisterActivity(w, Activity, ActivityName)
}
