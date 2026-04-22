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
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Nexus endpoint name for CreateAccess calls to infra worker
const secretsCreateAccessNexusEndpoint = "secrets-CreateAccess"

// Sentinel errors for fabric workflows
var (
	ErrBootstrapCertificateAuthorities = errors.New("bootstrap certificate authorities")
	ErrDeployCertificateStore          = errors.New("deploy certificate store")
	ErrCreateCertificateDatabase       = errors.New("create certificate database")
	ErrDeployCertificateAuthority      = errors.New("deploy certificate authority")
	ErrCreateCertificate               = errors.New("create certificate")
	ErrCreateBootstrapAdmin            = errors.New("create bootstrap admin")
	ErrCreateAdmin                     = errors.New("create admin")
	ErrCreateIdentity                  = errors.New("create identity")
	ErrCreateAccess                    = errors.New("create access")
	ErrInvalidCertType                 = errors.New("invalid cert type")
	ErrInvalidInput                    = errors.New("invalid input")
	ErrDeployCouchDB                   = errors.New("deploy couchdb")
	ErrDeployPeer                      = errors.New("deploy peer")
	ErrDeployOrderer                   = errors.New("deploy orderer")
	ErrDeployChainCode                 = errors.New("deploy chaincode")
	ErrCreateChannel                   = errors.New("create channel")
	ErrAddOrdererToChannel             = errors.New("add orderer to channel")
	ErrAddPeerToChannel                = errors.New("add peer to channel")
	ErrJoinOrdererToChannel            = errors.New("join orderer to channel")
	ErrJoinPeerToChannel               = errors.New("join peer to channel")
	ErrSetAnchorPeer                   = errors.New("set anchor peer")
	ErrAddOrdererToConsenterSet        = errors.New("add orderer to consenter set")
	ErrAddOrgToChannel                 = errors.New("add org to channel")
	ErrUpdateChannelConfig             = errors.New("update channel config")
)

// Workflows implements the FabricWorkflows interface
type Workflows struct {
	fabricConfig fabricconfig.FabricConfig
}

// NewWorkflows creates a new Workflows instance
func NewWorkflows(fabricConfig fabricconfig.FabricConfig) *Workflows {
	return &Workflows{fabricConfig: fabricConfig}
}

// =============================================================================
// BootstrapShard - Entry point for org bootstrap
// =============================================================================

func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return newBootstrapShardWorkflow(input, w.fabricConfig)
}

type bootstrapShardWorkflow struct {
	input        *pb.BootstrapShardWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newBootstrapShardWorkflow creates and validates a new bootstrapShardWorkflow
func newBootstrapShardWorkflow(input *pb.BootstrapShardWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*bootstrapShardWorkflow, error) {
	w := &bootstrapShardWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *bootstrapShardWorkflow) validate() error {
	// No required field validation - workflow has no required inputs
	return nil
}

func (w *bootstrapShardWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapShardOutput, error) {
	logger := workflow.GetLogger(ctx)

	// Step 1: Bootstrap all certificate authorities
	logger.Info("Bootstrapping certificate authorities")
	if _, err := pb.BootstrapCertificateAuthoritiesChild(ctx, &pb.BootstrapCertificateAuthoritiesInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBootstrapCertificateAuthorities, err)
	}

	// Step 2: Deploy peers (at least 3)
	// Each peer gets its own CouchDB instance for state storage
	peerIDs := []string{"peer0", "peer1", "peer2"}
	for _, peerID := range peerIDs {
		logger.Info("Deploying CouchDB for peer", "peerId", peerID)
		if _, err := pb.DeployCouchDbChild(ctx, &pb.DeployCouchDbInput{PeerId: peerID}); err != nil {
			return nil, fmt.Errorf("deploy couchdb for %s: %w", peerID, err)
		}

		logger.Info("Deploying peer", "peerId", peerID)
		if _, err := pb.DeployPeerChild(ctx, &pb.DeployPeerInput{PeerId: peerID}); err != nil {
			return nil, fmt.Errorf("deploy peer %s: %w", peerID, err)
		}
	}

	// Step 3: Deploy orderers (at least 1)
	ordererIDs := []string{"orderer0"}
	for _, ordererID := range ordererIDs {
		logger.Info("Deploying orderer", "ordererId", ordererID)
		if _, err := pb.DeployOrdererChild(ctx, &pb.DeployOrdererInput{OrdererId: ordererID}); err != nil {
			return nil, fmt.Errorf("deploy orderer %s: %w", ordererID, err)
		}
	}

	// Step 4: Deploy chaincode (at least 2)
	chainCodeIDs := []string{"asset-transfer", "fabcar"}
	for _, chainCodeID := range chainCodeIDs {
		logger.Info("Deploying chaincode", "chainCodeId", chainCodeID)
		if _, err := pb.DeployChainCodeChild(ctx, &pb.DeployChainCodeInput{ChainCodeId: chainCodeID}); err != nil {
			return nil, fmt.Errorf("deploy chaincode %s: %w", chainCodeID, err)
		}
	}

	// TODO: Hook into existing chains/channels

	logger.Info("Fabric shard bootstrap complete")
	return &pb.BootstrapShardOutput{}, nil
}

// =============================================================================
// BootstrapCertificateAuthorities - Orchestrates all CA deployments
// =============================================================================

func (w *Workflows) BootstrapCertificateAuthorities(ctx workflow.Context, input *pb.BootstrapCertificateAuthoritiesWorkflowInput) (pb.BootstrapCertificateAuthoritiesWorkflow, error) {
	return newBootstrapCertificateAuthoritiesWorkflow(input, w.fabricConfig)
}

