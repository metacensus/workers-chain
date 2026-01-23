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
	// Example: "fabric-metacensus" (operator) or "fabric-org1" (org)
	KubernetesNamespace() string
	// VaultBasePath returns the base Vault path for this owner's secrets
	// Example: "fabric/metacensus" (operator) or "fabric/org1" (org)
	VaultBasePath() string
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

func (c OperatorConfig) OwnerID() string             { return c.operatorID }
func (c OperatorConfig) KubernetesNamespace() string { return "fabric-" + c.operatorID }
func (c OperatorConfig) VaultBasePath() string       { return "fabric/" + c.operatorID }

// WorkerServiceAccountName returns the service account name for operator's fabric worker
func (c OperatorConfig) WorkerServiceAccountName() string {
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
