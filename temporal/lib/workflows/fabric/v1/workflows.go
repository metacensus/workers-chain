package fabric

import (
	"errors"
	"fmt"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	cryptopb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	postgrespb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/vault/v1/vaultv1nexustemporal"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Nexus endpoint names for cross-namespace calls to infra worker
const (
	vaultCreateAccessNexusEndpoint = "vault-CreateAccess"
)

// Sentinel errors for fabric workflows
var (
	ErrBootstrapCertificateAuthorities = errors.New("bootstrap certificate authorities")
	ErrDeployCertificateStore          = errors.New("deploy certificate store")
	ErrCreateCertificateDatabase       = errors.New("create certificate database")
	ErrDeployRootCertificateAuthority  = errors.New("deploy root certificate authority")
	ErrDeployCertificateAuthority      = errors.New("deploy certificate authority")
	ErrCreateVaultAccess               = errors.New("create vault access via nexus")
	ErrInvalidCertType                 = errors.New("invalid cert type")
)

// Workflows implements the FabricWorkflows interface from generated code
type Workflows struct {
	orgConfig    config.OrgConfig
	fabricConfig fabricconfig.FabricConfig
}

// NewWorkflows creates a new Workflows instance with org configuration
func NewWorkflows(orgConfig config.OrgConfig, fabricConfig fabricconfig.FabricConfig) *Workflows {
	return &Workflows{
		orgConfig:    orgConfig,
		fabricConfig: fabricConfig,
	}
}

// BootstrapShard implements the BootstrapShard workflow
func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return &bootstrapShardWorkflow{
		input:        input,
		orgConfig:    w.orgConfig,
		fabricConfig: w.fabricConfig,
	}, nil
}

// bootstrapShardWorkflow implements pb.BootstrapShardWorkflow
type bootstrapShardWorkflow struct {
	input        *pb.BootstrapShardWorkflowInput
	orgConfig    config.OrgConfig
	fabricConfig fabricconfig.FabricConfig
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
	return &bootstrapCertificateAuthoritiesWorkflow{
		input:        input,
		orgConfig:    w.orgConfig,
		fabricConfig: w.fabricConfig,
	}, nil
}

// bootstrapCertificateAuthoritiesWorkflow implements pb.BootstrapCertificateAuthoritiesWorkflow
type bootstrapCertificateAuthoritiesWorkflow struct {
	input        *pb.BootstrapCertificateAuthoritiesWorkflowInput
	orgConfig    config.OrgConfig
	fabricConfig fabricconfig.FabricConfig
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
	return &deployCertificateStoreWorkflow{
		input:        input,
		orgConfig:    w.orgConfig,
		fabricConfig: w.fabricConfig,
	}, nil
}

// deployCertificateStoreWorkflow implements pb.DeployCertificateStoreWorkflow
type deployCertificateStoreWorkflow struct {
	input        *pb.DeployCertificateStoreWorkflowInput
	orgConfig    config.OrgConfig
	fabricConfig fabricconfig.FabricConfig
}

