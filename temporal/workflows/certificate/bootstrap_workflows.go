package certificate

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	BootstrapMetacensusOrgWorkflowName = "certificate.BootstrapMetacensusOrgWorkflow"
)

// BootstrapMetacensusOrgWorkflow bootstraps the metacensus organization with admin identity
func BootstrapMetacensusOrgWorkflow(ctx workflow.Context, config BootstrapConfig) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting metacensus organization bootstrap",
		"ca_endpoint", config.CAEndpoint)

	// Step 1: Bootstrap admin identity
	var adminSecret string
	err := workflow.ExecuteActivity(ctx, GenerateSecureSecretActivity, 32).Get(ctx, &adminSecret)
	if err != nil {
		return fmt.Errorf("failed to generate admin secret: %w", err)
	}

	var bootstrapResult *BootstrapAdminResult
	err = workflow.ExecuteChildWorkflow(ctx, BootstrapAdminWorkflow, BootstrapAdminParams{
		CAEndpoint:     config.CAEndpoint,
		AdminID:        "metacensus-admin",
		AdminSecret:    adminSecret,
		DefaultAdminID: "admin",
		DefaultSecret:  "adminpw", // Default Fabric CA admin password
		Attributes: map[string]string{
			"hf.Registrar.Roles":         "admin",
			"hf.Registrar.Attributes":    "*",
			"hf.Registrar.DelegateRoles": "admin",
		},
		VaultPath: "fabric/metacensus/admin",
	}).Get(ctx, &bootstrapResult)
	if err != nil {
		return fmt.Errorf("failed to bootstrap admin identity: %w", err)
	}

	if !bootstrapResult.Success {
		return fmt.Errorf("admin bootstrap failed: %s", bootstrapResult.Message)
	}

	logger.Info("Admin identity bootstrapped successfully",
		"admin_id", bootstrapResult.AdminIdentity.ID)

	// Step 2: Create organization structure
	// Create affiliations for the organization
	err = workflow.ExecuteActivity(ctx, CreateAffiliationsActivity, CreateAffiliationsParams{
		CAEndpoint:    config.CAEndpoint,
		AdminIdentity: bootstrapResult.AdminIdentity,
		Affiliations: []string{
			"metacensus",
			"metacensus.department1",
			"metacensus.department2",
		},
	}).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create affiliations: %w", err)
	}

	// Step 3: Create initial peers
	peerConfigs := []PeerConfig{
		{
			PeerID:       "peer0",
			Organization: "metacensus",
			EnrollmentCA: config.CAEndpoint,
			TLSCA:        config.CAEndpoint,
			OrgNodeOUs:   config.OrgNodeOUs,
		},
		{
			PeerID:       "peer1",
			Organization: "metacensus",
			EnrollmentCA: config.CAEndpoint,
			TLSCA:        config.CAEndpoint,
			OrgNodeOUs:   config.OrgNodeOUs,
		},
	}

	// Provision peers in parallel
	for _, peerConfig := range peerConfigs {
		pc := peerConfig // Capture the variable
		workflow.Go(ctx, func(ctx workflow.Context) {
			workflow.ExecuteChildWorkflow(ctx, ProvisionPeerWorkflow, pc).Get(ctx, nil)
		})
	}

	// Step 4: Create initial orderers
	ordererConfigs := []OrdererConfig{
		{
			OrdererID:    "orderer0",
			Organization: "metacensus",
			EnrollmentCA: config.CAEndpoint,
			TLSCA:        config.CAEndpoint,
			OrgNodeOUs:   config.OrgNodeOUs,
		},
		{
			OrdererID:    "orderer1",
			Organization: "metacensus",
			EnrollmentCA: config.CAEndpoint,
			TLSCA:        config.CAEndpoint,
			OrgNodeOUs:   config.OrgNodeOUs,
		},
		{
			OrdererID:    "orderer2",
			Organization: "metacensus",
			EnrollmentCA: config.CAEndpoint,
			TLSCA:        config.CAEndpoint,
			OrgNodeOUs:   config.OrgNodeOUs,
		},
	}

	// Provision orderers in parallel
	for _, ordererConfig := range ordererConfigs {
		oc := ordererConfig // Capture the variable
		workflow.Go(ctx, func(ctx workflow.Context) {
			workflow.ExecuteChildWorkflow(ctx, ProvisionOrdererWorkflow, oc).Get(ctx, nil)
		})
	}

	logger.Info("Metacensus organization bootstrap completed successfully")
	return nil
}

// CreateAffiliationsParams contains parameters for creating affiliations
type CreateAffiliationsParams struct {
	CAEndpoint    string         `json:"ca_endpoint"`
	AdminIdentity *AdminIdentity `json:"admin_identity"`
	Affiliations  []string       `json:"affiliations"`
}

// CreateAffiliationsActivity creates affiliations in the Fabric CA
func CreateAffiliationsActivity(ctx context.Context, params CreateAffiliationsParams) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating affiliations",
		"affiliations", params.Affiliations)

	// TODO: Implement affiliation creation using the admin identity
	// This would require using the stored admin identity from Vault
	// For now, we'll just log that we would create them
	logger.Info("Would create affiliations using admin identity",
		"affiliations", params.Affiliations,
		"admin_id", params.AdminIdentity.ID)

	return nil
}
