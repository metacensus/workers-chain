package certificate

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	TestFabricCAWorkflowName = "certificate.TestFabricCAWorkflow"
)

// TestFabricCAWorkflow tests the Fabric CA integration with admin bootstrap
func TestFabricCAWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Fabric CA integration test")

	// Test the admin bootstrap workflow
	caEndpoint := "http://localhost:7054" // Use localhost for testing
	logger.Info("Setting CA endpoint", "url", caEndpoint, "url_length", len(caEndpoint))

	bootstrapConfig := BootstrapConfig{
		CAEndpoint: caEndpoint,
		OrgNodeOUs: NodeOUConfig{
			Enable: true,
			Client: map[string]string{
				"certificateAuthorities": "true",
				"clientOU":               "true",
			},
			Peer: map[string]string{
				"certificateAuthorities": "true",
				"peerOU":                 "true",
			},
			Admin: map[string]string{
				"certificateAuthorities": "true",
				"adminOU":                "true",
			},
			Orderer: map[string]string{
				"certificateAuthorities": "true",
				"ordererOU":              "true",
			},
		},
		Namespace: "fabric",
	}

	// Execute the bootstrap workflow
	err := workflow.ExecuteChildWorkflow(ctx, BootstrapMetacensusOrgWorkflow, bootstrapConfig).Get(ctx, nil)
	if err != nil {
		logger.Error("Bootstrap workflow failed", "error", err)
		return fmt.Errorf("bootstrap workflow failed: %w", err)
	}

	logger.Info("Fabric CA integration test completed successfully")
	return nil
}
