package helm

import "github.com/metacensus/infra/temporal/lib/workers/config"

// NamespaceValues represents Helm values for the fabric/namespace chart
type NamespaceValues struct {
	Name         string
	IsOperator   bool
	CommonLabels CommonLabels
}

// NewNamespaceValues creates values for a fabric namespace
// The isOperator flag determines whether network policies are created
// to allow cross-namespace access to TLS CA services
func NewNamespaceValues(owner config.OwnerConfig, isOperator bool) *NamespaceValues {
	return &NamespaceValues{
		Name:         owner.KubernetesNamespace(),
		IsOperator:   isOperator,
		CommonLabels: DefaultCommonLabels(),
	}
}
