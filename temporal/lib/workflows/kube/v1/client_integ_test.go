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

// TestKubeClient_RealCluster tests against a k3s cluster running in a container
// Run with: go test -tags=integration ./lib/workflows/kube/v1
func TestKubeClient_RealCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create temporary directory for kubeconfig
	tmpDir, err := os.MkdirTemp("", "k3s-kubeconfig-*")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig.yaml")

	// Start k3s container (lightweight Kubernetes)
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
		Privileged: true, // k3s requires privileged mode
		WaitingFor: wait.ForLog("Node controller sync successful").
			WithStartupTimeout(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start k3s container: %v", err)
	}
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	// Give k3s a moment to fully initialize
	time.Sleep(5 * time.Second)

	// Copy kubeconfig from container
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

	// Get the mapped port for k3s API server
	mappedPort, err := container.MappedPort(ctx, "6443")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	// Load and modify kubeconfig to use localhost and mapped port
	kubeconfigBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		t.Fatalf("failed to read kubeconfig: %v", err)
	}

	kubeconfig, err := clientcmd.Load(kubeconfigBytes)
	if err != nil {
		t.Fatalf("failed to parse kubeconfig: %v", err)
	}

	// Update server URL to use mapped port
	for _, cluster := range kubeconfig.Clusters {
		cluster.Server = fmt.Sprintf("https://127.0.0.1:%s", mappedPort.Port())
		cluster.InsecureSkipTLSVerify = true // For test purposes
		cluster.CertificateAuthorityData = nil // Must clear CA data when using InsecureSkipTLSVerify
	}

	// Write updated kubeconfig
	if err := clientcmd.WriteToFile(*kubeconfig, kubeconfigPath); err != nil {
		t.Fatalf("failed to write updated kubeconfig: %v", err)
	}

	// Create Kubernetes client using the kubeconfig
	client, err := NewKubeClient(&config.KubernetesConfig{
		KubeconfigPath: kubeconfigPath,
	})
	if err != nil {
		t.Fatalf("failed to create kube client: %v", err)
	}

	t.Run("create and delete service", func(t *testing.T) {
		// Create a namespace first
		namespaceManifest := []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: test-ns
`)
		decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		obj := &unstructured.Unstructured{}
		if _, _, err := decoder.Decode(namespaceManifest, nil, obj); err != nil {
			t.Fatalf("failed to parse namespace manifest: %v", err)
		}

		nsID := &pb.ResourceIdentifier{
			ApiVersion: "v1",
			Kind:       "Namespace",
			Name:       "test-ns",
		}

		if err := client.CreateResource(ctx, nsID, obj); err != nil {
			t.Fatalf("failed to create namespace: %v", err)
		}

		// Create a service
		serviceManifest := []byte(`
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
`)
		obj = &unstructured.Unstructured{}
		if _, _, err := decoder.Decode(serviceManifest, nil, obj); err != nil {
			t.Fatalf("failed to parse service manifest: %v", err)
		}

		serviceID := &pb.ResourceIdentifier{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "test-service",
			Namespace:  "test-ns",
		}

		// Create
		if err := client.CreateResource(ctx, serviceID, obj); err != nil {
			t.Fatalf("failed to create service: %v", err)
		}

		// Get status
		status, err := client.GetResourceStatus(ctx, serviceID)
		if err != nil {
			t.Fatalf("failed to get service status: %v", err)
		}
		if status == nil {
			t.Errorf("expected status, got nil")
		}

		// Update the service
		obj.Object["spec"].(map[string]interface{})["ports"] = []map[string]interface{}{
			{"port": int64(443), "targetPort": int64(8443)},
		}

		if err := client.UpdateResource(ctx, serviceID, obj); err != nil {
			t.Fatalf("failed to update service: %v", err)
		}

		// Delete
		if err := client.DeleteResource(ctx, serviceID); err != nil {
			t.Fatalf("failed to delete service: %v", err)
		}

		// Verify deletion (status should error)
		_, err = client.GetResourceStatus(ctx, serviceID)
		if err == nil {
			t.Error("expected error when getting status of deleted service, got nil")
		}
	})
}
