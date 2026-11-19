package fabric

import (
	"slices"
	"testing"

	fabricpb "github.com/metacensus/workers-chain/gen/fabric/v1"
	"github.com/metacensus/workers-chain/lib/workers/config"
	"github.com/metacensus/workers-chain/lib/workers/workerstest"
)

func TestWorkflows(t *testing.T) {
	got := workerstest.Registered(t, config.NewOperatorConfig("operator"), workflows)
	if want := workerstest.AllExcept(fabricpb.BootstrapShardWorkflowName); !slices.Equal(got, want) {
		t.Errorf("workflows = %v, want %v", got, want)
	}
}
