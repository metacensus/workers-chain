package helm

import "github.com/metacensus/infra/temporal/lib/workers/config"

type ServiceAccountValues struct {
	Namespace          string
	ServiceAccountName string
	RoleName           string
	CommonLabels       CommonLabels
}

func NewServiceAccountValues(orgCfg config.OrgConfig) *ServiceAccountValues {
	return &ServiceAccountValues{
		Namespace:          orgCfg.KubernetesNamespace(),
		ServiceAccountName: orgCfg.WorkerServiceAccountName(),
		RoleName:           orgCfg.WorkerServiceAccountName(),
		CommonLabels:       DefaultCommonLabels(),
	}
}
