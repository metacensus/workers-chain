package helm

// ============================================================================
// Reusable Common Structs
// ============================================================================

// ImageConfig represents container image configuration
type ImageConfig struct {
	Repository   string        `yaml:"repository"`
	Tag          string        `yaml:"tag"`
	PullPolicy   string        `yaml:"pullPolicy"`
	PullSecrets  []interface{} `yaml:"pullSecrets"`
}

// ServiceConfig represents Kubernetes service configuration
type ServiceConfig struct {
	Type string `yaml:"type"`
	Port int    `yaml:"port"`
}

// ResourceRequirements represents CPU and memory requests/limits
type ResourceRequirements struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type ResourceConfig struct {
	Requests ResourceRequirements `yaml:"requests"`
	Limits   ResourceRequirements `yaml:"limits"`
}

// ReadinessProbeConfig represents Kubernetes readiness probe configuration
type ReadinessProbeConfig struct {
	HTTPGet struct {
		Path string `yaml:"path"`
		Port int    `yaml:"port"`
	} `yaml:"httpGet"`
	InitialDelaySeconds int `yaml:"initialDelaySeconds"`
	PeriodSeconds       int `yaml:"periodSeconds"`
}

// MSPConfig represents MSP (Membership Service Provider) configuration
type MSPConfig struct {
	BasePath string `yaml:"basePath"`
	Paths    struct {
		Signcerts string `yaml:"signcerts"`
		Keystore  string `yaml:"keystore"`
		Cacerts   string `yaml:"cacerts"`
	} `yaml:"paths"`
	CSI struct {
		Enabled                  bool   `yaml:"enabled"`
		SecretProviderClassName  string `yaml:"secretProviderClassName"`
	} `yaml:"csi"`
}

// NamespaceConfig represents namespace configuration
type NamespaceConfig struct {
	Create bool   `yaml:"create"`
	Name   string `yaml:"name"`
}

// CommonLabels represents common Kubernetes labels
type CommonLabels struct {
	AppManagedBy string `yaml:"app.kubernetes.io/managed-by"`
	AppPartOf    string `yaml:"app.kubernetes.io/part-of"`
}

// ============================================================================
// Default Helper Functions (Common across components)
// ============================================================================

// DefaultImageConfig returns sensible defaults for Hyperledger Fabric images
func DefaultImageConfig(component, tag string) ImageConfig {
	repos := map[string]string{
		"peer":    "hyperledger/fabric-peer",
		"orderer": "hyperledger/fabric-orderer",
		"ca":      "hyperledger/fabric-ca",
		"couchdb": "couchdb",
	}

	repo, ok := repos[component]
	if !ok {
		repo = "hyperledger/fabric-" + component
	}

	if tag == "" {
		tag = "2.5.0"
		if component == "ca" {
			tag = "1.5.5"
		}
		if component == "couchdb" {
			tag = "3.3.2"
		}
	}

	return ImageConfig{
		Repository:  repo,
		Tag:         tag,
		PullPolicy:  "IfNotPresent",
		PullSecrets: []interface{}{},
	}
}

// DefaultServiceConfig returns sensible defaults for Fabric component services
func DefaultServiceConfig(component string) ServiceConfig {
	ports := map[string]int{
		"peer":    7051,
		"orderer": 7050,
		"ca":      7054,
		"couchdb": 5984,
	}

	port, ok := ports[component]
	if !ok {
		port = 8080
	}

	return ServiceConfig{
		Type: "ClusterIP",
		Port: port,
	}
}

// DefaultResourceConfig returns sensible defaults for Fabric component resources
func DefaultResourceConfig(component string) ResourceConfig {
	resourceLimits := map[string]struct {
		requestsCPU    string
		requestsMem    string
		limitsCPU      string
		limitsMem      string
	}{
		"peer": {
			requestsCPU:   "250m",
			requestsMem:   "512Mi",
			limitsCPU:     "1000m",
			limitsMem:     "2Gi",
		},
		"orderer": {
			requestsCPU:   "250m",
			requestsMem:   "512Mi",
			limitsCPU:     "1000m",
			limitsMem:     "2Gi",
		},
		"ca": {
			requestsCPU:   "100m",
			requestsMem:   "128Mi",
			limitsCPU:     "500m",
			limitsMem:     "512Mi",
		},
		"couchdb": {
			requestsCPU:   "100m",
			requestsMem:   "256Mi",
			limitsCPU:     "500m",
			limitsMem:     "1Gi",
		},
	}

	limits, ok := resourceLimits[component]
	if !ok {
		limits = resourceLimits["peer"]
	}

	return ResourceConfig{
		Requests: ResourceRequirements{
			CPU:    limits.requestsCPU,
			Memory: limits.requestsMem,
		},
		Limits: ResourceRequirements{
			CPU:    limits.limitsCPU,
			Memory: limits.limitsMem,
		},
	}
}

// DefaultReadinessProbeConfig returns sensible defaults for readiness probes
func DefaultReadinessProbeConfig(component string, port int) ReadinessProbeConfig {
	path := "/"
	if component == "couchdb" {
		path = "/_up"
	}

	probe := ReadinessProbeConfig{}
	probe.HTTPGet.Path = path
	probe.HTTPGet.Port = port
	probe.InitialDelaySeconds = 10
	probe.PeriodSeconds = 5

	return probe
}

// DefaultMSPConfig returns sensible defaults for MSP configuration
func DefaultMSPConfig() MSPConfig {
	msp := MSPConfig{
		BasePath: "/etc/hyperledger/fabric/msp",
	}
	msp.Paths.Signcerts = "/etc/hyperledger/fabric/msp/signcerts/cert.pem"
	msp.Paths.Keystore = "/etc/hyperledger/fabric/msp/keystore/key.pem"
	msp.Paths.Cacerts = "/etc/hyperledger/fabric/msp/cacerts/ca.pem"
	msp.CSI.Enabled = false
	msp.CSI.SecretProviderClassName = ""

	return msp
}

// DefaultCommonLabels returns sensible defaults for common Kubernetes labels
func DefaultCommonLabels() CommonLabels {
	return CommonLabels{
		AppManagedBy: "helm",
		AppPartOf:    "fabric",
	}
}

