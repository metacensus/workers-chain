package workers

import (
	"github.com/metacensus/infra/temporal/lib/workflows/helm/v1"
	"github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	"github.com/metacensus/infra/temporal/lib/workflows/temporal/v1"
	"github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.uber.org/fx"
)

var InfraModule = fx.Options(
	// Provide worker
	fx.Provide(NewWorker),

	vault.Module,
	fx.Provide(NewVaultConfig),

	helm.Module,

	kube.Module,
	fx.Provide(NewKubeConfig),

	temporal.Module,
	fx.Provide(NewTemporalClient),

	// TODO: "cluster" module
	// TODO: "cell" module
)

var FabricModule = fx.Options(
	fx.Provide(NewWorker),

	vault.Module,
	fx.Provide(NewVaultConfig),

	helm.Module,

	kube.Module,
	fx.Provide(NewKubeConfig),
)
