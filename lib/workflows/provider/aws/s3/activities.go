package s3

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/workflows/provider/aws/s3"
)

// Sentinel errors for S3 activities
var (
	ErrS3GetObject    = errors.New("S3 get object")
	ErrS3PutObject    = errors.New("S3 put object")
	ErrS3DeleteObject = errors.New("S3 delete object")
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
		return nil, fmt.Errorf("%w: %w", ErrS3GetObject, err)
	}

	return result, nil
}

// PutObject writes an object to S3
func (a *Activities) PutObject(ctx context.Context, req *pb.S3Object) (*pb.S3Object, error) {
	result, err := a.client.PutObject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrS3PutObject, err)
	}

	return result, nil
}

// DeleteObject deletes an object from S3
func (a *Activities) DeleteObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3DeleteOutput, error) {
	result, err := a.client.DeleteObject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrS3DeleteObject, err)
	}

	return result, nil
}
