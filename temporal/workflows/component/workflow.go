package component

import (
	"fmt"
	"time"

	"github.com/metacensus/infra/temporal/workflows/util"
	"go.temporal.io/sdk/workflow"
)

const BuildAndApplyName = "component.BuildAndApply"

// BuildAndApplyInput defines the input for the deployment workflow
type BuildAndApplyInput struct {
	RepoURL         string
	TargetComponent string
	Version         string
	Namespace       string
}

// BuildAndApplyOutput defines the output from the deployment workflow
type BuildAndApplyOutput struct {
	DeploymentTime time.Time
	Success        bool
	Message        string
}

// BuildAndApply orchestrates building and applying kubernetes manifests
func BuildAndApply(ctx workflow.Context, input BuildAndApplyInput) (*BuildAndApplyOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting deployment workflow",
		"component", input.TargetComponent,
		"version", input.Version)

	output := &BuildAndApplyOutput{}

	// Step 1: Prepare and build manifests
	prepareInput := BuildInput{
		RepoURL:         input.RepoURL,
		TargetComponent: input.TargetComponent,
		Version:         input.Version,
	}

	buildCtx := workflow.WithActivityOptions(ctx, util.ActivityOptions())
	err := workflow.ExecuteActivity(buildCtx, BuildActivityName, prepareInput).Get(ctx, nil)
	if err != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Failed to prepare manifests: %v", err)
		return output, err
	}

	logger.Info("Manifests prepared and stored")

	// Step 2: Apply manifests to kubernetes
	applyInput := ApplyInput{
		Component: input.TargetComponent,
		Version:   input.Version,
		Namespace: input.Namespace,
	}

	// Use a different timeout for apply as it might take longer
	applyCtx := workflow.WithActivityOptions(ctx, util.ActivityOptions())
	err = workflow.ExecuteActivity(applyCtx, ApplyActivityName, applyInput).Get(ctx, nil)
	if err != nil {
		output.Success = false
		output.Message = fmt.Sprintf("Failed to apply manifests: %v", err)
		return output, err
	}

	// Record deployment time using workflow time (deterministic)
	output.DeploymentTime = workflow.Now(ctx)
	output.Success = true
	output.Message = fmt.Sprintf("Successfully deployed %s version %s to namespace %s",
		input.TargetComponent, input.Version, input.Namespace)

	logger.Info("Deployment completed successfully",
		"component", input.TargetComponent,
		"version", input.Version,
		"namespace", input.Namespace)

	return output, nil
}
