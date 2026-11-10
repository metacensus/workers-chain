// Package serverfx assembles the Temporal server graph with MetaCensus
// components that cannot be injected through ServerOptions.
package serverfx

import (
	"github.com/metacensus/workers-chain/temporal-server/authz"
	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/dynamicconfig"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/pprof"
	"go.temporal.io/server/common/primitives"
	"go.temporal.io/server/common/resource"
	"go.temporal.io/server/service/frontend"
	"go.temporal.io/server/service/history"
	"go.temporal.io/server/service/history/replication"
	"go.temporal.io/server/temporal"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// TopLevelModule mirrors temporal.TopLevelModule with three substitutions:
// the history service is built by historyServiceProvider, which decorates the
// Nexus endpoint registry with our caller→endpoint allowlist, and both
// frontends by frontendServiceProvider, which decorates their matching client
// with the endpoint→payload binding.
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
		frontendServiceProvider,
		internalFrontendServiceProvider,
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

func frontendServiceProvider(params temporal.ServiceProviderParamsCommon) (temporal.ServicesGroupOut, error) {
	return genericFrontendServiceProvider(params, primitives.FrontendService)
}

func internalFrontendServiceProvider(params temporal.ServiceProviderParamsCommon) (temporal.ServicesGroupOut, error) {
	return genericFrontendServiceProvider(params, primitives.InternalFrontendService)
}

// genericFrontendServiceProvider is temporal's of the same name plus the
// matching client decoration. The frontend's Nexus handler reads the request
// payload only after it authorizes, so dispatch to matching is the first place
// the payload can be checked. Both frontends serve Nexus over HTTP, so both are
// decorated.
func genericFrontendServiceProvider(
	params temporal.ServiceProviderParamsCommon,
	serviceName primitives.ServiceName,
) (temporal.ServicesGroupOut, error) {
	if _, ok := params.ServiceNames[serviceName]; !ok {
		params.Logger.Info("Service is not requested, skipping initialization.", tag.Service(serviceName))
		return temporal.ServicesGroupOut{}, nil
	}

	app := fx.New(
		params.GetCommonServiceOptions(serviceName),
		fx.Supply(params.CustomFrontendInterceptors),
		fx.Supply([]grpc.StreamServerInterceptor{}),
		fx.Decorate(func() authorization.ClaimMapper {
			switch serviceName {
			case primitives.FrontendService:
				return params.ClaimMapper
			case primitives.InternalFrontendService:
				return authorization.NewNoopClaimMapper()
			default:
				panic("Unexpected frontend service name")
			}
		}),
		fx.Decorate(func() log.SnTaggedLogger {
			// Use "frontend" for logs even if serviceName is "internal-frontend", but add an
			// extra tag to differentiate.
			tags := []tag.Tag{tag.Service(primitives.FrontendService)}
			if serviceName == primitives.InternalFrontendService {
				tags = append(tags, tag.NewBoolTag("internal-frontend", true))
			}
			return log.With(params.Logger, tags...)
		}),
		frontend.Module,
		fx.Decorate(authz.NexusMatchingClient),
	)

	return temporal.NewService(app, serviceName, params.Logger), app.Err()
}
