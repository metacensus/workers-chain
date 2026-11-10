// Package serverfx assembles the Temporal server graph with MetaCensus
// components that cannot be injected through ServerOptions.
package serverfx

import (
	"github.com/metacensus/workers-chain/temporal-server/authz"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/pprof"
	"go.temporal.io/server/common/primitives"
	"go.temporal.io/server/common/resource"
	"go.temporal.io/server/service/history"
	"go.temporal.io/server/service/history/replication"
	"go.temporal.io/server/temporal"
	"go.uber.org/fx"
)

// TopLevelModule mirrors temporal.TopLevelModule with one substitution: the
// history service is built by historyServiceProvider, which decorates the Nexus
// endpoint registry with our caller→endpoint allowlist.
//
// Replacing the provider is the only available seam. ServerOptions has no
// per-service fx hook (upstream's WithFxOptionsForService exists only in
// Temporal's integration-test harness), and each service is a nested fx.New
// inside its provider, so a decoration at this level would not reach it.
//
// Re-diff this against temporal.TopLevelModule whenever go.temporal.io/server
// is bumped.
var TopLevelModule = fx.Options(
	fx.Provide(
		temporal.NewServerFxImpl,
		temporal.ServerOptionsProvider,
		resource.ArchivalMetadataProvider,
		temporal.TaskCategoryRegistryProvider,
		temporal.PersistenceFactoryProvider,
		historyServiceProvider,
		temporal.MatchingServiceProvider,
		temporal.FrontendServiceProvider,
		temporal.InternalFrontendServiceProvider,
		temporal.WorkerServiceProvider,
		temporal.ApplyClusterMetadataConfigProvider,
	),
	dynamicconfig.Module,
	pprof.Module,
	temporal.TraceExportModule,
	temporal.FxLogAdapter,
	fx.Invoke(temporal.ServerLifetimeHooks),
)

// historyServiceProvider is temporal.HistoryServiceProvider plus the endpoint
// registry decoration. The history service is where an in-workflow
// ScheduleNexusOperation resolves its endpoint, and the only place the calling
// workflow's namespace is known.
func historyServiceProvider(params temporal.ServiceProviderParamsCommon) (temporal.ServicesGroupOut, error) {
	serviceName := primitives.HistoryService

	if _, ok := params.ServiceNames[serviceName]; !ok {
		params.Logger.Info("Service is not requested, skipping initialization.", tag.Service(serviceName))
		return temporal.ServicesGroupOut{}, nil
	}

	app := fx.New(
		params.GetCommonServiceOptions(serviceName),
		history.QueueModule,
		history.Module,
		replication.Module,
		fx.Decorate(authz.NexusEndpointRegistry),
	)

	return temporal.NewService(app, serviceName, params.Logger), app.Err()
}
