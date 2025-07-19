package founder

import (
	"context"
	"time"

	"github.com/metacensus/infra/temporal/workflows/certificate"
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

	// Trigger Fabric CA bootstrap using the new workflow
	err = workflow.ExecuteChildWorkflow(ctx, certificate.BootstrapMetacensusOrgWorkflow).Get(ctx, nil)
	if err != nil {
		logger.Error("Fabric CA bootstrap failed", "error", err)
		return "", err
	}

	logger.Info("Fabric CA bootstrap completed successfully")

	logger.Info("founder workflow completed.", "result", result)

	return result, nil
}

const ActivityName = "founder.Activity"

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)
	return "Hello " + name + "!", nil
}
