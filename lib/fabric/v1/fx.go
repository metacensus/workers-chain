package fabric

import (
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/fabric/v1"
	"github.com/metacensus/workers-chain/gen/fabric/v1/fabricv1nexus"
	"github.com/metacensus/workers-chain/gen/fabric/v1/fabricv1temporal"
	"github.com/metacensus/workers-chain/lib/workers/config"
	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides fabric v1 deployment orchestration workflows for dependency injection.
//
// CA client access (BootstrapClient, cached Client via MultiClientProvider) lives in
// the certauth module (lib/fabric/certauth/v1); fabric workflows call
// certauth workflows as children or via Nexus rather than talking to CAs directly.
var Module = fx.Options(
	// Provide Workflows implementing pb.FabricWorkflows interface.
	// Workflows act AS the worker's owner, injected at startup
	// (operator on the fabric worker, org on shard workers), with the
	// operator as platform context.
	fx.Provide(func(owner config.OwnerConfig, operator *config.OperatorConfig) pb.FabricWorkflows {
		return NewWorkflows(owner, operator)
	}),
)

// RegisterWorkflows registers the workflows every owner runs locally;
// RegisterBootstrapFabric and RegisterBootstrapShard register the rest.
func RegisterWorkflows(w worker.Worker, workflows pb.FabricWorkflows) {
	pb.RegisterAddOrdererToChannelWorkflow(w, workflows.AddOrdererToChannel)
	pb.RegisterAddOrdererToConsenterSetWorkflow(w, workflows.AddOrdererToConsenterSet)
	pb.RegisterAddOrgToChannelWorkflow(w, workflows.AddOrgToChannel)
	pb.RegisterAddPeerToChannelWorkflow(w, workflows.AddPeerToChannel)
	pb.RegisterCreateCertificateDatabaseWorkflow(w, workflows.CreateCertificateDatabase)
	pb.RegisterCreateChannelWorkflow(w, workflows.CreateChannel)
	pb.RegisterDeployCertificateAuthorityWorkflow(w, workflows.DeployCertificateAuthority)
	pb.RegisterDeployCertificateStoreWorkflow(w, workflows.DeployCertificateStore)
	pb.RegisterDeployChainCodeWorkflow(w, workflows.DeployChainCode)
	pb.RegisterDeployCouchDbWorkflow(w, workflows.DeployCouchDb)
	pb.RegisterDeployOrdererWorkflow(w, workflows.DeployOrderer)
	pb.RegisterDeployPeerWorkflow(w, workflows.DeployPeer)
	pb.RegisterJoinOrdererToChannelWorkflow(w, workflows.JoinOrdererToChannel)
	pb.RegisterJoinPeerToChannelWorkflow(w, workflows.JoinPeerToChannel)
	pb.RegisterSetAnchorPeerWorkflow(w, workflows.SetAnchorPeer)
	pb.RegisterUpdateChannelConfigWorkflow(w, workflows.UpdateChannelConfig)
}

// RegisterBootstrapFabric registers BootstrapFabric's workflow and Nexus operation.
func RegisterBootstrapFabric(w worker.Worker, workflows pb.FabricWorkflows) error {
	pb.RegisterBootstrapFabricWorkflow(w, workflows.BootstrapFabric)
	return registerFabricNexusOps(w, func(h *fabricv1temporal.FabricNexusHandler) nexus.RegisterableOperation {
		return h.BootstrapFabric(fabricv1nexus.FabricBootstrapFabricOperationName)
	})
}

// RegisterBootstrapShard registers BootstrapShard's workflow and Nexus operation.
func RegisterBootstrapShard(w worker.Worker, workflows pb.FabricWorkflows) error {
	pb.RegisterBootstrapShardWorkflow(w, workflows.BootstrapShard)
	return registerFabricNexusOps(w, func(h *fabricv1temporal.FabricNexusHandler) nexus.RegisterableOperation {
		return h.BootstrapShard(fabricv1nexus.FabricBootstrapShardOperationName)
	})
}

func registerFabricNexusOps(r worker.NexusServiceRegistry, ops ...func(*fabricv1temporal.FabricNexusHandler) nexus.RegisterableOperation) error {
	h := &fabricv1temporal.FabricNexusHandler{}
	svc := nexus.NewService(fabricv1nexus.FabricServiceName)
	registered := make([]nexus.RegisterableOperation, 0, len(ops))
	for _, op := range ops {
		registered = append(registered, op(h))
	}
	if err := svc.Register(registered...); err != nil {
		return fmt.Errorf("register fabric nexus ops: %w", err)
	}
	r.RegisterNexusService(svc)
	return nil
}
