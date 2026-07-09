package fabric

import (
	"github.com/metacensus/infra/temporal/lib/workers"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	clusterfabric "github.com/metacensus/infra/temporal/lib/workflows/cluster/fabric/v1"
	certauth "github.com/metacensus/infra/temporal/lib/workflows/fabric/certauth/v1"
	fabric "github.com/metacensus/infra/temporal/lib/workflows/fabric/v1"
	helm "github.com/metacensus/infra/temporal/lib/workflows/helm/v1"
	kube "github.com/metacensus/infra/temporal/lib/workflows/kube/v1"
	postgres "github.com/metacensus/infra/temporal/lib/workflows/postgres/v1"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	temporal "github.com/metacensus/infra/temporal/lib/workflows/temporal/v1"
	vault "github.com/metacensus/infra/temporal/lib/workflows/vault/v1"
	"go.uber.org/fx"
)

// Module provides all dependencies for the shared fabric worker.
// The fabric worker is operator-owned (OPERATOR_ID only, no ORG_ID) and runs in
// the shared "fabric" Temporal namespace. It owns the shared TLS certificate
// authority chain and the certificate store, receives BootstrapFabric via Nexus
// from the central infra worker, and serves CreateCertificate Nexus calls from
// shard workers for cross-namespace TLS issuance.
var Module = fx.Module("fabric",
	// Load environment config (operator-owned: no org config)
	fx.Provide(
		config.GetTemporal,
		config.GetVault,
		config.GetWorker,
		config.GetKubernetes,
		config.GetOperator,
	),

	// Provide CertificateStoreConfig for postgres activities
	// The fabric worker uses OperatorConfig because the certificate store and
	// TLS CAs belong to the operator
	fx.Provide(func(operatorConfig *config.OperatorConfig) fabricconfig.CertificateStoreConfig {
		return fabricconfig.NewCertificateStoreConfig(operatorConfig)
	}),

	// Provide temporal worker
	fx.Provide(workers.NewWorker),

	// Provide temporal clients (no admin activities/workflows needed)
	temporal.ClientModule,

	// Include workflow modules (provides workflows/activities)
	helm.Module,
	kube.Module,
	vault.Module,
	secrets.Module,
	clusterfabric.Module,
	certauth.Module,
	fabric.Module,
	postgres.Module,

	// Register all workflows and activities with the worker
	fx.Invoke(
		vault.RegisterActivities,
		secrets.RegisterActivities,
		helm.RegisterActivities,
		kube.RegisterActivities,
		clusterfabric.RegisterWorkflows,
		certauth.RegisterActivities,
		certauth.RegisterWorkflows,
		// Handles CreateCertificate (and other CertificateAuthority) Nexus calls
		// from shard workers
		certauth.RegisterNexusService,
		fabric.RegisterWorkflows,
		// Handles BootstrapFabric Nexus calls from the central infra worker
		fabric.RegisterNexusService,
		postgres.RegisterActivities,
	),
)
