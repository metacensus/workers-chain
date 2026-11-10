package authz

import (
	"context"
	"errors"
	"testing"

	"go.temporal.io/api/serviceerror"
	persistencespb "go.temporal.io/server/api/persistence/v1"
	"go.temporal.io/server/common/namespace"
	"go.temporal.io/server/common/nexus/nexustest"
	"go.uber.org/mock/gomock"
)

// The history service resolves an in-workflow Nexus endpoint through
// GetByName with the calling workflow's namespace, and Temporal's schedule
// command handler turns PermissionDenied into a workflow task failure. These
// tests pin that a denied caller never reaches the wrapped registry.
func TestGuardedEndpointRegistryGetByName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		callerNS    string
		endpoint    string
		wantAllowed bool
	}{
		{"shard reaches its own certificate endpoint", "fabric-shard-acme", "fabric-shard-acme-CreateCertificate", true},
		{"shard cannot reach another org's certificate endpoint", "fabric-shard-acme", "fabric-shard-other-CreateCertificate", false},
		{"shard cannot bootstrap fabric", "fabric-shard-acme", EndpointFabricBootstrapFabric, false},
		{"infra bootstraps a shard", NamespaceInfra, "fabric-shard-acme-BootstrapShard", true},
		{"unknown namespace is denied", "someone-else", EndpointSecretsCreateAccess, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			const callerID = "namespace-id"

			namespaces := namespace.NewMockRegistry(gomock.NewController(t))
			namespaces.EXPECT().GetNamespaceName(namespace.ID(callerID)).
				Return(namespace.Name(tc.callerNS), nil).AnyTimes()

			wantEntry := &persistencespb.NexusEndpointEntry{Id: "entry-id"}
			reached := false
			inner := nexustest.FakeEndpointRegistry{
				OnGetByName: func(context.Context, namespace.ID, string) (*persistencespb.NexusEndpointEntry, error) {
					reached = true
					return wantEntry, nil
				},
			}

			entry, err := NexusEndpointRegistry(inner, namespaces, nil).
				GetByName(context.Background(), namespace.ID(callerID), tc.endpoint)

			if tc.wantAllowed {
				if err != nil {
					t.Fatalf("GetByName() error = %v, want nil", err)
				}
				if entry != wantEntry {
					t.Errorf("GetByName() entry = %v, want the wrapped registry's entry", entry)
				}
				return
			}

			if !errors.As(err, new(*serviceerror.PermissionDenied)) {
				t.Fatalf("GetByName() error = %v, want PermissionDenied", err)
			}
			if reached {
				t.Error("denied lookup reached the wrapped registry")
			}
		})
	}
}

// A namespace that cannot be resolved must not resolve an endpoint either.
func TestGuardedEndpointRegistryUnknownNamespace(t *testing.T) {
	t.Parallel()
	namespaces := namespace.NewMockRegistry(gomock.NewController(t))
	namespaces.EXPECT().GetNamespaceName(gomock.Any()).
		Return(namespace.EmptyName, serviceerror.NewNamespaceNotFound("gone"))

	inner := nexustest.FakeEndpointRegistry{
		OnGetByName: func(context.Context, namespace.ID, string) (*persistencespb.NexusEndpointEntry, error) {
			t.Error("wrapped registry called for an unresolvable namespace")
			return nil, nil
		},
	}

	_, err := NexusEndpointRegistry(inner, namespaces, nil).
		GetByName(context.Background(), namespace.ID("missing"), EndpointSecretsCreateAccess)
	if !errors.As(err, new(*serviceerror.NamespaceNotFound)) {
		t.Fatalf("GetByName() error = %v, want NamespaceNotFound", err)
	}
}
