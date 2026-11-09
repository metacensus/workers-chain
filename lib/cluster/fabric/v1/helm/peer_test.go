package helm

import (
	"context"
	"os"
	"testing"

	helmpb "github.com/metacensus/library-temporal-kube/gen/helm/v1"
	helmv1 "github.com/metacensus/library-temporal-kube/lib/helm/v1"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

// TestPeerChart_NoDockerVMEndpoint renders the peer chart as the worker does
// and asserts the peer container sets an empty VM endpoint. With the image's
// default (unix:///var/run/docker.sock) the peer registers a Docker health
// check, and /healthz answers 503 in a pod with no Docker socket (#38).
func TestPeerChart_NoDockerVMEndpoint(t *testing.T) {
	values, err := os.ReadFile("../../../../../helm/fabric/peer/values.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	out, err := helmv1.NewActivities(os.DirFS("../../../../../helm")).RenderChart(context.Background(), &helmpb.RenderChartInput{
		Chart:       "fabric/peer",
		ReleaseName: "fabric-peer",
		Namespace:   "fabric-shard-org1",
		ValuesYaml:  string(values),
	})
	if err != nil {
		t.Fatalf("render peer chart: %v", err)
	}

	var statefulSets int
	for _, resource := range out.GetResources() {
		var sts appsv1.StatefulSet
		if err := yaml.Unmarshal(resource.GetManifestYaml(), &sts); err != nil || sts.Kind != "StatefulSet" {
			continue
		}
		statefulSets++
		for _, container := range sts.Spec.Template.Spec.Containers {
			if container.Name != "peer" {
				continue
			}
			var set bool
			for _, env := range container.Env {
				if env.Name == "CORE_VM_ENDPOINT" {
					set = true
					if env.Value != "" || env.ValueFrom != nil {
						t.Errorf("CORE_VM_ENDPOINT = %q, want empty", env.Value)
					}
				}
			}
			if !set {
				t.Error("peer container does not set CORE_VM_ENDPOINT; the image default adds a Docker health check")
			}
		}
	}
	if statefulSets != 1 {
		t.Fatalf("rendered %d StatefulSets, want 1", statefulSets)
	}
}
