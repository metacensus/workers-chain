package s3

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	pb "github.com/metacensus/infra/temporal/gen/workflows/provider/aws/s3"
)

// S3Client defines the interface for S3 operations
type S3Client interface {
	GetObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3Object, error)
	PutObject(ctx context.Context, req *pb.S3Object) (*pb.S3Object, error)
	DeleteObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3DeleteOutput, error)
}

// s3Client implements S3Client
type s3Client struct {
	client *s3.Client
}

// NewS3Client creates a new S3 client with AWS configuration from environment
// This is the constructor that fx will use to provide S3Client
func NewS3Client(ctx context.Context) (S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &s3Client{
		client: s3.NewFromConfig(cfg),
	}, nil
}

// GetObject retrieves an object from S3
func (sc *s3Client) GetObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3Object, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := sc.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(req.GetBucket()),
		Key:    aws.String(req.GetKey()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the object data
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(result.Body); err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return &pb.S3Object{
		Bucket:      req.GetBucket(),
		Key:         req.GetKey(),
		Data:        buf.Bytes(),
		ContentType: aws.ToString(result.ContentType),
		Etag:        aws.ToString(result.ETag),
	}, nil
}

// PutObject writes an object to S3
func (sc *s3Client) PutObject(ctx context.Context, req *pb.S3Object) (*pb.S3Object, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	putInput := &s3.PutObjectInput{
		Bucket: aws.String(req.GetBucket()),
		Key:    aws.String(req.GetKey()),
		Body:   bytes.NewReader(req.GetData()),
	}

	// Set content type if provided
	if req.GetContentType() != "" {
		putInput.ContentType = aws.String(req.GetContentType())
	}

	result, err := sc.client.PutObject(ctx, putInput)
	if err != nil {
		return nil, fmt.Errorf("failed to put object to S3: %w", err)
	}

	return &pb.S3Object{
		Bucket:      req.GetBucket(),
		Key:         req.GetKey(),
		Data:        req.GetData(),
		ContentType: req.GetContentType(),
		Etag:        aws.ToString(result.ETag),
		VersionId:   aws.ToString(result.VersionId),
	}, nil
}

// DeleteObject deletes an object from S3
func (sc *s3Client) DeleteObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3DeleteOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := sc.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(req.GetBucket()),
		Key:    aws.String(req.GetKey()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return &pb.S3DeleteOutput{
		DeleteMarker: aws.ToBool(result.DeleteMarker),
		VersionId:    aws.ToString(result.VersionId),
	}, nil
}
