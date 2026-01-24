package fabric

import (
	"errors"
	"fmt"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	fabricv1nexustemporal "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1/fabricv1nexustemporal"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	postgrespb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/secrets/v1/secretsv1nexustemporal"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Nexus endpoint name for CreateAccess calls to infra worker
const secretsCreateAccessNexusEndpoint = "secrets-CreateAccess"

// computeTLSCommonName returns the CN for a TLS certificate, adding "-tls" suffix
// when both signing and TLS certs come from the same CA to avoid registration collision.
func computeTLSCommonName(baseName string, signingIssuerRef, tlsIssuerRef *pb.CertificateAuthorityRef) string {
	// Self-signed cases: no collision (not registering with a CA)
	if signingIssuerRef == nil || tlsIssuerRef == nil {
		return baseName
	}
	// Same CA issues both - need different identity name
	if signingIssuerRef.GetCertType() == tlsIssuerRef.GetCertType() &&
		signingIssuerRef.GetId() == tlsIssuerRef.GetId() {
		return baseName + "-tls"
	}
	return baseName
}

// Sentinel errors for fabric workflows
var (
	ErrDeployCertificateStore     = errors.New("deploy certificate store")
	ErrCreateCertificateDatabase  = errors.New("create certificate database")
	ErrDeployCertificateAuthority = errors.New("deploy certificate authority")
	ErrCreateCertificate          = errors.New("create certificate")
	ErrCreateBootstrapAdmin       = errors.New("create bootstrap admin")
	ErrCreateAdmin                = errors.New("create admin")
	ErrCreateIdentity             = errors.New("create identity")
	ErrCreateAccess               = errors.New("create access")
	ErrInvalidCertType            = errors.New("invalid cert type")
	ErrInvalidInput               = errors.New("invalid input")
	ErrDeployCouchDB              = errors.New("deploy couchdb")
	ErrDeployPeer                 = errors.New("deploy peer")
	ErrDeployOrderer              = errors.New("deploy orderer")
	ErrDeployChainCode            = errors.New("deploy chaincode")
)

// Workflows implements the FabricWorkflows interface
// Workflows no longer hold global orgConfig - they get org from each workflow's input
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

// =============================================================================
// BootstrapShard - Entry point for org bootstrap
// =============================================================================

func (w *Workflows) BootstrapShard(ctx workflow.Context, input *pb.BootstrapShardWorkflowInput) (pb.BootstrapShardWorkflow, error) {
	return newBootstrapShardWorkflow(input)
}

type bootstrapShardWorkflow struct {
	input *pb.BootstrapShardWorkflowInput
}

// newBootstrapShardWorkflow creates and validates a new bootstrapShardWorkflow
func newBootstrapShardWorkflow(input *pb.BootstrapShardWorkflowInput) (*bootstrapShardWorkflow, error) {
	w := &bootstrapShardWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *bootstrapShardWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	deployment := w.input.Req.GetDeployment()
	if deployment == nil {
		return fmt.Errorf("%w: deployment is required", ErrInvalidInput)
	}
	if len(deployment.GetCertificateAuthorities()) == 0 {
		return fmt.Errorf("%w: at least one certificate authority is required", ErrInvalidInput)
	}
	if len(deployment.GetPeers()) == 0 {
		return fmt.Errorf("%w: at least one peer is required", ErrInvalidInput)
	}
	if len(deployment.GetOrderers()) == 0 {
		return fmt.Errorf("%w: at least one orderer is required", ErrInvalidInput)
	}
	return nil
}

func (w *bootstrapShardWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapShardOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	deployment := w.input.Req.GetDeployment()

	// Step 1: Deploy certificate store and CAs
	logger.Info("Deploying certificate infrastructure")
	if _, err := pb.BootstrapCertificateAuthoritiesChild(ctx, &pb.BootstrapCertificateAuthoritiesInput{
		Org:                    org,
		CertificateAuthorities: deployment.GetCertificateAuthorities(),
	}); err != nil {
		return nil, fmt.Errorf("bootstrap certificate authorities: %w", err)
	}

	// Step 3: Deploy peers (using deployment spec)
	for _, peerSpec := range deployment.GetPeers() {
		logger.Info("Deploying CouchDB for peer", "peerId", peerSpec.GetId())
		if _, err := pb.DeployCouchDbChild(ctx, &pb.DeployCouchDbInput{
			Org:    org,
			PeerId: peerSpec.GetId(),
		}); err != nil {
			return nil, fmt.Errorf("deploy couchdb for %s: %w", peerSpec.GetId(), err)
		}

		logger.Info("Deploying peer", "peerId", peerSpec.GetId())
		if _, err := pb.DeployPeerChild(ctx, &pb.DeployPeerInput{
			Org:  org,
			Spec: peerSpec,
		}); err != nil {
			return nil, fmt.Errorf("deploy peer %s: %w", peerSpec.GetId(), err)
		}
	}

	// Step 4: Deploy orderers (using deployment spec)
	for _, ordererSpec := range deployment.GetOrderers() {
		logger.Info("Deploying orderer", "ordererId", ordererSpec.GetId())
		if _, err := pb.DeployOrdererChild(ctx, &pb.DeployOrdererInput{
			Org:  org,
			Spec: ordererSpec,
		}); err != nil {
			return nil, fmt.Errorf("deploy orderer %s: %w", ordererSpec.GetId(), err)
		}
	}

	// TODO: Deploy chaincode, hook into channels

	logger.Info("Fabric shard bootstrap complete")
	return &pb.BootstrapShardOutput{}, nil
}

// =============================================================================
// BootstrapCertificateAuthorities - Deploys certificate store and TLS CA chain
// =============================================================================

func (w *Workflows) BootstrapCertificateAuthorities(ctx workflow.Context, input *pb.BootstrapCertificateAuthoritiesWorkflowInput) (pb.BootstrapCertificateAuthoritiesWorkflow, error) {
	return newBootstrapCertificateAuthoritiesWorkflow(input)
}

type bootstrapCertificateAuthoritiesWorkflow struct {
	input *pb.BootstrapCertificateAuthoritiesWorkflowInput
}

// newBootstrapCertificateAuthoritiesWorkflow creates and validates a new bootstrapCertificateAuthoritiesWorkflow
func newBootstrapCertificateAuthoritiesWorkflow(input *pb.BootstrapCertificateAuthoritiesWorkflowInput) (*bootstrapCertificateAuthoritiesWorkflow, error) {
	w := &bootstrapCertificateAuthoritiesWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *bootstrapCertificateAuthoritiesWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	cas := w.input.Req.GetCertificateAuthorities()
	if len(cas) == 0 {
		return fmt.Errorf("%w: at least one certificate authority is required", ErrInvalidInput)
	}
	for i, caSpec := range cas {
		if caSpec.GetSelf() == nil {
			return fmt.Errorf("%w: certificate_authorities[%d].self is required", ErrInvalidInput, i)
		}
		// BootstrapCertificateAuthorities is for operator-level TLS infrastructure only
		if caSpec.GetSelf().GetCertType() != "tls" {
			return fmt.Errorf("%w: certificate_authorities[%d] must have cert_type='tls', got '%s'",
				ErrInvalidInput, i, caSpec.GetSelf().GetCertType())
		}
	}
	return nil
}

func (w *bootstrapCertificateAuthoritiesWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapCertificateAuthoritiesOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()

	// Step 1: Deploy certificate store (shared PostgreSQL)
	logger.Info("Deploying certificate store")
	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{
		Org: org,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	// Step 2: Deploy CAs in order using specialized workflows
	// Order is critical: TLS Root → TLS Int → Org Root → Org Int
	logger.Info("Deploying certificate authorities")

	// Build map of CA specs by certType/id for easy lookup
	caSpecs := make(map[string]*pb.CertificateAuthoritySpec)
	for _, spec := range w.input.Req.GetCertificateAuthorities() {
		key := fmt.Sprintf("%s/%s", spec.GetSelf().GetCertType(), spec.GetSelf().GetId())
		caSpecs[key] = spec
	}

	// Deploy TLS Root CA
	if tlsRootSpec, ok := caSpecs["tls/root"]; ok {
		logger.Info("Deploying TLS Root CA")
		if _, err := pb.DeployTlsRootCertificateAuthorityChild(ctx, &pb.DeployTlsRootCertificateAuthorityInput{
			Org:              org,
			ValidityDuration: tlsRootSpec.GetValidityDuration(),
		}); err != nil {
			return nil, fmt.Errorf("deploy TLS Root CA: %w", err)
		}
	}

	// Deploy TLS Intermediate CA
	if tlsIntSpec, ok := caSpecs["tls/int"]; ok {
		logger.Info("Deploying TLS Intermediate CA")
		if _, err := pb.DeployTlsIntermediateCertificateAuthorityChild(ctx, &pb.DeployTlsIntermediateCertificateAuthorityInput{
			Org:              org,
			ValidityDuration: tlsIntSpec.GetValidityDuration(),
		}); err != nil {
			return nil, fmt.Errorf("deploy TLS Intermediate CA: %w", err)
		}
	}

	// Deploy Org Root CA
	if orgRootSpec, ok := caSpecs["org/root"]; ok {
		logger.Info("Deploying Org Root CA")
		if _, err := pb.DeployOrgRootCertificateAuthorityChild(ctx, &pb.DeployOrgRootCertificateAuthorityInput{
			Org:              org,
			ValidityDuration: orgRootSpec.GetValidityDuration(),
			TlsIssuer:        &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
		}); err != nil {
			return nil, fmt.Errorf("deploy Org Root CA: %w", err)
		}
	}

	// Deploy Org Intermediate CA
	if orgIntSpec, ok := caSpecs["org/int"]; ok {
		logger.Info("Deploying Org Intermediate CA")
		if _, err := pb.DeployOrgIntermediateCertificateAuthorityChild(ctx, &pb.DeployOrgIntermediateCertificateAuthorityInput{
			Org:              org,
			ValidityDuration: orgIntSpec.GetValidityDuration(),
			TlsIssuer:        &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
		}); err != nil {
			return nil, fmt.Errorf("deploy Org Intermediate CA: %w", err)
		}
	}

	logger.Info("All certificate authorities deployed successfully")

	return &pb.BootstrapCertificateAuthoritiesOutput{}, nil
}

// =============================================================================
// DeployCertificateStore - Deploys shared PostgreSQL
// =============================================================================

func (w *Workflows) DeployCertificateStore(ctx workflow.Context, input *pb.DeployCertificateStoreWorkflowInput) (pb.DeployCertificateStoreWorkflow, error) {
	return newDeployCertificateStoreWorkflow(input)
}

type deployCertificateStoreWorkflow struct {
	input *pb.DeployCertificateStoreWorkflowInput
}

// newDeployCertificateStoreWorkflow creates and validates a new deployCertificateStoreWorkflow
func newDeployCertificateStoreWorkflow(input *pb.DeployCertificateStoreWorkflowInput) (*deployCertificateStoreWorkflow, error) {
	w := &deployCertificateStoreWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployCertificateStoreWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployCertificateStoreWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateStoreOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	certStore := fabricconfig.NewCertificateStoreConfig(orgConfig)

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
		Org:       org,
		Component: certStore.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate store deployment: %w", err)
	}

	logger.Info("Certificate store deployment initiated", "org", org)
	return &pb.DeployCertificateStoreOutput{}, nil
}

// =============================================================================
// DeployCertificateAuthority - Unified workflow for root and intermediate CAs
// =============================================================================

func (w *Workflows) DeployCertificateAuthority(ctx workflow.Context, input *pb.DeployCertificateAuthorityWorkflowInput) (pb.DeployCertificateAuthorityWorkflow, error) {
	return newDeployCertificateAuthorityWorkflow(input)
}

type deployCertificateAuthorityWorkflow struct {
	input *pb.DeployCertificateAuthorityWorkflowInput
}

// newDeployCertificateAuthorityWorkflow creates and validates a new deployCertificateAuthorityWorkflow
func newDeployCertificateAuthorityWorkflow(input *pb.DeployCertificateAuthorityWorkflowInput) (*deployCertificateAuthorityWorkflow, error) {
	w := &deployCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployCertificateAuthorityWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	spec := w.input.Req.GetSpec()
	if spec == nil {
		return fmt.Errorf("%w: spec is required", ErrInvalidInput)
	}
	caRef := spec.GetSelf()
	if caRef == nil {
		return fmt.Errorf("%w: spec.self is required", ErrInvalidInput)
	}
	certType := caRef.GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	return nil
}

func (w *deployCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateAuthorityOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	spec := w.input.Req.GetSpec()
	caRef := spec.GetSelf()
	certType := caRef.GetCertType()
	caID := caRef.GetId()

	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	if certType == "tls" {
		ownerConfig = operatorConfig
	} else {
		ownerConfig = orgConfig
	}

	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, certType, caID)
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
		Org:   org,
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	// Step 3: Create signing certificate
	// Use explicit issuer refs from spec - nil means self-signed (root CAs)
	signingIssuerRef := spec.GetSigningIssuer()
	signingPurpose := fmt.Sprintf("%s-%s-signing", certType, caID)
	validityDuration := spec.GetValidityDuration().AsDuration()

	if signingIssuerRef == nil || signingIssuerRef.GetCertType() == "" {
		// Self-signed (root CA)
		logger.Info("Creating self-signed signing certificate", "ca", caConfig.Description())

		err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
			SecretRef: caConfig.VaultSigningKeyRef(),
		}).Get(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate signing key pair: %w", err)
		}

		err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
			KeyRef:             caConfig.VaultSigningKeyRef(),
			CertRef:            caConfig.VaultSigningCertRef(),
			CommonName:         caConfig.ComponentName(),
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			ValidityDuration:   durationpb.New(validityDuration),
			MaxPathLength:      1,
			DnsNames:           nil, // No SANs for CA signing certs
		}).Get(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed signing certificate: %w", err)
		}
	} else {
		// Issued by parent CA - get config for the issuer to access its admin credentials
		// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
		var signingIssuerOwner config.OwnerConfig
		if signingIssuerRef.GetCertType() == "tls" {
			signingIssuerOwner = operatorConfig
		} else {
			signingIssuerOwner = orgConfig
		}
		signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(signingIssuerOwner, signingIssuerRef.GetCertType(), signingIssuerRef.GetId())
		logger.Info("Creating CA-signed signing certificate", "ca", caConfig.Description(), "issuer", signingIssuerConfig.Description())
		_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
			Org:       org,
			IssuerRef: signingIssuerRef,
			Profile:   "ca",
			Subject: &pb.CertificateSubject{
				CommonName:         caConfig.ComponentName(),
				Organization:       org,
				OrganizationalUnit: caConfig.OrganizationalUnit(),
			},
			Storage: &pb.IdentityRef{
				KeyRef:  caConfig.VaultSigningKeyRef(),
				CertRef: caConfig.VaultSigningCertRef(),
			},
			AdminIdentity: &pb.IdentityRef{
				KeyRef:  signingIssuerConfig.VaultAdminSigningKeyRef(),
				CertRef: signingIssuerConfig.VaultAdminSigningCertRef(),
			},
			Purpose: signingPurpose,
			Attributes: map[string]string{
				"hf.IntermediateCA": "true",
			},
			IssuerCertRef: signingIssuerConfig.VaultSigningCertRef(),
		})
		if err != nil {
			return nil, fmt.Errorf("%w (signing): %w", ErrCreateCertificate, err)
		}
	}

	// Step 4: Create TLS certificate
	// Use explicit issuer refs from spec - nil means self-signed (TLS Root CA only)
	tlsIssuerRef := spec.GetTlsIssuer()
	tlsPurpose := fmt.Sprintf("%s-%s-tls", certType, caID)

	if tlsIssuerRef == nil || tlsIssuerRef.GetCertType() == "" {
		// Self-signed TLS (TLS Root CA only)
		logger.Info("Creating self-signed TLS certificate", "ca", caConfig.Description())

		err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
			SecretRef: caConfig.VaultTLSKeyRef(),
		}).Get(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate TLS key pair: %w", err)
		}

		err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
			KeyRef:             caConfig.VaultTLSKeyRef(),
			CertRef:            caConfig.VaultTLSCertRef(),
			CommonName:         caConfig.ComponentName(),
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			ValidityDuration:   durationpb.New(validityDuration),
			MaxPathLength:      0,                      // Not a CA cert
			DnsNames:           caConfig.TLSDNSNames(), // SANs required for TLS
		}).Get(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed TLS certificate: %w", err)
		}
	} else {
		// Issued by TLS CA - use Nexus for cross-namespace certificate issuance
		tlsCommonName := computeTLSCommonName(caConfig.ComponentName(), signingIssuerRef, tlsIssuerRef)

		if tlsIssuerRef.GetCertType() == "tls" {
			// TLS CA is operator-owned - use Nexus to issue from operator worker
			logger.Info("Creating CA TLS certificate via Nexus", "ca", caConfig.Description(), "issuer", tlsIssuerRef)

			tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, tlsIssuerRef.GetCertType(), tlsIssuerRef.GetId())
			nexusClient := fabricv1nexustemporal.NewFabricNexusClient("fabric-IssueTlsCertificate")
			_, err = nexusClient.IssueTlsCertificate(ctx, &pb.IssueTlsCertificateInput{
				Org:       org,
				IssuerRef: tlsIssuerRef,
				Profile:   "tls",
				Subject: &pb.CertificateSubject{
					CommonName:         tlsCommonName,
					Organization:       org,
					OrganizationalUnit: caConfig.OrganizationalUnit(),
					DnsNames:           caConfig.TLSDNSNames(),
				},
				Storage: &pb.IdentityRef{
					KeyRef:  caConfig.VaultTLSKeyRef(),
					CertRef: caConfig.VaultTLSCertRef(),
				},
				Purpose:       tlsPurpose,
				IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
			}, workflow.NexusOperationOptions{
				ScheduleToCloseTimeout: 5 * time.Minute,
			})
		} else {
			// Org CA issuer - use direct call with org config
			tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(orgConfig, tlsIssuerRef.GetCertType(), tlsIssuerRef.GetId())
			logger.Info("Creating CA-signed TLS certificate", "ca", caConfig.Description(), "issuer", tlsIssuerConfig.Description())
			_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
				Org:       org,
				IssuerRef: tlsIssuerRef,
				Profile:   "tls",
				Subject: &pb.CertificateSubject{
					CommonName:         tlsCommonName,
					Organization:       org,
					OrganizationalUnit: caConfig.OrganizationalUnit(),
					DnsNames:           caConfig.TLSDNSNames(),
				},
				Storage: &pb.IdentityRef{
					KeyRef:  caConfig.VaultTLSKeyRef(),
					CertRef: caConfig.VaultTLSCertRef(),
				},
				AdminIdentity: &pb.IdentityRef{
					KeyRef:  tlsIssuerConfig.VaultAdminSigningKeyRef(),
					CertRef: tlsIssuerConfig.VaultAdminSigningCertRef(),
				},
				Purpose:       tlsPurpose,
				IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
			})
		}
		if err != nil {
			return nil, fmt.Errorf("%w (tls): %w", ErrCreateCertificate, err)
		}
	}

	// Step 5: Generate bootstrap admin credentials (username/password)
	logger.Info("Generating bootstrap admin credentials", "ca", caConfig.Description())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 6: Deploy CA pod
	logger.Info("Deploying CA pod", "ca", caConfig.Description())
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: caConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	// Step 7: Create bootstrap admin
	// Admin TLS issuer: TLS CAs use themselves, Org CAs use input's TLS issuer (TLS Int)
	// See docs/fabric/certs.csv rows 6-13 for admin certificate issuers
	var adminTLSIssuer *pb.CertificateAuthorityRef
	if certType == "tls" {
		adminTLSIssuer = caRef // TLS CA admins get TLS from their own CA (rows 7, 9)
	} else {
		adminTLSIssuer = tlsIssuerRef // Org CA admins get TLS from TLS Int (rows 12, 13)
	}

	logger.Info("Creating bootstrap admin", "ca", caConfig.Description())
	_, err = pb.CreateBootstrapAdminChild(ctx, &pb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        adminTLSIssuer,
		ValidityDuration: spec.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	// Step 8: Create permanent admin
	// Signing issuer is always the CA itself (per Fabric CA constraints)
	logger.Info("Creating permanent admin", "ca", caConfig.Description())
	_, err = pb.CreateAdminChild(ctx, &pb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef,          // Admin signing from the CA itself
		TlsIssuer:        adminTLSIssuer, // Same TLS issuer as bootstrap admin
		ValidityDuration: spec.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	// TODO Step 9: Revoke bootstrap admin

	logger.Info("Certificate authority deployment complete", "ca", caConfig.Description())
	return &pb.DeployCertificateAuthorityOutput{}, nil
}

// =============================================================================
// DeployTlsRootCertificateAuthority - TLS Root CA (operator-owned, self-signed)
// =============================================================================

func (w *Workflows) DeployTlsRootCertificateAuthority(ctx workflow.Context, input *pb.DeployTlsRootCertificateAuthorityWorkflowInput) (pb.DeployTlsRootCertificateAuthorityWorkflow, error) {
	return newDeployTlsRootCertificateAuthorityWorkflow(input)
}

type deployTlsRootCertificateAuthorityWorkflow struct {
	input *pb.DeployTlsRootCertificateAuthorityWorkflowInput
}

func newDeployTlsRootCertificateAuthorityWorkflow(input *pb.DeployTlsRootCertificateAuthorityWorkflowInput) (*deployTlsRootCertificateAuthorityWorkflow, error) {
	w := &deployTlsRootCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *deployTlsRootCertificateAuthorityWorkflow) validate() error {
	if w.input == nil || w.input.Req == nil {
		return fmt.Errorf("%w: input must not be nil", ErrInvalidInput)
	}
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org must not be empty", ErrInvalidInput)
	}
	return nil
}

