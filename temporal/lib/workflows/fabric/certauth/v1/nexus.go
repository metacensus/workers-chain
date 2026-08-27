package certauth

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1/certauthv1nexustemporal"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/workflow"
)

// NexusSignCertificate invokes the issuer-owner's SignCertificate workflow via
// Nexus. Only the public CSR crosses out and only the public certificate +
// chain cross back.
//
// The endpoint is the caller's own (fabric-shard-{org}-CreateCertificate), which
// is what lets the server authorize the request: only this org's namespace may
// use that endpoint, and the endpoint's org must match the requester declared
// in the input. The issuer scopes the subject to that requester.
func NexusSignCertificate(ctx workflow.Context, owner config.OwnerConfig, input *pb.SignCertificateInput) (*pb.SignCertificateOutput, error) {
	if owner.IsOperator() {
		// Only orgs reach a foreign issuer: the operator owns the shared chain.
		return nil, fmt.Errorf("%w: %q cannot request cross-owner issuance", ErrSignCertificate, owner.OwnerID())
	}
	orgCfg := config.NewOrgConfig(owner.OwnerID())

	input.Requester = owner.OwnerProto()
	fabricNexusClient := certauthv1nexustemporal.NewCertificateAuthorityNexusClient(orgCfg.NexusEndpointCreateCertificate())
	output, err := fabricNexusClient.SignCertificate(ctx, input, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSignCertificate, err)
	}
	return output, nil
}
