package kube

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides kube workflow activities and their dependencies
var Module = fx.Module("kube",
	// Provide KubeClient implementation
	fx.Provide(NewKubeClient),

	// Provide Activities
	fx.Provide(NewActivities),

	// Register activities with worker
	fx.Invoke(func(w worker.Worker, activities *Activities) {
		pb.RegisterKubernetesActivities(w, activities)
	}),
)
