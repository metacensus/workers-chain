package identity

import (
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	"go.temporal.io/sdk/workflow"
)

// Workflows implements the CertificateWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// BootstrapAdmin implements the BootstrapAdmin workflow
func (w *Workflows) BootstrapAdmin(ctx workflow.Context, input *pb.BootstrapAdminWorkflowInput) (pb.BootstrapAdminWorkflow, error) {
	return &bootstrapAdminWorkflow{input: input}, nil
}

// ProvisionPeer implements the ProvisionPeer workflow
func (w *Workflows) ProvisionPeer(ctx workflow.Context, input *pb.ProvisionPeerWorkflowInput) (pb.ProvisionPeerWorkflow, error) {
	return &provisionPeerWorkflow{input: input}, nil
}

// ProvisionOrderer implements the ProvisionOrderer workflow
func (w *Workflows) ProvisionOrderer(ctx workflow.Context, input *pb.ProvisionOrdererWorkflowInput) (pb.ProvisionOrdererWorkflow, error) {
	return &provisionOrdererWorkflow{input: input}, nil
}

// ProvisionClient implements the ProvisionClient workflow
func (w *Workflows) ProvisionClient(ctx workflow.Context, input *pb.ProvisionClientWorkflowInput) (pb.ProvisionClientWorkflow, error) {
	return &provisionClientWorkflow{input: input}, nil
}

// ProvisionOrg implements the ProvisionOrg workflow
func (w *Workflows) ProvisionOrg(ctx workflow.Context, input *pb.ProvisionOrgWorkflowInput) (pb.ProvisionOrgWorkflow, error) {
	return &provisionOrgWorkflow{input: input}, nil
}
