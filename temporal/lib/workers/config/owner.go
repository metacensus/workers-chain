package config

import (
	"fmt"

	fabricv1 "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/spf13/viper"
)

// OwnerConfig defines the contract for resource ownership (operator or org).
// This interface enables the same component configs (CertificateAuthorityConfig,
// PeerConfig, etc.) to work with both operator-owned and org-owned resources.
//
// Workers that act AS an owner (fabric worker = operator, shard worker = org)
// receive their OwnerConfig via dependency injection at startup; workflows on
// those workers must never guess ownership from inputs.
type OwnerConfig interface {
	// OwnerID returns the unique identifier for this owner (operator or org)
	OwnerID() string
	// IsOperator reports whether this owner is the operator.
	// The operator owns shared infrastructure (TLS CA chain); orgs own
	// their per-shard infrastructure (org CA chain, peers, orderers).
	IsOperator() bool
	// OwnerProto returns the typed fabric.v1.Owner message identifying this
	// owner, for workflows that act on behalf of an explicit owner.
	OwnerProto() *fabricv1.Owner
	// KubernetesNamespace returns the Kubernetes namespace for this owner's resources
	// Example: "fabric" (operator) or "fabric-shard-org1" (org)
	KubernetesNamespace() string
	// VaultBasePath returns the base Vault path for this owner's secrets
	// Example: "fabric/metacensus" (operator) or "fabric/shard/org1" (org)
	VaultBasePath() string
}

// WorkerOwnerConfig extends OwnerConfig with worker-specific naming.
// Both OperatorConfig (shared fabric worker) and OrgConfig (shard workers)
// implement this, letting cluster workflows provision either worker type.
type WorkerOwnerConfig interface {
	OwnerConfig
	// TemporalNamespace returns the Temporal namespace the worker polls
	TemporalNamespace() string
	// WorkerServiceAccountName returns the Kubernetes service account name for the worker
	WorkerServiceAccountName() string
	// WorkerDeploymentName returns the Kubernetes deployment name for the worker
	WorkerDeploymentName() string
	// WorkerVaultPolicyName returns the Vault policy name for the worker
	WorkerVaultPolicyName() string
	// WorkerVaultAuthRole returns the Vault Kubernetes auth role for the worker
	WorkerVaultAuthRole() string
}

// NewOwnerConfig converts a typed fabric.v1.Owner message into the matching
// concrete owner config. Workflows that act on behalf of an explicit owner
// (cluster.fabric.v1) resolve their input Owner exactly once with this.
func NewOwnerConfig(owner *fabricv1.Owner) (WorkerOwnerConfig, error) {
	switch id := owner.GetId().(type) {
	case *fabricv1.Owner_OperatorId:
		if id.OperatorId == "" {
			return nil, fmt.Errorf("owner.operator_id must not be empty")
		}
		return NewOperatorConfig(id.OperatorId), nil
	case *fabricv1.Owner_OrgId:
		if id.OrgId == "" {
			return nil, fmt.Errorf("owner.org_id must not be empty")
		}
		return NewOrgConfig(id.OrgId), nil
	default:
		return nil, fmt.Errorf("owner must specify operator_id or org_id")
	}
}

// OperatorConfig implements OwnerConfig for operator-owned resources.
// Operator-owned resources include shared infrastructure like TLS CAs.
type OperatorConfig struct {
	operatorID string
}

// LoadOperatorConfig reads the OPERATOR_ID environment variable and builds
// the OperatorConfig. Provided to workers via fx at startup.
func LoadOperatorConfig() (*OperatorConfig, error) {
	operatorID := viper.GetString(operatorIDName)
	if operatorID == "" {
		return nil, fmt.Errorf("OPERATOR_ID environment variable is required")
	}
	return &OperatorConfig{operatorID: operatorID}, nil
}

// NewOperatorConfig creates an OperatorConfig from an operator ID.
// This constructor is used when operator ID comes from a typed Owner input or testing.
func NewOperatorConfig(operatorID string) OperatorConfig {
	return OperatorConfig{operatorID: operatorID}
}

func (c OperatorConfig) OwnerID() string { return c.operatorID }

// IsOperator implements OwnerConfig: the operator owns shared infrastructure.
func (c OperatorConfig) IsOperator() bool { return true }

// OwnerProto returns the typed Owner message identifying the operator.
func (c OperatorConfig) OwnerProto() *fabricv1.Owner {
	return &fabricv1.Owner{Id: &fabricv1.Owner_OperatorId{OperatorId: c.operatorID}}
}

// KubernetesNamespace returns the shared "fabric" namespace where operator-owned
// infrastructure (fabric worker, TLS CAs, certificate store) is deployed.
func (c OperatorConfig) KubernetesNamespace() string { return OperatorKubernetesNamespace }

// TemporalNamespace returns the shared "fabric" Temporal namespace the fabric worker polls.
func (c OperatorConfig) TemporalNamespace() string { return OperatorKubernetesNamespace }

func (c OperatorConfig) VaultBasePath() string { return operatorVaultPrefix + c.operatorID }

// WorkerServiceAccountName returns the service account name for operator's fabric worker
func (c OperatorConfig) WorkerServiceAccountName() string {
	return "fabric-worker-" + c.operatorID
}

// WorkerDeploymentName returns the deployment name for operator's fabric worker
func (c OperatorConfig) WorkerDeploymentName() string {
	return "fabric-worker-" + c.operatorID
}

// WorkerVaultPolicyName returns the Vault policy name for operator's fabric worker
func (c OperatorConfig) WorkerVaultPolicyName() string {
	return "fabric-" + c.operatorID + "-worker"
}

// WorkerVaultAuthRole returns the Vault auth role name for operator's fabric worker
func (c OperatorConfig) WorkerVaultAuthRole() string {
	return "fabric-" + c.operatorID + "-worker"
}

// LoadOrgConfig reads the ORG_ID environment variable and builds the OrgConfig.
// Provided to shard workers via fx at startup.
func LoadOrgConfig() (*OrgConfig, error) {
	orgID := viper.GetString(orgIDName)
	if orgID == "" {
		return nil, fmt.Errorf("ORG_ID environment variable is required")
	}
	return &OrgConfig{OrgID: orgID}, nil
}

// NewOrgConfig creates an OrgConfig from an org ID.
// This constructor is used when org ID comes from a typed Owner input or testing.
func NewOrgConfig(orgID string) OrgConfig {
	return OrgConfig{OrgID: orgID}
}

// OwnerID implements OwnerConfig interface for OrgConfig.
func (c OrgConfig) OwnerID() string { return c.OrgID }

// IsOperator implements OwnerConfig: orgs are never the operator.
func (c OrgConfig) IsOperator() bool { return false }

// OwnerProto returns the typed Owner message identifying this org.
func (c OrgConfig) OwnerProto() *fabricv1.Owner {
	return &fabricv1.Owner{Id: &fabricv1.Owner_OrgId{OrgId: c.OrgID}}
}

// Note: KubernetesNamespace() and VaultBasePath() are already defined on OrgConfig in all.go