// Execute deploys the shared certificate store (PostgreSQL) for the organization.
func (w *deployCertificateStoreWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateStoreOutput, error) {
	logger := workflow.GetLogger(ctx)

	// Get certificate store config
	certStore := w.fabricConfig.CertificateStore()

	// Step 1: Request Vault access via Nexus call to infra worker
	// This creates the policy and auth role needed for the certificate store pod to access its secrets
	vaultClient := vaultv1nexustemporal.NewVaultNexusClient(vaultCreateAccessNexusEndpoint)
	_, err := vaultClient.CreateAccess(ctx, &vaultpb.CreateAccessInput{
		PolicyName:          certStore.PolicyName(),
		PolicyDocument:      certStore.VaultPolicy(),
		AuthRoleName:        certStore.AuthRoleName(),
		ServiceAccountName:  certStore.ServiceAccountName(),
		KubernetesNamespace: certStore.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateVaultAccess, err)
	}

	// Step 2: Generate PostgreSQL credentials (username + password)
	_, err = cryptopb.GenerateAndStorePasswordChild(ctx, &cryptopb.GenerateAndStorePasswordInput{
		VaultPath: certStore.VaultAdminCredentialsPath(),
		Username:  "admin",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate postgres credentials: %w", err)
	}

	// Step 3: Deploy PostgreSQL StatefulSet for certificate storage
	// The StatefulSet will use Vault CSI driver to fetch credentials at runtime
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.orgConfig.OrgID,
		Component: certStore.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate store deployment: %w", err)
	}

	// Step 4: Wait for certificate store to be ready
	// Use activity retry policy to poll until ready (GetStatus fails if not ready)
	logger.Info("Waiting for certificate store to be ready", "org", w.orgConfig.OrgID)
	waitOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second, // Start checking after 5s
			BackoffCoefficient: 1.0,             // No backoff - consistent 5s interval
			MaximumInterval:    5 * time.Second, // Keep at 5s
			MaximumAttempts:    120,             // 120 * 5s = 10 minutes max
		},
	}
	waitCtx := workflow.WithActivityOptions(ctx, waitOptions)

	err = workflow.ExecuteActivity(waitCtx, kubepb.GetStatusActivityName, &kubepb.GetStatusInput{
		Resources: map[string]*kubepb.ResourceIdentifier{
			"certificate-store": certStore.ResourceIdentifier(),
		},
	}).Get(waitCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("certificate store did not become ready: %w", err)
	}

	logger.Info("Certificate store deployed and ready", "org", w.orgConfig.OrgID)
	return &pb.DeployCertificateStoreOutput{}, nil
}

// CreateCertificateDatabase implements the CreateCertificateDatabase workflow
func (w *Workflows) CreateCertificateDatabase(ctx workflow.Context, input *pb.CreateCertificateDatabaseWorkflowInput) (pb.CreateCertificateDatabaseWorkflow, error) {
	return &createCertificateDatabaseWorkflow{
		input:        input,
		orgConfig:    w.orgConfig,
		fabricConfig: w.fabricConfig,
	}, nil
}

// createCertificateDatabaseWorkflow implements pb.CreateCertificateDatabaseWorkflow
type createCertificateDatabaseWorkflow struct {
	input        *pb.CreateCertificateDatabaseWorkflowInput
	orgConfig    config.OrgConfig
	fabricConfig fabricconfig.FabricConfig
}

