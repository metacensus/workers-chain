package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

// OwnerConfig defines the contract for resource ownership (operator or org).
// This interface enables the same component configs (CertificateAuthorityConfig,
// PeerConfig, etc.) to work with both operator-owned and org-owned resources.
type OwnerConfig interface {
	// OwnerID returns the unique identifier for this owner (operator or org)
	OwnerID() string
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

// OperatorConfig implements OwnerConfig for operator-owned resources.
// Operator-owned resources include shared infrastructure like TLS CAs.
type OperatorConfig struct {
	operatorID string
}

var (
	operatorGlobal *OperatorConfig
	operatorOnce   sync.Once
	operatorMu     sync.RWMutex
)

// InitializeOperator initializes the global OperatorConfig from the OPERATOR_ID env var.
// This should be called once at worker startup.
func InitializeOperator() error {
	var err error
	operatorOnce.Do(func() {
		operatorID := viper.GetString(operatorIDName)
		if operatorID == "" {
			err = fmt.Errorf("OPERATOR_ID environment variable is required")
			return
		}
		operatorGlobal = &OperatorConfig{operatorID: operatorID}
	})
	return err
}

// GetOperator returns the global OperatorConfig singleton.
// Panics if InitializeOperator() has not been called.
func GetOperator() *OperatorConfig {
	operatorMu.RLock()
	defer operatorMu.RUnlock()
	if operatorGlobal == nil {
		panic("OperatorConfig not initialized - call InitializeOperator() at startup")
	}
	return operatorGlobal
}

// NewOperatorConfig creates an OperatorConfig from an operator ID.
// This constructor is used when operator ID comes from workflow input or testing.
func NewOperatorConfig(operatorID string) OperatorConfig {
	return OperatorConfig{operatorID: operatorID}
}

func (c OperatorConfig) OwnerID() string { return c.operatorID }

// KubernetesNamespace returns the shared "fabric" namespace where operator-owned
// infrastructure (fabric worker, TLS CAs, certificate store) is deployed.
func (c OperatorConfig) KubernetesNamespace() string { return "fabric" }

// TemporalNamespace returns the shared "fabric" Temporal namespace the fabric worker polls.
func (c OperatorConfig) TemporalNamespace() string { return "fabric" }

func (c OperatorConfig) VaultBasePath() string { return "fabric/" + c.operatorID }

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

// NewOrgConfig creates an OrgConfig from an org ID.
// This constructor is used when org ID comes from workflow input.
func NewOrgConfig(orgID string) OrgConfig {
	return OrgConfig{OrgID: orgID}
}

// OwnerID implements OwnerConfig interface for OrgConfig.
func (c OrgConfig) OwnerID() string { return c.OrgID }

// Note: KubernetesNamespace() and VaultBasePath() are already defined on OrgConfig in all.go
