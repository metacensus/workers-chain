package fabric

import (
	"errors"
	"fmt"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	postgrespb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/secrets/v1/secretsv1nexustemporal"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Nexus endpoint name for CreateAccess calls to infra worker
const secretsCreateAccessNexusEndpoint = "secrets-CreateAccess"

// Sentinel errors for fabric workflows
var (
	ErrBootstrapCertificateAuthorities    = errors.New("bootstrap certificate authorities")
	ErrDeployCertificateStore             = errors.New("deploy certificate store")
	ErrCreateCertificateDatabase          = errors.New("create certificate database")
	ErrDeployRootCertificateAuthority     = errors.New("deploy root certificate authority")
	ErrDeployCertificateAuthority         = errors.New("deploy certificate authority")
	ErrBootstrapCertificateAuthorityAdmin = errors.New("bootstrap certificate authority admin")
	ErrCreateIdentity                     = errors.New("create identity")
	ErrCreateAccess                       = errors.New("create access")
	ErrInvalidCertType                    = errors.New("invalid cert type")
)

// Workflows implements the FabricWorkflows interface
type Workflows struct {
	fabricConfig fabricconfig.FabricConfig
}

// NewWorkflows creates a new Workflows instance
func NewWorkflows(fabricConfig fabricconfig.FabricConfig) *Workflows {
	return &Workflows{fabricConfig: fabricConfig}
}

// BootstrapShard implements the BootstrapShard workflow
func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return &bootstrapShardWorkflow{
		input:        input,
		fabricConfig: w.fabricConfig,
	}, nil
}

type bootstrapShardWorkflow struct {
	input        *pb.BootstrapShardWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

func (w *bootstrapShardWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapShardOutput, error) {
	if _, err := pb.BootstrapCertificateAuthoritiesChild(ctx, &pb.BootstrapCertificateAuthoritiesInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapCertificateAuthorities, err)
	}

	// TODO: Deploy peers, orderers, chaincode

	return &pb.BootstrapShardOutput{}, nil
}

// BootstrapCertificateAuthorities implements the workflow
func (w *Workflows) BootstrapCertificateAuthorities(ctx workflow.Context, input *pb.BootstrapCertificateAuthoritiesWorkflowInput) (pb.BootstrapCertificateAuthoritiesWorkflow, error) {
	return &bootstrapCertificateAuthoritiesWorkflow{
		input:        input,
		fabricConfig: w.fabricConfig,
	}, nil
}

type bootstrapCertificateAuthoritiesWorkflow struct {
	input        *pb.BootstrapCertificateAuthoritiesWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

func (w *bootstrapCertificateAuthoritiesWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapCertificateAuthoritiesOutput, error) {
	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	for _, certType := range []string{"tls", "org"} {
		if _, err := pb.DeployRootCertificateAuthorityChild(ctx, &pb.DeployRootCertificateAuthorityInput{
			CertType: certType,
		}); err != nil {
			return nil, fmt.Errorf("%w (certType=%s): %w", ErrDeployRootCertificateAuthority, certType, err)
		}

		if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{
			CertType: certType,
			Id:       "v1",
		}); err != nil {
			return nil, fmt.Errorf("%w (certType=%s): %w", ErrDeployCertificateAuthority, certType, err)
		}
	}

	return &pb.BootstrapCertificateAuthoritiesOutput{}, nil
}

// DeployCertificateStore implements the workflow
func (w *Workflows) DeployCertificateStore(ctx workflow.Context, input *pb.DeployCertificateStoreWorkflowInput) (pb.DeployCertificateStoreWorkflow, error) {
	return &deployCertificateStoreWorkflow{
		input:        input,
		fabricConfig: w.fabricConfig,
	}, nil
}

type deployCertificateStoreWorkflow struct {
	input        *pb.DeployCertificateStoreWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

func (w *deployCertificateStoreWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateStoreOutput, error) {
	logger := workflow.GetLogger(ctx)
	certStore := w.fabricConfig.CertificateStore()

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

	// Step 1: Create Vault access via Nexus call to infra worker
	nexusClient := secretsv1nexustemporal.NewSecretsNexusClient(secretsCreateAccessNexusEndpoint)
	_, err := nexusClient.CreateAccess(ctx, &secretspb.CreateAccessInput{
		PolicyName:          certStore.PolicyName(),
		PolicyDocument:      certStore.VaultPolicy(),
		AuthRoleName:        certStore.AuthRoleName(),
		ServiceAccountName:  certStore.ServiceAccountName(),
		KubernetesNamespace: certStore.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAccess, err)
	}

	// Step 2: Generate PostgreSQL credentials
	credentialsRef := secrets.NewVaultRef(certStore.VaultAdminCredentialsPath())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: credentialsRef,
		Username:  "admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate postgres credentials: %w", err)
	}

	// Step 3: Deploy PostgreSQL StatefulSet
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.fabricConfig.OrgID,
		Component: certStore.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate store deployment: %w", err)
	}

	// Step 4: Wait for certificate store to be ready
	logger.Info("Waiting for certificate store to be ready", "org", w.fabricConfig.OrgID)
	waitOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    120,
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

	logger.Info("Certificate store deployed and ready", "org", w.fabricConfig.OrgID)
	return &pb.DeployCertificateStoreOutput{}, nil
}

