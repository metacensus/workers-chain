package fabric

import (
	"errors"
	"fmt"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/vault/v1/vaultv1nexustemporal"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/workflow"
)

// Nexus endpoint names for cross-namespace calls to infra worker
const (
	vaultCreateAccessNexusEndpoint = "vault-CreateAccess"
)

// Sentinel errors for fabric workflows
var (
	ErrBootstrapCertificateAuthorities = errors.New("bootstrap certificate authorities")
	ErrDeployCertificateStore          = errors.New("deploy certificate store")
	ErrDeployRootCertificateAuthority  = errors.New("deploy root certificate authority")
	ErrDeployCertificateAuthority      = errors.New("deploy certificate authority")
	ErrCreateVaultAccess               = errors.New("create vault access via nexus")
	ErrInvalidCertType                 = errors.New("invalid cert type")
)

// Workflows implements the FabricWorkflows interface from generated code
type Workflows struct {
	orgConfig config.OrgConfig
}

// NewWorkflows creates a new Workflows instance with org configuration
func NewWorkflows(orgConfig config.OrgConfig) *Workflows {
	return &Workflows{orgConfig: orgConfig}
}

// BootstrapShard implements the BootstrapShard workflow
func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return &bootstrapShardWorkflow{input: input, orgConfig: w.orgConfig}, nil
}

// bootstrapShardWorkflow implements pb.BootstrapShardWorkflow
type bootstrapShardWorkflow struct {
	input     *pb.BootstrapShardWorkflowInput
	orgConfig config.OrgConfig
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
	orgConfig config.OrgConfig
}

// Execute orchestrates the deployment of all certificate authorities for the organization.
func (w *bootstrapCertificateAuthoritiesWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapCertificateAuthoritiesOutput, error) {
	// Deploy the shared certificate store (PostgreSQL)
	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	// Deploy Certificate Authorities for both TLS and Org cert types
	for _, certType := range []string{"tls", "org"} {
		// Deploy root certificate authority for this cert type
		if _, err := pb.DeployRootCertificateAuthorityChild(ctx, &pb.DeployRootCertificateAuthorityInput{
			CertType: certType,
		}); err != nil {
			return nil, fmt.Errorf("%w (certType=%s): %w", ErrDeployRootCertificateAuthority, certType, err)
		}

		// Deploy intermediate certificate authority for this cert type
		// TODO: Generate unique ID for intermediate CA (e.g., timestamp-based or sequential)
		if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{
			CertType: certType,
			Id:       "v1", // Initial intermediate CA version
		}); err != nil {
			return nil, fmt.Errorf("%w (certType=%s): %w", ErrDeployCertificateAuthority, certType, err)
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
	orgConfig config.OrgConfig
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
	orgConfig config.OrgConfig
}

// Execute deploys the root certificate authority for the organization.
func (w *deployRootCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployRootCertificateAuthorityOutput, error) {
	logger := workflow.GetLogger(ctx)
	certType := w.input.Req.GetCertType()

	// Validate cert type
	if certType != "tls" && certType != "org" {
		return nil, ErrInvalidCertType
	}

	// Build naming helper for this root CA
	rootCA := NewRootCertificateAuthority(w.orgConfig, certType)

	logger.Info("Deploying root certificate authority",
		"org", w.orgConfig.OrgID,
		"certType", certType,
		"policyName", rootCA.PolicyName(),
	)

	// Request Vault access via Nexus call to infra worker
	// This creates the policy and auth role needed for the root CA pod to access its secrets
	vaultClient := vaultv1nexustemporal.NewVaultNexusClient(vaultCreateAccessNexusEndpoint)
	_, err := vaultClient.CreateAccess(ctx, &vaultpb.CreateAccessInput{
		PolicyName:          rootCA.PolicyName(),
		PolicyDocument:      buildRootCertificateAuthorityPolicy(rootCA),
		AuthRoleName:        rootCA.AuthRoleName(),
		ServiceAccountName:  rootCA.ServiceAccountName(),
		KubernetesNamespace: rootCA.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateVaultAccess, err)
	}

	logger.Info("Vault access created for root certificate authority",
		"org", w.orgConfig.OrgID,
		"certType", certType,
	)

	// TODO: Create certificate database
	// TODO: Generate root CA key pair
	// TODO: Generate root CA certificate
	// TODO: Deploy root CA pod
	// TODO: Bootstrap root CA admin

	return &pb.DeployRootCertificateAuthorityOutput{}, nil
}

// buildRootCertificateAuthorityPolicy creates a Vault policy document for a root CA.
// The policy grants read/write access to the CA's secrets path.
func buildRootCertificateAuthorityPolicy(ca *FabricCertificateAuthority) *vaultpb.PolicyDocument {
	return &vaultpb.PolicyDocument{
		Rules: []*vaultpb.PolicyRule{
			// KV v2 data path - read/write access to CA secrets
			{
				Path: ca.VaultSecretDataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_CREATE,
					vaultpb.Capability_CAPABILITY_READ,
					vaultpb.Capability_CAPABILITY_UPDATE,
					vaultpb.Capability_CAPABILITY_LIST,
				},
			},
			// KV v2 metadata path - list and read metadata
			{
				Path: ca.VaultSecretMetadataPath(),
				Capabilities: []vaultpb.Capability{
					vaultpb.Capability_CAPABILITY_LIST,
					vaultpb.Capability_CAPABILITY_READ,
				},
			},
		},
	}
}

// DeployCertificateAuthority implements the DeployCertificateAuthority workflow
func (w *Workflows) DeployCertificateAuthority(ctx workflow.Context, input *pb.DeployCertificateAuthorityWorkflowInput) (pb.DeployCertificateAuthorityWorkflow, error) {
	return &deployCertificateAuthorityWorkflow{input: input, orgConfig: w.orgConfig}, nil
}

// deployCertificateAuthorityWorkflow implements pb.DeployCertificateAuthorityWorkflow
type deployCertificateAuthorityWorkflow struct {
	input     *pb.DeployCertificateAuthorityWorkflowInput
	orgConfig config.OrgConfig
}

// Execute deploys an intermediate certificate authority for the organization.
func (w *deployCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateAuthorityOutput, error) {
	// TODO: Implement intermediate CA deployment logic

	return &pb.DeployCertificateAuthorityOutput{}, nil
}
