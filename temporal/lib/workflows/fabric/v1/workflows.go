package fabric

import (
	"errors"
	"fmt"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Sentinel errors for fabric workflows
var (
	ErrInvalidInput               = errors.New("invalid input")
	ErrDeployCouchDB              = errors.New("deploy couchdb")
	ErrDeployPeer                 = errors.New("deploy peer")
	ErrDeployOrderer              = errors.New("deploy orderer")
	ErrDeployChainCode            = errors.New("deploy chaincode")
	ErrDeployCertificateStore     = errors.New("deploy certificate store")
	ErrCreateCertificateDatabase  = errors.New("create certificate database")
	ErrDeployCertificateAuthority = errors.New("deploy certificate authority")
	ErrCreateCertificate          = errors.New("create certificate")
	ErrCreateBootstrapAdmin       = errors.New("create bootstrap admin")
	ErrCreateAdmin                = errors.New("create admin")
	ErrCreateIdentity             = errors.New("create identity")
	ErrCreateAccess               = errors.New("create access")
	ErrInvalidCertType            = errors.New("invalid cert type")
)

// Workflows implements the FabricWorkflows interface
// Workflows no longer hold global orgConfig - they get org from each workflow's input
type Workflows struct{}

// NewWorkflows creates a new Workflows instance
func NewWorkflows() *Workflows {
	return &Workflows{}
}

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
	org := w.input.Req.GetOrg()
	deployment := w.input.Req.GetDeployment()

	// Extract Org CA validities from deployment specs
	var orgRootValidity, orgIntValidity *durationpb.Duration
	for _, caSpec := range deployment.GetCertificateAuthorities() {
		if caSpec.GetSelf().GetCertType() == "org" {
			if caSpec.GetSelf().GetId() == "root" {
				orgRootValidity = caSpec.GetValidityDuration()
			} else if caSpec.GetSelf().GetId() == "int" {
				orgIntValidity = caSpec.GetValidityDuration()
			}
		}
	}

	if _, err := pb.BootstrapOrgCertificateAuthoritiesChild(ctx, &pb.BootstrapOrgCertificateAuthoritiesInput{
		Org:                     org,
		OrgRootValidity:         orgRootValidity,
		OrgIntermediateValidity: orgIntValidity,
	}); err != nil {
		return nil, fmt.Errorf("bootstrap Org certificate authorities: %w", err)
	}

	for _, peerSpec := range deployment.GetPeers() {
		if _, err := pb.DeployCouchDbChild(ctx, &pb.DeployCouchDbInput{
			Org:    org,
			PeerId: peerSpec.GetId(),
		}); err != nil {
			return nil, fmt.Errorf("deploy couchdb for %s: %w", peerSpec.GetId(), err)
		}

		if _, err := pb.DeployPeerChild(ctx, &pb.DeployPeerInput{
			Org:  org,
			Spec: peerSpec,
		}); err != nil {
			return nil, fmt.Errorf("deploy peer %s: %w", peerSpec.GetId(), err)
		}
	}

	for _, ordererSpec := range deployment.GetOrderers() {
		if _, err := pb.DeployOrdererChild(ctx, &pb.DeployOrdererInput{
			Org:  org,
			Spec: ordererSpec,
		}); err != nil {
			return nil, fmt.Errorf("deploy orderer %s: %w", ordererSpec.GetId(), err)
		}
	}

	// TODO: Deploy chaincode, hook into channels

	return &pb.BootstrapShardOutput{}, nil
}

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
	orgIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(orgIssuerOwner, orgIssuer)

	var tlsIssuerOwner config.OwnerConfig
	if tlsIssuer.GetCertType() == "tls" {
		tlsIssuerOwner = operatorConfig
	} else {
		tlsIssuerOwner = orgConfig
	}
	tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(tlsIssuerOwner, tlsIssuer)

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

	_, err = NexusFabricCreateCertificate(ctx, &pb.CreateCertificateInput{
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
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create TLS identity: %w", err)
	}

	return &pb.CreateIdentityOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}

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
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	peerID := w.input.Req.GetPeerId()
	couchDbConfig := fabricconfig.NewCouchDbConfig(orgConfig, peerID)

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

	_, err := NexusSecretsCreateAccess(ctx, couchDbConfig)
	if err != nil {
		return nil, err
	}

	credentialsRef := couchDbConfig.VaultAdminRef()
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: credentialsRef,
		Username:  "admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate couchdb credentials: %w", ErrDeployCouchDB, err)
	}

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: couchDbConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create couchdb deployment: %w", ErrDeployCouchDB, err)
	}

	// This ensures the peer can connect to CouchDB when it starts
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

	return &pb.DeployCouchDbOutput{}, nil
}

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
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	spec := w.input.Req.GetSpec()
	peerID := spec.GetId()
	peerConfig := fabricconfig.NewPeerConfig(orgConfig, peerID)

	_, err := NexusSecretsCreateAccess(ctx, peerConfig)
	if err != nil {
		return nil, err
	}

	// The CreateIdentity workflow handles both org and TLS certificates
	// IdentityType "peer" ensures the certificate gets OU=peer for NodeOUs classification
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

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: peerConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create peer deployment: %w", ErrDeployPeer, err)
	}

	return &pb.DeployPeerOutput{}, nil
}

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
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	spec := w.input.Req.GetSpec()
	ordererID := spec.GetId()
	ordererConfig := fabricconfig.NewOrdererConfig(orgConfig, ordererID)

	_, err := NexusSecretsCreateAccess(ctx, ordererConfig)
	if err != nil {
		return nil, err
	}

	// The CreateIdentity workflow handles both org and TLS certificates
	// IdentityType "orderer" ensures the certificate gets OU=orderer for NodeOUs classification
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

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: ordererConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create orderer deployment: %w", ErrDeployOrderer, err)
	}

	return &pb.DeployOrdererOutput{}, nil
}

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
	// chainCodeID := w.input.Req.GetChainCodeId()

	// TODO: Implement chaincode deployment
	// 1. Nexus Call: secrets.v1.CreateAccess
	// 2. for only ("org"):
	//    - Workflow: fabric.v1.CreateIdentity (with IdentityType: "client")
	// 3. Workflow: cluster.fabric.v1.CreateDeployment
	// NOTE: IdentityType should be "client" for chaincode identities

	return &pb.DeployChainCodeOutput{}, nil
}
