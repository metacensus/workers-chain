package helm

// ============================================================================
// Orderer Component Struct
// ============================================================================

type OrdererValues struct {
	Organization struct {
		ID    string `yaml:"id"`
		MSPId string `yaml:"mspId"`
	} `yaml:"organization"`

	Component string `yaml:"component"`
	Number    int    `yaml:"number"`

	Image     ImageConfig     `yaml:"image"`
	Service   ServiceConfig   `yaml:"service"`
	Resources ResourceConfig  `yaml:"resources"`

	OrdererMSPID string `yaml:"mspId"`
	LoggingSpec  string `yaml:"loggingSpec"`

	TLS struct {
		Enabled            bool `yaml:"enabled"`
		ClientAuthRequired bool `yaml:"clientAuthRequired"`
	} `yaml:"tls"`

	ReadinessProbe ReadinessProbeConfig `yaml:"readinessProbe"`

	MSP           MSPConfig       `yaml:"msp"`
	Namespace     NamespaceConfig `yaml:"namespace"`
	CommonLabels  CommonLabels    `yaml:"commonLabels"`
}

// ============================================================================
// Constructor and Initialization Methods
// ============================================================================

// NewOrdererValues creates a new OrdererValues instance with sensible defaults
// org: organization ID
// ordererID: orderer identity (e.g., "orderer-0")
// mspID: Organization MSP ID
// ordererMSPID: Orderer MSP ID (e.g., "OrdererMSP")
// number: orderer instance number
// namespace: Kubernetes namespace for the orderer
func NewOrdererValues(org, ordererID, mspID, ordererMSPID string, number int, namespace string) *OrdererValues {
	o := &OrdererValues{
		Component:    "orderer",
		Number:       number,
		OrdererMSPID: ordererMSPID,
		Image:        DefaultImageConfig("orderer", ""),
		Service:      DefaultServiceConfig("orderer"),
		Resources:    DefaultResourceConfig("orderer"),
		LoggingSpec:  "INFO",
		MSP:          DefaultMSPConfig(),
		CommonLabels: DefaultCommonLabels(),
	}

	o.setDefaults(org, mspID, namespace)

	return o
}

// setDefaults populates default values for the OrdererValues struct
func (o *OrdererValues) setDefaults(org, mspID, namespace string) {
	o.Organization.ID = org
	o.Organization.MSPId = mspID

	o.TLS.Enabled = true
	o.TLS.ClientAuthRequired = false

	o.ReadinessProbe = DefaultReadinessProbeConfig("orderer", 7050)

	o.Namespace.Create = false
	o.Namespace.Name = namespace
}

