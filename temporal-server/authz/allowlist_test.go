package authz

import "testing"

func TestAllowed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, caller, endpoint string
		want                   bool
	}{
		{"infra bootstrap fabric", NamespaceInfra, EndpointFabricBootstrapFabric, true},
		{"infra bootstrap shard", NamespaceInfra, "fabric-shard-acme-BootstrapShard", true},
		{"infra cannot create access", NamespaceInfra, EndpointSecretsCreateAccess, false},
		{"infra cannot sign cert", NamespaceInfra, "fabric-shard-acme-CreateCertificate", false},
		{"infra empty org bootstrap", NamespaceInfra, "fabric-shard--BootstrapShard", false},
		{"fabric create access", NamespaceFabric, EndpointSecretsCreateAccess, true},
		{"fabric cannot bootstrap", NamespaceFabric, EndpointFabricBootstrapFabric, false},
		{"fabric cannot sign", NamespaceFabric, "fabric-shard-acme-CreateCertificate", false},
		{"shard create access", "fabric-shard-acme", EndpointSecretsCreateAccess, true},
		{"shard sign cert on own endpoint", "fabric-shard-acme", "fabric-shard-acme-CreateCertificate", true},
		{"shard cannot sign on another org's endpoint", "fabric-shard-acme", "fabric-shard-other-CreateCertificate", false},
		{"shard cannot sign on empty-org endpoint", "fabric-shard-acme", "fabric-shard--CreateCertificate", false},
		{"shard cannot bootstrap fabric", "fabric-shard-acme", EndpointFabricBootstrapFabric, false},
		{"shard cannot bootstrap self", "fabric-shard-acme", "fabric-shard-acme-BootstrapShard", false},
		{"shard cannot bootstrap other", "fabric-shard-acme", "fabric-shard-other-BootstrapShard", false},
		{"unknown caller", "other", EndpointSecretsCreateAccess, false},
		{"empty endpoint", NamespaceInfra, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := Allowed(tc.caller, tc.endpoint); got != tc.want {
				t.Fatalf("Allowed(%q, %q) = %v, want %v", tc.caller, tc.endpoint, got, tc.want)
			}
		})
	}
}
