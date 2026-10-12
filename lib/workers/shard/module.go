package shard

import (
	"github.com/metacensus/workers-chain/lib/workers"
	"github.com/metacensus/workers-chain/lib/workers/config"
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
	clusterfabric "github.com/metacensus/workers-chain/lib/workflows/cluster/fabric/v1"
	certauth "github.com/metacensus/workers-chain/lib/workflows/fabric/certauth/v1"
	channel "github.com/metacensus/workers-chain/lib/workflows/fabric/channel/v1"
	orderer "github.com/metacensus/workers-chain/lib/workflows/fabric/orderer/v1"
	peer "github.com/metacensus/workers-chain/lib/workflows/fabric/peer/v1"
	fabric "github.com/metacensus/workers-chain/lib/workflows/fabric/v1"
	helm "github.com/metacensus/workers-chain/lib/workflows/helm/v1"
	kube "github.com/metacensus/workers-chain/lib/workflows/kube/v1"
	postgres "github.com/metacensus/workers-chain/lib/workflows/postgres/v1"
	secrets "github.com/metacensus/workers-chain/lib/workflows/secrets/v1"
	temporal "github.com/metacensus/workers-chain/lib/workflows/temporal/v1"
	vault "github.com/metacensus/workers-chain/lib/workflows/vault/v1"
	"go.uber.org/fx"
)

// Module provides all dependencies for a per-org shard worker.
// Shard workers run in the org's "fabric-shard-{org}" Temporal namespace and
// deploy org-owned Fabric components (org CAs, peers, orderers, couchdb).
// They receive BootstrapShard via Nexus from the central infra worker and call
// the shared fabric worker via Nexus for TLS certificate issuance.
var Module = fx.Module("shard",
	// Load environment config.
	// OPERATOR_ID stays available: cluster workflows need the operator's TLS
	// root CA config for component trust stores (platform-level context).
	fx.Provide(
		config.GetTemporal,
		config.GetVault,
		config.GetWorker,
		config.GetKubernetes,
		config.LoadOrgConfig,
		config.LoadOperatorConfig,
	),

	// Worker identity: a shard worker acts AS its org. All owned
	// workflows/activities (fabric.v1, certauth.v1) receive this OwnerConfig.
	fx.Provide(func(orgCfg *config.OrgConfig) config.OwnerConfig {
		return orgCfg
	}),

	// Provide CertificateStoreConfig for postgres activities, derived from the
	// worker's owner (each org owns its certificate store)
	fx.Provide(func(owner config.OwnerConfig) fabricconfig.CertificateStoreConfig {
		return fabricconfig.NewCertificateStoreConfig(owner)
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
	channel.Module,
	orderer.Module,
	peer.Module,
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
		// Channel-lifecycle activities (config/governance + peer/orderer membership)
		channel.RegisterActivities,
		orderer.RegisterActivities,
		peer.RegisterActivities,
		fabric.RegisterWorkflows,
		// Handles BootstrapShard Nexus calls from the central infra worker
		fabric.RegisterBootstrapShardNexusService,
		postgres.RegisterActivities,
	),
)
