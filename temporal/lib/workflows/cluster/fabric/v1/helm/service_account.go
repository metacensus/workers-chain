package helm

import "github.com/metacensus/infra/temporal/lib/workers/config"

type ServiceAccountValues struct {
	Namespace          string
	ServiceAccountName string
	RoleName           string
	CommonLabels       CommonLabels
}

func NewServiceAccountValues(owner config.WorkerOwnerConfig) *ServiceAccountValues {
	return &ServiceAccountValues{
		Namespace:          owner.KubernetesNamespace(),
		ServiceAccountName: owner.WorkerServiceAccountName(),
		RoleName:           owner.WorkerServiceAccountName(),
		CommonLabels:       DefaultCommonLabels(),
	}
}
