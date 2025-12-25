package helm

type ServiceAccountValues struct {
	Namespace          string
	ServiceAccountName string
	RoleName           string
	CommonLabels       CommonLabels
}

func NewServiceAccountValues(org string) *ServiceAccountValues {
	return &ServiceAccountValues{
		Namespace:          "fabric-" + org,
		ServiceAccountName: "fabric-worker-" + org,
		RoleName:           "fabric-worker-" + org,
		CommonLabels:       DefaultCommonLabels(),
	}
}
