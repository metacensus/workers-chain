package helm

// ============================================================================
// Peer Component Struct
// ============================================================================

type PeerValues struct {
	Organization struct {
		ID    string
		MSPId string
	}

	Component string
	Number    int

	Image     ImageConfig
	Service   ServiceConfig
	Resources ResourceConfig

	LoggingSpec string

	TLS struct {
		Enabled            bool
		ClientAuthRequired bool
	}

	ReadinessProbe ReadinessProbeConfig

	CouchDB struct {
		Enabled   bool
		Image     ImageConfig
		Service   ServiceConfig
		Resources ResourceConfig

		Credentials struct {
			Username string
			Password string
		}

		ReadinessProbe ReadinessProbeConfig
	}

	MSP          MSPConfig
	Namespace    NamespaceConfig
	CommonLabels CommonLabels
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

