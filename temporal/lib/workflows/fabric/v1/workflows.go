package fabric

import (
	"errors"
	"fmt"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors for fabric workflows
var (
	ErrOrgRequired                     = errors.New("org is required")
	ErrBootstrapCertificateAuthorities = errors.New("bootstrap certificate authorities")
	ErrDeployCertificateStore          = errors.New("deploy certificate store")
)

// Workflows implements the FabricWorkflows interface from generated code
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// BootstrapShard implements the BootstrapShard workflow
func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return &bootstrapShardWorkflow{input: input}, nil
}

// bootstrapShardWorkflow implements pb.BootstrapShardWorkflow
type bootstrapShardWorkflow struct {
	input *pb.BootstrapShardWorkflowInput
}

// Execute deploys all Fabric components (CAs, peers, orderers) for the organization.
func (w *bootstrapShardWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapShardOutput, error) {
	if w.input.Req.Org == "" {
		return nil, ErrOrgRequired
	}

	// Bootstrap all certificate authorities (store, root CA, intermediate CAs)
	if _, err := pb.BootstrapCertificateAuthoritiesChild(ctx, &pb.BootstrapCertificateAuthoritiesInput{
		Org: w.input.Req.Org,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapCertificateAuthorities, err)
	}

	// TODO: Deploy peers, orderers, chaincode

	return &pb.BootstrapShardOutput{}, nil
}

// BootstrapCertificateAuthorities implements the BootstrapCertificateAuthorities workflow
func (w *Workflows) BootstrapCertificateAuthorities(ctx workflow.Context, input *pb.BootstrapCertificateAuthoritiesWorkflowInput) (pb.BootstrapCertificateAuthoritiesWorkflow, error) {
	return &bootstrapCertificateAuthoritiesWorkflow{input: input}, nil
}

// bootstrapCertificateAuthoritiesWorkflow implements pb.BootstrapCertificateAuthoritiesWorkflow
type bootstrapCertificateAuthoritiesWorkflow struct {
	input *pb.BootstrapCertificateAuthoritiesWorkflowInput
}

// Execute orchestrates the deployment of all certificate authorities for the organization.
func (w *bootstrapCertificateAuthoritiesWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapCertificateAuthoritiesOutput, error) {
	if w.input.Req.Org == "" {
		return nil, ErrOrgRequired
	}

	// Deploy the shared certificate store (PostgreSQL)
	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{
		Org: w.input.Req.Org,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	// TODO: Deploy root CA and intermediate CAs

	return &pb.BootstrapCertificateAuthoritiesOutput{}, nil
}

// DeployCertificateStore implements the DeployCertificateStore workflow
func (w *Workflows) DeployCertificateStore(ctx workflow.Context, input *pb.DeployCertificateStoreWorkflowInput) (pb.DeployCertificateStoreWorkflow, error) {
	return &deployCertificateStoreWorkflow{input: input}, nil
}

// deployCertificateStoreWorkflow implements pb.DeployCertificateStoreWorkflow
type deployCertificateStoreWorkflow struct {
	input *pb.DeployCertificateStoreWorkflowInput
}

// Execute deploys the shared certificate store (PostgreSQL) for the organization.
func (w *deployCertificateStoreWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateStoreOutput, error) {
	logger := workflow.GetLogger(ctx)
	if w.input.Req.Org == "" {
		return nil, ErrOrgRequired
	}

	logger.Info("Deploying certificate store", "org", w.input.Req.Org)

	// Deploy PostgreSQL StatefulSet for certificate storage
	_, err := clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.input.Req.Org,
		Component: "certificate-store",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate store deployment: %w", err)
	}

	logger.Info("Certificate store deployed successfully", "org", w.input.Req.Org)
	return &pb.DeployCertificateStoreOutput{}, nil
}
