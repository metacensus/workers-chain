package certauth

import (
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/workers-chain/gen/workflows/fabric/certauth/v1/certauthv1nexus"
	"github.com/metacensus/workers-chain/gen/workflows/fabric/certauth/v1/certauthv1temporal"
	"github.com/metacensus/workers-chain/lib/workers/config"
	"github.com/metacensus/workers-chain/lib/workflows/common"
	secrets "github.com/metacensus/workers-chain/lib/workflows/secrets/v1"
	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

var Module = fx.Options(
	// Provide Workflows implementing pb.CertificateAuthorityWorkflows interface.
	// Workflows act AS the worker's owner, injected at startup
	// (operator on the fabric worker, org on shard workers).
	fx.Provide(func(owner config.OwnerConfig) pb.CertificateAuthorityWorkflows {
		return NewWorkflows(owner)
	}),

	// Provide Client multi-client provider
	// Clients are cached because they have stable keypair/cert auth from secrets.
	// Uses MultiClientProvider because there are many CAs (tls-root, org-root, etc.)
	// and the specific CA is only known at activity call time.
	fx.Provide(func(secretProvider secrets.SecretProvider) *common.MultiClientProvider[Client, ClientConfig] {
		return common.NewMultiClientProvider(
			secretProvider,
			NewClient,
		)
	}),

	// Provide Activities with CA client providers
	// Activities get:
	// - secretProvider: for fetching/storing secrets with type safety // pragma: allowlist secret
	// - clientProvider: for cached Client access (Register/Revoke operations)
	// - owner: the worker's injected owner; every CA referenced in an
	//   activity input belongs to this owner
	fx.Provide(func(
		secretProvider secrets.SecretProvider,
		clientProvider *common.MultiClientProvider[Client, ClientConfig],
		owner config.OwnerConfig,
	) *Activities {
		return NewActivities(secretProvider, clientProvider, owner)
	}),
)

// RegisterWorkflows registers fabric workflows with the worker
func RegisterWorkflows(w worker.Worker, workflows pb.CertificateAuthorityWorkflows) {
	pb.RegisterCertificateAuthorityWorkflows(w, workflows)
}

// RegisterActivities registers fabric activities with the worker
func RegisterActivities(w worker.Worker, activities *Activities) {
	pb.RegisterIdentityActivities(w, activities)
}

// RegisterSignCertificateNexusService registers only SignCertificate on the
// CertificateAuthority Nexus service. Bootstrap/admin/identity workflows stay
// local to the owner that runs them.
//
// Takes worker.Worker (not NexusServiceRegistry) so uber/fx can inject the
// provided Worker; the registry interface alone is not in the DI graph.
func RegisterSignCertificateNexusService(w worker.Worker) error {
	return registerSignCertificateNexusService(w)
}

func registerSignCertificateNexusService(r worker.NexusServiceRegistry) error {
	h := &certauthv1temporal.CertificateAuthorityNexusHandler{}
	svc := nexus.NewService(certauthv1nexus.CertificateAuthorityServiceName)
	if err := svc.Register(h.SignCertificate(certauthv1nexus.CertificateAuthoritySignCertificateOperationName)); err != nil {
		return fmt.Errorf("register SignCertificate nexus op: %w", err)
	}
	r.RegisterNexusService(svc)
	return nil
}
