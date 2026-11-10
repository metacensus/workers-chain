package helm

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	helmpb "github.com/metacensus/library-temporal-kube/gen/helm/v1"
	helmv1 "github.com/metacensus/library-temporal-kube/lib/helm/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// coreYAMLVM mirrors the vm section of hyperledger/fabric-peer:2.5.0's
// core.yaml: the endpoint the peer turns into a Docker health check, and a
// comment naming the same socket that must be left alone.
const coreYAMLVM = `vm:

    # Endpoint of the vm management system.  For docker can be one of the following in general
    # unix:///var/run/docker.sock
    endpoint: unix:///var/run/docker.sock

    docker:
        tls:
            enabled: false
`

// TestPeerChart_NoDockerHealthCheck renders the peer chart as the worker does
// and runs the peer container's startup script against a core.yaml shaped like
// the image's: vm.endpoint must come out empty, or the peer registers a Docker
// health check and /healthz answers 503 in a pod with no Docker socket (#38).
// An empty CORE_VM_ENDPOINT does not do it: Fabric's Viper ignores empty
// variables.
func TestPeerChart_NoDockerHealthCheck(t *testing.T) {
	peer := renderContainer(t, "fabric/peer", "peer")

	var cfgPath string
	for _, env := range peer.Env {
		if env.Name == "FABRIC_CFG_PATH" {
			cfgPath = env.Value
		}
	}
	if cfgPath == "" {
		t.Fatal("peer container does not set FABRIC_CFG_PATH")
	}
	if !mounts(peer, cfgPath) {
		t.Fatalf("FABRIC_CFG_PATH %q is not a mounted volume", cfgPath)
	}
	if len(peer.Command) != 3 || peer.Command[0] != "sh" || peer.Command[1] != "-c" {
		t.Fatalf("command = %q, want sh -c <script>", peer.Command)
	}
	script := peer.Command[2]
	if !strings.HasSuffix(strings.TrimSpace(script), "exec peer node start") {
		t.Fatalf("script does not end by starting the peer:\n%s", script)
	}

	run := func(t *testing.T, core string) (string, error) {
		dir := t.TempDir()
		src := filepath.Join(dir, "image", "core.yaml")
		dst := filepath.Join(dir, "config")
		if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(dst, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(src, []byte(core), 0o644); err != nil {
			t.Fatal(err)
		}
		local := strings.NewReplacer(
			"/var/hyperledger/fabric/config/core.yaml", src,
			cfgPath+"/", dst+"/",
			"exec peer node start", "true",
		).Replace(script)
		out, err := exec.Command("sh", "-c", local).CombinedOutput()
		if err != nil {
			return string(out), err
		}
		written, err := os.ReadFile(filepath.Join(dst, "core.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		return string(written), nil
	}

	t.Run("blanks vm.endpoint", func(t *testing.T) {
		got, err := run(t, coreYAMLVM)
		if err != nil {
			t.Fatalf("script failed: %v\n%s", err, got)
		}
		if !strings.Contains(got, "\n    endpoint: \"\"\n") {
			t.Errorf("vm.endpoint not blanked:\n%s", got)
		}
		if !strings.Contains(got, "    # unix:///var/run/docker.sock\n") {
			t.Errorf("comment changed:\n%s", got)
		}
		if strings.Count(got, "docker.sock") != 1 {
			t.Errorf("want only the comment to name the socket:\n%s", got)
		}
	})
	t.Run("refuses to start when the endpoint line moved", func(t *testing.T) {
		moved := strings.Replace(coreYAMLVM, "    endpoint: unix:///var/run/docker.sock", "    endpoint: tcp://docker:2375", 1)
		if out, err := run(t, moved); err == nil {
			t.Errorf("script succeeded on a core.yaml it could not fix:\n%s", out)
		}
	})
}

// renderContainer renders chart with its values.example.yaml, as the worker
// does, and returns the named container of its StatefulSet.
func renderContainer(t *testing.T, chart, name string) corev1.Container {
	t.Helper()
	values, err := os.ReadFile("../../../../../helm/" + chart + "/values.example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	out, err := helmv1.NewActivities(os.DirFS("../../../../../helm")).RenderChart(context.Background(), &helmpb.RenderChartInput{
		Chart:       chart,
		ReleaseName: strings.ReplaceAll(chart, "/", "-"),
		Namespace:   "fabric-shard-org1",
		ValuesYaml:  string(values),
	})
	if err != nil {
		t.Fatalf("render %s chart: %v", chart, err)
	}
	for _, resource := range out.GetResources() {
		var sts appsv1.StatefulSet
		if err := yaml.Unmarshal(resource.GetManifestYaml(), &sts); err != nil || sts.Kind != "StatefulSet" {
			continue
		}
		for _, container := range sts.Spec.Template.Spec.Containers {
			if container.Name == name {
				return container
			}
		}
	}
	t.Fatalf("%s rendered no %s container", chart, name)
	return corev1.Container{}
}

func mounts(container corev1.Container, path string) bool {
	for _, mount := range container.VolumeMounts {
		if mount.MountPath == path {
			return true
		}
	}
	return false
}
