package workers

import (
	cell "github.com/metacensus/infra/temporal/lib/workflows/cell/v1"
	clusterfabric "github.com/metacensus/infra/temporal/lib/workflows/cluster/fabric/v1"
	helm "github.com/metacensus/infra/temporal/lib/workflows/helm/v1"
	kube "github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	temporal "github.com/metacensus/infra/temporal/lib/workflows/temporal/v1"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.uber.org/fx"
)

var InfraModule = fx.Options(
	// Provide worker and clients
	fx.Provide(NewWorker),
	fx.Provide(NewTemporalClient),
	fx.Provide(NewNamespaceClient),
	fx.Provide(NewVaultConfig),
	fx.Provide(NewKubeConfig),

	// Include modules (provides workflows/activities)
	vault.Module,
	helm.Module,
	kube.Module,
	temporal.Module,
	cell.Module,
	clusterfabric.Module,

	// Register all workflows and activities with the worker
	fx.Invoke(
		vault.RegisterActivities,
		vault.RegisterWorkflows,
		helm.RegisterActivities,
		kube.RegisterActivities,
		temporal.RegisterActivities,
		temporal.RegisterWorkflows,
		cell.RegisterWorkflows,
		clusterfabric.RegisterWorkflows,
	),
)

var FabricModule = fx.Options(
	fx.Provide(NewWorker),

	vault.Module,
	fx.Provide(NewVaultConfig),

	helm.Module,

	kube.Module,
	fx.Provide(NewKubeConfig),
)
