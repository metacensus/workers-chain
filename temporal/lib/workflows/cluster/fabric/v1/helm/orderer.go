package helm

// ============================================================================
// Orderer Component Struct
// ============================================================================

type OrdererValues struct {
	Organization struct {
		ID    string
		MSPId string
	}

	Component string
	Number    int

	Image     ImageConfig
	Service   ServiceConfig
	Resources ResourceConfig

	OrdererMSPID string
	LoggingSpec  string

	TLS struct {
		Enabled            bool
		ClientAuthRequired bool
	}

	ReadinessProbe ReadinessProbeConfig

	MSP          MSPConfig
	Namespace    NamespaceConfig
	CommonLabels CommonLabels
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

