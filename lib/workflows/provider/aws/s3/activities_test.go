package s3

import (
	"context"
	"testing"

	pb "github.com/metacensus/workers-chain/gen/workflows/provider/aws/s3"
	s3mocks "github.com/metacensus/workers-chain/mocks/workflows/provider/aws/s3"
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
			mockClient := s3mocks.NewMockS3Client(t)
			mockClient.EXPECT().GetObject(mock.Anything, tt.input).Return(tt.mockOutput, tt.mockErr)

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
			mockClient := s3mocks.NewMockS3Client(t)
			mockClient.EXPECT().PutObject(mock.Anything, tt.input).Return(tt.mockOutput, tt.mockErr)

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
			mockClient := s3mocks.NewMockS3Client(t)
			mockClient.EXPECT().DeleteObject(mock.Anything, tt.input).Return(tt.mockOutput, tt.mockErr)

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
		})
	}
}
