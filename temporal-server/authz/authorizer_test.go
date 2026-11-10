package authz

import (
	"context"
	"testing"

	"go.temporal.io/server/common/authorization"
)

func TestAuthorizer_NexusAllowDeny(t *testing.T) {
	t.Parallel()
	a := NewAuthorizer(nil)
	ctx := context.Background()

	claims := &authorization.Claims{
		Subject: "fabric-shard-acme",
		Namespaces: map[string]authorization.Role{
			"fabric-shard-acme": authorization.RoleWriter | authorization.RoleWorker,
		},
	}

	allow, err := a.Authorize(ctx, claims, &authorization.CallTarget{
		APIName:           "/temporal.api.nexusservice.v1.NexusService/DispatchNexusTask",
		NexusEndpointName: EndpointSecretsCreateAccess,
		Namespace:         NamespaceInfra,
	})
	if err != nil || allow.Decision != authorization.DecisionAllow {
		t.Fatalf("expected allow, got %+v err=%v", allow, err)
	}

	deny, err := a.Authorize(ctx, claims, &authorization.CallTarget{
		APIName:           "/temporal.api.nexusservice.v1.NexusService/DispatchNexusTask",
		NexusEndpointName: EndpointFabricBootstrapFabric,
		Namespace:         NamespaceFabric,
	})
	if err != nil || deny.Decision != authorization.DecisionDeny {
		t.Fatalf("expected deny, got %+v err=%v", deny, err)
	}
}

func TestAuthorizer_UnauthenticatedDenied(t *testing.T) {
	t.Parallel()
	a := NewAuthorizer(nil)
	res, err := a.Authorize(context.Background(), nil, &authorization.CallTarget{
		Namespace: NamespaceInfra,
		APIName:   "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution",
	})
	if err != nil || res.Decision != authorization.DecisionDeny {
		t.Fatalf("expected deny, got %+v err=%v", res, err)
	}
}

func TestAuthorizer_OwnNamespaceAllowed(t *testing.T) {
	t.Parallel()
	a := NewAuthorizer(nil)
	claims := &authorization.Claims{
		Subject: NamespaceInfra,
		Namespaces: map[string]authorization.Role{
			NamespaceInfra: authorization.RoleWriter | authorization.RoleWorker,
		},
	}
	res, err := a.Authorize(context.Background(), claims, &authorization.CallTarget{
		Namespace: NamespaceInfra,
		APIName:   "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution",
	})
	if err != nil || res.Decision != authorization.DecisionAllow {
		t.Fatalf("expected allow, got %+v err=%v", res, err)
	}
}

func TestAuthorizer_InfraOperatorAPIs(t *testing.T) {
	t.Parallel()
	a := NewAuthorizer(nil)
	claims := func(ns string) *authorization.Claims {
		return &authorization.Claims{
			Subject: ns,
			Namespaces: map[string]authorization.Role{
				ns: authorization.RoleWriter | authorization.RoleWorker,
			},
		}
	}
	cases := []struct {
		name   string
		caller string
		target *authorization.CallTarget
		allow  bool
	}{
		{"infra register shard", NamespaceInfra, &authorization.CallTarget{APIName: apiRegisterNamespace, Namespace: "fabric-shard-acme"}, true},
		{"infra create nexus endpoint", NamespaceInfra, &authorization.CallTarget{APIName: apiCreateNexusEndpoint}, true},
		{"infra cannot register default", NamespaceInfra, &authorization.CallTarget{APIName: apiRegisterNamespace, Namespace: "default"}, false},
		{"infra cannot start on shard", NamespaceInfra, &authorization.CallTarget{APIName: "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution", Namespace: "fabric-shard-acme"}, false},
		{"shard cannot register", "fabric-shard-acme", &authorization.CallTarget{APIName: apiRegisterNamespace, Namespace: "fabric-shard-other"}, false},
		{"fabric cannot register shard", NamespaceFabric, &authorization.CallTarget{APIName: apiRegisterNamespace, Namespace: "fabric-shard-acme"}, false},
		{"shard cannot create nexus endpoint", "fabric-shard-acme", &authorization.CallTarget{APIName: apiCreateNexusEndpoint}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := a.Authorize(context.Background(), claims(tc.caller), tc.target)
			if err != nil {
				t.Fatal(err)
			}
			got := res.Decision == authorization.DecisionAllow
			if got != tc.allow {
				t.Fatalf("decision=%v want allow=%v reason=%q", res.Decision, tc.allow, res.Reason)
			}
		})
	}
}

func TestAuthorizer_GRPCHealthCheckAllowed(t *testing.T) {
	t.Parallel()
	claims := &authorization.Claims{
		Subject: NamespaceInfra,
		Namespaces: map[string]authorization.Role{
			NamespaceInfra: authorization.RoleWriter | authorization.RoleWorker,
		},
	}
	res, err := NewAuthorizer(nil).Authorize(context.Background(), claims, &authorization.CallTarget{
		APIName: "/grpc.health.v1.Health/Check",
	})
	if err != nil || res.Decision != authorization.DecisionAllow {
		t.Fatalf("expected allow, got %+v err=%v", res, err)
	}
}
