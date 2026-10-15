package helm

import (
	"testing"

	"github.com/metacensus/workers-chain/lib/workers/config"
)

func TestWorkerValues_ImageRepository(t *testing.T) {
	operator := config.NewOperatorConfig("metacensus")

	cases := []struct {
		name   string
		values *WorkerValues
		want   string
	}{
		{name: "success - shard worker", values: NewWorkerValues(operator, config.NewOrgConfig("org1")), want: "metacensus/worker-fabric-shard"},
		{name: "success - fabric worker", values: NewOperatorWorkerValues(operator), want: "metacensus/worker-fabric"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.values.Image.Repository; got != tc.want {
				t.Errorf("Image.Repository = %q, want %q", got, tc.want)
			}
		})
	}
}
