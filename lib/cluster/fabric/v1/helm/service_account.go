package helm

import "github.com/metacensus/workers-chain/lib/workers/config"

type ServiceAccountValues struct {
	Namespace          string       `json:"namespace"`
	ServiceAccountName string       `json:"serviceAccountName"`
	RoleName           string       `json:"roleName"`
	CommonLabels       CommonLabels `json:"commonLabels"`
}

func NewServiceAccountValues(owner config.WorkerOwnerConfig) *ServiceAccountValues {
	return &ServiceAccountValues{
		Namespace:          owner.KubernetesNamespace(),
		ServiceAccountName: owner.WorkerServiceAccountName(),
		RoleName:           owner.WorkerServiceAccountName(),
		CommonLabels:       DefaultCommonLabels(),
	}
}
