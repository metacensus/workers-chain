package s3

import (
	pb "github.com/metacensus/workers-chain/gen/workflows/provider/aws/s3"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides S3 provider v1 activities and their dependencies for dependency injection
//
// Dependency chain:
//
//	NewS3Client(context.Context) -> S3Client
//	NewActivities(S3Client) -> pb.S3ProviderActivities
//
// Usage in your main fx.App:
//
//	fx.New(
//	    s3v1.Module,
//	    fx.Invoke(s3v1.RegisterActivities),
//	)
var Module = fx.Options(
	// Provide S3 client (uses AWS config from environment)
	fx.Provide(NewS3Client),

	// Provide Activities implementing pb.S3ProviderActivities interface (depends on S3Client)
	fx.Provide(func(client S3Client) pb.S3ProviderActivities {
		return NewActivities(client)
	}),
)

// RegisterActivities registers S3 provider activities with the worker
func RegisterActivities(w worker.Worker, activities pb.S3ProviderActivities) {
	pb.RegisterS3ProviderActivities(w, activities)
}
