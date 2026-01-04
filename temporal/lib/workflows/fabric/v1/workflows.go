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
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
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
	credentialsRef := certStore.VaultAdminRef()
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

	logger.Info("Certificate store deployment initiated", "org", w.fabricConfig.OrgID)
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
	logger := workflow.GetLogger(ctx)
	certType := w.input.Req.GetCertType()
	caID := w.input.Req.GetCertificateAuthorityId()

	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	if caID == "" {
		return nil, fmt.Errorf("certificate_authority_id cannot be empty")
	}

	caConfig := w.fabricConfig.CertificateAuthority(certType, caID)
	certStore := w.fabricConfig.CertificateStore()

	// Step 0: Wait for certificate store to be ready
	logger.Info("Waiting for certificate store to be ready", "org", w.fabricConfig.OrgID)
	waitOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    0, // Infinite retry
		},
	}
	waitCtx := workflow.WithActivityOptions(ctx, waitOptions)

	err := workflow.ExecuteActivity(waitCtx, kubepb.GetStatusActivityName, &kubepb.GetStatusInput{
		Resources: map[string]*kubepb.ResourceIdentifier{
			"certificate-store": certStore.ResourceIdentifier(),
		},
	}).Get(waitCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("certificate store did not become ready: %w", err)
	}

	logger.Info("Certificate store ready", "org", w.fabricConfig.OrgID)

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
	err = workflow.ExecuteActivity(activityCtx, postgrespb.CreateDatabaseActivityName, &postgrespb.CreateDatabaseInput{
		DatabaseName: caConfig.DatabaseName(),
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Step 2: Generate and store credentials in Vault
	credentialsRef := caConfig.VaultDatabaseUserRef()
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
	logger := workflow.GetLogger(ctx)
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

	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		CertType:               certType,
		CertificateAuthorityId: rootCA.CAID(),
	}); err != nil {
		return nil, fmt.Errorf("%w (certType=%s): %w", ErrCreateCertificateDatabase, certType, err)
	}

	serverRef := rootCA.VaultServerRef()
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: serverRef,
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA key pair: %w", err)
	}

	certValidityDuration := 15 * 365 * 24 * time.Hour

	// Construct DNS SANs for TLS certificate validation
	// Pattern: {component-name}.{namespace}.svc.cluster.local
	serviceHostname := fmt.Sprintf("%s.%s.svc.cluster.local", rootCA.ComponentName(), rootCA.KubernetesNamespace())

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
		KeyRef:             serverRef,
		CertRef:            serverRef,
		CommonName:         rootCA.ComponentName(),
		Organization:       w.fabricConfig.OrgID,
		OrganizationalUnit: rootCA.OrganizationalUnit(),
		ValidityDuration:   durationpb.New(certValidityDuration),
		MaxPathLength:      1,
		DnsNames:           []string{serviceHostname}, // Required for TLS verification
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA certificate: %w", err)
	}

	// Generate bootstrap admin credentials (username/password) for initial CA enrollment
	bootstrapAdminRef := rootCA.VaultBootstrapAdminRef()
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: bootstrapAdminRef,
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.fabricConfig.OrgID,
		Component: rootCA.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy root CA: %w", err)
	}

	logger.Info("Certificate authority deployment initiated", "org", w.fabricConfig.OrgID, "certType", certType)

	_, err = pb.BootstrapCertificateAuthorityAdminChild(ctx, &pb.BootstrapCertificateAuthorityAdminInput{
		CertType:                certType,
		CertificateAuthorityId:  rootCA.CAID(),
		BootstrapCredentialsRef: bootstrapAdminRef,
	})
	if err != nil {
		return nil, fmt.Errorf("%w (certType=%s): %w", ErrBootstrapCertificateAuthorityAdmin, certType, err)
	}

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
	logger := workflow.GetLogger(ctx)
	certType := w.input.Req.GetCertType()
	caID := w.input.Req.GetId()

	if certType != "tls" && certType != "org" {
		return nil, ErrInvalidCertType
	}

	intermediateCA := w.fabricConfig.CertificateAuthority(certType, caID)
	rootCA := w.fabricConfig.CertificateAuthority(certType, "root")

	// Step 1: Create Vault access via Nexus call to infra worker
	logger.Info("Creating Vault access for intermediate CA", "certType", certType, "caID", caID)
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
	logger.Info("Creating certificate database for intermediate CA", "certType", certType, "caID", caID)
	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		CertType:               certType,
		CertificateAuthorityId: caID,
	}); err != nil {
		return nil, fmt.Errorf("%w (certType=%s, caID=%s): %w", ErrCreateCertificateDatabase, certType, caID, err)
	}

	// Step 3: TODO - Scale down root CA (cluster.fabric.v1.ScaleDeployment not yet implemented)
	// This is optional for production but recommended to avoid concurrent database access during intermediate CA cert creation
	// For now, we proceed without scaling down the root CA

	// Step 4: Create intermediate CA identity (certificate signed by root CA)
	// Use CreateIdentity workflow with storage_ref to:
	// - Register/enroll with the root CA (parent via certificate_authority_id="root")
	// - Store the certificate at the intermediate CA's VaultServerRef path (via storage_ref)
	logger.Info("Creating intermediate CA certificate from root CA", "certType", certType, "caID", caID)

	identityName := fmt.Sprintf("ca-%s", caID) // e.g., "ca-v1"
	serverRef := intermediateCA.VaultServerRef()

	_, err = pb.CreateIdentityChild(ctx, &pb.CreateIdentityInput{
		CertType:               certType,
		CertificateAuthorityId: "root", // Register/enroll with root CA
		IdentityName:           identityName,
		CommonName:             intermediateCA.ComponentName(),
		Organization:           w.fabricConfig.OrgID,
		OrganizationalUnit:     intermediateCA.OrganizationalUnit(),
		Affiliation:            "",
		Attributes: map[string]string{
			"hf.IntermediateCA": "true", // Mark as intermediate CA
		},
		AdminCredentialsRef: rootCA.VaultAdminRef(), // Use root's permanent admin
		StorageRef:          serverRef,              // Store at intermediate CA's server path
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create intermediate CA certificate: %w", err)
	}

	logger.Info("Intermediate CA certificate created and signed by root CA", "certType", certType, "caID", caID)

	// Step 5: TODO - Scale root CA back up (cluster.fabric.v1.ScaleDeployment not yet implemented)

	// Step 6: Generate bootstrap admin credentials (username/password) for initial CA enrollment
	logger.Info("Generating bootstrap admin credentials for intermediate CA", "certType", certType, "caID", caID)

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

	bootstrapAdminRef := intermediateCA.VaultBootstrapAdminRef()
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: bootstrapAdminRef,
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 7: Deploy intermediate CA pod
	logger.Info("Deploying intermediate CA", "certType", certType, "caID", caID)
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.fabricConfig.OrgID,
		Component: intermediateCA.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy intermediate CA: %w", err)
	}

	logger.Info("Intermediate CA deployment initiated", "org", w.fabricConfig.OrgID, "certType", certType, "caID", caID)

	// Step 8: Bootstrap intermediate CA admin
	logger.Info("Bootstrapping intermediate CA admin", "certType", certType, "caID", caID)
	_, err = pb.BootstrapCertificateAuthorityAdminChild(ctx, &pb.BootstrapCertificateAuthorityAdminInput{
		CertType:                certType,
		CertificateAuthorityId:  caID,
		BootstrapCredentialsRef: bootstrapAdminRef,
	})
	if err != nil {
		return nil, fmt.Errorf("%w (certType=%s, caID=%s): %w", ErrBootstrapCertificateAuthorityAdmin, certType, caID, err)
	}

	logger.Info("Intermediate CA deployment complete", "org", w.fabricConfig.OrgID, "certType", certType, "caID", caID)

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
	logger := workflow.GetLogger(ctx)
	certType := w.input.Req.GetCertType()
	caID := w.input.Req.GetCertificateAuthorityId()

	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}

	caConfig := w.fabricConfig.CertificateAuthority(certType, caID)

	// Step 0: Wait for certificate authority to be ready
	logger.Info("Waiting for certificate authority to be ready", "org", w.fabricConfig.OrgID, "certType", certType, "caID", caID)
	waitOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    0, // Infinite retry
		},
	}
	waitCtx := workflow.WithActivityOptions(ctx, waitOptions)

	err := workflow.ExecuteActivity(waitCtx, kubepb.GetStatusActivityName, &kubepb.GetStatusInput{
		Resources: map[string]*kubepb.ResourceIdentifier{
			caConfig.ComponentName(): caConfig.ResourceIdentifier(),
		},
	}).Get(waitCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("certificate authority did not become ready: %w", err)
	}

	logger.Info("Certificate authority ready", "org", w.fabricConfig.OrgID, "certType", certType, "caID", caID)

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

	// Step 1: Create a keypair for the bootstrap admin identity to use for initial enrollment
	bootstrapAdminRef := caConfig.VaultBootstrapAdminRef()
	var createKeyPairOutput secretspb.CreateKeyPairOutput
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: bootstrapAdminRef,
	}).Get(ctx, &createKeyPairOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin key pair: %w", err)
	}

	// Step 2: Bootstrap enroll using the bootstrap credentials
	// This obtains the bootstrap admin's certificate and stores it with the keypair
	caServerRef := caConfig.VaultServerRef()
	var bootstrapEnrollOutput pb.BootstrapEnrollIdentityOutput
	err = workflow.ExecuteActivity(activityCtx, pb.BootstrapEnrollIdentityActivityName, &pb.BootstrapEnrollIdentityInput{
		CaUrl:                   caConfig.URL(),
		BootstrapCredentialsRef: w.input.Req.GetBootstrapCredentialsRef(),
		KeyRef:                  bootstrapAdminRef,
		CertRef:                 bootstrapAdminRef, // Store cert with keypair at unified path
		CommonName:              caConfig.AdminCommonName(),
		Organization:            w.fabricConfig.OrgID,
		OrganizationalUnit:      caConfig.OrganizationalUnit(),
		CaTlsCertRef:            caServerRef,
	}).Get(ctx, &bootstrapEnrollOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap enroll admin identity: %w", err)
	}

	// Step 3: Create permanent admin identity using the bootstrap admin credentials
	// This creates a new admin identity with the full registration/enrollment flow
	// IMPORTANT: We pass the bootstrap admin credentials so RegisterIdentity can authenticate
	// Note: Bootstrap admin is "bootstrap-admin", permanent admin is "admin"
	// The admin gets full registrar permissions including the ability to register intermediate CAs
	_, err = pb.CreateIdentityChild(ctx, &pb.CreateIdentityInput{
		CertType:               certType,
		CertificateAuthorityId: caID,
		IdentityName:           "admin",
		CommonName:             caConfig.AdminCommonName(),
		Organization:           w.fabricConfig.OrgID,
		OrganizationalUnit:     caConfig.OrganizationalUnit(),
		Affiliation:            "",
		Attributes: map[string]string{
			// Full registrar permissions - matches bootstrap admin
			"hf.Registrar.Roles":         "*",    // Can register any identity type
			"hf.Registrar.DelegateRoles": "*",    // Can delegate any role
			"hf.Registrar.Attributes":    "*",    // Can assign any attribute (required for hf.IntermediateCA)
			"hf.Revoker":                 "true", // Can revoke certificates
			"hf.IntermediateCA":          "true", // Can create intermediate CA certificates
			"hf.GenCRL":                  "true", // Can generate CRLs
			"hf.AffiliationMgr":          "true", // Can manage affiliations
		},
		AdminCredentialsRef: bootstrapAdminRef, // Use bootstrap admin to create permanent admin
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateIdentity, err)
	}

	// TODO Step 4: Revoke the bootstrap admin credentials using RevokeIdentity activity
	// This ensures the bootstrap credentials can no longer be used after permanent admin is created

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
	caID := w.input.Req.GetCertificateAuthorityId()
	identityName := w.input.Req.GetIdentityName()

	if certType != "tls" && certType != "org" {
		return nil, fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
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

	// Step 1: Generate key pair for identity
	// Use explicit storage_ref if provided, otherwise auto-generate path
	var keyRef *secretspb.SecretRef
	if w.input.Req.GetStorageRef() != nil {
		keyRef = w.input.Req.GetStorageRef()
	} else {
		keyRef = secrets.NewVaultRef(fmt.Sprintf("%s/%s", caConfig.VaultBasePath(), identityName))
	}

	err := workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: keyRef,
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Step 2: Register identity with CA
	// Pass through admin_credentials_ref if provided (e.g., bootstrap admin during CA setup)
	var registerOutput pb.RegisterIdentityOutput
	err = workflow.ExecuteActivity(activityCtx, pb.RegisterIdentityActivityName, &pb.RegisterIdentityInput{
		CertType:               certType,
		CertificateAuthorityId: caID,
		IdentityName:           identityName,
		Affiliation:            w.input.Req.GetAffiliation(),
		Attributes:             w.input.Req.GetAttributes(),
		IdentityType:           "client", // Default type
		AdminCredentialsRef:    w.input.Req.GetAdminCredentialsRef(),
	}).Get(ctx, &registerOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to register identity: %w", err)
	}

	// Step 3: Enroll identity with CA
	// This updates the keyRef secret to include the certificate
	caServerRef := caConfig.VaultServerRef()
	var enrollOutput pb.EnrollIdentityOutput
	err = workflow.ExecuteActivity(activityCtx, pb.EnrollIdentityActivityName, &pb.EnrollIdentityInput{
		CertType:               certType,
		CertificateAuthorityId: caID,
		IdentityName:           identityName,
		EnrollmentSecret:       registerOutput.GetEnrollmentSecret(),
		KeyRef:                 keyRef,
		CommonName:             w.input.Req.GetCommonName(),
		Organization:           w.input.Req.GetOrganization(),
		OrganizationalUnit:     w.input.Req.GetOrganizationalUnit(),
		CaTlsCertRef:           caServerRef,
	}).Get(ctx, &enrollOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity: %w", err)
	}

	// Return a single credentials ref containing key + cert
	return &pb.CreateIdentityOutput{
		CredentialsRef: keyRef,
	}, nil
}
