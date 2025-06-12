package util

import (
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

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

func ActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		HeartbeatTimeout:    5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    5 * time.Minute,
			MaximumAttempts:    3,
		},
	}
}