func (w *deployTlsRootCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployTlsRootCertificateAuthorityOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	operatorConfig := config.GetOperator()
	validityDuration := w.input.Req.GetValidityDuration().AsDuration()

	// TLS Root CA is always operator-owned, certType="tls", id="root"
	caConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, "tls", "root")
	caRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	logger.Info("Deploying TLS Root Certificate Authority", "ca", caConfig.Description())

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
		Org:   org,
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	// Step 3: Create self-signed signing certificate
	logger.Info("Creating self-signed signing certificate", "ca", caConfig.Description())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: caConfig.VaultSigningKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key pair: %w", err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
		KeyRef:             caConfig.VaultSigningKeyRef(),
		CertRef:            caConfig.VaultSigningCertRef(),
		CommonName:         caConfig.ComponentName(),
		Organization:       org,
		OrganizationalUnit: caConfig.OrganizationalUnit(),
		ValidityDuration:   durationpb.New(validityDuration),
		MaxPathLength:      1,
		DnsNames:           nil, // No SANs for CA signing certs
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed signing certificate: %w", err)
	}

	// Step 4: Create self-signed TLS certificate
	logger.Info("Creating self-signed TLS certificate", "ca", caConfig.Description())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: caConfig.VaultTLSKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS key pair: %w", err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
		KeyRef:             caConfig.VaultTLSKeyRef(),
		CertRef:            caConfig.VaultTLSCertRef(),
		CommonName:         caConfig.ComponentName(),
		Organization:       org,
		OrganizationalUnit: caConfig.OrganizationalUnit(),
		ValidityDuration:   durationpb.New(validityDuration),
		MaxPathLength:      0,                      // Not a CA cert
		DnsNames:           caConfig.TLSDNSNames(), // SANs required for TLS
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed TLS certificate: %w", err)
	}

	// Step 5: Generate bootstrap admin credentials
	logger.Info("Generating bootstrap admin credentials", "ca", caConfig.Description())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 6: Deploy CA pod
	logger.Info("Deploying CA pod", "ca", caConfig.Description())
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: caConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	// Step 7: Create bootstrap admin (TLS from self)
	logger.Info("Creating bootstrap admin", "ca", caConfig.Description())
	_, err = pb.CreateBootstrapAdminChild(ctx, &pb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        caRef, // Admin TLS from self (TLS Root)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	// Step 8: Create permanent admin (TLS from self)
	logger.Info("Creating permanent admin", "ca", caConfig.Description())
	_, err = pb.CreateAdminChild(ctx, &pb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef, // Admin signing from self
		TlsIssuer:        caRef, // Admin TLS from self
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	logger.Info("TLS Root Certificate Authority deployment complete", "ca", caConfig.Description())
	return &pb.DeployTlsRootCertificateAuthorityOutput{}, nil
}

// =============================================================================
// DeployTlsIntermediateCertificateAuthority - TLS Int CA (operator-owned, issued by TLS Root)
// =============================================================================

func (w *Workflows) DeployTlsIntermediateCertificateAuthority(ctx workflow.Context, input *pb.DeployTlsIntermediateCertificateAuthorityWorkflowInput) (pb.DeployTlsIntermediateCertificateAuthorityWorkflow, error) {
	return newDeployTlsIntermediateCertificateAuthorityWorkflow(input)
}

type deployTlsIntermediateCertificateAuthorityWorkflow struct {
	input *pb.DeployTlsIntermediateCertificateAuthorityWorkflowInput
}

func newDeployTlsIntermediateCertificateAuthorityWorkflow(input *pb.DeployTlsIntermediateCertificateAuthorityWorkflowInput) (*deployTlsIntermediateCertificateAuthorityWorkflow, error) {
	w := &deployTlsIntermediateCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *deployTlsIntermediateCertificateAuthorityWorkflow) validate() error {
	if w.input == nil || w.input.Req == nil {
		return fmt.Errorf("%w: input must not be nil", ErrInvalidInput)
	}
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org must not be empty", ErrInvalidInput)
	}
	return nil
}

