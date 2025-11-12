package helm

// ============================================================================
// Peer Component Struct
// ============================================================================

type PeerValues struct {
	Organization struct {
		ID    string `yaml:"id"`
		MSPId string `yaml:"mspId"`
	} `yaml:"organization"`

	Component string `yaml:"component"`
	Number    int    `yaml:"number"`

	Image     ImageConfig     `yaml:"image"`
	Service   ServiceConfig   `yaml:"service"`
	Resources ResourceConfig  `yaml:"resources"`

	LoggingSpec string `yaml:"loggingSpec"`

	TLS struct {
		Enabled              bool `yaml:"enabled"`
		ClientAuthRequired   bool `yaml:"clientAuthRequired"`
	} `yaml:"tls"`

	ReadinessProbe ReadinessProbeConfig `yaml:"readinessProbe"`

	CouchDB struct {
		Enabled   bool                 `yaml:"enabled"`
		Image     ImageConfig          `yaml:"image"`
		Service   ServiceConfig        `yaml:"service"`
		Resources ResourceConfig       `yaml:"resources"`

		Credentials struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"credentials"`

		ReadinessProbe ReadinessProbeConfig `yaml:"readinessProbe"`
	} `yaml:"couchdb"`

	MSP           MSPConfig       `yaml:"msp"`
	Namespace     NamespaceConfig `yaml:"namespace"`
	CommonLabels  CommonLabels    `yaml:"commonLabels"`
}

// ============================================================================
// Constructor and Initialization Methods
// ============================================================================

// NewPeerValues creates a new PeerValues instance with sensible defaults
// org: organization ID (e.g., "org1")
// peerID: peer identity (e.g., "peer-0")
// mspID: MSP ID (e.g., "Org1MSP")
// number: peer instance number (e.g., 0)
// namespace: Kubernetes namespace for the peer
func NewPeerValues(org, peerID, mspID string, number int, namespace string) *PeerValues {
	p := &PeerValues{
		Component: "peer",
		Number:    number,
		Image:     DefaultImageConfig("peer", ""),
		Service:   DefaultServiceConfig("peer"),
		Resources: DefaultResourceConfig("peer"),
		LoggingSpec: "INFO",
		MSP:       DefaultMSPConfig(),
		CommonLabels: DefaultCommonLabels(),
	}

	p.setDefaults(org, mspID, namespace)

	return p
}

// setDefaults populates default values for the PeerValues struct
func (p *PeerValues) setDefaults(org, mspID, namespace string) {
	p.Organization.ID = org
	p.Organization.MSPId = mspID

	p.TLS.Enabled = true
	p.TLS.ClientAuthRequired = false

	p.ReadinessProbe = DefaultReadinessProbeConfig("peer", 7051)

	// CouchDB defaults
	p.CouchDB.Enabled = true
	p.CouchDB.Image = DefaultImageConfig("couchdb", "")
	p.CouchDB.Service = DefaultServiceConfig("couchdb")
	p.CouchDB.Resources = DefaultResourceConfig("couchdb")
	p.CouchDB.ReadinessProbe = DefaultReadinessProbeConfig("couchdb", 5984)

	p.Namespace.Create = false
	p.Namespace.Name = namespace
}

