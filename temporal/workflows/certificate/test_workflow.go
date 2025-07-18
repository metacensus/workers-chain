package certificate

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	TestFabricCAWorkflowName = "certificate.TestFabricCAWorkflow"
)

// TestFabricCAWorkflow tests the basic Fabric CA integration
func TestFabricCAWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Fabric CA test workflow started")

	// Test identity registration
	var reg Registration
	err := workflow.ExecuteChildWorkflow(ctx, RegisterIdentityWorkflow, IdentityParams{
		Type:       "client",
		Name:       "test-client",
		CAEndpoint: "http://ca0.fabric.svc.cluster.local:7054",
		Attributes: map[string]string{
			"hf.Registrar.Roles": "client",
		},
		Affiliation: "org1.department1",
		Secret:      "testpw",
	}).Get(ctx, &reg)
	if err != nil {
		logger.Error("Test identity registration failed", "error", err)
		return fmt.Errorf("test registration failed: %w", err)
	}

	logger.Info("Test identity registration successful", "id", reg.ID)

	// Test enrollment
	var enrollment Enrollment
	err = workflow.ExecuteChildWorkflow(ctx, EnrollIdentityWorkflow, EnrollmentParams{
		Registration: reg,
		Profile:      "sign",
		CAEndpoint:   "http://ca0.fabric.svc.cluster.local:7054",
		CSRConfig: CSRConfig{
			CommonName:   "test-client",
			Organization: "Metacensus",
			Country:      "US",
			Province:     "CA",
			Locality:     "San Francisco",
		},
	}).Get(ctx, &enrollment)
	if err != nil {
		logger.Error("Test enrollment failed", "error", err)
		return fmt.Errorf("test enrollment failed: %w", err)
	}

	logger.Info("Test enrollment successful", "profile", enrollment.Profile)

	// Test secret storage
	var secret SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, StoreSecretWorkflow, SecretParams{
		Name: "test-client-secret",
		Data: enrollment.ToSecretData(),
		Labels: map[string]string{
			"fabric-component": "client",
			"organization":     "metacensus",
			"managed-by":       "temporal",
			"test":             "true",
		},
		Version: time.Now().Unix(),
	}).Get(ctx, &secret)
	if err != nil {
		logger.Error("Test secret storage failed", "error", err)
		return fmt.Errorf("test secret storage failed: %w", err)
	}

	logger.Info("Test secret storage successful", "secret", secret.Name)

	logger.Info("Fabric CA test workflow completed successfully")
	return nil
}
