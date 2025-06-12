package component

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/terraform-exec/tfexec"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

const BuildActivityName = "component.Build"

// BuildInput defines input for PrepareAndBuildManifests activity
type BuildInput struct {
	RepoURL         string
	TargetComponent string
	Version         string
}

// PrepareAndBuildManifests clones repo, builds manifests via terraform, and uploads to S3
func (a *Activities) Build(ctx context.Context, input BuildInput) error {
	logger := activity.GetLogger(ctx)

	// Generate working directory with random hash
	hashBytes := make([]byte, 16)
	if _, err := rand.Read(hashBytes); err != nil {
		return temporal.NewApplicationError("failed to generate random hash", "RANDOM_ERROR", err)
	}
	workingDirHash := hex.EncodeToString(hashBytes)

	workDir := filepath.Join(terraformWorkspacesPath, workingDirHash)
	logger.Info("Creating working directory", "path", workDir)

	// Ensure cleanup happens even on error
	defer func() {
		if _, err := os.Stat(workDir); err == nil {
			logger.Info("Cleaning up working directory", "path", workDir)
			if err := os.RemoveAll(workDir); err != nil {
				logger.Error("Failed to cleanup working directory", "error", err)
			}
		}
	}()

	if err := os.MkdirAll(workDir, 0755); err != nil {
		return temporal.NewApplicationError("failed to create working directory", "DIR_ERROR", err)
	}

	logger.Info("Cloning repository", "url", input.RepoURL, "version", input.Version)
	if err := cloneRepo(ctx, workDir, input.RepoURL, input.Version); err != nil {
		return temporal.NewApplicationError("failed to clone repository", "GIT_CLONE_ERROR", err)
	}

	// Find terraform directory for component
	tfDir := filepath.Join(workDir, "terraform", "components", input.TargetComponent)
	if _, err := os.Stat(tfDir); os.IsNotExist(err) {
		return temporal.NewApplicationError("terraform directory not found", "TF_DIR_NOT_FOUND",
			fmt.Errorf("looked in %s", tfDir))
	}

	// Initialize and build manifests with terraform
	logger.Info("Building kubernetes manifests with terraform", "dir", tfDir)
	manifestContent, err := buildManifests(ctx, workDir, tfDir)
	if err != nil {
		return temporal.NewApplicationError("failed to build manifests", "TF_BUILD_ERROR", err)
	}

	err = a.store.Put(ctx, input.TargetComponent, input.Version, manifestContent)
	if err != nil {
		return temporal.NewApplicationError("failed to upload to store", "STORE_UPLOAD_ERROR", err)
	}

	return nil
}

func cloneRepo(ctx context.Context, workDir, repoURL, version string) error {
	// Clone repository
	repo, err := git.PlainClone(workDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Checkout specific version/tag
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try as tag first
	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewTagReferenceName(version),
	}); err != nil {
		// Try as branch next
		if err := worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(version),
		}); err != nil {
			// Finally try as commit hash
			if err := worktree.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(version),
			}); err != nil {
				return fmt.Errorf("failed to checkout version: %w", err)
			}
		}
	}
	return nil
}

func buildManifests(ctx context.Context, workDir, tfDir string) ([]byte, error) {
	tf, err := tfexec.NewTerraform(tfDir, terraformPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create terraform instance: %w", err)
	}

	// Initialize terraform
	if err := tf.Init(ctx, tfexec.Upgrade(true)); err != nil {
		return nil, fmt.Errorf("terraform init failed: %w", err)
	}

	// Plan to generate manifests
	planFile := filepath.Join(tfDir, "tfplan")
	if _, err := tf.Plan(ctx, tfexec.Out(planFile)); err != nil {
		return nil, fmt.Errorf("terraform plan failed: %w", err)
	}

	// Apply to generate the manifests
	if err := tf.Apply(ctx, tfexec.DirOrPlan(planFile)); err != nil {
		return nil, fmt.Errorf("terraform apply failed: %w", err)
	}

	// Get output containing manifest content
	outputRaw, err := tf.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get terraform output: %w", err)
	}

	return []byte(outputRaw["kubernetes_manifests"].Value), nil
}
