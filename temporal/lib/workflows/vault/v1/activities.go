package vault

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

// policyDocumentToHCL converts a structured PolicyDocument with rules and capabilities
// into Vault HCL policy format suitable for the Vault API.
//
// Example output:
//
//	path "secret/data/fabric/*" {
//	  capabilities = ["read", "list"]
//	}
//	path "auth/kubernetes/login" {
//	  capabilities = ["create"]
//	}
func policyDocumentToHCL(doc *pb.PolicyDocument) (string, error) {
	if doc == nil {
		return "", errors.New("policy document is nil")
	}
	if len(doc.Rules) == 0 {
		return "", errors.New("policy document must contain at least one rule")
	}

	// Map proto enum name to lowercase capability name (e.g., "CAPABILITY_READ" -> "read")
	capabilityToVaultName := func(cap pb.Capability) (string, error) {
		if cap == pb.Capability_CAPABILITY_UNSPECIFIED {
			return "", errors.New("capability cannot be unspecified")
		}
		// Use generated proto name map, then lowercase and trim prefix
		protoName, ok := pb.Capability_name[int32(cap)]
		if !ok {
			return "", fmt.Errorf("unknown capability: %d", cap)
		}
		// "CAPABILITY_READ" -> "read"
		vaultName := strings.ToLower(strings.TrimPrefix(protoName, "CAPABILITY_"))
		return vaultName, nil
	}

	var hcl strings.Builder

	for i, rule := range doc.Rules {
		if rule.Path == "" {
			return "", fmt.Errorf("rule %d: path cannot be empty", i)
		}
		if len(rule.Capabilities) == 0 {
			return "", fmt.Errorf("rule %d (%s): must have at least one capability", i, rule.Path)
		}

		// Write path block
		hcl.WriteString(fmt.Sprintf(`path "%s" {`, rule.Path))
		hcl.WriteString("\n")

		// Write capabilities array
		hcl.WriteString("  capabilities = [")
		for j, cap := range rule.Capabilities {
			vaultName, err := capabilityToVaultName(cap)
			if err != nil {
				return "", fmt.Errorf("rule %d (%s): %w", i, rule.Path, err)
			}
			hcl.WriteString(fmt.Sprintf(`"%s"`, vaultName))
			if j < len(rule.Capabilities)-1 {
				hcl.WriteString(", ")
			}
		}
		hcl.WriteString("]\n")
		hcl.WriteString("}\n")

		// Add blank line between rules (except after last rule)
		if i < len(doc.Rules)-1 {
			hcl.WriteString("\n")
		}
	}

	return hcl.String(), nil
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

	// Validate required fields
	if req.Organization == "" {
		return nil, errors.New("organization cannot be empty")
	}
	if req.ServiceAccountName == "" {
		return nil, errors.New("service account name cannot be empty")
	}
	if req.KubernetesNamespace == "" {
		return nil, errors.New("kubernetes namespace cannot be empty")
	}
	if len(req.PolicyNames) == 0 {
		return nil, errors.New("policy names list cannot be empty")
	}

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

	if req.Name == "" {
		return nil, errors.New("role name cannot be empty")
	}

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

	if req.Name == "" {
		return nil, errors.New("policy name cannot be empty")
	}

	// Convert structured PolicyDocument to HCL format
	hcl, err := policyDocumentToHCL(req.Document)
	if err != nil {
		logger.Error("Failed to convert policy document to HCL", "error", err)
		return nil, fmt.Errorf("create policy: %w", err)
	}

	path := buildPolicyPath(req.Name)

	// Prepare request body for Vault policy
	body := map[string]any{
		vaultFieldPolicy: hcl,
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

	if req.Name == "" {
		return nil, errors.New("policy name cannot be empty")
	}

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