type bootstrapCertificateAuthoritiesWorkflow struct {
	input        *pb.BootstrapCertificateAuthoritiesWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newBootstrapCertificateAuthoritiesWorkflow creates and validates a new bootstrapCertificateAuthoritiesWorkflow
func newBootstrapCertificateAuthoritiesWorkflow(input *pb.BootstrapCertificateAuthoritiesWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*bootstrapCertificateAuthoritiesWorkflow, error) {
	w := &bootstrapCertificateAuthoritiesWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *bootstrapCertificateAuthoritiesWorkflow) validate() error {
	// No required field validation - workflow has no required inputs
	return nil
}

func (w *bootstrapCertificateAuthoritiesWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapCertificateAuthoritiesOutput, error) {
	logger := workflow.GetLogger(ctx)

	// Step 1: Deploy certificate store (shared PostgreSQL)
	logger.Info("Deploying certificate store")
	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	// Step 2: Deploy TLS chain FIRST (Org chain depends on TLS Int CA for TLS certs)
	logger.Info("Deploying TLS chain")

	// TLS Root CA
	if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{
		CaRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
	}); err != nil {
		return nil, fmt.Errorf("deploy tls/root: %w", err)
	}

	// TLS Int CA
	if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{
		CaRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
	}); err != nil {
		return nil, fmt.Errorf("deploy tls/v1: %w", err)
	}

	// Step 3: Deploy Org chain (can now use TLS Int CA for TLS certs)
	logger.Info("Deploying Org chain")

	// Org Root CA
	if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{
		CaRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "root"},
	}); err != nil {
		return nil, fmt.Errorf("deploy org/root: %w", err)
	}

	// Org Int CA
	if _, err := pb.DeployCertificateAuthorityChild(ctx, &pb.DeployCertificateAuthorityInput{
		CaRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
	}); err != nil {
		return nil, fmt.Errorf("deploy org/v1: %w", err)
	}

	logger.Info("All certificate authorities deployed successfully")
	return &pb.BootstrapCertificateAuthoritiesOutput{}, nil
}

// =============================================================================
// DeployCertificateStore - Deploys shared PostgreSQL
// =============================================================================

func (w *Workflows) DeployCertificateStore(ctx workflow.Context, input *pb.DeployCertificateStoreWorkflowInput) (pb.DeployCertificateStoreWorkflow, error) {
	return newDeployCertificateStoreWorkflow(input, w.fabricConfig)
}