// Execute creates a dedicated database and user for a certificate authority.
func (w *createCertificateDatabaseWorkflow) Execute(ctx workflow.Context) (*pb.CreateCertificateDatabaseOutput, error) {
	// Extract input parameters
	certType := w.input.Req.GetCertType()
	caID := w.input.Req.GetCertificateAuthorityId()

	// Validate input
	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	if caID == "" {
		return nil, fmt.Errorf("certificate_authority_id cannot be empty")
	}

	// Get CA config
	caConfig := w.fabricConfig.CertificateAuthority(certType, caID)

	// Set up activity options with timeouts
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Step 1: Create the PostgreSQL database
	var createDBOutput *postgrespb.CreateDatabaseOutput
	err := workflow.ExecuteActivity(activityCtx, postgrespb.CreateDatabaseActivityName, &postgrespb.CreateDatabaseInput{
		DatabaseName: caConfig.DatabaseName(),
	}).Get(activityCtx, &createDBOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Step 2: Generate and store credentials (username + password) in Vault
	credentialsOutput, err := cryptopb.GenerateAndStorePasswordChild(ctx, &cryptopb.GenerateAndStorePasswordInput{
		VaultPath: caConfig.VaultDatabaseCredentialsPath(),
		Username:  caConfig.DatabaseUsername(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate credentials: %w", err)
	}

	// Step 3: Fetch the generated credentials from Vault
	// We need both username and password values to create the PostgreSQL user
	var fetchOutput *vaultpb.FetchSecretOutput
	err = workflow.ExecuteActivity(activityCtx, vaultpb.FetchSecretActivityName, &vaultpb.FetchSecretInput{
		Path:    credentialsOutput.GetSecretRef().GetPath(),
		Version: credentialsOutput.GetSecretRef().GetVersion(),
	}).Get(activityCtx, &fetchOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials from Vault: %w", err)
	}

	// Extract username and password from secret data
	usernameBytes, ok := fetchOutput.GetData()["username"]
	if !ok {
		return nil, fmt.Errorf("username not found in Vault secret")
	}
	username := string(usernameBytes)

	passwordBytes, ok := fetchOutput.GetData()["password"]
	if !ok {
		return nil, fmt.Errorf("password not found in Vault secret")
	}
	password := string(passwordBytes)

	// Step 4: Create the PostgreSQL user with the generated credentials
	var createUserOutput *postgrespb.CreateUserOutput
	err = workflow.ExecuteActivity(activityCtx, postgrespb.CreateUserActivityName, &postgrespb.CreateUserInput{
		Username: username,
		Password: password,
	}).Get(activityCtx, &createUserOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Step 5: Grant all privileges on the database to the user
	var grantOutput *postgrespb.GrantPrivilegesOutput
	err = workflow.ExecuteActivity(activityCtx, postgrespb.GrantPrivilegesActivityName, &postgrespb.GrantPrivilegesInput{
		Username:     username,
		DatabaseName: caConfig.DatabaseName(),
		Privileges:   []postgrespb.Privilege{postgrespb.Privilege_PRIVILEGE_ALL}, // Full access for CA to manage its certificates
	}).Get(activityCtx, &grantOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to grant privileges: %w", err)
	}

	return &pb.CreateCertificateDatabaseOutput{
		CredentialsRef: credentialsOutput.GetSecretRef(),
	}, nil
}

// DeployRootCertificateAuthority implements the DeployRootCertificateAuthority workflow
func (w *Workflows) DeployRootCertificateAuthority(ctx workflow.Context, input *pb.DeployRootCertificateAuthorityWorkflowInput) (pb.DeployRootCertificateAuthorityWorkflow, error) {
	return &deployRootCertificateAuthorityWorkflow{
		input:        input,
		orgConfig:    w.orgConfig,
		fabricConfig: w.fabricConfig,
	}, nil
}

// deployRootCertificateAuthorityWorkflow implements pb.DeployRootCertificateAuthorityWorkflow
type deployRootCertificateAuthorityWorkflow struct {
	input        *pb.DeployRootCertificateAuthorityWorkflowInput
	orgConfig    config.OrgConfig
	fabricConfig fabricconfig.FabricConfig
}

// Execute deploys the root certificate authority for the organization.
func (w *deployRootCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployRootCertificateAuthorityOutput, error) {
	certType := w.input.Req.GetCertType()

	// Validate cert type
	if certType != "tls" && certType != "org" {
		return nil, ErrInvalidCertType
	}

	// Get root CA config
	rootCA := w.fabricConfig.CertificateAuthority(certType, "root")

	// Request Vault access via Nexus call to infra worker
	// This creates the policy and auth role needed for the root CA pod to access its secrets
	vaultClient := vaultv1nexustemporal.NewVaultNexusClient(vaultCreateAccessNexusEndpoint)
	_, err := vaultClient.CreateAccess(ctx, &vaultpb.CreateAccessInput{
		PolicyName:          rootCA.PolicyName(),
		PolicyDocument:      rootCA.VaultPolicy(),
		AuthRoleName:        rootCA.AuthRoleName(),
		ServiceAccountName:  rootCA.ServiceAccountName(),
		KubernetesNamespace: rootCA.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateVaultAccess, err)
	}

	// Create certificate database for this root CA
	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		CertType:               certType,
		CertificateAuthorityId: "root",
	}); err != nil {
		return nil, fmt.Errorf("%w (certType=%s): %w", ErrCreateCertificateDatabase, certType, err)
	}

	// Generate root CA key pair and store in Vault
	_, err = cryptopb.GenerateAndStoreKeyPairChild(ctx, &cryptopb.GenerateAndStoreKeyPairInput{
		VaultPath: rootCA.VaultKeyPairPath(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA key pair: %w", err)
	}

	// Generate self-signed root CA certificate and store in Vault
	// Root CAs are valid for 15 years
	certValidityDuration := 15 * 365 * 24 * time.Hour
	_, err = cryptopb.CreateSelfSignedCertificateChild(ctx, &cryptopb.CreateSelfSignedCertificateInput{
		VaultKeyPath:         rootCA.VaultKeyPairPath(),
		VaultCertificatePath: rootCA.VaultCertificatePath(),
		CommonName:           rootCA.ComponentName(),
		Organization:         w.orgConfig.OrgID,
		OrganizationalUnit:   rootCA.OrganizationalUnit(),
		ValidityDuration:     durationpb.New(certValidityDuration),
		MaxPathLength:        1, // Can issue intermediate CAs
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA certificate: %w", err)
	}

	// Generate bootstrap admin credentials (username + password) for CA initial setup
	// These credentials will be used in the deployment to configure the CA's initial admin user
	_, err = cryptopb.GenerateAndStorePasswordChild(ctx, &cryptopb.GenerateAndStorePasswordInput{
		VaultPath: rootCA.VaultAdminCredentialsPath(),
		Username:  "admin",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Deploy root CA pod
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.orgConfig.OrgID,
		Component: rootCA.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy root CA: %w", err)
	}

	// TODO: Bootstrap root CA admin

	return &pb.DeployRootCertificateAuthorityOutput{}, nil
}

// DeployCertificateAuthority implements the DeployCertificateAuthority workflow
func (w *Workflows) DeployCertificateAuthority(ctx workflow.Context, input *pb.DeployCertificateAuthorityWorkflowInput) (pb.DeployCertificateAuthorityWorkflow, error) {
	return &deployCertificateAuthorityWorkflow{
		input:        input,
		orgConfig:    w.orgConfig,
		fabricConfig: w.fabricConfig,
	}, nil
}

// deployCertificateAuthorityWorkflow implements pb.DeployCertificateAuthorityWorkflow
type deployCertificateAuthorityWorkflow struct {
	input        *pb.DeployCertificateAuthorityWorkflowInput
	orgConfig    config.OrgConfig
	fabricConfig fabricconfig.FabricConfig
}

// Execute deploys an intermediate certificate authority for the organization.
func (w *deployCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateAuthorityOutput, error) {
	certType := w.input.Req.GetCertType()
	caID := w.input.Req.GetId()

	// Validate cert type
	if certType != "tls" && certType != "org" {
		return nil, ErrInvalidCertType
	}

	// Get intermediate CA config
	intermediateCA := w.fabricConfig.CertificateAuthority(certType, caID)

	// Request Vault access via Nexus call to infra worker
	// This creates the policy and auth role needed for the intermediate CA pod to access its secrets
	vaultClient := vaultv1nexustemporal.NewVaultNexusClient(vaultCreateAccessNexusEndpoint)
	_, err := vaultClient.CreateAccess(ctx, &vaultpb.CreateAccessInput{
		PolicyName:          intermediateCA.PolicyName(),
		PolicyDocument:      intermediateCA.VaultPolicy(),
		AuthRoleName:        intermediateCA.AuthRoleName(),
		ServiceAccountName:  intermediateCA.ServiceAccountName(),
		KubernetesNamespace: intermediateCA.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateVaultAccess, err)
	}

	// Create certificate database for this intermediate CA
	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		CertType:               certType,
		CertificateAuthorityId: caID,
	}); err != nil {
		return nil, fmt.Errorf("%w (certType=%s, caID=%s): %w", ErrCreateCertificateDatabase, certType, caID, err)
	}

	// TODO: Scale down root CA
	// TODO: Create intermediate CA identity via root CA
	// TODO: Scale root CA back up
	// TODO: Deploy intermediate CA pod
	// TODO: Bootstrap intermediate CA admin

	return &pb.DeployCertificateAuthorityOutput{}, nil
}
