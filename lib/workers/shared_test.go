package workers_test

import (
	"context"
	"testing"

	pb "github.com/metacensus/library-temporal-kube/gen/helm/v1"
	helm "github.com/metacensus/library-temporal-kube/lib/helm/v1"
	"github.com/metacensus/workers-chain/lib/workers"
	"github.com/stretchr/testify/require"
)

func TestCharts_RendersFromEmbeddedFS(t *testing.T) {
	out, err := helm.NewActivities(workers.Charts()).RenderChart(context.Background(), &pb.RenderChartInput{
		Chart:       "fabric/namespace",
		ReleaseName: "fabric-namespace",
		ValuesYaml:  "name: fabric-shard-example-org\ncommonLabels: {managedBy: helm, partOf: fabric}\n",
	})

	require.NoError(t, err)
	require.NotEmpty(t, out.GetResources())
}
