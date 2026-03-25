//go:build integration
// +build integration

package kube

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/tools/clientcmd"
)

// setupK3sCluster starts a k3s container and returns a KubeClient configured to use it.
// Caller must ensure the returned cleanup is invoked (e.g. defer).
func setupK3sCluster(t *testing.T) (KubeClient, func()) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	tmpDir, err := os.MkdirTemp("", "k3s-kubeconfig-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	cleanupDir := func() { os.RemoveAll(tmpDir) }

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig.yaml")

	var container testcontainers.Container
	cleanup := func() {
		if container != nil {
			if err := container.Terminate(ctx); err != nil {
				t.Logf("failed to terminate container: %v", err)
			}
		}
		cleanupDir()
	}

	success := false
	defer func() {
		if !success {
			cleanup()
		}
	}()

	req := testcontainers.ContainerRequest{
		Image:        "rancher/k3s:v1.28.5-k3s1",
		ExposedPorts: []string{"6443/tcp"},
		Env: map[string]string{
			"K3S_KUBECONFIG_OUTPUT": "/output/kubeconfig.yaml",
			"K3S_KUBECONFIG_MODE":   "666",
		},
		Cmd: []string{
			"server",
			"--disable=traefik",
			"--disable=metrics-server",
			"--tls-san=127.0.0.1",
		},
		Privileged: true,
		WaitingFor: wait.ForLog("Node controller sync successful").
			WithStartupTimeout(120 * time.Second),
	}

	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start k3s container: %v", err)
	}

	time.Sleep(5 * time.Second)

	reader, err := container.CopyFileFromContainer(ctx, "/output/kubeconfig.yaml")
	if err != nil {
		t.Fatalf("failed to copy kubeconfig from container: %v", err)
	}
	defer reader.Close()

	kubeconfigFile, err := os.Create(kubeconfigPath)
	if err != nil {
		t.Fatalf("failed to create kubeconfig file: %v", err)
	}
	defer kubeconfigFile.Close()

	if _, err := kubeconfigFile.ReadFrom(reader); err != nil {
		t.Fatalf("failed to write kubeconfig: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "6443")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	kubeconfigBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		t.Fatalf("failed to read kubeconfig: %v", err)
	}

	kubeconfig, err := clientcmd.Load(kubeconfigBytes)
	if err != nil {
		t.Fatalf("failed to parse kubeconfig: %v", err)
	}

	for _, cluster := range kubeconfig.Clusters {
		cluster.Server = fmt.Sprintf("https://127.0.0.1:%s", mappedPort.Port())
		cluster.InsecureSkipTLSVerify = true
		cluster.CertificateAuthorityData = nil
	}

	if err := clientcmd.WriteToFile(*kubeconfig, kubeconfigPath); err != nil {
		t.Fatalf("failed to write updated kubeconfig: %v", err)
	}

	client, err := NewKubeClient(&config.KubernetesConfig{KubeconfigPath: kubeconfigPath})
	if err != nil {
		t.Fatalf("failed to create kube client: %v", err)
	}

	success = true
	return client, cleanup
}

func decodeManifest(t *testing.T, manifest []byte) *unstructured.Unstructured {
	t.Helper()
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	if _, _, err := decoder.Decode(manifest, nil, obj); err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}
	return obj
}

// TestKubeClient_RealCluster tests against a k3s cluster running in a container
// Run with: go test -tags=integration ./temporal/lib/workflows/kube/v1
func TestKubeClient_RealCluster(t *testing.T) {
	client, cleanup := setupK3sCluster(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	t.Run("create and delete service", func(t *testing.T) {
		nsObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: test-ns
`))
		nsID := &pb.ResourceIdentifier{
			ApiVersion: "v1",
			Kind:       "Namespace",
			Name:       "test-ns",
		}
		if err := client.CreateResource(ctx, nsID, nsObj); err != nil {
			t.Fatalf("failed to create namespace: %v", err)
		}

		serviceObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: test-ns
spec:
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: test
`))

		serviceID := &pb.ResourceIdentifier{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "test-service",
			Namespace:  "test-ns",
		}

		if err := client.CreateResource(ctx, serviceID, serviceObj); err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		status, err := client.GetResourceStatus(ctx, serviceID)
		if err != nil {
			t.Fatalf("failed to get service status: %v", err)
		}
		if status == nil {
			t.Errorf("expected status, got nil")
		}

		serviceObj.Object["spec"].(map[string]interface{})["ports"] = []map[string]interface{}{
			{"port": int64(443), "targetPort": int64(8443)},
		}
		if err := client.UpdateResource(ctx, serviceID, serviceObj); err != nil {
			t.Fatalf("failed to update service: %v", err)
		}

		if err := client.DeleteResource(ctx, serviceID); err != nil {
			t.Fatalf("failed to delete service: %v", err)
		}

		_, err = client.GetResourceStatus(ctx, serviceID)
		if err == nil {
			t.Error("expected error when getting status of deleted service, got nil")
		}
	})
}