func (w *deployTlsIntermediateCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployTlsIntermediateCertificateAuthorityOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	operatorConfig := config.GetOperator()

	// TLS Int CA is always operator-owned, certType="tls", id="int"
	caConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, "tls", "int")
	caRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"}
	
	// Parent is always TLS Root (operator-owned, direct call)
	signingIssuerRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	tlsIssuerRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, "tls", "root")
	
	logger.Info("Deploying TLS Intermediate Certificate Authority", "ca", caConfig.Description())

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
		Org:   org,
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	// Step 3: Create signing certificate issued by TLS Root (direct workflow call, not Nexus)
	logger.Info("Creating CA-signed signing certificate via direct call", "ca", caConfig.Description(), "issuer", signingIssuerConfig.Description())
	signingPurpose := fmt.Sprintf("tls-int-signing")
	_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		Org:       org,
		IssuerRef: signingIssuerRef,
		Profile:   "ca",
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.ComponentName(),
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultSigningKeyRef(),
			CertRef: caConfig.VaultSigningCertRef(),
		},
		AdminIdentity: &pb.IdentityRef{
			KeyRef:  signingIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: signingIssuerConfig.VaultAdminSigningCertRef(),
		},
		Purpose: signingPurpose,
		Attributes: map[string]string{
			"hf.IntermediateCA": "true",
		},
		IssuerCertRef: signingIssuerConfig.VaultSigningCertRef(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w (signing): %w", ErrCreateCertificate, err)
	}

	// Step 4: Create TLS certificate issued by TLS Root (direct workflow call, not Nexus)
	// CRITICAL FIX: This was using Nexus before, causing the bug
	logger.Info("Creating CA-signed TLS certificate via direct call", "ca", caConfig.Description(), "issuer", signingIssuerConfig.Description())
	tlsPurpose := fmt.Sprintf("tls-int-tls")
	tlsCommonName := computeTLSCommonName(caConfig.ComponentName(), signingIssuerRef, tlsIssuerRef)
	_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		Org:       org,
		IssuerRef: tlsIssuerRef,
		Profile:   "tls",
		Subject: &pb.CertificateSubject{
			CommonName:         tlsCommonName,
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			DnsNames:           caConfig.TLSDNSNames(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultTLSKeyRef(),
			CertRef: caConfig.VaultTLSCertRef(),
		},
		AdminIdentity: &pb.IdentityRef{
			KeyRef:  signingIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: signingIssuerConfig.VaultAdminSigningCertRef(),
		},
		Purpose:       tlsPurpose,
		IssuerCertRef: signingIssuerConfig.VaultSigningCertRef(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w (tls): %w", ErrCreateCertificate, err)
	}

	// Step 5: Generate bootstrap admin credentials
	logger.Info("Generating bootstrap admin credentials", "ca", caConfig.Description())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 6: Deploy CA pod
	logger.Info("Deploying CA pod", "ca", caConfig.Description())
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: caConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	// Step 7: Create bootstrap admin (TLS from self)
	logger.Info("Creating bootstrap admin", "ca", caConfig.Description())
	_, err = pb.CreateBootstrapAdminChild(ctx, &pb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        caRef, // Admin TLS from self (TLS Int)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	// Step 8: Create permanent admin (TLS from self)
	logger.Info("Creating permanent admin", "ca", caConfig.Description())
	_, err = pb.CreateAdminChild(ctx, &pb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef, // Admin signing from self
		TlsIssuer:        caRef, // Admin TLS from self
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	logger.Info("TLS Intermediate Certificate Authority deployment complete", "ca", caConfig.Description())
	return &pb.DeployTlsIntermediateCertificateAuthorityOutput{}, nil
}

// =============================================================================
// DeployOrgRootCertificateAuthority - Org Root CA (org-owned, self-signed signing, TLS via Nexus)
// =============================================================================

func (w *Workflows) DeployOrgRootCertificateAuthority(ctx workflow.Context, input *pb.DeployOrgRootCertificateAuthorityWorkflowInput) (pb.DeployOrgRootCertificateAuthorityWorkflow, error) {
	return newDeployOrgRootCertificateAuthorityWorkflow(input)
}

type deployOrgRootCertificateAuthorityWorkflow struct {
	input *pb.DeployOrgRootCertificateAuthorityWorkflowInput
}

func newDeployOrgRootCertificateAuthorityWorkflow(input *pb.DeployOrgRootCertificateAuthorityWorkflowInput) (*deployOrgRootCertificateAuthorityWorkflow, error) {
	w := &deployOrgRootCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *deployOrgRootCertificateAuthorityWorkflow) validate() error {
	if w.input == nil || w.input.Req == nil {
		return fmt.Errorf("%w: input must not be nil", ErrInvalidInput)
	}
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org must not be empty", ErrInvalidInput)
	}
	if req.GetTlsIssuer() == nil {
		return fmt.Errorf("%w: tls_issuer must not be nil", ErrInvalidInput)
	}
	return nil
}

func (w *deployOrgRootCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployOrgRootCertificateAuthorityOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	validityDuration := w.input.Req.GetValidityDuration().AsDuration()

	// Org Root CA is always org-owned, certType="org", id="root"
	caConfig := fabricconfig.NewCertificateAuthorityConfig(orgConfig, "org", "root")
	caRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}
	
	// TLS issuer is always TLS Int (from input)
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, tlsIssuerRef.GetCertType(), tlsIssuerRef.GetId())
	
	logger.Info("Deploying Org Root Certificate Authority", "ca", caConfig.Description())

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
		Org:   org,
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	// Step 3: Create self-signed signing certificate
	logger.Info("Creating self-signed signing certificate", "ca", caConfig.Description())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: caConfig.VaultSigningKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key pair: %w", err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
		KeyRef:             caConfig.VaultSigningKeyRef(),
		CertRef:            caConfig.VaultSigningCertRef(),
		CommonName:         caConfig.ComponentName(),
		Organization:       org,
		OrganizationalUnit: caConfig.OrganizationalUnit(),
		ValidityDuration:   durationpb.New(validityDuration),
		MaxPathLength:      1,
		DnsNames:           nil, // No SANs for CA signing certs
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed signing certificate: %w", err)
	}

	// Step 4: Create TLS certificate via Nexus from operator's TLS Int CA
	logger.Info("Creating CA TLS certificate via Nexus", "ca", caConfig.Description(), "issuer", tlsIssuerRef)
	tlsPurpose := fmt.Sprintf("org-root-tls")
	tlsCommonName := computeTLSCommonName(caConfig.ComponentName(), nil, tlsIssuerRef)
	
	fabricNexusClient := fabricv1nexustemporal.NewFabricNexusClient("fabric-IssueTlsCertificate")
	_, err = fabricNexusClient.IssueTlsCertificate(ctx, &pb.IssueTlsCertificateInput{
		Org:       org,
		IssuerRef: tlsIssuerRef,
		Profile:   "tls",
		Subject: &pb.CertificateSubject{
			CommonName:         tlsCommonName,
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			DnsNames:           caConfig.TLSDNSNames(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultTLSKeyRef(),
			CertRef: caConfig.VaultTLSCertRef(),
		},
		Purpose:       tlsPurpose,
		IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w (tls): %w", ErrCreateCertificate, err)
	}

	// Step 5: Generate bootstrap admin credentials
	logger.Info("Generating bootstrap admin credentials", "ca", caConfig.Description())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 6: Deploy CA pod
	logger.Info("Deploying CA pod", "ca", caConfig.Description())
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: caConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	// Step 7: Create bootstrap admin (TLS from TLS Int, not self)
	logger.Info("Creating bootstrap admin", "ca", caConfig.Description())
	_, err = pb.CreateBootstrapAdminChild(ctx, &pb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        tlsIssuerRef, // Admin TLS from TLS Int (not self)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	// Step 8: Create permanent admin (TLS from TLS Int, not self)
	logger.Info("Creating permanent admin", "ca", caConfig.Description())
	_, err = pb.CreateAdminChild(ctx, &pb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef,        // Admin signing from self
		TlsIssuer:        tlsIssuerRef, // Admin TLS from TLS Int (not self)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	logger.Info("Org Root Certificate Authority deployment complete", "ca", caConfig.Description())
	return &pb.DeployOrgRootCertificateAuthorityOutput{}, nil
}

// =============================================================================
// DeployOrgIntermediateCertificateAuthority - Org Int CA (org-owned, signing from Org Root, TLS via Nexus)
// =============================================================================

func (w *Workflows) DeployOrgIntermediateCertificateAuthority(ctx workflow.Context, input *pb.DeployOrgIntermediateCertificateAuthorityWorkflowInput) (pb.DeployOrgIntermediateCertificateAuthorityWorkflow, error) {
	return newDeployOrgIntermediateCertificateAuthorityWorkflow(input)
}

type deployOrgIntermediateCertificateAuthorityWorkflow struct {
	input *pb.DeployOrgIntermediateCertificateAuthorityWorkflowInput
}

func newDeployOrgIntermediateCertificateAuthorityWorkflow(input *pb.DeployOrgIntermediateCertificateAuthorityWorkflowInput) (*deployOrgIntermediateCertificateAuthorityWorkflow, error) {
	w := &deployOrgIntermediateCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *deployOrgIntermediateCertificateAuthorityWorkflow) validate() error {
	if w.input == nil || w.input.Req == nil {
		return fmt.Errorf("%w: input must not be nil", ErrInvalidInput)
	}
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org must not be empty", ErrInvalidInput)
	}
	if req.GetTlsIssuer() == nil {
		return fmt.Errorf("%w: tls_issuer must not be nil", ErrInvalidInput)
	}
	return nil
}

func (w *deployOrgIntermediateCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployOrgIntermediateCertificateAuthorityOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()

	// Org Int CA is always org-owned, certType="org", id="int"
	caConfig := fabricconfig.NewCertificateAuthorityConfig(orgConfig, "org", "int")
	caRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "int"}
	
	// Signing issuer is always Org Root (org-owned, direct call)
	signingIssuerRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}
	signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(orgConfig, "org", "root")
	
	// TLS issuer is always TLS Int (from input, via Nexus)
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, tlsIssuerRef.GetCertType(), tlsIssuerRef.GetId())
	
	logger.Info("Deploying Org Intermediate Certificate Authority", "ca", caConfig.Description())

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
		Org:   org,
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	// Step 3: Create signing certificate issued by Org Root (direct workflow call)
	logger.Info("Creating CA-signed signing certificate via direct call", "ca", caConfig.Description(), "issuer", signingIssuerConfig.Description())
	signingPurpose := fmt.Sprintf("org-int-signing")
	_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		Org:       org,
		IssuerRef: signingIssuerRef,
		Profile:   "ca",
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.ComponentName(),
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultSigningKeyRef(),
			CertRef: caConfig.VaultSigningCertRef(),
		},
		AdminIdentity: &pb.IdentityRef{
			KeyRef:  signingIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: signingIssuerConfig.VaultAdminSigningCertRef(),
		},
		Purpose: signingPurpose,
		Attributes: map[string]string{
			"hf.IntermediateCA": "true",
		},
		IssuerCertRef: signingIssuerConfig.VaultSigningCertRef(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w (signing): %w", ErrCreateCertificate, err)
	}

	// Step 4: Create TLS certificate via Nexus from operator's TLS Int CA
	logger.Info("Creating CA TLS certificate via Nexus", "ca", caConfig.Description(), "issuer", tlsIssuerRef)
	tlsPurpose := fmt.Sprintf("org-int-tls")
	tlsCommonName := computeTLSCommonName(caConfig.ComponentName(), signingIssuerRef, tlsIssuerRef)
	
	fabricNexusClient := fabricv1nexustemporal.NewFabricNexusClient("fabric-IssueTlsCertificate")
	_, err = fabricNexusClient.IssueTlsCertificate(ctx, &pb.IssueTlsCertificateInput{
		Org:       org,
		IssuerRef: tlsIssuerRef,
		Profile:   "tls",
		Subject: &pb.CertificateSubject{
			CommonName:         tlsCommonName,
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			DnsNames:           caConfig.TLSDNSNames(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultTLSKeyRef(),
			CertRef: caConfig.VaultTLSCertRef(),
		},
		Purpose:       tlsPurpose,
		IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
	}, workflow.NexusOperationOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("%w (tls): %w", ErrCreateCertificate, err)
	}

	// Step 5: Generate bootstrap admin credentials
	logger.Info("Generating bootstrap admin credentials", "ca", caConfig.Description())
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 6: Deploy CA pod
	logger.Info("Deploying CA pod", "ca", caConfig.Description())
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: caConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	// Step 7: Create bootstrap admin (TLS from TLS Int, not self)
	logger.Info("Creating bootstrap admin", "ca", caConfig.Description())
	_, err = pb.CreateBootstrapAdminChild(ctx, &pb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        tlsIssuerRef, // Admin TLS from TLS Int (not self)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	// Step 8: Create permanent admin (TLS from TLS Int, not self)
	logger.Info("Creating permanent admin", "ca", caConfig.Description())
	_, err = pb.CreateAdminChild(ctx, &pb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef,        // Admin signing from self
		TlsIssuer:        tlsIssuerRef, // Admin TLS from TLS Int (not self)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	logger.Info("Org Intermediate Certificate Authority deployment complete", "ca", caConfig.Description())
	return &pb.DeployOrgIntermediateCertificateAuthorityOutput{}, nil
}

// =============================================================================
// CreateCertificateDatabase - Creates PostgreSQL database for a CA
// =============================================================================

func (w *Workflows) CreateCertificateDatabase(ctx workflow.Context, input *pb.CreateCertificateDatabaseWorkflowInput) (pb.CreateCertificateDatabaseWorkflow, error) {
	return newCreateCertificateDatabaseWorkflow(input)
}

type createCertificateDatabaseWorkflow struct {
	input *pb.CreateCertificateDatabaseWorkflowInput
}

// newCreateCertificateDatabaseWorkflow creates and validates a new createCertificateDatabaseWorkflow
func newCreateCertificateDatabaseWorkflow(input *pb.CreateCertificateDatabaseWorkflowInput) (*createCertificateDatabaseWorkflow, error) {
	w := &createCertificateDatabaseWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createCertificateDatabaseWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
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
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()
	caID := caRef.GetId()

	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	if certType == "tls" {
		ownerConfig = operatorConfig
	} else {
		ownerConfig = orgConfig
	}

	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, certType, caID)
	certStore := fabricconfig.NewCertificateStoreConfig(orgConfig)

	// Wait for certificate store to be ready
	logger.Info("Waiting for certificate store to be ready", "org", org)
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

	logger.Info("Certificate store ready", "org", org)

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
	err = workflow.ExecuteActivity(activityCtx, postgrespb.CreateDatabaseActivityName, &postgrespb.CreateDatabaseInput{
		DatabaseName: caConfig.DatabaseName(),
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Generate credentials
	credentialsRef := caConfig.VaultDatabaseUserRef()
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: credentialsRef,
		Username:  caConfig.DatabaseUsername(),
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate credentials: %w", err)
	}

	// Create user
	err = workflow.ExecuteActivity(activityCtx, postgrespb.CreateUserActivityName, &postgrespb.CreateUserInput{
		CredentialsRef: credentialsRef,
	}).Get(activityCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Grant privileges
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

// =============================================================================
// CreateCertificate - Atomic unit of certificate issuance
// =============================================================================

func (w *Workflows) CreateCertificate(ctx workflow.Context, input *pb.CreateCertificateWorkflowInput) (pb.CreateCertificateWorkflow, error) {
	return newCreateCertificateWorkflow(input)
}

type createCertificateWorkflow struct {
	input *pb.CreateCertificateWorkflowInput
}

// newCreateCertificateWorkflow creates and validates a new createCertificateWorkflow
func newCreateCertificateWorkflow(input *pb.CreateCertificateWorkflowInput) (*createCertificateWorkflow, error) {
	w := &createCertificateWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createCertificateWorkflow) validate() error {
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	if req.GetIssuerRef() == nil {
		return fmt.Errorf("%w: issuer_ref is required (use secrets.CreateSelfSignedCertificate for self-signed)", ErrInvalidInput)
	}
	if req.GetStorage() == nil {
		return fmt.Errorf("%w: storage is required", ErrInvalidInput)
	}
	if req.GetStorage().GetKeyRef() == nil {
		return fmt.Errorf("%w: storage.key_ref is required", ErrInvalidInput)
	}
	if req.GetStorage().GetCertRef() == nil {
		return fmt.Errorf("%w: storage.cert_ref is required", ErrInvalidInput)
	}
	if req.GetIssuerCertRef() == nil {
		return fmt.Errorf("%w: issuer_cert_ref is required", ErrInvalidInput)
	}
	certType := req.GetIssuerRef().GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: issuer cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	return nil
}

func (w *createCertificateWorkflow) Execute(ctx workflow.Context) (*pb.CreateCertificateOutput, error) {
	req := w.input.Req
	org := req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	issuerRef := req.GetIssuerRef()
	certType := issuerRef.GetCertType()
	caID := issuerRef.GetId()
	storage := req.GetStorage()

	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	if certType == "tls" {
		ownerConfig = operatorConfig
	} else {
		ownerConfig = orgConfig
	}

	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, certType, caID)

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

	// Step 1: Generate key pair at storage.key_ref (keys/ path)
	err := workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: storage.GetKeyRef(),
	}).Get(ctx, nil)
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
	err = workflow.ExecuteActivity(activityCtx, pb.RegisterIdentityActivityName, &pb.RegisterIdentityInput{
		Org:           org,
		CaRef:         issuerRef,
		IdentityName:  identityName,
		Affiliation:   "",
		Attributes:    req.GetAttributes(),
		IdentityType:  identityType,
		AdminIdentity: req.GetAdminIdentity(),
	}).Get(ctx, &registerOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to register identity: %w", err)
	}

	// Step 3: Enroll identity with CA
	// Keys and certs are stored at separate paths via storage IdentityRef
	var enrollOutput pb.EnrollIdentityOutput
	err = workflow.ExecuteActivity(activityCtx, pb.EnrollIdentityActivityName, &pb.EnrollIdentityInput{
		Org:              org,
		CaRef:            issuerRef,
		IdentityName:     identityName,
		EnrollmentSecret: registerOutput.GetEnrollmentSecret(),
		Storage:          storage, // IdentityRef with key and cert refs
		Subject:          subject,
		CaTlsCertRef:     caConfig.VaultTLSCertRef(), // CA's TLS cert for mTLS
		IssuerCertRef:    req.GetIssuerCertRef(),     // Issuer's cert path for chain building
		Profile:          req.GetProfile(),
	}).Get(ctx, &enrollOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity: %w", err)
	}

	return &pb.CreateCertificateOutput{
		Identity: storage,
	}, nil
}

// =============================================================================
// IssueTlsCertificate - Nexus operation for cross-namespace TLS cert issuance
// =============================================================================

func (w *Workflows) IssueTlsCertificate(ctx workflow.Context, input *pb.IssueTlsCertificateWorkflowInput) (pb.IssueTlsCertificateWorkflow, error) {
	return newIssueTlsCertificateWorkflow(input)
}

type IssueTlsCertificateWorkflow struct {
	input *pb.IssueTlsCertificateWorkflowInput
}

// newIssueTlsCertificateWorkflow creates and validates a new IssueTlsCertificateWorkflow
func newIssueTlsCertificateWorkflow(input *pb.IssueTlsCertificateWorkflowInput) (*IssueTlsCertificateWorkflow, error) {
	w := &IssueTlsCertificateWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *IssueTlsCertificateWorkflow) validate() error {
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	if req.GetIssuerRef() == nil {
		return fmt.Errorf("%w: issuer_ref is required", ErrInvalidInput)
	}
	if req.GetIssuerRef().GetCertType() != "tls" {
		return fmt.Errorf("%w: issuer must be TLS CA", ErrInvalidCertType)
	}
	if req.GetStorage() == nil {
		return fmt.Errorf("%w: storage is required", ErrInvalidInput)
	}
	if req.GetSubject() == nil {
		return fmt.Errorf("%w: subject is required", ErrInvalidInput)
	}
	return nil
}

func (w *IssueTlsCertificateWorkflow) Execute(ctx workflow.Context) (*pb.IssueTlsCertificateOutput, error) {
	req := w.input.Req
	operatorConfig := config.GetOperator()
	issuerRef := req.GetIssuerRef()
	storage := req.GetStorage()

	// Get TLS CA config (operator-owned)
	tlsCAConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, issuerRef.GetCertType(), issuerRef.GetId())

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

	// Step 1: Generate key pair (stored in org namespace)
	err := workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: storage.GetKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Step 2: Register identity with TLS CA (using operator admin credentials)
	var registerOutput pb.RegisterIdentityOutput
	err = workflow.ExecuteActivity(activityCtx, pb.RegisterIdentityActivityName, &pb.RegisterIdentityInput{
		Org:          operatorConfig.OwnerID(), // Operator owns the TLS CA
		CaRef:        issuerRef,
		IdentityName: req.GetSubject().GetCommonName(),
		Affiliation:  "",
		Attributes:   req.GetAttributes(),
		IdentityType: "client", // Default type
		AdminIdentity: &pb.IdentityRef{
			KeyRef:  tlsCAConfig.VaultAdminSigningKeyRef(),
			CertRef: tlsCAConfig.VaultAdminSigningCertRef(),
		},
	}).Get(ctx, &registerOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to register identity: %w", err)
	}

	// Step 3: Enroll identity (stores cert in org namespace)
	err = workflow.ExecuteActivity(activityCtx, pb.EnrollIdentityActivityName, &pb.EnrollIdentityInput{
		Org:              operatorConfig.OwnerID(), // CA is operator-owned
		CaRef:            issuerRef,
		IdentityName:     req.GetSubject().GetCommonName(),
		EnrollmentSecret: registerOutput.GetEnrollmentSecret(),
		Storage:          storage, // Org namespace storage
		Subject:          req.GetSubject(),
		CaTlsCertRef:     tlsCAConfig.VaultTLSCertRef(),
		IssuerCertRef:    req.GetIssuerCertRef(),
		Profile:          req.GetProfile(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity: %w", err)
	}

	return &pb.IssueTlsCertificateOutput{
		Identity: storage,
	}, nil
}

// =============================================================================
// CreateBootstrapAdmin - Creates bootstrap admin after CA deployment
// =============================================================================

func (w *Workflows) CreateBootstrapAdmin(ctx workflow.Context, input *pb.CreateBootstrapAdminWorkflowInput) (pb.CreateBootstrapAdminWorkflow, error) {
	return newCreateBootstrapAdminWorkflow(input)
}

type createBootstrapAdminWorkflow struct {
	input *pb.CreateBootstrapAdminWorkflowInput
}

// newCreateBootstrapAdminWorkflow creates and validates a new createBootstrapAdminWorkflow
func newCreateBootstrapAdminWorkflow(input *pb.CreateBootstrapAdminWorkflowInput) (*createBootstrapAdminWorkflow, error) {
	w := &createBootstrapAdminWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createBootstrapAdminWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	if w.input.Req.GetCaRef() == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	return nil
}

func (w *createBootstrapAdminWorkflow) Execute(ctx workflow.Context) (*pb.CreateBootstrapAdminOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()
	caID := caRef.GetId()

	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	if certType == "tls" {
		ownerConfig = operatorConfig
	} else {
		ownerConfig = orgConfig
	}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, certType, caID)

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

	// Create keypair for signing identity (keys/ path)
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: caConfig.VaultBootstrapAdminSigningKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key pair: %w", err)
	}

	// Bootstrap enroll signing identity from this CA
	// Storage IdentityRef holds both key and cert refs
	signingStorage := &pb.IdentityRef{
		KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
		CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
	}
	err = workflow.ExecuteActivity(activityCtx, pb.BootstrapEnrollIdentityActivityName, &pb.BootstrapEnrollIdentityInput{
		CaUrl:                   caConfig.URL(),
		BootstrapCredentialsRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Storage:                 signingStorage,
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.BootstrapAdminCommonName(),
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		CaTlsCertRef:  caConfig.VaultTLSCertRef(),
		IssuerCertRef: caConfig.VaultSigningCertRef(), // CA's signing cert for chain building
		Profile:       "",                             // Default profile for admin identity
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap enroll signing identity: %w", err)
	}

	// Create keypair for TLS client cert (keys/ path)
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: caConfig.VaultBootstrapAdminTLSKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS key pair: %w", err)
	}

	// Bootstrap enroll TLS identity from the appropriate CA
	// Use explicit tls_issuer from input - if nil, use the CA itself (TLS Root CA case)
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	var tlsIssuerConfig fabricconfig.CertificateAuthorityConfig
	if tlsIssuerRef == nil || tlsIssuerRef.GetCertType() == "" {
		// No explicit TLS issuer - use the CA itself (TLS Root CA case)
		tlsIssuerConfig = caConfig
		tlsIssuerRef = caRef // For collision detection
	} else {
		// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
		var tlsIssuerOwner config.OwnerConfig
		if tlsIssuerRef.GetCertType() == "tls" {
			tlsIssuerOwner = operatorConfig
		} else {
			tlsIssuerOwner = orgConfig
		}
		tlsIssuerConfig = fabricconfig.NewCertificateAuthorityConfig(tlsIssuerOwner, tlsIssuerRef.GetCertType(), tlsIssuerRef.GetId())
	}

	// Compute TLS common name - signing identity always comes from this CA (bootstrap enrollment)
	// For collision detection, signing issuer is effectively this CA (caRef)
	tlsCommonName := computeTLSCommonName(
		fmt.Sprintf("bootstrap-admin@%s", caConfig.ComponentName()),
		caRef, // signing from this CA
		tlsIssuerRef,
	)

	tlsStorage := &pb.IdentityRef{
		KeyRef:  caConfig.VaultBootstrapAdminTLSKeyRef(),
		CertRef: caConfig.VaultBootstrapAdminTLSCertRef(),
	}
	err = workflow.ExecuteActivity(activityCtx, pb.BootstrapEnrollIdentityActivityName, &pb.BootstrapEnrollIdentityInput{
		CaUrl:                   tlsIssuerConfig.URL(),
		BootstrapCredentialsRef: tlsIssuerConfig.VaultBootstrapAdminCredentialsRef(),
		Storage:                 tlsStorage,
		Subject: &pb.CertificateSubject{
			CommonName:         tlsCommonName,
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		CaTlsCertRef:  tlsIssuerConfig.VaultTLSCertRef(),
		IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(), // TLS issuer's signing cert for chain building
		Profile:       "tls",                                 // TLS profile for client cert
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap enroll TLS identity: %w", err)
	}

	logger.Info("Bootstrap admin created", "ca", caConfig.Description())

	return &pb.CreateBootstrapAdminOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}

// =============================================================================
// CreateAdmin - Creates permanent admin using bootstrap admin
// =============================================================================

func (w *Workflows) CreateAdmin(ctx workflow.Context, input *pb.CreateAdminWorkflowInput) (pb.CreateAdminWorkflow, error) {
	return newCreateAdminWorkflow(input)
}

type createAdminWorkflow struct {
	input *pb.CreateAdminWorkflowInput
}

// newCreateAdminWorkflow creates and validates a new createAdminWorkflow
func newCreateAdminWorkflow(input *pb.CreateAdminWorkflowInput) (*createAdminWorkflow, error) {
	w := &createAdminWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createAdminWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	if w.input.Req.GetCaRef() == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	return nil
}

func (w *createAdminWorkflow) Execute(ctx workflow.Context) (*pb.CreateAdminOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()
	caID := caRef.GetId()

	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	if certType == "tls" {
		ownerConfig = operatorConfig
	} else {
		ownerConfig = orgConfig
	}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, certType, caID)

	logger.Info("Creating permanent admin", "ca", caConfig.Description())

	// Use explicit signing issuer from input - if nil, use the CA itself
	// (Per Fabric CA constraints, admin must be in CA's database)
	signingIssuerRef := w.input.Req.GetSigningIssuer()
	if signingIssuerRef == nil || signingIssuerRef.GetCertType() == "" {
		signingIssuerRef = caRef // Default: signing from this CA itself
	}
	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var signingIssuerOwner config.OwnerConfig
	if signingIssuerRef.GetCertType() == "tls" {
		signingIssuerOwner = operatorConfig
	} else {
		signingIssuerOwner = orgConfig
	}
	signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(signingIssuerOwner, signingIssuerRef.GetCertType(), signingIssuerRef.GetId())
	signingPurpose := fmt.Sprintf("%s-%s-admin-signing", certType, caID)

	// Determine which credentials to use for registration with the signing issuer:
	// - If signing issuer is this CA itself → use this CA's bootstrap admin
	// - If signing issuer is a different CA → use that CA's permanent admin
	var signingAdminIdentity *pb.IdentityRef
	if signingIssuerRef.GetCertType() == caRef.GetCertType() && signingIssuerRef.GetId() == caRef.GetId() {
		signingAdminIdentity = &pb.IdentityRef{
			KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
			CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
		}
	} else {
		signingAdminIdentity = &pb.IdentityRef{
			KeyRef:  signingIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: signingIssuerConfig.VaultAdminSigningCertRef(),
		}
	}

	signingStorage := &pb.IdentityRef{
		KeyRef:  caConfig.VaultAdminSigningKeyRef(),
		CertRef: caConfig.VaultAdminSigningCertRef(),
	}
	_, err := pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		Org:       org,
		IssuerRef: signingIssuerRef,
		Profile:   "", // Default profile for admin identity
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.AdminCommonName(),
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		Storage:       signingStorage,
		AdminIdentity: signingAdminIdentity,
		Purpose:       signingPurpose,
		IssuerCertRef: signingIssuerConfig.VaultSigningCertRef(),
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

	// Use explicit TLS issuer from input - if nil, use the CA itself
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	if tlsIssuerRef == nil || tlsIssuerRef.GetCertType() == "" {
		tlsIssuerRef = caRef // Default: TLS from this CA itself
	}
	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var tlsIssuerOwner config.OwnerConfig
	if tlsIssuerRef.GetCertType() == "tls" {
		tlsIssuerOwner = operatorConfig
	} else {
		tlsIssuerOwner = orgConfig
	}
	tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(tlsIssuerOwner, tlsIssuerRef.GetCertType(), tlsIssuerRef.GetId())
	tlsPurpose := fmt.Sprintf("%s-%s-admin-tls", certType, caID)

	// Compute TLS common name - add "-tls" suffix if same CA issues both to avoid collision
	tlsCommonName := computeTLSCommonName(caConfig.AdminCommonName(), signingIssuerRef, tlsIssuerRef)

	// Determine which credentials to use for registration with the TLS issuer:
	// - If TLS issuer is this CA itself → use this CA's bootstrap admin
	// - If TLS issuer is a different CA → use that CA's permanent admin
	var tlsAdminIdentity *pb.IdentityRef
	if tlsIssuerRef.GetCertType() == caRef.GetCertType() && tlsIssuerRef.GetId() == caRef.GetId() {
		tlsAdminIdentity = &pb.IdentityRef{
			KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
			CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
		}
	} else {
		tlsAdminIdentity = &pb.IdentityRef{
			KeyRef:  tlsIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: tlsIssuerConfig.VaultAdminSigningCertRef(),
		}
	}

	tlsStorage := &pb.IdentityRef{
		KeyRef:  caConfig.VaultAdminTLSKeyRef(),
		CertRef: caConfig.VaultAdminTLSCertRef(),
	}

	if tlsIssuerRef.GetCertType() == "tls" {
		// TLS CA issuer - use Nexus for cross-namespace certificate issuance
		nexusClient := fabricv1nexustemporal.NewFabricNexusClient("fabric-IssueTlsCertificate")
		_, err = nexusClient.IssueTlsCertificate(ctx, &pb.IssueTlsCertificateInput{
			Org:       org,
			IssuerRef: tlsIssuerRef,
			Profile:   "tls",
			Subject: &pb.CertificateSubject{
				CommonName:         tlsCommonName,
				Organization:       org,
				OrganizationalUnit: caConfig.OrganizationalUnit(),
			},
			Storage:       tlsStorage,
			Purpose:       tlsPurpose,
			IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
		}, workflow.NexusOperationOptions{
			ScheduleToCloseTimeout: 5 * time.Minute,
		})
	} else {
		// Org CA or self - use direct CreateCertificate with appropriate admin identity
		_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
			Org:       org,
			IssuerRef: tlsIssuerRef,
			Profile:   "tls",
			Subject: &pb.CertificateSubject{
				CommonName:         tlsCommonName, // Different name if same CA issues both
				Organization:       org,
				OrganizationalUnit: caConfig.OrganizationalUnit(),
			},
			Storage:       tlsStorage,
			AdminIdentity: tlsAdminIdentity,
			Purpose:       tlsPurpose,
			IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create admin TLS identity: %w", err)
	}

	logger.Info("Permanent admin created", "ca", caConfig.Description())

	return &pb.CreateAdminOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}

// =============================================================================
// CreateIdentity - Creates full Fabric identity with both org and TLS certs
// =============================================================================

func (w *Workflows) CreateIdentity(ctx workflow.Context, input *pb.CreateIdentityWorkflowInput) (pb.CreateIdentityWorkflow, error) {
	return newCreateIdentityWorkflow(input)
}

type createIdentityWorkflow struct {
	input *pb.CreateIdentityWorkflowInput
}

// newCreateIdentityWorkflow creates and validates a new createIdentityWorkflow
func newCreateIdentityWorkflow(input *pb.CreateIdentityWorkflowInput) (*createIdentityWorkflow, error) {
	w := &createIdentityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createIdentityWorkflow) validate() error {
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	if req.GetIdentityName() == "" {
		return fmt.Errorf("%w: identity_name is required", ErrInvalidInput)
	}
	if req.GetOrgIssuerRef() == nil {
		return fmt.Errorf("%w: org_issuer_ref is required", ErrInvalidInput)
	}
	if req.GetTlsIssuerRef() == nil {
		return fmt.Errorf("%w: tls_issuer_ref is required", ErrInvalidInput)
	}
	if req.GetSigningStorage() == nil {
		return fmt.Errorf("%w: signing_storage is required", ErrInvalidInput)
	}
	if req.GetSigningStorage().GetKeyRef() == nil {
		return fmt.Errorf("%w: signing_storage.key_ref is required", ErrInvalidInput)
	}
	if req.GetSigningStorage().GetCertRef() == nil {
		return fmt.Errorf("%w: signing_storage.cert_ref is required", ErrInvalidInput)
	}
	if req.GetTlsStorage() == nil {
		return fmt.Errorf("%w: tls_storage is required", ErrInvalidInput)
	}
	if req.GetTlsStorage().GetKeyRef() == nil {
		return fmt.Errorf("%w: tls_storage.key_ref is required", ErrInvalidInput)
	}
	if req.GetTlsStorage().GetCertRef() == nil {
		return fmt.Errorf("%w: tls_storage.cert_ref is required", ErrInvalidInput)
	}
	return nil
}

func (w *createIdentityWorkflow) Execute(ctx workflow.Context) (*pb.CreateIdentityOutput, error) {
	req := w.input.Req
	org := req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	identityName := req.GetIdentityName()
	orgIssuer := req.GetOrgIssuerRef()
	tlsIssuer := req.GetTlsIssuerRef()
	signingStorage := req.GetSigningStorage()
	tlsStorage := req.GetTlsStorage()

	// Get issuer configs to access their admin credentials
	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var orgIssuerOwner config.OwnerConfig
	if orgIssuer.GetCertType() == "tls" {
		orgIssuerOwner = operatorConfig
	} else {
		orgIssuerOwner = orgConfig
	}
	orgIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(orgIssuerOwner, orgIssuer.GetCertType(), orgIssuer.GetId())

	var tlsIssuerOwner config.OwnerConfig
	if tlsIssuer.GetCertType() == "tls" {
		tlsIssuerOwner = operatorConfig
	} else {
		tlsIssuerOwner = orgConfig
	}
	tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(tlsIssuerOwner, tlsIssuer.GetCertType(), tlsIssuer.GetId())

	// Create org identity cert (uses identity_type for proper OU classification)
	signingPurpose := fmt.Sprintf("identity-%s-signing", identityName)
	_, err := pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		Org:       org,
		IssuerRef: orgIssuer,
		Profile:   "", // Default profile
		Subject:   req.GetSubject(),
		Storage:   signingStorage,
		AdminIdentity: &pb.IdentityRef{
			KeyRef:  orgIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: orgIssuerConfig.VaultAdminSigningCertRef(),
		},
		Purpose:       signingPurpose,
		IssuerCertRef: orgIssuerConfig.VaultSigningCertRef(),
		Attributes:    req.GetAttributes(),
		IdentityType:  req.GetIdentityType(), // Pass through for NodeOUs classification
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	// Create TLS client cert
	// TLS certs don't need identity_type since NodeOUs only applies to org/signing certs
	tlsPurpose := fmt.Sprintf("identity-%s-tls", identityName)

	if tlsIssuer.GetCertType() == "tls" {
		// TLS CA issuer - use Nexus for cross-namespace certificate issuance
		nexusClient := fabricv1nexustemporal.NewFabricNexusClient("fabric-IssueTlsCertificate")
		_, err = nexusClient.IssueTlsCertificate(ctx, &pb.IssueTlsCertificateInput{
			Org:       org,
			IssuerRef: tlsIssuer,
			Profile:   req.GetTlsProfile(),
			Subject: &pb.CertificateSubject{
				CommonName:         req.GetSubject().GetCommonName(),
				Organization:       req.GetSubject().GetOrganization(),
				OrganizationalUnit: req.GetSubject().GetOrganizationalUnit(),
				DnsNames:           req.GetSubject().GetDnsNames(),
			},
			Storage:       tlsStorage,
			Purpose:       tlsPurpose,
			IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
		}, workflow.NexusOperationOptions{
			ScheduleToCloseTimeout: 5 * time.Minute,
		})
	} else {
		// Org CA - use direct CreateCertificate
		_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
			Org:       org,
			IssuerRef: tlsIssuer,
			Profile:   req.GetTlsProfile(),
			Subject: &pb.CertificateSubject{
				CommonName:         req.GetSubject().GetCommonName(),
				Organization:       req.GetSubject().GetOrganization(),
				OrganizationalUnit: req.GetSubject().GetOrganizationalUnit(),
				DnsNames:           req.GetSubject().GetDnsNames(),
			},
			Storage: tlsStorage,
			AdminIdentity: &pb.IdentityRef{
				KeyRef:  tlsIssuerConfig.VaultAdminSigningKeyRef(),
				CertRef: tlsIssuerConfig.VaultAdminSigningCertRef(),
			},
			Purpose:       tlsPurpose,
			IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS identity: %w", err)
	}

	return &pb.CreateIdentityOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}

// =============================================================================
// DeployCouchDb - Deploys CouchDB instance for a peer
// =============================================================================

func (w *Workflows) DeployCouchDb(ctx workflow.Context, input *pb.DeployCouchDbWorkflowInput) (pb.DeployCouchDbWorkflow, error) {
	return newDeployCouchDbWorkflow(input)
}

type deployCouchDbWorkflow struct {
	input *pb.DeployCouchDbWorkflowInput
}

// newDeployCouchDbWorkflow creates and validates a new deployCouchDbWorkflow
func newDeployCouchDbWorkflow(input *pb.DeployCouchDbWorkflowInput) (*deployCouchDbWorkflow, error) {
	w := &deployCouchDbWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployCouchDbWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployCouchDbWorkflow) Execute(ctx workflow.Context) (*pb.DeployCouchDbOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	peerID := w.input.Req.GetPeerId()
	couchDbConfig := fabricconfig.NewCouchDbConfig(orgConfig, peerID)

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
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: credentialsRef,
		Username:  "admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate couchdb credentials: %w", ErrDeployCouchDB, err)
	}

	// Step 3: Deploy CouchDB StatefulSet
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
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

	logger.Info("CouchDB deployment ready", "peerId", peerID, "org", org)
	return &pb.DeployCouchDbOutput{}, nil
}

