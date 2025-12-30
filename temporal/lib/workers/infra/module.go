package infra

import (
	"github.com/metacensus/infra/temporal/lib/workers"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	cell "github.com/metacensus/infra/temporal/lib/workflows/cell/v1"
	clusterfabric "github.com/metacensus/infra/temporal/lib/workflows/cluster/fabric/v1"
	crypto "github.com/metacensus/infra/temporal/lib/workflows/crypto/v1"
	helm "github.com/metacensus/infra/temporal/lib/workflows/helm/v1"
	kube "github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	temporal "github.com/metacensus/infra/temporal/lib/workflows/temporal/v1"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.uber.org/fx"
)

// Module provides all dependencies for the infra worker
var Module = fx.Module("infra",
	// Load environment config
	fx.Provide(
		config.GetTemporal,
		config.GetVault,
		config.GetWorker,
		config.GetKubernetes,
	),

	// Provide temporal worker
	fx.Provide(workers.NewWorker),

	// Include workflow modules (provides workflows/activities)
	vault.Module,
	crypto.Module,
	helm.Module,
	kube.Module,
	temporal.Module, // Provides Client, NamespaceClient, and admin activities/workflows
	cell.Module,
	clusterfabric.Module,

	// Register all workflows and activities with the worker
	fx.Invoke(
		vault.RegisterActivities,
		vault.RegisterWorkflows,
		vault.RegisterNexusService,
		crypto.RegisterActivities,
		crypto.RegisterWorkflows,
		helm.RegisterActivities,
		kube.RegisterActivities,
		temporal.RegisterActivities,
		temporal.RegisterWorkflows,
		cell.RegisterWorkflows,
		clusterfabric.RegisterWorkflows,
	),
)