// CreateCertificateDatabase implements the workflow
func (w *Workflows) CreateCertificateDatabase(ctx workflow.Context, input *pb.CreateCertificateDatabaseWorkflowInput) (pb.CreateCertificateDatabaseWorkflow, error) {
	return &createCertificateDatabaseWorkflow{
		input:        input,
		fabricConfig: w.fabricConfig,
	}, nil
}

type createCertificateDatabaseWorkflow struct {
	input        *pb.CreateCertificateDatabaseWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

func (w *createCertificateDatabaseWorkflow) Execute(ctx workflow.Context) (*pb.CreateCertificateDatabaseOutput, error) {
	certType := w.input.Req.GetCertType()
	caID := w.input.Req.GetCertificateAuthorityId()

	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	if caID == "" {
		return nil, fmt.Errorf("certificate_authority_id cannot be empty")
	}

	caConfig := w.fabricConfig.CertificateAuthority(certType, caID)

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
	err := workflow.ExecuteActivity(activityCtx, postgrespb.CreateDatabaseActivityName, &postgrespb.CreateDatabaseInput{
		DatabaseName: caConfig.DatabaseName(),
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Step 2: Generate and store credentials in Vault
	credentialsRef := secrets.NewVaultRef(caConfig.VaultDatabaseCredentialsPath())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: credentialsRef,
		Username:  caConfig.DatabaseUsername(),
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate credentials: %w", err)
	}

	// Step 3: Create the PostgreSQL user (fetches credentials internally)
	err = workflow.ExecuteActivity(activityCtx, postgrespb.CreateUserActivityName, &postgrespb.CreateUserInput{
		CredentialsRef: credentialsRef,
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Step 4: Grant privileges
	err = workflow.ExecuteActivity(activityCtx, postgrespb.GrantPrivilegesActivityName, &postgrespb.GrantPrivilegesInput{
		Username:     caConfig.DatabaseUsername(),
		DatabaseName: caConfig.DatabaseName(),
		Privileges:   []postgrespb.Privilege{postgrespb.Privilege_PRIVILEGE_ALL},
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to grant privileges: %w", err)
	}

	return &pb.CreateCertificateDatabaseOutput{
		CredentialsRef: credentialsRef,
	}, nil
}

// DeployRootCertificateAuthority implements the workflow
func (w *Workflows) DeployRootCertificateAuthority(ctx workflow.Context, input *pb.DeployRootCertificateAuthorityWorkflowInput) (pb.DeployRootCertificateAuthorityWorkflow, error) {
	return &deployRootCertificateAuthorityWorkflow{
		input:        input,
		fabricConfig: w.fabricConfig,
	}, nil
}

type deployRootCertificateAuthorityWorkflow struct {
	input        *pb.DeployRootCertificateAuthorityWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

func (w *deployRootCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployRootCertificateAuthorityOutput, error) {
	certType := w.input.Req.GetCertType()

	if certType != "tls" && certType != "org" {
		return nil, ErrInvalidCertType
	}

	rootCA := w.fabricConfig.CertificateAuthority(certType, "root")

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

	// Step 1: Create Vault access via Nexus call to infra worker
	nexusClient := secretsv1nexustemporal.NewSecretsNexusClient(secretsCreateAccessNexusEndpoint)
	_, err := nexusClient.CreateAccess(ctx, &secretspb.CreateAccessInput{
		PolicyName:          rootCA.PolicyName(),
		PolicyDocument:      rootCA.VaultPolicy(),
		AuthRoleName:        rootCA.AuthRoleName(),
		ServiceAccountName:  rootCA.ServiceAccountName(),
		KubernetesNamespace: rootCA.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAccess, err)
	}

	// Step 2: Create certificate database
	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		CertType:               certType,
		CertificateAuthorityId: "root",
	}); err != nil {
		return nil, fmt.Errorf("%w (certType=%s): %w", ErrCreateCertificateDatabase, certType, err)
	}

	// Step 3: Generate root CA key pair
	keyPairRef := secrets.NewVaultRef(rootCA.VaultKeyPairPath())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: keyPairRef,
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA key pair: %w", err)
	}

	// Step 4: Generate self-signed root CA certificate
	certRef := secrets.NewVaultRef(rootCA.VaultCertificatePath())
	certValidityDuration := 15 * 365 * 24 * time.Hour
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
		KeyRef:             keyPairRef,
		CertRef:            certRef,
		CommonName:         rootCA.ComponentName(),
		Organization:       w.fabricConfig.OrgID,
		OrganizationalUnit: rootCA.OrganizationalUnit(),
		ValidityDuration:   durationpb.New(certValidityDuration),
		MaxPathLength:      1,
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA certificate: %w", err)
	}

	// Step 5: Generate bootstrap admin credentials
	adminCredentialsRef := secrets.NewVaultRef(rootCA.VaultAdminCredentialsPath())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: adminCredentialsRef,
		Username:  "admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 6: Deploy root CA pod
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.fabricConfig.OrgID,
		Component: rootCA.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy root CA: %w", err)
	}

	// TODO: Bootstrap root CA admin

	return &pb.DeployRootCertificateAuthorityOutput{}, nil
}

// DeployCertificateAuthority implements the workflow
func (w *Workflows) DeployCertificateAuthority(ctx workflow.Context, input *pb.DeployCertificateAuthorityWorkflowInput) (pb.DeployCertificateAuthorityWorkflow, error) {
	return &deployCertificateAuthorityWorkflow{
		input:        input,
		fabricConfig: w.fabricConfig,
	}, nil
}

type deployCertificateAuthorityWorkflow struct {
	input        *pb.DeployCertificateAuthorityWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

func (w *deployCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateAuthorityOutput, error) {
	certType := w.input.Req.GetCertType()
	caID := w.input.Req.GetId()

	if certType != "tls" && certType != "org" {
		return nil, ErrInvalidCertType
	}

	intermediateCA := w.fabricConfig.CertificateAuthority(certType, caID)

	// Step 1: Create Vault access via Nexus call to infra worker
	nexusClient := secretsv1nexustemporal.NewSecretsNexusClient(secretsCreateAccessNexusEndpoint)
	_, err := nexusClient.CreateAccess(ctx, &secretspb.CreateAccessInput{
		PolicyName:          intermediateCA.PolicyName(),
		PolicyDocument:      intermediateCA.VaultPolicy(),
		AuthRoleName:        intermediateCA.AuthRoleName(),
		ServiceAccountName:  intermediateCA.ServiceAccountName(),
		KubernetesNamespace: intermediateCA.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAccess, err)
	}

	// Step 2: Create certificate database
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

// BootstrapCertificateAuthorityAdmin implements the workflow
func (w *Workflows) BootstrapCertificateAuthorityAdmin(ctx workflow.Context, input *pb.BootstrapCertificateAuthorityAdminWorkflowInput) (pb.BootstrapCertificateAuthorityAdminWorkflow, error) {
	return &bootstrapCertificateAuthorityAdminWorkflow{
		input:        input,
		fabricConfig: w.fabricConfig,
	}, nil
}

type bootstrapCertificateAuthorityAdminWorkflow struct {
	input        *pb.BootstrapCertificateAuthorityAdminWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

func (w *bootstrapCertificateAuthorityAdminWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapCertificateAuthorityAdminOutput, error) {
	certType := w.input.Req.GetCertType()
	caID := w.input.Req.GetCertificateAuthorityId()

	if certType != "tls" && certType != "org" {
		return nil, ErrInvalidCertType
	}

	// TODO: Create admin identities
	// TODO: Revoke bootstrap admin credentials

	_ = caID

	return &pb.BootstrapCertificateAuthorityAdminOutput{}, nil
}

// CreateIdentity implements the workflow
func (w *Workflows) CreateIdentity(ctx workflow.Context, input *pb.CreateIdentityWorkflowInput) (pb.CreateIdentityWorkflow, error) {
	return &createIdentityWorkflow{
		input:        input,
		fabricConfig: w.fabricConfig,
	}, nil
}

type createIdentityWorkflow struct {
	input        *pb.CreateIdentityWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

func (w *createIdentityWorkflow) Execute(ctx workflow.Context) (*pb.CreateIdentityOutput, error) {
	certType := w.input.Req.GetCertType()
	identityName := w.input.Req.GetIdentityName()

	if certType != "tls" && certType != "org" {
		return nil, ErrInvalidCertType
	}

	// TODO: Generate key pair for identity using secrets.v1.CreateKeyPair
	// TODO: Register identity with CA
	// TODO: Enroll identity with CA
	// TODO: Store identity certificate

	_ = identityName

	return &pb.CreateIdentityOutput{}, nil
}