// TestKubeClient_CreateResource integration tests for CreateResource
func TestKubeClient_CreateResource(t *testing.T) {
	client, cleanup := setupK3sCluster(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		nsObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: create-test-ns
`))
		id := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Namespace", Name: "create-test-ns"}
		if err := client.CreateResource(ctx, id, nsObj); err != nil {
			t.Fatalf("CreateResource: %v", err)
		}
	})

	t.Run("error - already exists", func(t *testing.T) {
		nsObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: default
`))
		id := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Namespace", Name: "default"}
		err := client.CreateResource(ctx, id, nsObj)
		if err == nil {
			t.Error("expected error when creating existing namespace, got nil")
		}
	})
}

// TestKubeClient_UpdateResource integration tests for UpdateResource
func TestKubeClient_UpdateResource(t *testing.T) {
	client, cleanup := setupK3sCluster(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	t.Run("success - create then update", func(t *testing.T) {
		nsObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: update-test-ns
`))
		nsID := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Namespace", Name: "update-test-ns"}
		if err := client.CreateResource(ctx, nsID, nsObj); err != nil {
			t.Fatalf("create namespace: %v", err)
		}

		cmObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: update-test-ns
data:
  key1: value1
`))
		cmID := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "ConfigMap", Name: "test-cm", Namespace: "update-test-ns"}
		if err := client.CreateResource(ctx, cmID, cmObj); err != nil {
			t.Fatalf("create configmap: %v", err)
		}

		cmObj.Object["data"] = map[string]interface{}{"key1": "updated"}
		if err := client.UpdateResource(ctx, cmID, cmObj); err != nil {
			t.Fatalf("UpdateResource: %v", err)
		}
	})

	t.Run("error - resource not found", func(t *testing.T) {
		cmObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: nonexistent-cm
  namespace: default
data:
  key: value
`))
		id := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "ConfigMap", Name: "nonexistent-cm", Namespace: "default"}
		err := client.UpdateResource(ctx, id, cmObj)
		if err == nil {
			t.Error("expected error when updating nonexistent resource, got nil")
		}
	})
}

// TestKubeClient_ApplyResource integration tests for ApplyResource
func TestKubeClient_ApplyResource(t *testing.T) {
	client, cleanup := setupK3sCluster(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	t.Run("create new resource", func(t *testing.T) {
		nsObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: apply-create-ns
`))
		nsID := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Namespace", Name: "apply-create-ns"}
		if err := client.ApplyResource(ctx, nsID, nsObj); err != nil {
			t.Fatalf("ApplyResource create: %v", err)
		}
	})

	t.Run("update existing resource", func(t *testing.T) {
		cmObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-cm
  namespace: default
data:
  key: initial
`))
		id := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "ConfigMap", Name: "apply-cm", Namespace: "default"}
		if err := client.ApplyResource(ctx, id, cmObj); err != nil {
			t.Fatalf("ApplyResource create: %v", err)
		}

		cmObj.Object["data"] = map[string]interface{}{"key": "updated-via-apply"}
		if err := client.ApplyResource(ctx, id, cmObj); err != nil {
			t.Fatalf("ApplyResource update: %v", err)
		}
	})
}

// TestKubeClient_GetResourceStatus integration tests for GetResourceStatus
func TestKubeClient_GetResourceStatus(t *testing.T) {
	client, cleanup := setupK3sCluster(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	t.Run("success - service", func(t *testing.T) {
		nsObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: status-test-ns
`))
		nsID := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Namespace", Name: "status-test-ns"}
		if err := client.CreateResource(ctx, nsID, nsObj); err != nil {
			t.Fatalf("create namespace: %v", err)
		}

		svcObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Service
metadata:
  name: status-svc
  namespace: status-test-ns
spec:
  ports:
  - port: 80
  selector:
    app: test
