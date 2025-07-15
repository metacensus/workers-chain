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

	// Trigger Fabric certificates generation
	certReq := certificate.FabricCertificatesRequest{
		Namespace:    "fabric",
		Organization: "Metacensus",
		Country:      "US",
		Province:     "CA",
		Locality:     "San Francisco",
		ValidDays:    365,
	}

	var certResult certificate.FabricCertificatesResult
	err = workflow.ExecuteChildWorkflow(ctx, certificate.FabricCertificatesWorkflow, certReq).Get(ctx, &certResult)
	if err != nil {
		logger.Error("Fabric certificates generation failed", "error", err)
		return "", err
	}

	logger.Info("Fabric certificates generated successfully",
		"orderers", len(certResult.OrdererResults),
		"peers", len(certResult.PeerResults),
		"cas", len(certResult.CAResults))

	logger.Info("founder workflow completed.", "result", result)

	return result, nil
}

const ActivityName = "founder.Activity"

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)
	return "Hello " + name + "!", nil
}