type deployCertificateStoreWorkflow struct {
	input        *pb.DeployCertificateStoreWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newDeployCertificateStoreWorkflow creates and validates a new deployCertificateStoreWorkflow
func newDeployCertificateStoreWorkflow(input *pb.DeployCertificateStoreWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*deployCertificateStoreWorkflow, error) {
	w := &deployCertificateStoreWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployCertificateStoreWorkflow) validate() error {
	// No required field validation - workflow has no required inputs
	return nil
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

// =============================================================================
// DeployCertificateAuthority - Unified workflow for root and intermediate CAs
// =============================================================================

func (w *Workflows) DeployCertificateAuthority(ctx workflow.Context, input *pb.DeployCertificateAuthorityWorkflowInput) (pb.DeployCertificateAuthorityWorkflow, error) {
	return newDeployCertificateAuthorityWorkflow(input, w.fabricConfig)
}

type deployCertificateAuthorityWorkflow struct {
	input        *pb.DeployCertificateAuthorityWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newDeployCertificateAuthorityWorkflow creates and validates a new deployCertificateAuthorityWorkflow
func newDeployCertificateAuthorityWorkflow(input *pb.DeployCertificateAuthorityWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*deployCertificateAuthorityWorkflow, error) {
	w := &deployCertificateAuthorityWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployCertificateAuthorityWorkflow) validate() error {
	caRef := w.input.Req.GetCaRef()
	if caRef == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	certType := caRef.GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	return nil
}

func (w *deployCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateAuthorityOutput, error) {
	logger := workflow.GetLogger(ctx)
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()
	caID := caRef.GetId()

	caConfig := w.fabricConfig.CertificateAuthority(certType, caID)
	logger.Info("Deploying certificate authority", "ca", caConfig.Description())

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

	// Step 1: Create Vault access
	nexusClient := secretsv1nexustemporal.NewSecretsNexusClient(secretsCreateAccessNexusEndpoint)
	_, err := nexusClient.CreateAccess(ctx, &secretspb.CreateAccessInput{
		PolicyName:          caConfig.PolicyName(),
		PolicyDocument:      caConfig.VaultPolicy(),
		AuthRoleName:        caConfig.AuthRoleName(),
		ServiceAccountName:  caConfig.ServiceAccountName(),
		KubernetesNamespace: caConfig.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAccess, err)
	}

	// Step 2: Create certificate database
	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	// Step 3: Create signing certificate
	signingIssuer := caConfig.SigningIssuer()
	signingPurpose := fmt.Sprintf("%s-%s-signing", certType, caID)

	if signingIssuer == nil {
		// Self-signed (root CA)
		logger.Info("Creating self-signed signing certificate", "ca", caConfig.Description())
		certValidityDuration := 15 * 365 * 24 * time.Hour

		if err != nil {
			return nil, fmt.Errorf("failed to generate signing key pair: %w", err)
		}

			CommonName:         caConfig.ComponentName(),
			Organization:       w.fabricConfig.OrgID,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			ValidityDuration:   durationpb.New(certValidityDuration),
			MaxPathLength:      1,
			DnsNames:           nil, // No SANs for CA signing certs
		}).Get(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed signing certificate: %w", err)
		}
	} else {
		// Issued by parent CA
		logger.Info("Creating CA-signed signing certificate", "ca", caConfig.Description(), "issuer", signingIssuer.Description())
		_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
			IssuerRef: &pb.CertificateAuthorityRef{
				CertType: signingIssuer.CertType(),
				Id:       signingIssuer.CAID(),
			},
			Profile: "ca",
			Subject: &pb.CertificateSubject{
				CommonName:         caConfig.ComponentName(),
				Organization:       w.fabricConfig.OrgID,
				OrganizationalUnit: caConfig.OrganizationalUnit(),
			},
			StorageRef:          caConfig.VaultSigningRef(),
			AdminCredentialsRef: signingIssuer.VaultAdminSigningRef(),
			Purpose:             signingPurpose,
			Attributes: map[string]string{
				"hf.IntermediateCA": "true",
			},
			// Store parent chain separately for x509 verification
			// This enables intermediate CA to verify client certs it issues
			ChainRef: caConfig.VaultSigningChainRef(),
		})
		if err != nil {
			return nil, fmt.Errorf("%w (signing): %w", ErrCreateCertificate, err)
		}
	}

	// Step 4: Create TLS certificate
	tlsIssuer := caConfig.TLSIssuer()
	tlsPurpose := fmt.Sprintf("%s-%s-tls", certType, caID)

	if tlsIssuer == nil {
		// Self-signed TLS (TLS Root CA only)
		logger.Info("Creating self-signed TLS certificate", "ca", caConfig.Description())
		certValidityDuration := 15 * 365 * 24 * time.Hour

		if err != nil {
			return nil, fmt.Errorf("failed to generate TLS key pair: %w", err)
		}

			CommonName:         caConfig.ComponentName(),
			Organization:       w.fabricConfig.OrgID,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			ValidityDuration:   durationpb.New(certValidityDuration),
			MaxPathLength:      0,                      // Not a CA cert
			DnsNames:           caConfig.TLSDNSNames(), // SANs required for TLS
		}).Get(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed TLS certificate: %w", err)
		}
	} else {
		// Issued by TLS CA
		// Use TLSCommonName() which differs from ComponentName() when both signing and TLS certs
		// come from the same CA (e.g., TLS Int CA) to avoid registration collision
		logger.Info("Creating CA-signed TLS certificate", "ca", caConfig.Description(), "issuer", tlsIssuer.Description())
		_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
			IssuerRef: &pb.CertificateAuthorityRef{
				CertType: tlsIssuer.CertType(),
				Id:       tlsIssuer.CAID(),
			},
			Profile: "tls",
			Subject: &pb.CertificateSubject{
				CommonName:         caConfig.TLSCommonName(), // Different name if same CA issues both
				Organization:       w.fabricConfig.OrgID,
				OrganizationalUnit: caConfig.OrganizationalUnit(),
				DnsNames:           caConfig.TLSDNSNames(),
			},
			StorageRef:          caConfig.VaultTLSRef(),
			AdminCredentialsRef: tlsIssuer.VaultAdminSigningRef(),
			Purpose:             tlsPurpose,
		})
		if err != nil {
			return nil, fmt.Errorf("%w (tls): %w", ErrCreateCertificate, err)
		}
	}

	// Step 5: Generate bootstrap admin credentials (username/password)
	logger.Info("Generating bootstrap admin credentials", "ca", caConfig.Description())
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 6: Deploy CA pod
	logger.Info("Deploying CA pod", "ca", caConfig.Description())
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.fabricConfig.OrgID,
		Component: caConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	// Step 7: Create bootstrap admin
	logger.Info("Creating bootstrap admin", "ca", caConfig.Description())
	_, err = pb.CreateBootstrapAdminChild(ctx, &pb.CreateBootstrapAdminInput{
		CaRef: caRef,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	// Step 8: Create permanent admin
	logger.Info("Creating permanent admin", "ca", caConfig.Description())
	_, err = pb.CreateAdminChild(ctx, &pb.CreateAdminInput{
		CaRef: caRef,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	// TODO Step 9: Revoke bootstrap admin

	logger.Info("Certificate authority deployment complete", "ca", caConfig.Description())
	return &pb.DeployCertificateAuthorityOutput{}, nil
}

// =============================================================================
// CreateCertificateDatabase - Creates PostgreSQL database for a CA
// =============================================================================

func (w *Workflows) CreateCertificateDatabase(ctx workflow.Context, input *pb.CreateCertificateDatabaseWorkflowInput) (pb.CreateCertificateDatabaseWorkflow, error) {
	return newCreateCertificateDatabaseWorkflow(input, w.fabricConfig)
}

type createCertificateDatabaseWorkflow struct {
	input        *pb.CreateCertificateDatabaseWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newCreateCertificateDatabaseWorkflow creates and validates a new createCertificateDatabaseWorkflow
func newCreateCertificateDatabaseWorkflow(input *pb.CreateCertificateDatabaseWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*createCertificateDatabaseWorkflow, error) {
	w := &createCertificateDatabaseWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createCertificateDatabaseWorkflow) validate() error {
	caRef := w.input.Req.GetCaRef()
	if caRef == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	certType := caRef.GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	return nil
}

func (w *createCertificateDatabaseWorkflow) Execute(ctx workflow.Context) (*pb.CreateCertificateDatabaseOutput, error) {
	logger := workflow.GetLogger(ctx)
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()
	caID := caRef.GetId()

	caConfig := w.fabricConfig.CertificateAuthority(certType, caID)
	certStore := w.fabricConfig.CertificateStore()

	// Wait for certificate store to be ready
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

	// Create the database
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Generate credentials
	credentialsRef := caConfig.VaultDatabaseUserRef()
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate credentials: %w", err)
	}

	// Create user
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Grant privileges
		Privileges:   []postgrespb.Privilege{postgrespb.Privilege_PRIVILEGE_ALL},
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to grant privileges: %w", err)
	}

	return &pb.CreateCertificateDatabaseOutput{
		CredentialsRef: credentialsRef,
	}, nil
}

// =============================================================================
// CreateCertificate - Atomic unit of certificate issuance
// =============================================================================

func (w *Workflows) CreateCertificate(ctx workflow.Context, input *pb.CreateCertificateWorkflowInput) (pb.CreateCertificateWorkflow, error) {
	return newCreateCertificateWorkflow(input, w.fabricConfig)
}

type createCertificateWorkflow struct {
	input        *pb.CreateCertificateWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newCreateCertificateWorkflow creates and validates a new createCertificateWorkflow
func newCreateCertificateWorkflow(input *pb.CreateCertificateWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*createCertificateWorkflow, error) {
	w := &createCertificateWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createCertificateWorkflow) validate() error {
	req := w.input.Req
	if req.GetIssuerRef() == nil {
		return fmt.Errorf("%w: issuer_ref is required (use secrets.CreateSelfSignedCertificate for self-signed)", ErrInvalidInput)
	}
	if req.GetStorageRef() == nil {
		return fmt.Errorf("%w: storage_ref is required", ErrInvalidInput)
	}
	certType := req.GetIssuerRef().GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: issuer cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	return nil
}

func (w *createCertificateWorkflow) Execute(ctx workflow.Context) (*pb.CreateCertificateOutput, error) {
	req := w.input.Req
	issuerRef := req.GetIssuerRef()
	certType := issuerRef.GetCertType()
	caID := issuerRef.GetId()

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

	// Step 1: Generate key pair
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Step 2: Register identity with CA
	subject := req.GetSubject()
	identityName := subject.GetCommonName()

	// Default to "client" if identity_type not specified
	// Valid values: "peer", "orderer", "admin", "client"
	// This determines the OU field in the certificate, critical for NodeOUs classification
	identityType := req.GetIdentityType()
	if identityType == "" {
		identityType = "client"
	}

	var registerOutput pb.RegisterIdentityOutput
		Affiliation:         "",
		Attributes:          req.GetAttributes(),
		IdentityType:        identityType,
		AdminCredentialsRef: req.GetAdminCredentialsRef(),
	}).Get(ctx, &registerOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to register identity: %w", err)
	}

	// Step 3: Enroll identity with CA
	var enrollOutput pb.EnrollIdentityOutput
		EnrollmentSecret: registerOutput.GetEnrollmentSecret(),
		KeyRef:           req.GetStorageRef(),
		Subject:          subject,
		CaTlsCertRef:     caConfig.VaultTLSRef(),
		Profile:          req.GetProfile(),
		ChainRef:         req.GetChainRef(), // Pass through for intermediate CA certs
	}).Get(ctx, &enrollOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity: %w", err)
	}

	return &pb.CreateCertificateOutput{
		CredentialsRef: req.GetStorageRef(),
	}, nil
}

