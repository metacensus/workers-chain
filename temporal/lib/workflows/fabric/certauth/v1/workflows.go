package certauth

import (
	"errors"
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
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
		AdminIdentity: &fabricpb.IdentityRef{
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

	_, err = NexusCreateCertificate(ctx, &pb.CreateCertificateInput{
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
	storage := req.GetStorage()

	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	if certType == "tls" {
		ownerConfig = operatorConfig
	} else {
		ownerConfig = orgConfig
	}

	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, issuerRef)

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

	err := workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: storage.GetKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

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
