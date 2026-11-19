package shard

import (
	"slices"
	"testing"

	certauthpb "github.com/metacensus/workers-chain/gen/fabric/certauth/v1"
	fabricpb "github.com/metacensus/workers-chain/gen/fabric/v1"
	"github.com/metacensus/workers-chain/lib/workers/config"
	"github.com/metacensus/workers-chain/lib/workers/workerstest"
)

func TestWorkflows(t *testing.T) {
	got := workerstest.Registered(t, config.NewOrgConfig("org1"), workflows)
	if want := workerstest.AllExcept(certauthpb.SignCertificateWorkflowName, fabricpb.BootstrapFabricWorkflowName); !slices.Equal(got, want) {
		t.Errorf("workflows = %v, want %v", got, want)
	}
}
