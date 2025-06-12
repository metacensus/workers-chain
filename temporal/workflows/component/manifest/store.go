package manifest

import (
	"bytes"
	"context"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type S3Store struct {
	client *s3.S3
	bucket string
}

func NewS3Store(client *s3.S3, bucket string) *S3Store {
	return &S3Store{
		client: client,
		bucket: bucket,
	}
}

func (s *S3Store) Get(ctx context.Context, component, version string) ([]*unstructured.Unstructured, error) {

	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3key(component, version)),
	})
	if err != nil {
		return nil, temporal.NewApplicationError("failed to download from S3", "S3_DOWNLOAD_ERROR", err)
	}
	defer result.Body.Close()

	manifestBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, temporal.NewApplicationError("failed to read manifest content", "READ_ERROR", err)
	}

	// Parse and collect all resources
	resources, err := parseManifests(manifestBytes)
	if err != nil {
		return nil, temporal.NewApplicationError("failed to parse manifests", "PARSE_ERROR", err)
	}

	// Apply resources in order (namespaces first, then others)
	return sortResourcesByKind(resources), nil
}

func (s *S3Store) Put(ctx context.Context, component, version string, rawManifest []byte) error {

	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s3key(component, version)),
		Body:        bytes.NewReader(rawManifest),
		ContentType: aws.String("application/x-yaml"),
		Metadata: map[string]*string{
			"timestamp":    aws.String(time.Now().Format(time.RFC3339)),
			"component":    aws.String(component),
			"version":      aws.String(version),
			"workflow-run": aws.String(activity.GetInfo(ctx).WorkflowExecution.RunID),
		},
	})
	return err
}

func s3key(component, version string) string {
	return path.Join(version, component, "manifest.yaml")
}
