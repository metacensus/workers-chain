package fabric

import (
	clusterfabric "github.com/metacensus/workers-chain/lib/cluster/fabric/v1"
	certauth "github.com/metacensus/workers-chain/lib/fabric/certauth/v1"
	channel "github.com/metacensus/workers-chain/lib/fabric/channel/v1"
	orderer "github.com/metacensus/workers-chain/lib/fabric/orderer/v1"
	peer "github.com/metacensus/workers-chain/lib/fabric/peer/v1"
	fabric "github.com/metacensus/workers-chain/lib/fabric/v1"
	postgres "github.com/metacensus/workers-chain/lib/postgres/v1"
	secrets "github.com/metacensus/workers-chain/lib/secrets/v1"
	temporal "github.com/metacensus/workers-chain/lib/temporal/v1"
	vault "github.com/metacensus/workers-chain/lib/vault/v1"
	"github.com/metacensus/workers-chain/lib/workers"
	"github.com/metacensus/workers-chain/lib/workers/config"
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
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
		config.LoadOperatorConfig,
	),

	// Worker identity: the fabric worker acts AS the operator. All owned
	// workflows/activities (fabric.v1, certauth.v1) receive this OwnerConfig.
	fx.Provide(func(operatorConfig *config.OperatorConfig) config.OwnerConfig {
		return operatorConfig
	}),

	// Provide CertificateStoreConfig for postgres activities, derived from the
	// worker's owner (the certificate store and TLS CAs belong to the operator)
	fx.Provide(func(owner config.OwnerConfig) fabricconfig.CertificateStoreConfig {
		return fabricconfig.NewCertificateStoreConfig(owner)
	}),

	// Provide temporal worker
	fx.Provide(workers.NewWorker),

	// Provide temporal clients (no admin activities/workflows needed)
	temporal.ClientModule,

	// Include workflow modules (provides workflows/activities)
	workers.ClusterModule,
	vault.Module,
	secrets.Module,
	clusterfabric.Module,
	certauth.Module,
	channel.Module,
	orderer.Module,
	peer.Module,
	fabric.Module,
	postgres.Module,

	// Register all workflows and activities with the worker
	fx.Invoke(
		vault.RegisterActivities,
		secrets.RegisterActivities,
		workers.RegisterClusterActivities,
		clusterfabric.RegisterWorkflows,
		certauth.RegisterActivities,
		certauth.RegisterWorkflows,
		// Channel-lifecycle activities (config/governance + peer/orderer membership)
		channel.RegisterActivities,
		orderer.RegisterActivities,
		peer.RegisterActivities,
		// Handles SignCertificate Nexus calls from shard workers
		certauth.RegisterSignCertificateNexusService,
		fabric.RegisterWorkflows,
		// Handles BootstrapFabric Nexus calls from the central infra worker
		fabric.RegisterBootstrapFabricNexusService,
		postgres.RegisterActivities,
	),
)
