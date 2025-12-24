package helm

// ============================================================================
// CA Component Struct
// ============================================================================

type CAValues struct {
	Organization struct {
		ID    string `yaml:"id"`
		MSPId string `yaml:"mspId"`
	} `yaml:"organization"`

	Component string `yaml:"component"`
	Number    int    `yaml:"number"`

	Image     ImageConfig     `yaml:"image"`
	Service   ServiceConfig   `yaml:"service"`
	Resources ResourceConfig  `yaml:"resources"`

	LoggingSpec    string                   `yaml:"loggingSpec"`
	ReadinessProbe ReadinessProbeConfig     `yaml:"readinessProbe"`

	MSP           MSPConfig       `yaml:"msp"`
	Namespace     NamespaceConfig `yaml:"namespace"`
	CommonLabels  CommonLabels    `yaml:"commonLabels"`
}

// ============================================================================
// Constructor and Initialization Methods
// ============================================================================

// NewCAValues creates a new CAValues instance with sensible defaults
// org: organization ID
// caID: CA identity (e.g., "ca-0")
// mspID: MSP ID
// number: CA instance number
// namespace: Kubernetes namespace for the CA
func NewCAValues(org, caID, mspID string, number int, namespace string) *CAValues {
	c := &CAValues{
		Component:   "ca",
		Number:      number,
		Image:       DefaultImageConfig("ca", ""),
		Service:     DefaultServiceConfig("ca"),
		Resources:   DefaultResourceConfig("ca"),
		LoggingSpec: "INFO",
		MSP:         DefaultMSPConfig(),
		CommonLabels: DefaultCommonLabels(),
	}

	c.setDefaults(org, mspID, namespace)

	return c
}

// setDefaults populates default values for the CAValues struct
func (c *CAValues) setDefaults(org, mspID, namespace string) {
	c.Organization.ID = org
	c.Organization.MSPId = mspID

	c.ReadinessProbe = DefaultReadinessProbeConfig("ca", 7054)

	c.Namespace.Create = false
	c.Namespace.Name = namespace
}

