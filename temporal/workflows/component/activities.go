package component

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/metacensus/infra/temporal/workflows/component/kubernetes"
	"github.com/metacensus/infra/temporal/workflows/component/manifest"
	"github.com/metacensus/infra/temporal/workflows/util"
	"go.temporal.io/sdk/worker"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	terraformPath           = "/usr/local/bin/terraform"
	terraformWorkspacesPath = "/tmp/terraform-workspaces"

	ServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

	S3Bucket = "metacensus-infra-manifests"
)

// Activities struct holds dependencies for activities
type Activities struct {
	store      ManifestStore
	kubernetes *kubernetes.Wrapper
}

type ManifestStore interface {
	Get(ctx context.Context, component, version string) ([]*unstructured.Unstructured, error)
	Put(ctx context.Context, component, version string, rawManifest []byte) error
}

func RegisterActivities(w worker.Worker) {
	config, err := createK8sConfig()
	if err != nil {
		log.Fatal("Failed to create Kubernetes config:", err)
	}

	kubernetes, err := kubernetes.NewWrapper(config)

	// Create AWS session
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	if err != nil {
		log.Fatal("Failed to create AWS session:", err)
	}

	// Create S3 client
	s3Client := s3.New(awsSession)

	// Create activities instance with dependencies
	activities := &Activities{
		store:      manifest.NewS3Store(s3Client, S3Bucket),
		kubernetes: kubernetes,
	}
	// Register activities - using struct methods
	util.RegisterActivity(w, activities.Apply, ApplyActivityName)
	util.RegisterActivity(w, activities.Build, BuildActivityName)
}

// createK8sConfig creates a kubernetes config using in-cluster config or kubeconfig
func createK8sConfig() (*rest.Config, error) {
	var config *rest.Config

	// Try in-cluster config first (when running in a pod)
	if _, err := os.Stat(ServiceAccountTokenPath); err == nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		// Fall back to kubeconfig
		// TODO: does this make sense? Useful for development?
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create config from kubeconfig: %w", err)
		}
	}
	return config, nil
}