// =============================================================================
// CreateBootstrapAdmin - Creates bootstrap admin after CA deployment
// =============================================================================

func (w *Workflows) CreateBootstrapAdmin(ctx workflow.Context, input *pb.CreateBootstrapAdminWorkflowInput) (pb.CreateBootstrapAdminWorkflow, error) {
	return newCreateBootstrapAdminWorkflow(input, w.fabricConfig)
}

type createBootstrapAdminWorkflow struct {
	input        *pb.CreateBootstrapAdminWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newCreateBootstrapAdminWorkflow creates and validates a new createBootstrapAdminWorkflow
func newCreateBootstrapAdminWorkflow(input *pb.CreateBootstrapAdminWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*createBootstrapAdminWorkflow, error) {
	w := &createBootstrapAdminWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createBootstrapAdminWorkflow) validate() error {
	if w.input.Req.GetCaRef() == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	return nil
}

func (w *createBootstrapAdminWorkflow) Execute(ctx workflow.Context) (*pb.CreateBootstrapAdminOutput, error) {
	logger := workflow.GetLogger(ctx)
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()
	caID := caRef.GetId()
	caConfig := w.fabricConfig.CertificateAuthority(certType, caID)

	// Wait for CA to be ready
	logger.Info("Waiting for CA to be ready", "ca", caConfig.Description())
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
		return nil, fmt.Errorf("CA did not become ready: %w", err)
	}

	logger.Info("CA ready, creating bootstrap admin", "ca", caConfig.Description())

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

	// Create keypair for signing identity
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key pair: %w", err)
	}

	// Bootstrap enroll signing identity from this CA
		KeyRef:                  caConfig.VaultBootstrapAdminSigningRef(),
		CertRef:                 caConfig.VaultBootstrapAdminSigningRef(),
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.BootstrapAdminCommonName(),
			Organization:       w.fabricConfig.OrgID,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		CaTlsCertRef: caConfig.VaultTLSRef(),
		Profile:      "", // Default profile for admin identity
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap enroll signing identity: %w", err)
	}

	// Create keypair for TLS client cert
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS key pair: %w", err)
	}

	// Bootstrap enroll TLS identity from the appropriate CA
	// Use BootstrapAdminTLSCommonName() for consistency with admin naming pattern
	tlsIssuer := caConfig.AdminTLSIssuer()
		KeyRef:                  caConfig.VaultBootstrapAdminTLSRef(),
		CertRef:                 caConfig.VaultBootstrapAdminTLSRef(),
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.BootstrapAdminTLSCommonName(),
			Organization:       w.fabricConfig.OrgID,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		CaTlsCertRef: tlsIssuer.VaultTLSRef(),
		Profile:      "tls", // TLS profile for client cert
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap enroll TLS identity: %w", err)
	}

	logger.Info("Bootstrap admin created", "ca", caConfig.Description())

	return &pb.CreateBootstrapAdminOutput{
		SigningCredentialsRef: caConfig.VaultBootstrapAdminSigningRef(),
		TlsCredentialsRef:     caConfig.VaultBootstrapAdminTLSRef(),
	}, nil
}

// =============================================================================
// CreateAdmin - Creates permanent admin using bootstrap admin
// =============================================================================

func (w *Workflows) CreateAdmin(ctx workflow.Context, input *pb.CreateAdminWorkflowInput) (pb.CreateAdminWorkflow, error) {
	return newCreateAdminWorkflow(input, w.fabricConfig)
}

type createAdminWorkflow struct {
	input        *pb.CreateAdminWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newCreateAdminWorkflow creates and validates a new createAdminWorkflow
func newCreateAdminWorkflow(input *pb.CreateAdminWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*createAdminWorkflow, error) {
	w := &createAdminWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createAdminWorkflow) validate() error {
	if w.input.Req.GetCaRef() == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	return nil
}

func (w *createAdminWorkflow) Execute(ctx workflow.Context) (*pb.CreateAdminOutput, error) {
	logger := workflow.GetLogger(ctx)
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()
	caID := caRef.GetId()
	caConfig := w.fabricConfig.CertificateAuthority(certType, caID)

	logger.Info("Creating permanent admin", "ca", caConfig.Description())

	// Create signing identity from the appropriate CA
	signingIssuer := caConfig.AdminSigningIssuer()
	signingPurpose := fmt.Sprintf("%s-%s-admin-signing", certType, caID)

	// Determine which credentials to use for registration with the signing issuer:
	// - If signing issuer is this CA itself → use this CA's bootstrap admin
	// - If signing issuer is a different CA → use that CA's permanent admin
	var signingAdminCreds *secretspb.SecretRef
	if signingIssuer.CertType() == caConfig.CertType() && signingIssuer.CAID() == caConfig.CAID() {
		signingAdminCreds = caConfig.VaultBootstrapAdminSigningRef()
	} else {
		signingAdminCreds = signingIssuer.VaultAdminSigningRef()
	}

	_, err := pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		IssuerRef: &pb.CertificateAuthorityRef{
			CertType: signingIssuer.CertType(),
			Id:       signingIssuer.CAID(),
		},
		Profile: "", // Default profile for admin identity
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.AdminCommonName(),
			Organization:       w.fabricConfig.OrgID,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		StorageRef:          caConfig.VaultAdminSigningRef(),
		AdminCredentialsRef: signingAdminCreds,
		Purpose:             signingPurpose,
		Attributes: map[string]string{
			"hf.Registrar.Roles":         "*",
			"hf.Registrar.DelegateRoles": "*",
			"hf.Registrar.Attributes":    "*",
			"hf.Revoker":                 "true",
			"hf.IntermediateCA":          "true",
			"hf.GenCRL":                  "true",
			"hf.AffiliationMgr":          "true",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create admin signing identity: %w", err)
	}

	// Create TLS client cert from the appropriate CA
	// Use AdminTLSCommonName() which differs from AdminCommonName() when both certs
	// come from the same CA (e.g., TLS Root CA Admin) to avoid registration collision
	tlsIssuer := caConfig.AdminTLSIssuer()
	tlsPurpose := fmt.Sprintf("%s-%s-admin-tls", certType, caID)

	// Determine which credentials to use for registration with the TLS issuer:
	// - If TLS issuer is this CA itself → use this CA's bootstrap admin
	// - If TLS issuer is a different CA → use that CA's permanent admin
	var tlsAdminCreds *secretspb.SecretRef
	if tlsIssuer.CertType() == caConfig.CertType() && tlsIssuer.CAID() == caConfig.CAID() {
		tlsAdminCreds = caConfig.VaultBootstrapAdminSigningRef()
	} else {
		tlsAdminCreds = tlsIssuer.VaultAdminSigningRef()
	}

	_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		IssuerRef: &pb.CertificateAuthorityRef{
			CertType: tlsIssuer.CertType(),
			Id:       tlsIssuer.CAID(),
		},
		Profile: "tls",
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.AdminTLSCommonName(), // Different name if same CA issues both
			Organization:       w.fabricConfig.OrgID,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		StorageRef:          caConfig.VaultAdminTLSRef(),
		AdminCredentialsRef: tlsAdminCreds,
		Purpose:             tlsPurpose,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create admin TLS identity: %w", err)
	}

	logger.Info("Permanent admin created", "ca", caConfig.Description())

	return &pb.CreateAdminOutput{
		SigningCredentialsRef: caConfig.VaultAdminSigningRef(),
		TlsCredentialsRef:     caConfig.VaultAdminTLSRef(),
	}, nil
}

