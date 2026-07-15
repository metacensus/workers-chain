package certauth

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1/certauthv1nexustemporal"
	cellv1 "github.com/metacensus/infra/temporal/lib/workflows/cell/v1"
	"go.temporal.io/sdk/workflow"
)

// NexusSignCertificate invokes the issuer-owner's SignCertificate workflow via
// Nexus. Only the public CSR crosses out and only the public certificate +
// chain cross back.
func NexusSignCertificate(ctx workflow.Context, input *pb.SignCertificateInput) (*pb.SignCertificateOutput, error) {
	fabricNexusClient := certauthv1nexustemporal.NewCertificateAuthorityNexusClient(cellv1.FabricCreateCertificateNexusEndpoint)
	output, err := fabricNexusClient.SignCertificate(ctx, input, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSignCertificate, err)
	}
	return output, nil
}
