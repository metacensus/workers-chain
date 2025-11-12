package helm

// ============================================================================
// CouchDB Component Struct
// ============================================================================

type CouchDBValues struct {
	Organization struct {
		ID string `yaml:"id"`
	} `yaml:"organization"`

	Component string `yaml:"component"`
	Number    int    `yaml:"number"`

	Image     ImageConfig     `yaml:"image"`
	Service   ServiceConfig   `yaml:"service"`
	Resources ResourceConfig  `yaml:"resources"`

	Credentials struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"credentials"`

	ReadinessProbe ReadinessProbeConfig `yaml:"readinessProbe"`

	Namespace    NamespaceConfig `yaml:"namespace"`
	CommonLabels CommonLabels    `yaml:"commonLabels"`
}

// ============================================================================
// Constructor and Initialization Methods
// ============================================================================

// NewCouchDBValues creates a new CouchDBValues instance with sensible defaults
// org: organization ID
// couchdbID: CouchDB identity (e.g., "couchdb-0")
// number: CouchDB instance number
// namespace: Kubernetes namespace for CouchDB
func NewCouchDBValues(org, couchdbID string, number int, namespace string) *CouchDBValues {
	c := &CouchDBValues{
		Component:   "couchdb",
		Number:      number,
		Image:       DefaultImageConfig("couchdb", ""),
		Service:     DefaultServiceConfig("couchdb"),
		Resources:   DefaultResourceConfig("couchdb"),
		CommonLabels: DefaultCommonLabels(),
	}

	c.setDefaults(org, namespace)

	return c
}

// setDefaults populates default values for the CouchDBValues struct
func (c *CouchDBValues) setDefaults(org, namespace string) {
	c.Organization.ID = org

	c.ReadinessProbe = DefaultReadinessProbeConfig("couchdb", 5984)

	c.Namespace.Create = false
	c.Namespace.Name = namespace
}