// =============================================================================
// CreateIdentity - Creates full Fabric identity with both org and TLS certs
// =============================================================================

func (w *Workflows) CreateIdentity(ctx workflow.Context, input *pb.CreateIdentityWorkflowInput) (pb.CreateIdentityWorkflow, error) {
	return newCreateIdentityWorkflow(input, w.fabricConfig)
}

type createIdentityWorkflow struct {
	input        *pb.CreateIdentityWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newCreateIdentityWorkflow creates and validates a new createIdentityWorkflow
func newCreateIdentityWorkflow(input *pb.CreateIdentityWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*createIdentityWorkflow, error) {
	w := &createIdentityWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createIdentityWorkflow) validate() error {
	req := w.input.Req
	if req.GetIdentityName() == "" {
		return fmt.Errorf("%w: identity_name is required", ErrInvalidInput)
	}
	if req.GetOrgIssuerRef() == nil {
		return fmt.Errorf("%w: org_issuer_ref is required", ErrInvalidInput)
	}
	if req.GetTlsIssuerRef() == nil {
		return fmt.Errorf("%w: tls_issuer_ref is required", ErrInvalidInput)
	}
	if req.GetStorageSigningRef() == nil {
		return fmt.Errorf("%w: storage_signing_ref is required", ErrInvalidInput)
	}
	if req.GetStorageTlsRef() == nil {
		return fmt.Errorf("%w: storage_tls_ref is required", ErrInvalidInput)
	}
	return nil
}

func (w *createIdentityWorkflow) Execute(ctx workflow.Context) (*pb.CreateIdentityOutput, error) {
	req := w.input.Req
	identityName := req.GetIdentityName()
	orgIssuer := req.GetOrgIssuerRef()
	tlsIssuer := req.GetTlsIssuerRef()

	// Get issuer configs to access their admin credentials
	orgIssuerConfig := w.fabricConfig.CertificateAuthority(orgIssuer.GetCertType(), orgIssuer.GetId())
	tlsIssuerConfig := w.fabricConfig.CertificateAuthority(tlsIssuer.GetCertType(), tlsIssuer.GetId())

	// Create org identity cert (uses identity_type for proper OU classification)
	signingPurpose := fmt.Sprintf("identity-%s-signing", identityName)
	_, err := pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		IssuerRef:           orgIssuer,
		Profile:             "", // Default profile
		Subject:             req.GetSubject(),
		StorageRef:          req.GetStorageSigningRef(),
		AdminCredentialsRef: orgIssuerConfig.VaultAdminSigningRef(),
		Purpose:             signingPurpose,
		Attributes:          req.GetAttributes(),
		IdentityType:        req.GetIdentityType(), // Pass through for NodeOUs classification
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	// Create TLS client cert
	// TLS certs don't need identity_type since NodeOUs only applies to org/signing certs
	tlsPurpose := fmt.Sprintf("identity-%s-tls", identityName)
	_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		IssuerRef: tlsIssuer,
		Profile:   req.GetTlsProfile(),
		Subject: &pb.CertificateSubject{
			CommonName:         req.GetSubject().GetCommonName(),
			Organization:       req.GetSubject().GetOrganization(),
			OrganizationalUnit: req.GetSubject().GetOrganizationalUnit(),
			DnsNames:           req.GetSubject().GetDnsNames(),
		},
		StorageRef:          req.GetStorageTlsRef(),
		AdminCredentialsRef: tlsIssuerConfig.VaultAdminSigningRef(),
		Purpose:             tlsPurpose,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS identity: %w", err)
	}

	return &pb.CreateIdentityOutput{
		SigningCredentialsRef: req.GetStorageSigningRef(),
		TlsCredentialsRef:     req.GetStorageTlsRef(),
	}, nil
}

// =============================================================================
// DeployCouchDb - Deploys CouchDB instance for a peer
// =============================================================================

func (w *Workflows) DeployCouchDb(ctx workflow.Context, input *pb.DeployCouchDbWorkflowInput) (pb.DeployCouchDbWorkflow, error) {
	return newDeployCouchDbWorkflow(input, w.fabricConfig)
}

