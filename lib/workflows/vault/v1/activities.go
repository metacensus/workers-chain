package vault

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pb "github.com/metacensus/workers-chain/gen/workflows/vault/v1"
)

// Sentinel errors for vault activities
var (
	ErrCreateAuthRole     = errors.New("create auth role")
	ErrDeleteAuthRole     = errors.New("delete auth role")
	ErrConvertPolicyToHCL = errors.New("convert policy to HCL")
	ErrCreatePolicy       = errors.New("create policy")
	ErrDeletePolicy       = errors.New("delete policy")
)

// Vault path components
const (
	authKubernetesRolePath = "auth/kubernetes/role"
	sysPoliciesACLPath     = "sys/policies/acl"
)

// Vault API field names
const (
	vaultFieldPolicy                        = "policy"
	vaultFieldBoundServiceAccountNames      = "bound_service_account_names"
	vaultFieldBoundServiceAccountNamespaces = "bound_service_account_namespaces"
	vaultFieldPolicies                      = "policies"
)

func buildAuthRolePath(roleName string) string {
	return fmt.Sprintf("%s/%s", authKubernetesRolePath, roleName)
}

func buildPolicyPath(policyName string) string {
	return fmt.Sprintf("%s/%s", sysPoliciesACLPath, policyName)
}

// policyDocumentToHCL converts a structured PolicyDocument to Vault HCL format
func policyDocumentToHCL(doc *pb.PolicyDocument) (string, error) {
	if doc == nil {
		return "", errors.New("policy document is nil")
	}
	if len(doc.Rules) == 0 {
		return "", errors.New("policy document must contain at least one rule")
	}

	capabilityToVaultName := func(cap pb.Capability) (string, error) {
		if cap == pb.Capability_CAPABILITY_UNSPECIFIED {
			return "", errors.New("capability cannot be unspecified")
		}
		protoName, ok := pb.Capability_name[int32(cap)]
		if !ok {
			return "", fmt.Errorf("unknown capability: %d", cap)
		}
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

		hcl.WriteString(fmt.Sprintf(`path "%s" {`, rule.Path))
		hcl.WriteString("\n")

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

		if i < len(doc.Rules)-1 {
			hcl.WriteString("\n")
		}
	}

	return hcl.String(), nil
}

// Activities implements the VaultActivities interface
type Activities struct {
	client VaultClient
}

// NewActivities creates a new Activities instance
func NewActivities(client VaultClient) *Activities {
	return &Activities{client: client}
}

// CreateAuthRole creates a Kubernetes auth role in Vault
func (a *Activities) CreateAuthRole(ctx context.Context, req *pb.CreateAuthRoleInput) (*pb.CreateAuthRoleOutput, error) {
	if req.RoleName == "" {
		return nil, errors.New("role name cannot be empty")
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

	path := buildAuthRolePath(req.RoleName)

	body := map[string]any{
		vaultFieldBoundServiceAccountNames:      []string{req.ServiceAccountName},
		vaultFieldBoundServiceAccountNamespaces: []string{req.KubernetesNamespace},
		vaultFieldPolicies:                      req.PolicyNames,
	}

	if err := a.client.Write(ctx, path, body); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAuthRole, err)
	}

	return &pb.CreateAuthRoleOutput{Name: req.RoleName}, nil
}

// DeleteAuthRole deletes a Kubernetes auth role from Vault
func (a *Activities) DeleteAuthRole(ctx context.Context, req *pb.DeleteAuthRoleInput) (*pb.DeleteAuthRoleOutput, error) {
	if req.Name == "" {
		return nil, errors.New("role name cannot be empty")
	}

	path := buildAuthRolePath(req.Name)

	if err := a.client.Delete(ctx, path); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeleteAuthRole, err)
	}

	return &pb.DeleteAuthRoleOutput{}, nil
}

// CreatePolicy creates a new Vault policy
func (a *Activities) CreatePolicy(ctx context.Context, req *pb.CreatePolicyInput) (*pb.CreatePolicyOutput, error) {
	if req.Name == "" {
		return nil, errors.New("policy name cannot be empty")
	}

	hcl, err := policyDocumentToHCL(req.Document)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConvertPolicyToHCL, err)
	}

	path := buildPolicyPath(req.Name)

	body := map[string]any{
		vaultFieldPolicy: hcl,
	}

	if err := a.client.Write(ctx, path, body); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreatePolicy, err)
	}

	return &pb.CreatePolicyOutput{Name: req.Name}, nil
}

// DeletePolicy deletes a Vault policy
func (a *Activities) DeletePolicy(ctx context.Context, req *pb.DeletePolicyInput) (*pb.DeletePolicyOutput, error) {
	if req.Name == "" {
		return nil, errors.New("policy name cannot be empty")
	}

	path := buildPolicyPath(req.Name)

	if err := a.client.Delete(ctx, path); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeletePolicy, err)
	}

	return &pb.DeletePolicyOutput{}, nil
}
