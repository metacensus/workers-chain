package helm

// NamespaceValues represents Helm values for the fabric/namespace chart
type NamespaceValues struct {
	Name         string       `yaml:"name"`
	CommonLabels CommonLabels `yaml:"commonLabels"`
}

// NewNamespaceValues creates values for a fabric namespace
func NewNamespaceValues(org string) *NamespaceValues {
	return &NamespaceValues{
		Name:         "fabric-" + org,
		CommonLabels: DefaultCommonLabels(),
	}
}
