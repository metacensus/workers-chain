package certificate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	BootstrapAdminWorkflowName = "certificate.BootstrapAdminWorkflow"
	CreateAdminIdentityName    = "certificate.CreateAdminIdentity"
	StoreAdminIdentityName     = "certificate.StoreAdminIdentity"
	RevokeDefaultAdminName     = "certificate.RevokeDefaultAdmin"
)

// BootstrapAdminWorkflow creates a new admin identity and revokes the default one
func BootstrapAdminWorkflow(ctx workflow.Context, params BootstrapAdminParams) (*BootstrapAdminResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting admin bootstrap workflow",
		"ca_endpoint", params.CAEndpoint,
		"admin_id", params.AdminID)

	// Step 1: Create new admin identity using default admin
	var adminIdentity *AdminIdentity
	err := workflow.ExecuteActivity(ctx, CreateAdminIdentityActivity, CreateAdminIdentityParams{
		CAEndpoint:     params.CAEndpoint,
		AdminID:        params.AdminID,
		AdminSecret:    params.AdminSecret,
		DefaultAdminID: params.DefaultAdminID,
		DefaultSecret:  params.DefaultSecret,
		Attributes:     params.Attributes,
	}).Get(ctx, &adminIdentity)
	if err != nil {
		return &BootstrapAdminResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create admin identity: %v", err),
		}, err
	}

	// Step 2: Store admin identity in Vault
	err = workflow.ExecuteActivity(ctx, StoreAdminIdentityActivity, StoreAdminIdentityParams{
		AdminIdentity: adminIdentity,
		VaultPath:     params.VaultPath,
	}).Get(ctx, nil)
	if err != nil {
		return &BootstrapAdminResult{
			Success: false,
			Message: fmt.Sprintf("Failed to store admin identity: %v", err),
		}, err
	}

	// Step 3: Revoke the default admin identity
	err = workflow.ExecuteActivity(ctx, RevokeDefaultAdminActivity, RevokeDefaultAdminParams{
		CAEndpoint:     params.CAEndpoint,
		DefaultAdminID: params.DefaultAdminID,
		DefaultSecret:  params.DefaultSecret,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to revoke default admin, but continuing", "error", err)
		// Don't fail the workflow if we can't revoke the default admin
	}

	logger.Info("Admin bootstrap completed successfully",
		"admin_id", params.AdminID)

	return &BootstrapAdminResult{
		AdminIdentity: adminIdentity,
		Success:       true,
		Message:       "Admin bootstrap completed successfully",
	}, nil
}

// CreateAdminIdentityParams contains parameters for creating admin identity
type CreateAdminIdentityParams struct {
	CAEndpoint     string            `json:"ca_endpoint"`
	AdminID        string            `json:"admin_id"`
	AdminSecret    string            `json:"admin_secret"`
	DefaultAdminID string            `json:"default_admin_id"`
	DefaultSecret  string            `json:"default_secret"`
	Attributes     map[string]string `json:"attributes"`
}

// StoreAdminIdentityParams contains parameters for storing admin identity
type StoreAdminIdentityParams struct {
	AdminIdentity *AdminIdentity `json:"admin_identity"`
	VaultPath     string         `json:"vault_path"`
}

// RevokeDefaultAdminParams contains parameters for revoking default admin
type RevokeDefaultAdminParams struct {
	CAEndpoint     string `json:"ca_endpoint"`
	DefaultAdminID string `json:"default_admin_id"`
	DefaultSecret  string `json:"default_secret"`
}

