package infra

import (
	"github.com/metacensus/infra/temporal/lib/workers"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	cell "github.com/metacensus/infra/temporal/lib/workflows/cell/v1"
	clusterfabric "github.com/metacensus/infra/temporal/lib/workflows/cluster/fabric/v1"
	fabric "github.com/metacensus/infra/temporal/lib/workflows/fabric/v1"
	helm "github.com/metacensus/infra/temporal/lib/workflows/helm/v1"
	kube "github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	postgres "github.com/metacensus/infra/temporal/lib/workflows/postgres/v1"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
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
		config.GetOperator,
	),

	// Provide CertificateStoreConfig for postgres activities
	// The infra worker uses OperatorConfig because TLS CAs belong to the operator
	fx.Provide(func(operatorConfig *config.OperatorConfig) fabricconfig.CertificateStoreConfig {
		return fabricconfig.NewCertificateStoreConfig(operatorConfig)
	}),

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
	fabric.Module,
	postgres.Module,

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
		fabric.RegisterActivities,
		fabric.RegisterWorkflows,
		postgres.RegisterActivities,
	),
)
