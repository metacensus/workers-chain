package component

import (
	"github.com/metacensus/infra/temporal/workflows/util"
	"go.temporal.io/sdk/worker"
)

// RegisterWorkflows registers all workflows with the worker
func RegisterWorkflows(w worker.Worker) {

	// Register workflows
	util.RegisterWorkflow(w, BuildAndApply, BuildAndApplyName)
}
