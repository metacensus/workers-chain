package vault

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/activity"
)

// Vault path components
const (
	authKubernetesRolePath = "auth/kubernetes/role"
	sysPoliciesACLPath     = "sys/policies/acl"
)

// Vault API field names
const (
	vaultFieldData                          = "data"
	vaultFieldMetadata                      = "metadata"
	vaultFieldVersion                       = "version"
	vaultFieldPolicy                        = "policy"
	vaultFieldBoundServiceAccountNames      = "bound_service_account_names"
	vaultFieldBoundServiceAccountNamespaces = "bound_service_account_namespaces"
	vaultFieldPolicies                      = "policies"
)

// buildAuthRolePath constructs the Vault path for a Kubernetes auth role
func buildAuthRolePath(roleName string) string {
	return fmt.Sprintf("%s/%s", authKubernetesRolePath, roleName)
}

// buildPolicyPath constructs the Vault path for a policy
func buildPolicyPath(policyName string) string {
	return fmt.Sprintf("%s/%s", sysPoliciesACLPath, policyName)
}

// Activities implements the VaultActivities interface from generated code
type Activities struct {
	client VaultClient
}

// NewActivities creates a new Activities instance
func NewActivities(client VaultClient) *Activities {
	return &Activities{
		client: client,
	}
}

// CreateAuthRole creates a Kubernetes auth role in Vault
func (a *Activities) CreateAuthRole(ctx context.Context, req *pb.CreateAuthRoleInput) (*pb.CreateAuthRoleOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating auth role",
		"organization", req.Organization,
		"service_account", req.ServiceAccountName,
		"namespace", req.KubernetesNamespace,
		"policies", req.PolicyNames)

	roleName := fmt.Sprintf("fabric-%s", req.Organization)
	path := buildAuthRolePath(roleName)

	// Prepare request body for Vault Kubernetes auth role
	body := map[string]any{
		vaultFieldBoundServiceAccountNames:      []string{req.ServiceAccountName},
		vaultFieldBoundServiceAccountNamespaces: []string{req.KubernetesNamespace},
		vaultFieldPolicies:                      req.PolicyNames,
	}

	if err := a.client.Write(ctx, path, body); err != nil {
		logger.Error("Failed to create auth role", "error", err)
		return nil, fmt.Errorf("create auth role: %w", err)
	}

	logger.Info("Auth role created successfully", "role_name", roleName)
	return &pb.CreateAuthRoleOutput{
		Name: roleName,
	}, nil
}

// DeleteAuthRole deletes a Kubernetes auth role from Vault
func (a *Activities) DeleteAuthRole(ctx context.Context, req *pb.DeleteAuthRoleInput) (*pb.DeleteAuthRoleOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Deleting auth role", "role_name", req.Name)

	path := buildAuthRolePath(req.Name)

	if err := a.client.Delete(ctx, path); err != nil {
		logger.Error("Failed to delete auth role", "error", err)
		return nil, fmt.Errorf("delete auth role: %w", err)
	}

	logger.Info("Auth role deleted successfully", "role_name", req.Name)
	return &pb.DeleteAuthRoleOutput{}, nil
}

// CreatePolicy creates a new Vault policy
func (a *Activities) CreatePolicy(ctx context.Context, req *pb.CreatePolicyInput) (*pb.CreatePolicyOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating policy", "policy_name", req.Name)

	path := buildPolicyPath(req.Name)

	// Prepare request body for Vault policy
	body := map[string]any{
		vaultFieldPolicy: req.Document,
	}

	if err := a.client.Write(ctx, path, body); err != nil {
		logger.Error("Failed to create policy", "error", err)
		return nil, fmt.Errorf("create policy: %w", err)
	}

	logger.Info("Policy created successfully", "policy_name", req.Name)
	return &pb.CreatePolicyOutput{
		Name: req.Name,
	}, nil
}

