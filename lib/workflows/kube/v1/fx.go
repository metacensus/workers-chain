package kube

import (
	pb "github.com/metacensus/workers-chain/gen/workflows/kube/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides kube workflow activities and their dependencies
var Module = fx.Module("kube",
	// Provide KubeClient implementation
	fx.Provide(NewKubeClient),

	// Provide Activities
	fx.Provide(NewActivities),
)

// RegisterActivities registers kubernetes activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) {
	pb.RegisterKubernetesActivities(w, activities)
	w.RegisterActivity(activities.EnsureTemporalMTLSClientCert)
}
