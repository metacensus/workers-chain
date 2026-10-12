package infra

import (
	"github.com/metacensus/workers-chain/lib/workers"
	"github.com/metacensus/workers-chain/lib/workers/config"
	cell "github.com/metacensus/workers-chain/lib/workflows/cell/v1"
	clusterfabric "github.com/metacensus/workers-chain/lib/workflows/cluster/fabric/v1"
	helm "github.com/metacensus/workers-chain/lib/workflows/helm/v1"
	kube "github.com/metacensus/workers-chain/lib/workflows/kube/v1"
	secrets "github.com/metacensus/workers-chain/lib/workflows/secrets/v1"
	temporal "github.com/metacensus/workers-chain/lib/workflows/temporal/v1"
	vault "github.com/metacensus/workers-chain/lib/workflows/vault/v1"
	"go.uber.org/fx"
)

// Module provides all dependencies for the infra worker.
// The infra worker runs in the "infra" Temporal namespace and orchestrates
// platform bootstrap (cell workflows), Temporal admin, secrets access, and
// cluster provisioning. Fabric-specific work (TLS CAs, certificate issuance,
// shard components) is delegated via Nexus to the fabric and shard workers.
var Module = fx.Module("infra",
	// Load environment config.
	// The infra worker acts ON BEHALF OF explicit owners; the operator config
	// provides platform-level context (operator identity for InitFabric and
	// TLS root trust for cluster deployments).
	fx.Provide(
		config.GetTemporal,
		config.GetVault,
		config.GetWorker,
		config.GetKubernetes,
		config.LoadOperatorConfig,
	),

	// Provide temporal worker
	fx.Provide(workers.NewWorker),

	// Include workflow modules (provides workflows/activities)
	vault.Module,
	secrets.Module,
	helm.Module,
	kube.Module,
	temporal.Module,
	cell.Module,
	clusterfabric.Module,

	// Register all workflows and activities with the worker
	fx.Invoke(
		vault.RegisterActivities,
		secrets.RegisterActivities,
		secrets.RegisterWorkflows,
		secrets.RegisterNexusService,
		helm.RegisterActivities,
		kube.RegisterActivities,
		temporal.RegisterActivities,
		temporal.RegisterWorkflows,
		cell.RegisterWorkflows,
		clusterfabric.RegisterWorkflows,
	),
)
