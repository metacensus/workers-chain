package workers

import (
	"fmt"

	"github.com/metacensus/infra/temporal/lib/env"
	kube "github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// NewTemporalClient creates a Temporal client from environment configuration
func NewTemporalClient() (client.Client, error) {
	return client.Dial(client.Options{
		HostPort:  env.TemporalAddress(),
		Namespace: env.TemporalNamespace(),
	})
}

// NewNamespaceClient creates a Temporal namespace client for namespace management operations
func NewNamespaceClient() (client.NamespaceClient, error) {
	return client.NewNamespaceClient(client.Options{
		HostPort: env.TemporalAddress(),
	})
}

// NewWorker creates a Temporal worker from the client and environment configuration
func NewWorker(c client.Client) worker.Worker {
	return worker.New(c, env.TaskQueue(), worker.Options{})
}

// NewVaultConfig creates VaultConfig from environment variables
func NewVaultConfig() vault.VaultConfig {
	return vault.VaultConfig{
		Address: env.VaultAddress(),
		Role:    env.VaultRole(),
	}
}

// NewKubeConfig creates KubeConfig from environment variables
func NewKubeConfig() kube.KubeConfig {
	return kube.KubeConfig{
		KubeconfigPath: env.KubeconfigPath(),
	}
}

// OrgConfig holds the fabric organization identity and provides derived values
// for Kubernetes namespaces, Temporal namespaces, Vault paths, etc.
type OrgConfig struct {
	OrgID string // Organization identifier (e.g., "metacensus", "org1")
}

// NewOrgConfig creates OrgConfig from environment variables
func NewOrgConfig() (OrgConfig, error) {
	orgID := env.OrgID()
	if orgID == "" {
		return OrgConfig{}, fmt.Errorf("ORG_ID environment variable is required for fabric worker")
	}
	return OrgConfig{OrgID: orgID}, nil
}

// KubeNamespace returns the Kubernetes namespace for this org (fabric-{org})
func (c OrgConfig) KubeNamespace() string {
	return "fabric-" + c.OrgID
}

// TemporalNamespace returns the Temporal namespace for this org (fabric-{org})
func (c OrgConfig) TemporalNamespace() string {
	return "fabric-" + c.OrgID
}

// VaultBasePath returns the base Vault path for this org's secrets
func (c OrgConfig) VaultBasePath() string {
	return "fabric/" + c.OrgID
}

// ServiceAccountName returns the Kubernetes service account name for this org
func (c OrgConfig) ServiceAccountName() string {
	return "fabric-" + c.OrgID
}

// WorkerDeploymentName returns the worker deployment name for this org
func (c OrgConfig) WorkerDeploymentName() string {
	return "fabric-" + c.OrgID + "-worker"
}
