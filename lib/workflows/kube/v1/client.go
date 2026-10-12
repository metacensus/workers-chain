package kube

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/httpstream"
	apiremotecommand "k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	k8sexec "k8s.io/client-go/util/exec"

	pb "github.com/metacensus/workers-chain/gen/workflows/kube/v1"
	"github.com/metacensus/workers-chain/lib/workers/config"
)

// ResourceStatus represents the current state of a resource
type ResourceStatus struct {
	Ready           bool
	Message         string
	DesiredReplicas int32
	ReadyReplicas   int32
}

// KubeClient defines the interface for Kubernetes operations
// This is a pure API wrapper - no YAML parsing, assumes caller provides correct data
type KubeClient interface {
	// CreateResource creates a new resource (fails if already exists)
	CreateResource(ctx context.Context, id *pb.ResourceIdentifier, obj *unstructured.Unstructured) error

	// UpdateResource updates an existing resource (fails if not found)
	UpdateResource(ctx context.Context, id *pb.ResourceIdentifier, obj *unstructured.Unstructured) error

	// ApplyResource creates or updates a resource (upsert behavior)
	// This is a convenience method that tries update first, then create
	ApplyResource(ctx context.Context, id *pb.ResourceIdentifier, obj *unstructured.Unstructured) error

	// DeleteResource removes a resource (succeeds even if not found)
	DeleteResource(ctx context.Context, id *pb.ResourceIdentifier) error

	// GetResourceStatus queries the current status of a resource
	// Returns error if resource doesn't exist or can't determine status
	GetResourceStatus(ctx context.Context, id *pb.ResourceIdentifier) (*ResourceStatus, error)

	// ExecPod executes a command in a pod container
	ExecPod(ctx context.Context, id *pb.ContainerIdentifier, container string, command []string, stdin []byte) (stdout, stderr []byte, exitCode int32, err error)

	// GetSecretData returns a copy of secret.Data for namespace/name.
	GetSecretData(ctx context.Context, namespace, name string) (map[string][]byte, error)

	// UpsertOpaqueSecret creates or updates an Opaque secret.
	UpsertOpaqueSecret(ctx context.Context, namespace, name string, data map[string][]byte) error
}

// kubeClient implements KubeClient using client-go
type kubeClient struct {
	dynamicClient dynamic.Interface
	typedClient   kubernetes.Interface
	restConfig    *rest.Config
}

// NewKubeClient creates a new Kubernetes client with the given configuration
func NewKubeClient(cfg *config.KubernetesConfig) (KubeClient, error) {
	var restConfig *rest.Config
	var err error

	if cfg.KubeconfigPath != "" {
		restConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", cfg.KubeconfigPath, err)
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
		}
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	typedClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create typed client: %w", err)
	}

	return &kubeClient{
		dynamicClient: dynamicClient,
		typedClient:   typedClient,
		restConfig:    restConfig,
	}, nil
}

// CreateResource creates a new resource in the cluster
func (kc *kubeClient) CreateResource(ctx context.Context, id *pb.ResourceIdentifier, obj *unstructured.Unstructured) error {
	gvr := kc.parseGVR(id.ApiVersion, id.Kind)
	resourceClient := kc.getResourceClient(gvr, id.Namespace)

	_, err := resourceClient.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	return nil
}

// UpdateResource updates an existing resource in the cluster
func (kc *kubeClient) UpdateResource(ctx context.Context, id *pb.ResourceIdentifier, obj *unstructured.Unstructured) error {
	gvr := kc.parseGVR(id.ApiVersion, id.Kind)
	resourceClient := kc.getResourceClient(gvr, id.Namespace)

	_, err := resourceClient.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}

	return nil
}

