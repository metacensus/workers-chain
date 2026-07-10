package cluster

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides cluster/fabric v1 workflows for dependency injection
// This module only provides workflows - it uses helm.v1 and kube.v1 activities
var Module = fx.Options(
	// Provide Workflows implementing pb.ClusterWorkflows interface.
	// The operator config is platform-level context (TLS root CA trust
	// anchors, shard worker env); resource ownership comes from each
	// workflow input's typed Owner.
	fx.Provide(func(operator *config.OperatorConfig) pb.ClusterWorkflows {
		return NewWorkflows(operator)
	}),
)

// RegisterWorkflows registers cluster/fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.ClusterWorkflows) {
	pb.RegisterClusterWorkflows(w, workflows)
}