type deployCouchDbWorkflow struct {
	input        *pb.DeployCouchDbWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newDeployCouchDbWorkflow creates and validates a new deployCouchDbWorkflow
func newDeployCouchDbWorkflow(input *pb.DeployCouchDbWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*deployCouchDbWorkflow, error) {
	w := &deployCouchDbWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployCouchDbWorkflow) validate() error {
	// No required field validation - peer_id is used but not strictly required
	return nil
}

func (w *deployCouchDbWorkflow) Execute(ctx workflow.Context) (*pb.DeployCouchDbOutput, error) {
	logger := workflow.GetLogger(ctx)
	peerID := w.input.Req.GetPeerId()
	couchDbConfig := w.fabricConfig.CouchDb(peerID)

	logger.Info("DeployCouchDb workflow started", "peerId", peerID)

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
		PolicyName:          couchDbConfig.PolicyName(),
		PolicyDocument:      couchDbConfig.VaultPolicy(),
		AuthRoleName:        couchDbConfig.AuthRoleName(),
		ServiceAccountName:  couchDbConfig.ServiceAccountName(),
		KubernetesNamespace: couchDbConfig.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCouchDB, err)
	}

	// Step 2: Generate CouchDB admin credentials
	credentialsRef := couchDbConfig.VaultAdminRef()
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate couchdb credentials: %w", ErrDeployCouchDB, err)
	}

	// Step 3: Deploy CouchDB StatefulSet
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.fabricConfig.OrgID,
		Component: couchDbConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create couchdb deployment: %w", ErrDeployCouchDB, err)
	}

	// Step 4: Wait for CouchDB to be ready before returning
	// This ensures the peer can connect to CouchDB when it starts
	logger.Info("Waiting for CouchDB to be ready", "peerId", peerID)
	waitOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    5 * time.Second,
			MaximumAttempts:    0, // Infinite retry until timeout
		},
	}
	waitCtx := workflow.WithActivityOptions(ctx, waitOptions)

	err = workflow.ExecuteActivity(waitCtx, kubepb.GetStatusActivityName, &kubepb.GetStatusInput{
		Resources: map[string]*kubepb.ResourceIdentifier{
			couchDbConfig.ComponentName(): couchDbConfig.ResourceIdentifier(),
		},
	}).Get(waitCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: CouchDB did not become ready: %w", ErrDeployCouchDB, err)
	}

	logger.Info("CouchDB deployment ready", "peerId", peerID, "org", w.fabricConfig.OrgID)
	return &pb.DeployCouchDbOutput{}, nil
}

// =============================================================================
// DeployPeer - Deploys Fabric peer node
// =============================================================================

func (w *Workflows) DeployPeer(ctx workflow.Context, input *pb.DeployPeerWorkflowInput) (pb.DeployPeerWorkflow, error) {
	return newDeployPeerWorkflow(input, w.fabricConfig)
}

