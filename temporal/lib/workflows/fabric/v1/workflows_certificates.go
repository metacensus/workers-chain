package fabric

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	postgrespb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// computeTLSCommonName returns the CN for a TLS certificate, adding "-tls" suffix
// when both signing and TLS certs come from the same CA to avoid registration collision.
func computeTLSCommonName(baseName string, signingIssuerRef, tlsIssuerRef *pb.CertificateAuthorityRef) string {
	// Self-signed cases: no collision (not registering with a CA)
	if signingIssuerRef == nil || tlsIssuerRef == nil {
		return baseName
	}
	// Same CA issues both - need different identity name
	if signingIssuerRef.GetCertType() == tlsIssuerRef.GetCertType() &&
		signingIssuerRef.GetId() == tlsIssuerRef.GetId() {
		return baseName + "-tls"
	}
	return baseName
}

func (w *Workflows) CreateCertificateDatabase(ctx workflow.Context, input *pb.CreateCertificateDatabaseWorkflowInput) (pb.CreateCertificateDatabaseWorkflow, error) {
	return newCreateCertificateDatabaseWorkflow(input)
}

type createCertificateDatabaseWorkflow struct {
	input *pb.CreateCertificateDatabaseWorkflowInput
}

// newCreateCertificateDatabaseWorkflow creates and validates a new createCertificateDatabaseWorkflow
func newCreateCertificateDatabaseWorkflow(input *pb.CreateCertificateDatabaseWorkflowInput) (*createCertificateDatabaseWorkflow, error) {
	w := &createCertificateDatabaseWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createCertificateDatabaseWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	caRef := w.input.Req.GetCaRef()
	if caRef == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	certType := caRef.GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	return nil
}

func (w *createCertificateDatabaseWorkflow) Execute(ctx workflow.Context) (*pb.CreateCertificateDatabaseOutput, error) {
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()

	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	if certType == "tls" {
		ownerConfig = operatorConfig
	} else {
		ownerConfig = orgConfig
	}

	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, caRef)
	certStore := fabricconfig.NewCertificateStoreConfig(orgConfig)

	// Wait for certificate store to be ready
	waitOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    0, // Infinite retry
		},
	}
	waitCtx := workflow.WithActivityOptions(ctx, waitOptions)

	err := workflow.ExecuteActivity(waitCtx, kubepb.GetStatusActivityName, &kubepb.GetStatusInput{
		Resources: map[string]*kubepb.ResourceIdentifier{
			"certificate-store": certStore.ResourceIdentifier(),
		},
	}).Get(waitCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("certificate store did not become ready: %w", err)
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Create the database
	err = workflow.ExecuteActivity(activityCtx, postgrespb.CreateDatabaseActivityName, &postgrespb.CreateDatabaseInput{
		DatabaseName: caConfig.DatabaseName(),
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Generate credentials
	credentialsRef := caConfig.VaultDatabaseUserRef()
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: credentialsRef,
		Username:  caConfig.DatabaseUsername(),
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate credentials: %w", err)
	}

	// Create user
	err = workflow.ExecuteActivity(activityCtx, postgrespb.CreateUserActivityName, &postgrespb.CreateUserInput{
		CredentialsRef: credentialsRef,
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Grant privileges
	err = workflow.ExecuteActivity(activityCtx, postgrespb.GrantPrivilegesActivityName, &postgrespb.GrantPrivilegesInput{
		Username:     caConfig.DatabaseUsername(),
		DatabaseName: caConfig.DatabaseName(),
		Privileges:   []postgrespb.Privilege{postgrespb.Privilege_PRIVILEGE_ALL},
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to grant privileges: %w", err)
	}

	return &pb.CreateCertificateDatabaseOutput{
		CredentialsRef: credentialsRef,
	}, nil
}
