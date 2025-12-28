package fabric

import (
	"errors"
	"fmt"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	"github.com/metacensus/infra/temporal/lib/env"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors for fabric workflows
var (
	ErrBootstrapCertificateAuthorities = errors.New("bootstrap certificate authorities")
	ErrDeployCertificateStore          = errors.New("deploy certificate store")
	ErrDeployRootCertificateAuthority  = errors.New("deploy root certificate authority")
	ErrDeployCertificateAuthority      = errors.New("deploy certificate authority")
)

// Workflows implements the FabricWorkflows interface from generated code
type Workflows struct {
	orgConfig env.OrgConfig
}

// NewWorkflows creates a new Workflows instance with org configuration
func NewWorkflows(orgConfig env.OrgConfig) *Workflows {
	return &Workflows{orgConfig: orgConfig}
}

// BootstrapShard implements the BootstrapShard workflow
func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return &bootstrapShardWorkflow{input: input, orgConfig: w.orgConfig}, nil
}

// bootstrapShardWorkflow implements pb.BootstrapShardWorkflow
type bootstrapShardWorkflow struct {
	input     *pb.BootstrapShardWorkflowInput
	orgConfig env.OrgConfig
}

// Execute deploys all Fabric components (CAs, peers, orderers) for the organization.
func (w *bootstrapShardWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapShardOutput, error) {
	// Bootstrap all certificate authorities (store, root CA, intermediate CAs)
	if _, err := pb.BootstrapCertificateAuthoritiesChild(ctx, &pb.BootstrapCertificateAuthoritiesInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapCertificateAuthorities, err)
	}

	// TODO: Deploy peers, orderers, chaincode

	return &pb.BootstrapShardOutput{}, nil
}

// BootstrapCertificateAuthorities implements the BootstrapCertificateAuthorities workflow
func (w *Workflows) BootstrapCertificateAuthorities(ctx workflow.Context, input *pb.BootstrapCertificateAuthoritiesWorkflowInput) (pb.BootstrapCertificateAuthoritiesWorkflow, error) {
	return &bootstrapCertificateAuthoritiesWorkflow{input: input, orgConfig: w.orgConfig}, nil
}

// bootstrapCertificateAuthoritiesWorkflow implements pb.BootstrapCertificateAuthoritiesWorkflow
type bootstrapCertificateAuthoritiesWorkflow struct {
	input     *pb.BootstrapCertificateAuthoritiesWorkflowInput
	orgConfig env.OrgConfig
}

// Execute orchestrates the deployment of all certificate authorities for the organization.
func (w *bootstrapCertificateAuthoritiesWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapCertificateAuthoritiesOutput, error) {
	// Deploy the shared certificate store (PostgreSQL)
	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	// Deploy root certificate authority
	if _, err := pb.DeployRootCertificateAuthorityChild(ctx, &pb.DeployRootCertificateAuthorityInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployRootCertificateAuthority, err)
	}

	// Deploy intermediate CAs for both "tls" and "org"
	for _, caType := range []string{"tls", "org"} {
		if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{}); err != nil {
			return nil, fmt.Errorf("%w (caType=%s): %w", ErrDeployCertificateAuthority, caType, err)
		}
	}

	return &pb.BootstrapCertificateAuthoritiesOutput{}, nil
}

// DeployCertificateStore implements the DeployCertificateStore workflow
func (w *Workflows) DeployCertificateStore(ctx workflow.Context, input *pb.DeployCertificateStoreWorkflowInput) (pb.DeployCertificateStoreWorkflow, error) {
	return &deployCertificateStoreWorkflow{input: input, orgConfig: w.orgConfig}, nil
}

// deployCertificateStoreWorkflow implements pb.DeployCertificateStoreWorkflow
type deployCertificateStoreWorkflow struct {
	input     *pb.DeployCertificateStoreWorkflowInput
	orgConfig env.OrgConfig
}

// Execute deploys the shared certificate store (PostgreSQL) for the organization.
func (w *deployCertificateStoreWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateStoreOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Deploying certificate store", "org", w.orgConfig.OrgID)

	// Deploy PostgreSQL StatefulSet for certificate storage
	_, err := clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.orgConfig.OrgID,
		Component: "certificate-store",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate store deployment: %w", err)
	}

	logger.Info("Certificate store deployed successfully", "org", w.orgConfig.OrgID)
	return &pb.DeployCertificateStoreOutput{}, nil
}

// DeployRootCertificateAuthority implements the DeployRootCertificateAuthority workflow
func (w *Workflows) DeployRootCertificateAuthority(ctx workflow.Context, input *pb.DeployRootCertificateAuthorityWorkflowInput) (pb.DeployRootCertificateAuthorityWorkflow, error) {
	return &deployRootCertificateAuthorityWorkflow{input: input, orgConfig: w.orgConfig}, nil
}

// deployRootCertificateAuthorityWorkflow implements pb.DeployRootCertificateAuthorityWorkflow
type deployRootCertificateAuthorityWorkflow struct {
	input     *pb.DeployRootCertificateAuthorityWorkflowInput
	orgConfig env.OrgConfig
}

// Execute deploys the root certificate authority for the organization.
func (w *deployRootCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployRootCertificateAuthorityOutput, error) {
	// TODO: Implement root CA deployment logic

	return &pb.DeployRootCertificateAuthorityOutput{}, nil
}

// DeployCertificateAuthority implements the DeployCertificateAuthority workflow
func (w *Workflows) DeployCertificateAuthority(ctx workflow.Context, input *pb.DeployCertificateAuthorityWorkflowInput) (pb.DeployCertificateAuthorityWorkflow, error) {
	return &deployCertificateAuthorityWorkflow{input: input, orgConfig: w.orgConfig}, nil
}

// deployCertificateAuthorityWorkflow implements pb.DeployCertificateAuthorityWorkflow
type deployCertificateAuthorityWorkflow struct {
	input     *pb.DeployCertificateAuthorityWorkflowInput
	orgConfig env.OrgConfig
}

// Execute deploys an intermediate certificate authority for the organization.
func (w *deployCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateAuthorityOutput, error) {
	// TODO: Implement intermediate CA deployment logic

	return &pb.DeployCertificateAuthorityOutput{}, nil
}
