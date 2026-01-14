package cluster

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides cluster/fabric v1 workflows for dependency injection
// This module only provides workflows - it uses helm.v1 and kube.v1 activities
var Module = fx.Options(
	// Provide Workflows implementing pb.ClusterWorkflows interface
	fx.Provide(func() pb.ClusterWorkflows {
		return NewWorkflows()
	}),
)

// RegisterWorkflows registers cluster/fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.ClusterWorkflows) {
	pb.RegisterClusterWorkflows(w, workflows)
}