// ApplyResource creates or updates a resource (upsert behavior)
func (kc *kubeClient) ApplyResource(ctx context.Context, id *pb.ResourceIdentifier, obj *unstructured.Unstructured) error {
	// Try update first, then create if not found
	err := kc.UpdateResource(ctx, id, obj)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return kc.CreateResource(ctx, id, obj)
		}
		return err
	}

	return nil
}

// DeleteResource removes a resource from the cluster
func (kc *kubeClient) DeleteResource(ctx context.Context, id *pb.ResourceIdentifier) error {
	gvr := kc.parseGVR(id.ApiVersion, id.Kind)
	resourceClient := kc.getResourceClient(gvr, id.Namespace)

	err := resourceClient.Delete(ctx, id.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

// GetResourceStatus queries the current status of a resource
func (kc *kubeClient) GetResourceStatus(ctx context.Context, id *pb.ResourceIdentifier) (*ResourceStatus, error) {
	// Use typed client for known resource types with rich status
	switch id.Kind {
	case "Deployment":
		return kc.getDeploymentStatus(ctx, id.Namespace, id.Name)
	case "StatefulSet":
		return kc.getStatefulSetStatus(ctx, id.Namespace, id.Name)
	case "DaemonSet":
		return kc.getDaemonSetStatus(ctx, id.Namespace, id.Name)
	case "Pod":
		return kc.getPodStatus(ctx, id.Namespace, id.Name)
	default:
		// For other resources, just check existence
		return kc.checkExistence(ctx, id)
	}
}

// Helper: Parse GVR from APIVersion and Kind
func (kc *kubeClient) parseGVR(apiVersion, kind string) schema.GroupVersionResource {
	group, version := parseAPIVersion(apiVersion)
	resource := kindToResource(kind)

	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
}

// Helper: Get resource client (namespaced or cluster-scoped)
func (kc *kubeClient) getResourceClient(gvr schema.GroupVersionResource, namespace string) dynamic.ResourceInterface {
	if namespace != "" {
		return kc.dynamicClient.Resource(gvr).Namespace(namespace)
	}
	return kc.dynamicClient.Resource(gvr)
}

// Status checking implementations

func (kc *kubeClient) getDeploymentStatus(ctx context.Context, namespace, name string) (*ResourceStatus, error) {
	deployment, err := kc.typedClient.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	status := &ResourceStatus{
		DesiredReplicas: *deployment.Spec.Replicas,
		ReadyReplicas:   deployment.Status.ReadyReplicas,
	}

	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas &&
		deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas {
		status.Ready = true
		status.Message = "all replicas ready"
	} else if deployment.Status.ReadyReplicas > 0 {
		status.Ready = false
		status.Message = fmt.Sprintf("%d/%d replicas ready", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
	} else {
		status.Ready = false
		status.Message = "no replicas ready"
	}

	return status, nil
}

func (kc *kubeClient) getStatefulSetStatus(ctx context.Context, namespace, name string) (*ResourceStatus, error) {
	statefulSet, err := kc.typedClient.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statefulset: %w", err)
	}

	status := &ResourceStatus{
		DesiredReplicas: *statefulSet.Spec.Replicas,
		ReadyReplicas:   statefulSet.Status.ReadyReplicas,
	}

	if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas &&
		statefulSet.Status.UpdatedReplicas == *statefulSet.Spec.Replicas {
		status.Ready = true
		status.Message = "all replicas ready"
	} else if statefulSet.Status.ReadyReplicas > 0 {
		status.Ready = false
		status.Message = fmt.Sprintf("%d/%d replicas ready", statefulSet.Status.ReadyReplicas, *statefulSet.Spec.Replicas)
	} else {
		status.Ready = false
		status.Message = "no replicas ready"
	}

	return status, nil
}

func (kc *kubeClient) getDaemonSetStatus(ctx context.Context, namespace, name string) (*ResourceStatus, error) {
	daemonSet, err := kc.typedClient.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get daemonset: %w", err)
	}

	status := &ResourceStatus{
		DesiredReplicas: daemonSet.Status.DesiredNumberScheduled,
		ReadyReplicas:   daemonSet.Status.NumberReady,
	}

	if daemonSet.Status.NumberReady == daemonSet.Status.DesiredNumberScheduled {
		status.Ready = true
		status.Message = "all pods ready"
	} else if daemonSet.Status.NumberReady > 0 {
		status.Ready = false
		status.Message = fmt.Sprintf("%d/%d pods ready", daemonSet.Status.NumberReady, daemonSet.Status.DesiredNumberScheduled)
	} else {
		status.Ready = false
		status.Message = "no pods ready"
	}

	return status, nil
}

func (kc *kubeClient) getPodStatus(ctx context.Context, namespace, name string) (*ResourceStatus, error) {
	pod, err := kc.typedClient.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	status := &ResourceStatus{}

	switch pod.Status.Phase {
	case corev1.PodRunning:
		allReady := len(pod.Status.ContainerStatuses) > 0
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				allReady = false
				break
			}
		}
		status.Ready = allReady
		if allReady {
			status.Message = "all containers running"
		} else {
			status.Message = "not all containers ready"
		}
	case corev1.PodSucceeded:
		status.Ready = true
		status.Message = "pod completed successfully"
	case corev1.PodFailed:
		status.Ready = false
		status.Message = fmt.Sprintf("pod failed: %s", pod.Status.Message)
	case corev1.PodPending:
		status.Ready = false
		status.Message = fmt.Sprintf("pod pending: %s", pod.Status.Message)
	default:
		status.Ready = false
		status.Message = "unknown pod phase"
	}

	return status, nil
}