// CreateAdminIdentityActivity creates a new admin identity using the default admin
func CreateAdminIdentityActivity(ctx context.Context, params CreateAdminIdentityParams) (*AdminIdentity, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating new admin identity",
		"admin_id", params.AdminID,
		"ca_endpoint", params.CAEndpoint)

	// Create Fabric CA client
	logger.Info("Creating Fabric CA client", "url", params.CAEndpoint)

	// Test direct HTTP connection first
	logger.Info("Testing direct HTTP connection", "url", params.CAEndpoint, "url_length", len(params.CAEndpoint))

	// Debug: Print each character of the URL
	logger.Info("URL characters", "url_bytes", []byte(params.CAEndpoint))

	// Try with hardcoded URL to test if parameter passing is the issue
	hardcodedURL := "http://localhost:7054"
	logger.Info("Testing with hardcoded URL", "url", hardcodedURL)

	// Create a simple HTTP client to test connectivity
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	testURL := hardcodedURL + "/cainfo"
	logger.Info("Full test URL", "url", testURL, "url_length", len(testURL))

	resp, err := httpClient.Get(testURL)
	if err != nil {
		logger.Error("Direct HTTP connection failed", "error", err, "url", testURL, "original_url", params.CAEndpoint)
		// Don't fail the workflow yet, just log the error and continue
		logger.Warn("Continuing with Fabric CA client despite HTTP test failure")
	} else {
		defer resp.Body.Close()
		logger.Info("Direct HTTP connection successful", "status", resp.StatusCode)
	}

	// Create client with direct URL (avoiding normalization issues)
	logger.Info("Creating Fabric CA client with URL", "url", params.CAEndpoint, "url_length", len(params.CAEndpoint))

	client := &lib.Client{
		Config: &lib.ClientConfig{
			URL: params.CAEndpoint,
		},
		HomeDir: "/tmp/fabric-ca-client", // Set a temporary home directory
	}

	// Debug: Check what URL is actually set in the config
	logger.Info("Client config URL after creation", "url", client.Config.URL, "url_length", len(client.Config.URL))

	err = client.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Fabric CA client: %w", err)
	}

	// Debug: Check what URL is set after Init
	logger.Info("Client config URL after Init", "url", client.Config.URL, "url_length", len(client.Config.URL))

	// Step 1: Enroll as default admin
	logger.Info("About to enroll as default admin", "admin_id", params.DefaultAdminID, "client_url", client.Config.URL)

	defaultAdminReq := &api.EnrollmentRequest{
		Name:   params.DefaultAdminID,
		Secret: params.DefaultSecret,
	}

	logger.Info("Enrollment request created", "request_name", defaultAdminReq.Name)

	defaultAdminResp, err := client.Enroll(defaultAdminReq)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll as default admin: %w", err)
	}

	defaultAdminIdentity := defaultAdminResp.Identity

	// Step 2: Register new admin identity
	// Convert attributes to Fabric CA format
	var attributes []api.Attribute
	for name, value := range params.Attributes {
		attributes = append(attributes, api.Attribute{
			Name:  name,
			Value: value,
			ECert: true,
		})
	}

	// Add admin-specific attributes
	attributes = append(attributes, api.Attribute{
		Name:  "hf.Registrar.Roles",
		Value: "admin",
		ECert: true,
	}, api.Attribute{
		Name:  "hf.Registrar.Attributes",
		Value: "*",
		ECert: true,
	})

	registerReq := &api.RegistrationRequest{
		Name:        params.AdminID,
		Type:        "client",
		Secret:      params.AdminSecret,
		Affiliation: "org1.department1",
		Attributes:  attributes,
	}

	registerResp, err := defaultAdminIdentity.Register(registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register new admin identity: %w", err)
	}

	// Step 3: Enroll the new admin identity
	enrollReq := &api.EnrollmentRequest{
		Name:   params.AdminID,
		Secret: registerResp.Secret,
	}

	enrollResp, err := client.Enroll(enrollReq)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll new admin identity: %w", err)
	}

	// Extract certificate and key
	identity := enrollResp.Identity
	signer := identity.GetECert()
	if signer == nil {
		return nil, fmt.Errorf("no signer found in enrollment response")
	}

	cert := signer.GetX509Cert()
	if cert == nil {
		return nil, fmt.Errorf("no certificate found in signer")
	}

	// TODO: Extract private key properly from BCCSP
	// For now, we'll use a placeholder
	keyBytes := []byte{} // TODO: Extract key properly

	logger.Info("Admin identity created successfully",
		"admin_id", params.AdminID)

	return &AdminIdentity{
		ID:          params.AdminID,
		Secret:      registerResp.Secret,
		Certificate: cert.Raw,
		PrivateKey:  keyBytes, // TODO: Extract properly
		CACert:      enrollResp.CAInfo.CAChain,
		Attributes:  params.Attributes,
		CAEndpoint:  params.CAEndpoint,
	}, nil
}

// StoreAdminIdentityActivity stores the admin identity in Vault
func StoreAdminIdentityActivity(ctx context.Context, params StoreAdminIdentityParams) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Storing admin identity in Vault",
		"admin_id", params.AdminIdentity.ID,
		"vault_path", params.VaultPath)

	// TODO: Implement Vault storage
	// For now, we'll just log that we would store it
	logger.Info("Would store admin identity in Vault",
		"path", params.VaultPath,
		"admin_id", params.AdminIdentity.ID)

	return nil
}

// RevokeDefaultAdminActivity revokes the default admin identity
func RevokeDefaultAdminActivity(ctx context.Context, params RevokeDefaultAdminParams) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Revoking default admin identity",
		"default_admin_id", params.DefaultAdminID)

	// Create Fabric CA client
	logger.Info("Creating Fabric CA client for revocation", "url", params.CAEndpoint)

	client := &lib.Client{
		Config: &lib.ClientConfig{
			URL: params.CAEndpoint,
		},
		HomeDir: "/tmp/fabric-ca-client", // Set a temporary home directory
	}

	err := client.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize Fabric CA client: %w", err)
	}

	// Enroll as default admin to get identity for revocation
	enrollReq := &api.EnrollmentRequest{
		Name:   params.DefaultAdminID,
		Secret: params.DefaultSecret,
	}

	enrollResp, err := client.Enroll(enrollReq)
	if err != nil {
		return fmt.Errorf("failed to enroll as default admin for revocation: %w", err)
	}

	// Revoke the default admin
	revokeReq := &api.RevocationRequest{
		Name: params.DefaultAdminID,
	}

	_, err = enrollResp.Identity.Revoke(revokeReq)
	if err != nil {
		return fmt.Errorf("failed to revoke default admin: %w", err)
	}

	logger.Info("Default admin identity revoked successfully",
		"default_admin_id", params.DefaultAdminID)

	return nil
}

// GenerateSecureSecretActivity generates a secure random secret
func GenerateSecureSecretActivity(ctx context.Context, length int) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Generating secure secret", "length", length)

	bytes := make([]byte, length/2) // hex encoding doubles the length
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateSecureSecret generates a secure random secret (for backward compatibility)
func GenerateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length/2) // hex encoding doubles the length
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