// =============================================================================
// DeployPeer - Deploys Fabric peer node
// =============================================================================

func (w *Workflows) DeployPeer(ctx workflow.Context, input *pb.DeployPeerWorkflowInput) (pb.DeployPeerWorkflow, error) {
	return newDeployPeerWorkflow(input)
}

type deployPeerWorkflow struct {
	input *pb.DeployPeerWorkflowInput
}

// newDeployPeerWorkflow creates and validates a new deployPeerWorkflow
func newDeployPeerWorkflow(input *pb.DeployPeerWorkflowInput) (*deployPeerWorkflow, error) {
	w := &deployPeerWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployPeerWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	spec := w.input.Req.GetSpec()
	if spec == nil {
		return fmt.Errorf("%w: spec is required", ErrInvalidInput)
	}
	if spec.GetOrgIssuer() == nil {
		return fmt.Errorf("%w: spec.org_issuer is required", ErrInvalidInput)
	}
	if spec.GetTlsIssuer() == nil {
		return fmt.Errorf("%w: spec.tls_issuer is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployPeerWorkflow) Execute(ctx workflow.Context) (*pb.DeployPeerOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	spec := w.input.Req.GetSpec()
	peerID := spec.GetId()
	peerConfig := fabricconfig.NewPeerConfig(orgConfig, peerID)

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
		Org:          org,
		IdentityName: peerConfig.IdentityName(),
		Subject: &pb.CertificateSubject{
			CommonName:         peerConfig.CommonName(),
			Organization:       org,
			OrganizationalUnit: peerConfig.OrganizationalUnit(),
			DnsNames:           peerConfig.TLSDNSNames(),
		},
		OrgIssuerRef: spec.GetOrgIssuer(),
		TlsIssuerRef: spec.GetTlsIssuer(),
		SigningStorage: &pb.IdentityRef{
			KeyRef:  peerConfig.VaultSigningKeyRef(),
			CertRef: peerConfig.VaultSigningCertRef(),
		},
		TlsStorage: &pb.IdentityRef{
			KeyRef:  peerConfig.VaultTLSKeyRef(),
			CertRef: peerConfig.VaultTLSCertRef(),
		},
		TlsProfile:   "tls",
		IdentityType: "peer", // Critical: sets OU=peer in certificate for NodeOUs
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create peer identity: %w", ErrDeployPeer, err)
	}

	// Step 3: Deploy peer StatefulSet
	logger.Info("Deploying peer", "peerId", peerID)
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: peerConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create peer deployment: %w", ErrDeployPeer, err)
	}

	logger.Info("Peer deployment initiated", "peerId", peerID, "org", org)
	return &pb.DeployPeerOutput{}, nil
}