// DeletePolicy deletes a Vault policy
func (a *Activities) DeletePolicy(ctx context.Context, req *pb.DeletePolicyInput) (*pb.DeletePolicyOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Deleting policy", "policy_name", req.Name)

	path := buildPolicyPath(req.Name)

	if err := a.client.Delete(ctx, path); err != nil {
		logger.Error("Failed to delete policy", "error", err)
		return nil, fmt.Errorf("delete policy: %w", err)
	}

	logger.Info("Policy deleted successfully", "policy_name", req.Name)
	return &pb.DeletePolicyOutput{}, nil
}

// StoreSecret stores a secret in Vault at the specified path
func (a *Activities) StoreSecret(ctx context.Context, req *pb.StoreSecretInput) (*pb.StoreSecretOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Storing secret", "path", req.Path)

	// Vault KV v2 path structure: <mount>/data/<path>
	vaultPath := fmt.Sprintf("secret/data/%s", req.Path)

	// Convert byte data to strings for Vault KV v2 storage
	dataStr := make(map[string]string)
	for k, v := range req.Data {
		dataStr[k] = string(v)
	}

	// Prepare request body with Vault KV v2 format
	body := map[string]any{
		vaultFieldData: dataStr,
	}

	if err := a.client.Write(ctx, vaultPath, body); err != nil {
		logger.Error("Failed to store secret", "error", err)
		return nil, fmt.Errorf("store secret: %w", err)
	}

	logger.Info("Secret stored successfully", "path", req.Path)

	return &pb.StoreSecretOutput{
		SecretRef: &pb.SecretRef{
			Path:    req.Path,
			Version: 1, // Initial version
		},
	}, nil
}

// FetchSecret retrieves a secret from Vault
func (a *Activities) FetchSecret(ctx context.Context, req *pb.FetchSecretInput) (*pb.FetchSecretOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Fetching secret", "path", req.Path)

	// Vault KV v2 path structure: <mount>/data/<path>
	vaultPath := fmt.Sprintf("secret/data/%s", req.Path)

	// Add version parameter to path if specified
	if req.Version > 0 {
		vaultPath = fmt.Sprintf("%s?version=%d", vaultPath, req.Version)
	}

	result, err := a.client.Read(ctx, vaultPath)
	if err != nil {
		// Check if it's a not found error
		if errors.Is(err, ErrNotFound) {
			logger.Warn("Secret not found", "path", req.Path)
			return nil, fmt.Errorf("fetch secret: %w", err)
		}
		logger.Error("Failed to fetch secret", "error", err)
		return nil, fmt.Errorf("fetch secret: %w", err)
	}

	// Parse Vault KV v2 response structure: {data: {data: {...}, metadata: {version: N}}}
	dataSection, ok := result[vaultFieldData].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("fetch secret: invalid response format")
	}

	secretData, ok := dataSection[vaultFieldData].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("fetch secret: invalid data format")
	}

	// Extract version from metadata if available
	version := req.Version
	if metadata, ok := dataSection[vaultFieldMetadata].(map[string]any); ok {
		if v, ok := metadata[vaultFieldVersion].(float64); ok {
			version = int64(v)
		}
	}

	// Convert strings to bytes
	data := make(map[string][]byte)
	for k, v := range secretData {
		if strVal, ok := v.(string); ok {
			data[k] = []byte(strVal)
		}
	}

	logger.Info("Secret fetched successfully",
		"path", req.Path,
		"data_present", len(data) > 0)

	return &pb.FetchSecretOutput{
		SecretRef: &pb.SecretRef{
			Path:    req.Path,
			Version: version,
		},
		Data: data,
	}, nil
}

// DeleteSecret deletes a secret from Vault
func (a *Activities) DeleteSecret(ctx context.Context, req *pb.DeleteSecretInput) (*pb.DeleteSecretOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Deleting secret", "path", req.Path)

	// Vault KV v2 path structure: <mount>/data/<path>
	vaultPath := fmt.Sprintf("secret/data/%s", req.Path)

	if err := a.client.Delete(ctx, vaultPath); err != nil {
		logger.Error("Failed to delete secret", "error", err)
		return nil, fmt.Errorf("delete secret: %w", err)
	}

	logger.Info("Secret deleted successfully", "path", req.Path)

	return &pb.DeleteSecretOutput{}, nil
}