`))
		svcID := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Service", Name: "status-svc", Namespace: "status-test-ns"}
		if err := client.CreateResource(ctx, svcID, svcObj); err != nil {
			t.Fatalf("create service: %v", err)
		}

		status, err := client.GetResourceStatus(ctx, svcID)
		if err != nil {
			t.Fatalf("GetResourceStatus: %v", err)
		}
		if status == nil {
			t.Fatal("expected status, got nil")
		}
		if !status.Ready {
			t.Errorf("service should be ready, got Ready=%v Message=%q", status.Ready, status.Message)
		}
	})

	t.Run("error - not found", func(t *testing.T) {
		id := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Service", Name: "nonexistent-svc", Namespace: "default"}
		_, err := client.GetResourceStatus(ctx, id)
		if err == nil {
			t.Error("expected error when getting status of nonexistent resource, got nil")
		}
	})
}

// TestKubeClient_ExecPod integration tests for ExecPod (uses ContainerIdentifier)
func TestKubeClient_ExecPod(t *testing.T) {
	client, cleanup := setupK3sCluster(t)
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Create namespace and a running pod
	nsObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: exec-test-ns
`))
	nsID := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Namespace", Name: "exec-test-ns"}
	if err := client.CreateResource(ctx, nsID, nsObj); err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	podObj := decodeManifest(t, []byte(`
apiVersion: v1
kind: Pod
metadata:
  name: exec-pod
  namespace: exec-test-ns
spec:
  containers:
  - name: main
    image: busybox:1.36
    command: ["sleep", "300"]
  restartPolicy: Never
`))
	podID := &pb.ResourceIdentifier{ApiVersion: "v1", Kind: "Pod", Name: "exec-pod", Namespace: "exec-test-ns"}
	if err := client.CreateResource(ctx, podID, podObj); err != nil {
		t.Fatalf("create pod: %v", err)
	}

	// Wait for pod to be running
	containerID := &pb.ContainerIdentifier{Namespace: "exec-test-ns", Name: "exec-pod"}
	for i := 0; i < 60; i++ {
		status, err := client.GetResourceStatus(ctx, podID)
		if err == nil && status != nil && status.Ready {
			break
		}
		time.Sleep(time.Second)
		if i == 59 {
			t.Fatal("pod did not become ready in time")
		}
	}

	t.Run("success - stdout", func(t *testing.T) {
		stdout, stderr, exitCode, err := client.ExecPod(ctx, containerID, "main", []string{"echo", "hello"}, nil)
		if err != nil {
			t.Fatalf("ExecPod: %v", err)
		}
		if string(stdout) != "hello\n" {
			t.Errorf("stdout: got %q, want %q", string(stdout), "hello\n")
		}
		if len(stderr) != 0 {
			t.Errorf("stderr: got %q, want empty", string(stderr))
		}
		if exitCode != 0 {
			t.Errorf("exit code: got %d, want 0", exitCode)
		}
	})

	t.Run("success - stderr", func(t *testing.T) {
		stdout, stderr, exitCode, err := client.ExecPod(ctx, containerID, "main", []string{"sh", "-c", "echo hello >&2"}, nil)
		if err != nil {
			t.Fatalf("ExecPod: %v", err)
		}
		if string(stderr) != "hello\n" {
			t.Errorf("stderr: got %q, want %q", string(stderr), "hello\n")
		}
		if exitCode != 0 {
			t.Errorf("exit code: got %d, want 0", exitCode)
		}
		_ = stdout
	})

	t.Run("success - exit code", func(t *testing.T) {
		_, _, exitCode, err := client.ExecPod(ctx, containerID, "main", []string{"sh", "-c", "exit 42"}, nil)
		if err == nil {
			t.Fatal("expected error for non-zero exit code, got nil")
		}
		if exitCode != 42 {
			t.Errorf("exit code: got %d, want 42", exitCode)
		}
	})

	t.Run("error - pod not found", func(t *testing.T) {
		badID := &pb.ContainerIdentifier{Namespace: "exec-test-ns", Name: "nonexistent-pod"}
		_, _, _, err := client.ExecPod(ctx, badID, "main", []string{"echo", "x"}, nil)
		if err == nil {
			t.Error("expected error for nonexistent pod, got nil")
		}
	})

	t.Run("error - container not found", func(t *testing.T) {
		_, _, _, err := client.ExecPod(ctx, containerID, "nonexistent-container", []string{"echo", "x"}, nil)
		if err == nil {
			t.Error("expected error for nonexistent container, got nil")
		}
	})
}
