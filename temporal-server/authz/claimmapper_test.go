package authz

import (
	"context"
	"crypto/x509/pkix"
	"testing"

	"go.temporal.io/server/common/authorization"
)

func TestClaimMapper_RequiresClientCert(t *testing.T) {
	t.Parallel()
	c := NewClaimMapper(nil)

	if _, err := c.GetClaims(nil); err == nil {
		t.Fatal("expected error for missing auth info")
	}
	if _, err := c.GetClaims(&authorization.AuthInfo{}); err == nil {
		t.Fatal("expected error for missing client certificate")
	}
}

func TestClaimMapper_WorkerCNIsNamespace(t *testing.T) {
	t.Parallel()
	c := NewClaimMapper(nil)

	claims, err := c.GetClaims(&authorization.AuthInfo{
		TLSSubject: &pkix.Name{CommonName: "fabric-shard-acme"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Subject != "fabric-shard-acme" {
		t.Fatalf("subject = %q, want fabric-shard-acme", claims.Subject)
	}
	if claims.System != authorization.RoleUndefined {
		t.Fatalf("worker cert must not carry a system role, got %v", claims.System)
	}
	if claims.Namespaces["fabric-shard-acme"]&authorization.RoleWriter == 0 {
		t.Fatalf("expected writer on own namespace, got %v", claims.Namespaces)
	}
}

func TestClaimMapper_SystemCNs(t *testing.T) {
	t.Parallel()
	c := NewClaimMapper(nil)

	for _, cn := range []string{"temporal-system-admin", "temporal-internode"} {
		claims, err := c.GetClaims(&authorization.AuthInfo{
			TLSSubject: &pkix.Name{CommonName: cn},
		})
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", cn, err)
		}
		if claims.System != authorization.RoleAdmin {
			t.Fatalf("%s: system role = %v, want admin", cn, claims.System)
		}

		res, err := NewAuthorizer(nil).Authorize(context.Background(), claims, &authorization.CallTarget{
			Namespace: NamespaceFabric,
			APIName:   "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution",
		})
		if err != nil || res.Decision != authorization.DecisionAllow {
			t.Fatalf("%s: expected allow, got %+v err=%v", cn, res, err)
		}
	}
}
