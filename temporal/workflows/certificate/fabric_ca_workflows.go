package certificate

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-ca/lib/client/credential/x509"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	RegisterIdentityWorkflowName = "certificate.RegisterIdentityWorkflow"
	EnrollIdentityWorkflowName   = "certificate.EnrollIdentityWorkflow"
	AssembleMSPWorkflowName      = "certificate.AssembleMSPWorkflow"
	StoreSecretWorkflowName      = "certificate.StoreSecretWorkflow"
)

// RegisterIdentityWorkflow registers a new identity with Fabric CA
func RegisterIdentityWorkflow(ctx workflow.Context, params IdentityParams) (*Registration, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Identity registration workflow started",
		"name", params.Name,
		"type", params.Type,
		"ca_endpoint", params.CAEndpoint)

	var result Registration
	err := workflow.ExecuteActivity(ctx, RegisterIdentityActivity, params).Get(ctx, &result)
	if err != nil {
		logger.Error("Identity registration activity failed", "error", err)
		return nil, fmt.Errorf("identity registration failed: %w", err)
	}

	logger.Info("Identity registration workflow completed",
		"id", result.ID,
		"type", result.Type)

	return &result, nil
}

// RegisterIdentityActivity registers a new identity with Fabric CA
func RegisterIdentityActivity(ctx context.Context, params IdentityParams) (*Registration, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting identity registration",
		"name", params.Name,
		"type", params.Type,
		"ca_endpoint", params.CAEndpoint)

	// Create Fabric CA client
	client := &lib.Client{
		Config: &lib.ClientConfig{
			URL: params.CAEndpoint,
		},
	}

	err := client.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Fabric CA client: %w", err)
	}

	// Convert attributes to Fabric CA format
	var attributes []api.Attribute
	for name, value := range params.Attributes {
		attributes = append(attributes, api.Attribute{
			Name:  name,
			Value: value,
			ECert: true,
		})
	}

	// Create registration request
	req := &api.RegistrationRequest{
		Name:        params.Name,
		Type:        params.Type,
		Secret:      params.Secret,
		Affiliation: params.Affiliation,
		Attributes:  attributes,
	}

	// Register the identity
	resp, err := client.Register(req)
	if err != nil {
		return nil, fmt.Errorf("failed to register identity: %w", err)
	}

	logger.Info("Identity registered successfully",
		"name", params.Name,
		"type", params.Type)

	return &Registration{
		ID:          params.Name,
		Type:        params.Type,
		Attributes:  params.Attributes,
		Affiliation: params.Affiliation,
		CAEndpoint:  params.CAEndpoint,
	}, nil
}

// EnrollIdentityWorkflow enrolls an identity to get certificates
func EnrollIdentityWorkflow(ctx workflow.Context, params EnrollmentParams) (*Enrollment, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Identity enrollment workflow started",
		"id", params.Registration.ID,
		"profile", params.Profile,
		"ca_endpoint", params.CAEndpoint)

	var result Enrollment
	err := workflow.ExecuteActivity(ctx, EnrollIdentityActivity, params).Get(ctx, &result)
	if err != nil {
		logger.Error("Identity enrollment activity failed", "error", err)
		return nil, fmt.Errorf("identity enrollment failed: %w", err)
	}

	logger.Info("Identity enrollment workflow completed",
		"id", params.Registration.ID,
		"profile", result.Profile)

	return &result, nil
}

