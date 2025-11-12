package helm

// ============================================================================
// SecretProviderClass (SPC) Component Struct
// ============================================================================

type SPCValues struct {
	Vault struct {
		Address string `yaml:"address"`
	} `yaml:"vault"`

	Organization string `yaml:"organization"`
	Identity     string `yaml:"identity"`
	Namespace    string `yaml:"namespace"`
	ResourceName string `yaml:"resourceName"`

	CommonLabels CommonLabels `yaml:"commonLabels"`
}

// ============================================================================
// Constructor and Initialization Methods
// ============================================================================

// NewSPCValues creates a new SPCValues instance with sensible defaults
// org: organization ID (e.g., "org1")
// identity: identity name (e.g., "peer-0", "orderer-0")
// namespace: Kubernetes namespace where SecretProviderClass will be created
// vaultAddress: Vault server address (default: http://vault.secrets.svc.cluster.local:8200)
func NewSPCValues(org, identity, namespace, vaultAddress string) *SPCValues {
	s := &SPCValues{
		Organization: org,
		Identity:     identity,
		Namespace:    namespace,
		ResourceName: "", // Will be generated as {org}-{identity}-spc
		CommonLabels: DefaultCommonLabels(),
	}

	s.setDefaults(vaultAddress)

	return s
}

// setDefaults populates default values for the SPCValues struct
func (s *SPCValues) setDefaults(vaultAddress string) {
	if vaultAddress == "" {
		vaultAddress = "http://vault.secrets.svc.cluster.local:8200"
	}
	s.Vault.Address = vaultAddress
}
