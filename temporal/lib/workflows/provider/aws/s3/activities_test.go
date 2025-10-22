package s3

import (
	"context"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/provider/aws/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetObject(t *testing.T) {
	tests := []struct {
		name          string
		input         *pb.S3ObjectRef
		mockOutput    *pb.S3Object
		mockErr       error
		expectErr     bool
		expectContent []byte
	}{
		{
			name: "successful get",
			input: &pb.S3ObjectRef{
				Bucket: "test-bucket",
				Key:    "test-key",
			},
			mockOutput: &pb.S3Object{
				Bucket:      "test-bucket",
				Key:         "test-key",
				Data:        []byte("test data"),
				ContentType: "text/plain",
				Etag:        "abc123",
			},
			expectErr:     false,
			expectContent: []byte("test data"),
		},
		{
			name: "get fails",
			input: &pb.S3ObjectRef{
				Bucket: "test-bucket",
				Key:    "missing-key",
			},
			mockErr:   assert.AnError,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockS3Client{}
			mockClient.On("GetObject", mock.Anything, tt.input).Return(tt.mockOutput, tt.mockErr)

			activities := NewActivities(mockClient)
			result, err := activities.GetObject(context.Background(), tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectContent, result.GetData())
				assert.Equal(t, "abc123", result.GetEtag())
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestPutObject(t *testing.T) {
	tests := []struct {
		name       string
		input      *pb.S3Object
		mockOutput *pb.S3Object
		mockErr    error
		expectErr  bool
	}{
		{
			name: "successful put",
			input: &pb.S3Object{
				Bucket:      "test-bucket",
				Key:         "test-key",
				Data:        []byte("test data"),
				ContentType: "text/plain",
			},
			mockOutput: &pb.S3Object{
				Bucket:      "test-bucket",
				Key:         "test-key",
				Data:        []byte("test data"),
				ContentType: "text/plain",
				Etag:        "abc123",
				VersionId:   "v1",
			},
			expectErr: false,
		},
		{
			name: "put fails",
			input: &pb.S3Object{
				Bucket: "test-bucket",
				Key:    "test-key",
				Data:   []byte("test data"),
			},
			mockErr:   assert.AnError,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockS3Client{}
			mockClient.On("PutObject", mock.Anything, tt.input).Return(tt.mockOutput, tt.mockErr)

			activities := NewActivities(mockClient)
			result, err := activities.PutObject(context.Background(), tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "abc123", result.GetEtag())
				assert.Equal(t, "v1", result.GetVersionId())
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestDeleteObject(t *testing.T) {
	tests := []struct {
		name       string
		input      *pb.S3ObjectRef
		mockOutput *pb.S3DeleteOutput
		mockErr    error
		expectErr  bool
	}{
		{
			name: "successful delete",
			input: &pb.S3ObjectRef{
				Bucket: "test-bucket",
				Key:    "test-key",
			},
			mockOutput: &pb.S3DeleteOutput{
				DeleteMarker: false,
				VersionId:    "v1",
			},
			expectErr: false,
		},
		{
			name: "delete fails",
			input: &pb.S3ObjectRef{
				Bucket: "test-bucket",
				Key:    "test-key",
			},
			mockErr:   assert.AnError,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockS3Client{}
			mockClient.On("DeleteObject", mock.Anything, tt.input).Return(tt.mockOutput, tt.mockErr)

			activities := NewActivities(mockClient)
			result, err := activities.DeleteObject(context.Background(), tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.False(t, result.GetDeleteMarker())
				assert.Equal(t, "v1", result.GetVersionId())
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// mockS3Client is a simple mock implementation for testing
type mockS3Client struct {
	mock.Mock
}

func (m *mockS3Client) GetObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3Object, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.S3Object), args.Error(1)
}

func (m *mockS3Client) PutObject(ctx context.Context, req *pb.S3Object) (*pb.S3Object, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.S3Object), args.Error(1)
}

func (m *mockS3Client) DeleteObject(ctx context.Context, req *pb.S3ObjectRef) (*pb.S3DeleteOutput, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.S3DeleteOutput), args.Error(1)
}
