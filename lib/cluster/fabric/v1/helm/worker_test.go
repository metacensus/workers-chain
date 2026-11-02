package helm

import (
	"testing"

	"github.com/metacensus/workers-chain/lib/workers/config"
)

func TestWorkerValues_ImageTag(t *testing.T) {
	config.Version = "v1.2.3"
	t.Cleanup(func() { config.Version = "" })
	operator := config.NewOperatorConfig("metacensus")

	cases := []struct {
		name   string
		values *WorkerValues
	}{
		{name: "success - shard worker", values: NewWorkerValues(operator, config.NewOrgConfig("org1"))},
		{name: "success - fabric worker", values: NewOperatorWorkerValues(operator)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.values.Image.Tag; got != "v1.2.3" {
				t.Errorf("Image.Tag = %q, want %q", got, "v1.2.3")
			}
		})
	}
}

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