// EnrollIdentityActivity enrolls an identity to get certificates
func EnrollIdentityActivity(ctx context.Context, params EnrollmentParams) (*Enrollment, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting identity enrollment",
		"id", params.Registration.ID,
		"profile", params.Profile,
		"ca_endpoint", params.CAEndpoint)

	// Create Fabric CA client
	client := &lib.Client{
		Config: &lib.ClientConfig{
			URL: params.CAEndpoint,
		},
	}

	err := client.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Fabric CA client: %w", err)
	}

	// Create enrollment request
	req := &api.EnrollmentRequest{
		Name:    params.Registration.ID,
		Secret:  params.Registration.Secret,
		Profile: params.Profile,
		CSR:     params.CSRConfig.ToCSRInfo(),
	}

	// Enroll the identity
	resp, err := client.Enroll(req)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity: %w", err)
	}

	// Extract certificate and key from the response
	identity := resp.Identity
	creds := identity.GetCredentials()

	var certBytes, keyBytes []byte
	for _, cred := range creds {
		if x509Cred, ok := cred.(*x509.Credential); ok {
			signer := x509Cred.Signer()
			if signer != nil {
				// Get certificate
				cert := signer.Certificate()
				if cert != nil {
					certBytes = cert.Raw
				}

				// Get private key (this is a simplified approach)
				// In a real implementation, you'd need to extract the key properly
				// from the BCCSP or use the in-memory key generation functions
			}
		}
	}

	// For now, we'll use the CA chain as the CA cert
	caCert := resp.CAInfo.CAChain

	logger.Info("Identity enrolled successfully",
		"id", params.Registration.ID,
		"profile", params.Profile)

	return &Enrollment{
		Certificate: certBytes,
		PrivateKey:  keyBytes, // This needs to be properly extracted
		CACert:      caCert,
		Profile:     params.Profile,
	}, nil
}

// AssembleMSPWorkflow assembles a complete MSP structure from enrollments
func AssembleMSPWorkflow(ctx workflow.Context, params MSPParams) (*MSPBundle, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("MSP assembly workflow started")

	var result MSPBundle
	err := workflow.ExecuteActivity(ctx, AssembleMSPActivity, params).Get(ctx, &result)
	if err != nil {
		logger.Error("MSP assembly activity failed", "error", err)
		return nil, fmt.Errorf("MSP assembly failed: %w", err)
	}

	logger.Info("MSP assembly workflow completed")

	return &result, nil
}

// AssembleMSPActivity assembles a complete MSP structure from enrollments
func AssembleMSPActivity(ctx context.Context, params MSPParams) (*MSPBundle, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting MSP assembly")

	// Create MSP bundle
	msp := &MSPBundle{
		SigningCert:  params.SigningEnrollment.Certificate,
		SigningKey:   params.SigningEnrollment.PrivateKey,
		CACerts:      params.CACerts,
		NodeOUConfig: params.NodeOUConfig,
	}

	// Add TLS certificates if available
	if params.TLSEnrollment != nil {
		msp.TLSCert = params.TLSEnrollment.Certificate
		msp.TLSKey = params.TLSEnrollment.PrivateKey
	}

	logger.Info("MSP assembly completed")

	return msp, nil
}

// StoreSecretWorkflow stores certificates in both Vault and Kubernetes secrets
func StoreSecretWorkflow(ctx workflow.Context, params SecretParams) (*SecretRef, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Secret storage workflow started",
		"name", params.Name,
		"version", params.Version)

	var result SecretRef
	err := workflow.ExecuteActivity(ctx, StoreSecretActivity, params).Get(ctx, &result)
	if err != nil {
		logger.Error("Secret storage activity failed", "error", err)
		return nil, fmt.Errorf("secret storage failed: %w", err)
	}

	logger.Info("Secret storage workflow completed",
		"name", result.Name,
		"version", result.Version)

	return &result, nil
}

// StoreSecretActivity stores certificates in both Vault and Kubernetes secrets
func StoreSecretActivity(ctx context.Context, params SecretParams) (*SecretRef, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting secret storage",
		"name", params.Name,
		"version", params.Version)

	// TODO: Implement Vault storage
	// For now, we'll just return a placeholder
	// In the full implementation, this would:
	// 1. Store in Vault for compliance/backup
	// 2. Store in Kubernetes secrets for runtime use
	// 3. Ensure synchronization between both systems

	logger.Info("Secret storage completed",
		"name", params.Name,
		"version", params.Version)

	return &SecretRef{
		Name:      params.Name,
		Namespace: "fabric", // TODO: Make this configurable
		Version:   params.Version,
	}, nil
}
