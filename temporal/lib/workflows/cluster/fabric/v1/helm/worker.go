package helm

// WorkerValues represents the Helm values for a fabric worker deployment
type WorkerValues struct {
	Org            string
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

// NewWorkerValues creates WorkerValues with sensible defaults for a fabric worker
func NewWorkerValues(org string) *WorkerValues {
	temporalNamespace := "fabric-" + org
	vaultRole := "fabric-" + org + "-worker"

	return &WorkerValues{
		Org: org,
		Image: ImageConfig{
			Repository: "metacensus/fabric-worker",
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