// =============================================================================
// DeployOrderer - Deploys Fabric orderer node
// =============================================================================

func (w *Workflows) DeployOrderer(ctx workflow.Context, input *pb.DeployOrdererWorkflowInput) (pb.DeployOrdererWorkflow, error) {
	return newDeployOrdererWorkflow(input)
}

type deployOrdererWorkflow struct {
	input *pb.DeployOrdererWorkflowInput
}

// newDeployOrdererWorkflow creates and validates a new deployOrdererWorkflow
func newDeployOrdererWorkflow(input *pb.DeployOrdererWorkflowInput) (*deployOrdererWorkflow, error) {
	w := &deployOrdererWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployOrdererWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	spec := w.input.Req.GetSpec()
	if spec == nil {
		return fmt.Errorf("%w: spec is required", ErrInvalidInput)
	}
	if spec.GetOrgIssuer() == nil {
		return fmt.Errorf("%w: spec.org_issuer is required", ErrInvalidInput)
	}
	if spec.GetTlsIssuer() == nil {
		return fmt.Errorf("%w: spec.tls_issuer is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployOrdererWorkflow) Execute(ctx workflow.Context) (*pb.DeployOrdererOutput, error) {
	logger := workflow.GetLogger(ctx)
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	spec := w.input.Req.GetSpec()
	ordererID := spec.GetId()
	ordererConfig := fabricconfig.NewOrdererConfig(orgConfig, ordererID)

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
		Org:          org,
		IdentityName: ordererConfig.IdentityName(),
		Subject: &pb.CertificateSubject{
			CommonName:         ordererConfig.CommonName(),
			Organization:       org,
			OrganizationalUnit: ordererConfig.OrganizationalUnit(),
			DnsNames:           ordererConfig.TLSDNSNames(),
		},
		OrgIssuerRef: spec.GetOrgIssuer(),
		TlsIssuerRef: spec.GetTlsIssuer(),
		SigningStorage: &pb.IdentityRef{
			KeyRef:  ordererConfig.VaultSigningKeyRef(),
			CertRef: ordererConfig.VaultSigningCertRef(),
		},
		TlsStorage: &pb.IdentityRef{
			KeyRef:  ordererConfig.VaultTLSKeyRef(),
			CertRef: ordererConfig.VaultTLSCertRef(),
		},
		TlsProfile:   "tls",
		IdentityType: "orderer", // Critical: sets OU=orderer in certificate for NodeOUs
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create orderer identity: %w", ErrDeployOrderer, err)
	}

	// Step 3: Deploy orderer StatefulSet
	logger.Info("Deploying orderer", "ordererId", ordererID)
	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: ordererConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create orderer deployment: %w", ErrDeployOrderer, err)
	}

	logger.Info("Orderer deployment initiated", "ordererId", ordererID, "org", org)
	return &pb.DeployOrdererOutput{}, nil
}

// =============================================================================
// DeployChainCode - Deploys chaincode (smart contract)
// =============================================================================

func (w *Workflows) DeployChainCode(ctx workflow.Context, input *pb.DeployChainCodeWorkflowInput) (pb.DeployChainCodeWorkflow, error) {
	return newDeployChainCodeWorkflow(input)
}

type deployChainCodeWorkflow struct {
	input *pb.DeployChainCodeWorkflowInput
}

// newDeployChainCodeWorkflow creates and validates a new deployChainCodeWorkflow
func newDeployChainCodeWorkflow(input *pb.DeployChainCodeWorkflowInput) (*deployChainCodeWorkflow, error) {
	w := &deployChainCodeWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployChainCodeWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
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
