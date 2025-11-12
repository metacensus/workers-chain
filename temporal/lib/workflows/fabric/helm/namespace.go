package helm

// ============================================================================
// Namespace Component Struct
// ============================================================================

type NamespaceValues struct {
	Namespace struct {
		Name string `yaml:"name"`
	} `yaml:"namespace"`

	CommonLabels CommonLabels `yaml:"commonLabels"`
}

// ============================================================================
// Constructor and Initialization Methods
// ============================================================================

// NewNamespaceValues creates a new NamespaceValues instance with sensible defaults
// namespaceName: name of the Kubernetes namespace
func NewNamespaceValues(namespaceName string) *NamespaceValues {
	n := &NamespaceValues{
		CommonLabels: DefaultCommonLabels(),
	}

	n.setDefaults(namespaceName)

	return n
}

// setDefaults populates default values for the NamespaceValues struct
func (n *NamespaceValues) setDefaults(namespaceName string) {
	n.Namespace.Name = namespaceName
}

