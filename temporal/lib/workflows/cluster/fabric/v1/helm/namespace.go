package helm

import "github.com/metacensus/infra/temporal/lib/workers/config"

// NamespaceValues represents Helm values for the fabric/namespace chart
type NamespaceValues struct {
	Name         string
	CommonLabels CommonLabels
}

// NewNamespaceValues creates values for a fabric namespace
func NewNamespaceValues(orgCfg config.OrgConfig) *NamespaceValues {
	return &NamespaceValues{
		Name:         orgCfg.KubernetesNamespace(),
		CommonLabels: DefaultCommonLabels(),
	}
}
