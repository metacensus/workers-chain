package s3

import (
	"context"
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/provider/aws/s3"
)

// Activities implements the S3ProviderActivities interface from generated code
type Activities struct {
	client S3Client
}

// NewActivities creates a new Activities instance
func NewActivities(client S3Client) *Activities {
	return &Activities{
		client: client,
	}
}

// GetObject retrieves an object from S3
func (a *Activities) GetObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3Object, error) {
	result, err := a.client.GetObject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("S3 client error: %w", err)
	}

	return result, nil
}

// PutObject writes an object to S3
func (a *Activities) PutObject(ctx context.Context, req *pb.S3Object) (*pb.S3Object, error) {
	result, err := a.client.PutObject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("S3 client error: %w", err)
	}

	return result, nil
}

// DeleteObject deletes an object from S3
func (a *Activities) DeleteObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3DeleteOutput, error) {
	result, err := a.client.DeleteObject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("S3 client error: %w", err)
	}

	return result, nil
}
