package founder

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

const WorkflowName = "founder.Workflow"

// Workflow is the main founder workflow that orchestrates infrastructure setup
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("founder workflow started", "name", name)

	// Execute the original activity
	var result string
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	// TODO: Trigger Fabric CA bootstrap using fabric/identity/v1 workflows
	// This will be implemented once the bootstrap workflow is ready

	logger.Info("founder workflow completed.", "result", result)

	return result, nil
}

const ActivityName = "founder.Activity"

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)
	return "Hello " + name + "!", nil
}
