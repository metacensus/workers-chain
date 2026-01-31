package certauth

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1/certauthv1nexustemporal"
	cellv1 "github.com/metacensus/infra/temporal/lib/workflows/cell/v1"
	"go.temporal.io/sdk/workflow"
)

func NexusCreateCertificate(ctx workflow.Context, input *pb.CreateCertificateInput) (*pb.CreateCertificateOutput, error) {
	fabricNexusClient := certauthv1nexustemporal.NewCertificateAuthorityNexusClient(cellv1.FabricCreateCertificateNexusEndpoint)
	output, err := fabricNexusClient.CreateCertificate(ctx, input, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificate, err)
	}
	return output, nil
}