type deployPeerWorkflow struct {
	input        *pb.DeployPeerWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newDeployPeerWorkflow creates and validates a new deployPeerWorkflow
func newDeployPeerWorkflow(input *pb.DeployPeerWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*deployPeerWorkflow, error) {
	w := &deployPeerWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployPeerWorkflow) validate() error {
	// No required field validation - peer_id is used but not strictly required
	return nil
}

func (w *deployPeerWorkflow) Execute(ctx workflow.Context) (*pb.DeployPeerOutput, error) {
	logger := workflow.GetLogger(ctx)
	peerID := w.input.Req.GetPeerId()
	peerConfig := w.fabricConfig.Peer(peerID)

	logger.Info("DeployPeer workflow started", "peerId", peerID)

	// Step 1: Create Vault access via Nexus call to infra worker
	nexusClient := secretsv1nexustemporal.NewSecretsNexusClient(secretsCreateAccessNexusEndpoint)
	_, err := nexusClient.CreateAccess(ctx, &secretspb.CreateAccessInput{
		PolicyName:          peerConfig.PolicyName(),
		PolicyDocument:      peerConfig.VaultPolicy(),
		AuthRoleName:        peerConfig.AuthRoleName(),
		ServiceAccountName:  peerConfig.ServiceAccountName(),
		KubernetesNamespace: peerConfig.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create vault access: %w", ErrDeployPeer, err)
	}

	// Step 2: Create peer identity (org signing + TLS certs)
	// The CreateIdentity workflow handles both org and TLS certificates
	// IdentityType "peer" ensures the certificate gets OU=peer for NodeOUs classification
	logger.Info("Creating peer identity", "peerId", peerID)
	_, err = pb.CreateIdentityChild(ctx, &pb.CreateIdentityInput{
		IdentityName: peerConfig.IdentityName(),
		Subject: &pb.CertificateSubject{
			CommonName:         peerConfig.CommonName(),
			Organization:       w.fabricConfig.OrgID,
			OrganizationalUnit: peerConfig.OrganizationalUnit(),
			DnsNames:           peerConfig.TLSDNSNames(),
		},
		OrgIssuerRef: &pb.CertificateAuthorityRef{
			CertType: peerConfig.OrgIssuerCertType(),
			Id:       peerConfig.OrgIssuerID(),
		},
		TlsIssuerRef: &pb.CertificateAuthorityRef{
			CertType: peerConfig.TLSIssuerCertType(),
			Id:       peerConfig.TLSIssuerID(),
		},
		StorageSigningRef: peerConfig.VaultSigningRef(),
		StorageTlsRef:     peerConfig.VaultTLSRef(),
		TlsProfile:        "tls",
		IdentityType:      "peer", // Critical: sets OU=peer in certificate for NodeOUs
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create peer identity: %w", ErrDeployPeer, err)
	}

	// Step 3: Deploy peer StatefulSet
	logger.Info("Deploying peer", "peerId", peerID)
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.fabricConfig.OrgID,
		Component: peerConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create peer deployment: %w", ErrDeployPeer, err)
	}

	logger.Info("Peer deployment initiated", "peerId", peerID, "org", w.fabricConfig.OrgID)
	return &pb.DeployPeerOutput{}, nil
}

// =============================================================================
// DeployOrderer - Deploys Fabric orderer node
// =============================================================================

func (w *Workflows) DeployOrderer(ctx workflow.Context, input *pb.DeployOrdererWorkflowInput) (pb.DeployOrdererWorkflow, error) {
	return newDeployOrdererWorkflow(input, w.fabricConfig)
}

type deployOrdererWorkflow struct {
	input        *pb.DeployOrdererWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newDeployOrdererWorkflow creates and validates a new deployOrdererWorkflow
func newDeployOrdererWorkflow(input *pb.DeployOrdererWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*deployOrdererWorkflow, error) {
	w := &deployOrdererWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployOrdererWorkflow) validate() error {
	// No required field validation - orderer_id is used but not strictly required
	return nil
}

func (w *deployOrdererWorkflow) Execute(ctx workflow.Context) (*pb.DeployOrdererOutput, error) {
	logger := workflow.GetLogger(ctx)
	ordererID := w.input.Req.GetOrdererId()
	ordererConfig := w.fabricConfig.Orderer(ordererID)

	logger.Info("DeployOrderer workflow started", "ordererId", ordererID)

	// Step 1: Create Vault access via Nexus call to infra worker
	nexusClient := secretsv1nexustemporal.NewSecretsNexusClient(secretsCreateAccessNexusEndpoint)
	_, err := nexusClient.CreateAccess(ctx, &secretspb.CreateAccessInput{
		PolicyName:          ordererConfig.PolicyName(),
		PolicyDocument:      ordererConfig.VaultPolicy(),
		AuthRoleName:        ordererConfig.AuthRoleName(),
		ServiceAccountName:  ordererConfig.ServiceAccountName(),
		KubernetesNamespace: ordererConfig.KubernetesNamespace(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create vault access: %w", ErrDeployOrderer, err)
	}

	// Step 2: Create orderer identity (org signing + TLS certs)
	// The CreateIdentity workflow handles both org and TLS certificates
	// IdentityType "orderer" ensures the certificate gets OU=orderer for NodeOUs classification
	logger.Info("Creating orderer identity", "ordererId", ordererID)
	_, err = pb.CreateIdentityChild(ctx, &pb.CreateIdentityInput{
		IdentityName: ordererConfig.IdentityName(),
		Subject: &pb.CertificateSubject{
			CommonName:         ordererConfig.CommonName(),
			Organization:       w.fabricConfig.OrgID,
			OrganizationalUnit: ordererConfig.OrganizationalUnit(),
			DnsNames:           ordererConfig.TLSDNSNames(),
		},
		OrgIssuerRef: &pb.CertificateAuthorityRef{
			CertType: ordererConfig.OrgIssuerCertType(),
			Id:       ordererConfig.OrgIssuerID(),
		},
		TlsIssuerRef: &pb.CertificateAuthorityRef{
			CertType: ordererConfig.TLSIssuerCertType(),
			Id:       ordererConfig.TLSIssuerID(),
		},
		StorageSigningRef: ordererConfig.VaultSigningRef(),
		StorageTlsRef:     ordererConfig.VaultTLSRef(),
		TlsProfile:        "tls",
		IdentityType:      "orderer", // Critical: sets OU=orderer in certificate for NodeOUs
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create orderer identity: %w", ErrDeployOrderer, err)
	}

	// Step 3: Deploy orderer StatefulSet
	logger.Info("Deploying orderer", "ordererId", ordererID)
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       w.fabricConfig.OrgID,
		Component: ordererConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create orderer deployment: %w", ErrDeployOrderer, err)
	}

	logger.Info("Orderer deployment initiated", "ordererId", ordererID, "org", w.fabricConfig.OrgID)
	return &pb.DeployOrdererOutput{}, nil
}

// =============================================================================
// DeployChainCode - Deploys chaincode (smart contract)
// =============================================================================

func (w *Workflows) DeployChainCode(ctx workflow.Context, input *pb.DeployChainCodeWorkflowInput) (pb.DeployChainCodeWorkflow, error) {
	return newDeployChainCodeWorkflow(input, w.fabricConfig)
}

type deployChainCodeWorkflow struct {
	input        *pb.DeployChainCodeWorkflowInput
	fabricConfig fabricconfig.FabricConfig
}

// newDeployChainCodeWorkflow creates and validates a new deployChainCodeWorkflow
func newDeployChainCodeWorkflow(input *pb.DeployChainCodeWorkflowInput, fabricConfig fabricconfig.FabricConfig) (*deployChainCodeWorkflow, error) {
	w := &deployChainCodeWorkflow{input: input, fabricConfig: fabricConfig}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployChainCodeWorkflow) validate() error {
	// No required field validation - chain_code_id is used but not strictly required
	return nil
}

func (w *deployChainCodeWorkflow) Execute(ctx workflow.Context) (*pb.DeployChainCodeOutput, error) {
	logger := workflow.GetLogger(ctx)
	chainCodeID := w.input.Req.GetChainCodeId()
	logger.Info("DeployChainCode workflow started", "chainCodeId", chainCodeID)

	// TODO: Implement chaincode deployment
	// Steps according to Fabric Bootstrap Workflows.md:
	// 1. Nexus Call: secrets.v1.CreateAccess
	// 2. for only ("org"):
	//    - Workflow: fabric.v1.CreateIdentity (with IdentityType: "client")
	// 3. Workflow: cluster.fabric.v1.CreateDeployment
	// NOTE: IdentityType should be "client" for chaincode identities

	return &pb.DeployChainCodeOutput{}, nil
}

// =============================================================================
// CreateChannel - Orchestrates creation of a new Fabric channel
// =============================================================================

func (w *Workflows) CreateChannel(ctx workflow.Context, input *pb.CreateChannelWorkflowInput) (pb.CreateChannelWorkflow, error) {
	return &createChannelWorkflow{input: input}, nil
}

type createChannelWorkflow struct {
	input *pb.CreateChannelWorkflowInput
}

func (w *createChannelWorkflow) Execute(ctx workflow.Context) (*pb.CreateChannelOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)




	// Join orderers sequentially
	if _, err := pb.JoinOrdererToChannelChild(ctx, &pb.JoinOrdererToChannelInput{}); err != nil {
		return nil, fmt.Errorf("%w: join orderer: %w", ErrCreateChannel, err)
	}

	// Join peers (parallel after orderers ready)
	if _, err := pb.JoinPeerToChannelChild(ctx, &pb.JoinPeerToChannelInput{}); err != nil {
		return nil, fmt.Errorf("%w: join peer: %w", ErrCreateChannel, err)
	}

	// Set anchor peers per org
	if _, err := pb.SetAnchorPeerChild(ctx, &pb.SetAnchorPeerInput{}); err != nil {
		return nil, fmt.Errorf("%w: set anchor peer: %w", ErrCreateChannel, err)
	}

	return &pb.CreateChannelOutput{}, nil
}

// =============================================================================
// AddOrdererToChannel - Adds a new orderer to an existing channel
// =============================================================================

func (w *Workflows) AddOrdererToChannel(ctx workflow.Context, input *pb.AddOrdererToChannelWorkflowInput) (pb.AddOrdererToChannelWorkflow, error) {
	return &addOrdererToChannelWorkflow{input: input}, nil
}

type addOrdererToChannelWorkflow struct {
	input *pb.AddOrdererToChannelWorkflowInput
}

func (w *addOrdererToChannelWorkflow) Execute(ctx workflow.Context) (*pb.AddOrdererToChannelOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Fetch latest config block (NOT genesis) for follower sync

	// Join as follower; polls until synced
	if _, err := pb.JoinOrdererToChannelChild(ctx, &pb.JoinOrdererToChannelInput{}); err != nil {
		return nil, fmt.Errorf("%w: join orderer: %w", ErrAddOrdererToChannel, err)
	}

	// Promote follower to consenter via config update
	if _, err := pb.AddOrdererToConsenterSetChild(ctx, &pb.AddOrdererToConsenterSetInput{}); err != nil {
		return nil, fmt.Errorf("%w: add to consenter set: %w", ErrAddOrdererToChannel, err)
	}

	return &pb.AddOrdererToChannelOutput{}, nil
}

// =============================================================================
// AddPeerToChannel - Adds a new peer to an existing channel
// =============================================================================

func (w *Workflows) AddPeerToChannel(ctx workflow.Context, input *pb.AddPeerToChannelWorkflowInput) (pb.AddPeerToChannelWorkflow, error) {
	return &addPeerToChannelWorkflow{input: input}, nil
}

type addPeerToChannelWorkflow struct {
	input *pb.AddPeerToChannelWorkflowInput
}

func (w *addPeerToChannelWorkflow) Execute(ctx workflow.Context) (*pb.AddPeerToChannelOutput, error) {
	// TODO: conditionally add org to channel if not already a member
	if _, err := pb.AddOrgToChannelChild(ctx, &pb.AddOrgToChannelInput{}); err != nil {
		return nil, fmt.Errorf("%w: add org: %w", ErrAddPeerToChannel, err)
	}

	if _, err := pb.JoinPeerToChannelChild(ctx, &pb.JoinPeerToChannelInput{}); err != nil {
		return nil, fmt.Errorf("%w: join peer: %w", ErrAddPeerToChannel, err)
	}

	return &pb.AddPeerToChannelOutput{}, nil
}

// =============================================================================
// JoinOrdererToChannel - Joins an orderer node to a channel via osnadmin
// =============================================================================

func (w *Workflows) JoinOrdererToChannel(ctx workflow.Context, input *pb.JoinOrdererToChannelWorkflowInput) (pb.JoinOrdererToChannelWorkflow, error) {
	return &joinOrdererToChannelWorkflow{input: input}, nil
}

type joinOrdererToChannelWorkflow struct {
	input *pb.JoinOrdererToChannelWorkflowInput
}

func (w *joinOrdererToChannelWorkflow) Execute(ctx workflow.Context) (*pb.JoinOrdererToChannelOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)



	return &pb.JoinOrdererToChannelOutput{}, nil
}

// =============================================================================
// JoinPeerToChannel - Joins a peer node to a channel
// =============================================================================

func (w *Workflows) JoinPeerToChannel(ctx workflow.Context, input *pb.JoinPeerToChannelWorkflowInput) (pb.JoinPeerToChannelWorkflow, error) {
	return &joinPeerToChannelWorkflow{input: input}, nil
}

type joinPeerToChannelWorkflow struct {
	input *pb.JoinPeerToChannelWorkflowInput
}

func (w *joinPeerToChannelWorkflow) Execute(ctx workflow.Context) (*pb.JoinPeerToChannelOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)



	return &pb.JoinPeerToChannelOutput{}, nil
}

// =============================================================================
// SetAnchorPeer - Configures the anchor peer for an org on a channel
// =============================================================================

func (w *Workflows) SetAnchorPeer(ctx workflow.Context, input *pb.SetAnchorPeerWorkflowInput) (pb.SetAnchorPeerWorkflow, error) {
	return &setAnchorPeerWorkflow{input: input}, nil
}

type setAnchorPeerWorkflow struct {
	input *pb.SetAnchorPeerWorkflowInput
}

func (w *setAnchorPeerWorkflow) Execute(ctx workflow.Context) (*pb.SetAnchorPeerOutput, error) {
	if _, err := pb.UpdateChannelConfigChild(ctx, &pb.UpdateChannelConfigInput{}); err != nil {
		return nil, fmt.Errorf("%w: update config: %w", ErrSetAnchorPeer, err)
	}

	return &pb.SetAnchorPeerOutput{}, nil
}

// =============================================================================
// AddOrdererToConsenterSet - Promotes an orderer from follower to consenter
// =============================================================================

func (w *Workflows) AddOrdererToConsenterSet(ctx workflow.Context, input *pb.AddOrdererToConsenterSetWorkflowInput) (pb.AddOrdererToConsenterSetWorkflow, error) {
	return &addOrdererToConsenterSetWorkflow{input: input}, nil
}

type addOrdererToConsenterSetWorkflow struct {
	input *pb.AddOrdererToConsenterSetWorkflowInput
}

func (w *addOrdererToConsenterSetWorkflow) Execute(ctx workflow.Context) (*pb.AddOrdererToConsenterSetOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	if _, err := pb.UpdateChannelConfigChild(ctx, &pb.UpdateChannelConfigInput{}); err != nil {
		return nil, fmt.Errorf("%w: update config: %w", ErrAddOrdererToConsenterSet, err)
	}


	return &pb.AddOrdererToConsenterSetOutput{}, nil
}

// =============================================================================
// AddOrgToChannel - Adds a new organization to an existing channel
// =============================================================================

func (w *Workflows) AddOrgToChannel(ctx workflow.Context, input *pb.AddOrgToChannelWorkflowInput) (pb.AddOrgToChannelWorkflow, error) {
	return &addOrgToChannelWorkflow{input: input}, nil
}

type addOrgToChannelWorkflow struct {
	input *pb.AddOrgToChannelWorkflowInput
}

func (w *addOrgToChannelWorkflow) Execute(ctx workflow.Context) (*pb.AddOrgToChannelOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)


	if _, err := pb.UpdateChannelConfigChild(ctx, &pb.UpdateChannelConfigInput{}); err != nil {
		return nil, fmt.Errorf("%w: update config: %w", ErrAddOrgToChannel, err)
	}


	return &pb.AddOrgToChannelOutput{}, nil
}

// =============================================================================
// UpdateChannelConfig - Common pattern for all channel configuration changes
// =============================================================================

func (w *Workflows) UpdateChannelConfig(ctx workflow.Context, input *pb.UpdateChannelConfigWorkflowInput) (pb.UpdateChannelConfigWorkflow, error) {
	return &updateChannelConfigWorkflow{input: input}, nil
}

type updateChannelConfigWorkflow struct {
	input *pb.UpdateChannelConfigWorkflowInput
}

func (w *updateChannelConfigWorkflow) Execute(ctx workflow.Context) (*pb.UpdateChannelConfigOutput, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)





	return &pb.UpdateChannelConfigOutput{}, nil
}
