package helm

import "github.com/metacensus/infra/temporal/lib/workers/config"

// WorkerValues represents the Helm values for a fabric worker deployment.
// The same chart deploys both the shared fabric worker (operator-owned, no org)
// and per-org shard workers.
type WorkerValues struct {
	Operator string
	// Org is the owner identifier used for resource naming
	// (org ID for shard workers, operator ID for the shared fabric worker)
	Org string
	// OrgId is the ORG_ID env value; empty for the shared fabric worker,
	// which owns no organization
	OrgId string
	// Namespace is the Kubernetes namespace the worker is deployed into
	Namespace      string
	Image          ImageConfig
	Replicas       int
	Temporal       TemporalConfig
	Vault          VaultConfig
	LogLevel       string
	HealthCheck    HealthCheckConfig
	Resources      ResourceConfig
	LivenessProbe  ProbeConfig
	ReadinessProbe ProbeConfig
	CommonLabels   CommonLabels
}

// TemporalConfig represents Temporal connection configuration
type TemporalConfig struct {
	Address   string
	Namespace string
	TaskQueue string
}

// VaultConfig represents Vault connection configuration
type VaultConfig struct {
	Address string
	Role    string
}

// HealthCheckConfig represents health check port configuration
type HealthCheckConfig struct {
	Port int
}

// ProbeConfig represents Kubernetes probe configuration
type ProbeConfig struct {
	Path                string
	InitialDelaySeconds int
	PeriodSeconds       int
}

// NewWorkerValues creates WorkerValues for a per-org shard worker
func NewWorkerValues(operatorCfg config.OperatorConfig, orgCfg config.OrgConfig) *WorkerValues {
	return newWorkerValues(
		operatorCfg.OwnerID(),
		orgCfg.OrgID,
		orgCfg.OrgID,
		orgCfg.KubernetesNamespace(),
		orgCfg.TemporalNamespace(),
		orgCfg.WorkerVaultAuthRole(),
		"metacensus/shard-worker",
	)
}

// NewOperatorWorkerValues creates WorkerValues for the shared fabric worker
// (operator-owned, deployed once into the "fabric" namespace)
func NewOperatorWorkerValues(operatorCfg config.OperatorConfig) *WorkerValues {
	return newWorkerValues(
		operatorCfg.OwnerID(),
		operatorCfg.OwnerID(),
		"", // no ORG_ID: the fabric worker owns shared infrastructure, not an org
		operatorCfg.KubernetesNamespace(),
		operatorCfg.TemporalNamespace(),
		operatorCfg.WorkerVaultAuthRole(),
		"metacensus/fabric-worker",
	)
}

func newWorkerValues(operator, org, orgID, kubeNamespace, temporalNamespace, vaultRole, imageRepository string) *WorkerValues {
	return &WorkerValues{
		Operator:  operator,
		Org:       org,
		OrgId:     orgID,
		Namespace: kubeNamespace,
		Image: ImageConfig{
			Repository: imageRepository,
			Tag:        "latest",
			PullPolicy: "IfNotPresent",
		},
		Replicas: 1,
		Temporal: TemporalConfig{
			Address:   "temporal-frontend.temporal.svc.cluster.local:7233",
			Namespace: temporalNamespace,
			TaskQueue: "infra",
		},
		Vault: VaultConfig{
			Address: "http://vault.secrets.svc.cluster.local:8200",
			Role:    vaultRole,
		},
		LogLevel: "info",
		HealthCheck: HealthCheckConfig{
			Port: 8080,
		},
		Resources: ResourceConfig{
			Requests: ResourceRequirements{
				CPU:    "100m",
				Memory: "128Mi",
			},
			Limits: ResourceRequirements{
				CPU:    "500m",
				Memory: "512Mi",
			},
		},
		LivenessProbe: ProbeConfig{
			Path:                "/health",
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
		},
		ReadinessProbe: ProbeConfig{
			Path:                "/ready",
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
		},
		CommonLabels: DefaultCommonLabels(),
	}
}