func (kc *kubeClient) checkExistence(ctx context.Context, id *pb.ResourceIdentifier) (*ResourceStatus, error) {
	gvr := kc.parseGVR(id.ApiVersion, id.Kind)
	resourceClient := kc.getResourceClient(gvr, id.Namespace)

	_, err := resourceClient.Get(ctx, id.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("resource not found or inaccessible: %w", err)
	}

	// Resource exists, but we don't have detailed status
	return &ResourceStatus{
		Ready:   true,
		Message: "resource exists",
	}, nil
}

// Utility functions

// parseAPIVersion splits "group/version" or "version" into components
func parseAPIVersion(apiVersion string) (group, version string) {
	parts := strings.SplitN(apiVersion, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", apiVersion // Core API group
}

// kindToResource converts Kind to resource name (lowercase plural)
// Uses explicit mapping for known types, falls back to simple pluralization
func kindToResource(kind string) string {
	// Explicit mappings for common Kubernetes resources
	commonResources := map[string]string{
		"Namespace":             "namespaces",
		"Pod":                   "pods",
		"Service":               "services",
		"Deployment":            "deployments",
		"StatefulSet":           "statefulsets",
		"DaemonSet":             "daemonsets",
		"ReplicaSet":            "replicasets",
		"Job":                   "jobs",
		"CronJob":               "cronjobs",
		"ConfigMap":             "configmaps",
		"Secret":                "secrets", // pragma: allowlist secret
		"PersistentVolume":      "persistentvolumes",
		"PersistentVolumeClaim": "persistentvolumeclaims",
		"ServiceAccount":        "serviceaccounts",
		"Role":                  "roles",
		"RoleBinding":           "rolebindings",
		"ClusterRole":           "clusterroles",
		"ClusterRoleBinding":    "clusterrolebindings",
		"Ingress":               "ingresses",
		"NetworkPolicy":         "networkpolicies",
		"Endpoints":             "endpoints",
		"SecretProviderClass":   "secretproviderclasses", // Vault CSI driver // pragma: allowlist secret
	}

	if resource, ok := commonResources[kind]; ok {
		return resource
	}

	// Fallback: simple lowercase + 's'
	return strings.ToLower(kind) + "s"
}

// ExecPod executes a command in a pod container.
// Non-zero exit statuses are reported via a non-nil error, and the exitCode is populated when available.
func (kc *kubeClient) ExecPod(ctx context.Context, id *pb.ContainerIdentifier, container string, command []string, stdin []byte) (stdout, stderr []byte, exitCode int32, err error) {
	if id == nil {
		return nil, nil, 0, fmt.Errorf("container identifier is required")
	}
	namespace := id.GetNamespace()
	podName := id.GetName()
	if namespace == "" || podName == "" {
		return nil, nil, 0, fmt.Errorf("namespace and name are required in container identifier")
	}
	// Get the pod to validate it exists and check its state
	pod, err := kc.typedClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get pod %s/%s: %w", namespace, podName, err)
	}

	// Validate pod is in a runnable state
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return nil, nil, 0, fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
	}
	if pod.Status.Phase != corev1.PodRunning {
		return nil, nil, 0, fmt.Errorf("cannot exec into a container in a non-running pod; current phase is %s", pod.Status.Phase)
	}

	// Validate container exists in pod
	containerExists := false
	for _, c := range pod.Spec.Containers {
		if c.Name == container {
			containerExists = true
			break
		}
	}
	if !containerExists {
		return nil, nil, 0, fmt.Errorf("container %s not found in pod %s/%s", container, namespace, podName)
	}

	if kc.restConfig == nil {
		return nil, nil, 0, fmt.Errorf("failed to create executor: rest config is required")
	}

	// Get REST client from typed client (reuses existing connection)
	restClient := kc.typedClient.CoreV1().RESTClient()

	// Build exec request URL
	// Note: We build the request to get the URL, but don't send it directly.
	// Exec operations require a streaming SPDY connection (not a regular HTTP request),
	// so we pass the URL to remotecommand.NewSPDYExecutor which handles the streaming protocol.
	req := restClient.Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     len(stdin) > 0,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	// Prefer WebSocket exec (works in more environments), fall back to SPDY.
	wsExecutor, err := remotecommand.NewWebSocketExecutorForProtocols(
		kc.restConfig,
		"GET",
		req.URL().String(),
		apiremotecommand.StreamProtocolV5Name,
		apiremotecommand.StreamProtocolV4Name,
		apiremotecommand.StreamProtocolV3Name,
		apiremotecommand.StreamProtocolV2Name,
		apiremotecommand.StreamProtocolV1Name,
	)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create websocket executor: %w", err)
	}
	spdyExecutor, err := remotecommand.NewSPDYExecutor(kc.restConfig, "POST", req.URL())
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create spdy executor: %w", err)
	}
	executor, err := remotecommand.NewFallbackExecutor(
		wsExecutor,
		spdyExecutor,
		func(err error) bool {
			var upgradeErr *httpstream.UpgradeFailureError
			return errors.As(err, &upgradeErr)
		},
	)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create executor: %w", err)
	}

	// Capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	var stdinReader io.Reader
	if len(stdin) > 0 {
		stdinReader = bytes.NewReader(stdin)
	}

	// Execute command
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  stdinReader,
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
		Tty:    false,
	})

	// Handle execution error
	if err != nil {
		exitCode = 1
		var exitErr k8sexec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = int32(exitErr.ExitStatus())
		}
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), exitCode, fmt.Errorf("command execution failed: %w", err)
	}

	// Success - exit code 0
	exitCode = 0
	return stdoutBuf.Bytes(), stderrBuf.Bytes(), exitCode, nil
}

// TODO: Implement wait functionality
// func (kc *kubeClient) waitForResource(ctx context.Context, id ResourceIdentifier, timeout time.Duration) error {
//   - Poll GetResourceStatus until Ready or timeout
//   - Return error if timeout expires
// }

// TODO: Implement parallel operations at activity layer
// - Activities can spawn goroutines to call ApplyResource/DeleteResource concurrently
// - Collect results and aggregate into response
