package certificate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	logger.Info("Creating Fabric CA client for registration", "url", params.CAEndpoint)

	client := &lib.Client{
		Config: &lib.ClientConfig{
			URL: params.CAEndpoint,
		},
		HomeDir: "/tmp/fabric-ca-client", // Set a temporary home directory
	}

	err := client.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Fabric CA client: %w", err)
	}

	// First, we need to enroll as admin to get the admin identity
	adminEnrollReq := &api.EnrollmentRequest{
		Name:   "admin",
		Secret: "adminpw", // Default Fabric CA admin password
	}

	adminResp, err := client.Enroll(adminEnrollReq)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll as admin: %w", err)
	}

	adminIdentity := adminResp.Identity

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

	// Register the identity using the admin identity
	resp, err := adminIdentity.Register(req)
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
		Secret:      resp.Secret,
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
	logger.Info("Creating Fabric CA client for enrollment", "url", params.CAEndpoint)

	client := &lib.Client{
		Config: &lib.ClientConfig{
			URL: params.CAEndpoint,
		},
		HomeDir: "/tmp/fabric-ca-client", // Set a temporary home directory
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

	// Get the X509 credential
	x509Cred := identity.GetX509Credential()
	if x509Cred == nil {
		return nil, fmt.Errorf("no X509 credential found in enrollment response")
	}

	// Get the signer
	signer := identity.GetECert()
	if signer == nil {
		return nil, fmt.Errorf("no signer found in enrollment response")
	}

	// Get certificate
	cert := signer.GetX509Cert()
	if cert == nil {
		return nil, fmt.Errorf("no certificate found in signer")
	}

	certBytes := cert.Raw

	// Get private key (this is a simplified approach)
	// In a real implementation, you'd need to extract the key properly
	// from the BCCSP or use the in-memory key generation functions
	keyBytes := []byte{} // TODO: Extract key properly

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

	// Create Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create MSP directory structure in the secret
	secretData := make(map[string][]byte)

	// Add the provided data
	for key, value := range params.Data {
		secretData[key] = value
	}

	// Create the secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: "fabric", // TODO: Make this configurable
			Labels:    params.Labels,
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}

	// Create or update the secret
	_, err = clientset.CoreV1().Secrets("fabric").Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		// If secret already exists, update it
		if strings.Contains(err.Error(), "already exists") {
			_, err = clientset.CoreV1().Secrets("fabric").Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to update existing secret: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create secret: %w", err)
		}
	}

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
