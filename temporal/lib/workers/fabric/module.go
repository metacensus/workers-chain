package fabric

import (
	"github.com/metacensus/infra/temporal/lib/workers"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	clusterfabric "github.com/metacensus/infra/temporal/lib/workflows/cluster/fabric/v1"
	crypto "github.com/metacensus/infra/temporal/lib/workflows/crypto/v1"
	fabric "github.com/metacensus/infra/temporal/lib/workflows/fabric/v1"
	helm "github.com/metacensus/infra/temporal/lib/workflows/helm/v1"
	kube "github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	postgres "github.com/metacensus/infra/temporal/lib/workflows/postgres/v1"
	temporal "github.com/metacensus/infra/temporal/lib/workflows/temporal/v1"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.uber.org/fx"
)

// Module provides all dependencies for the fabric worker
var Module = fx.Module("fabric",
	// Load environment config
	fx.Provide(
		config.GetTemporal,
		config.GetVault,
		config.GetWorker,
		config.GetKubernetes,
		config.GetOrg,
	),

	// Provide fabric-specific configs
	fx.Provide(func(orgConfig *config.OrgConfig) fabricconfig.FabricConfig {
		return fabricconfig.NewFabricConfig(*orgConfig)
	}),
	fx.Provide(func(fabricConfig fabricconfig.FabricConfig) fabricconfig.CertificateStoreConfig {
		return fabricConfig.CertificateStore()
	}),

	// Provide temporal worker
	fx.Provide(workers.NewWorker),

	// Provide temporal clients (no admin activities/workflows needed)
	temporal.ClientModule,

	// Include workflow modules (provides workflows/activities)
	helm.Module,
	kube.Module,
	vault.Module,
	crypto.Module,
	clusterfabric.Module,
	fabric.Module,
	postgres.Module,

	// Register all workflows and activities with the worker
	fx.Invoke(
		vault.RegisterActivities,
		vault.RegisterWorkflows,
		crypto.RegisterActivities,
		crypto.RegisterWorkflows,
		helm.RegisterActivities,
		kube.RegisterActivities,
		clusterfabric.RegisterWorkflows,
		fabric.RegisterActivities,
		fabric.RegisterWorkflows,
		fabric.RegisterNexusService,
		postgres.RegisterActivities,
	),
)
