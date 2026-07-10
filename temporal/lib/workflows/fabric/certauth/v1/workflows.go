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

// Sentinel errors for certificate authority workflows
var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrNotOwned          = errors.New("certificate authority not owned by this worker")
	ErrCreateCertificate = errors.New("create certificate")
	ErrInvalidCertType   = errors.New("invalid cert type")
)

// Workflows implements the CertificateAuthorityWorkflows interface.
//
// Certauth workflows act AS their worker's owner: the owner is injected at
// worker startup and never carried in workflow inputs. A locally-referenced
// issuer always belongs to the executing worker's owner; foreign issuers
// (an org needing a TLS certificate from the operator's TLS chain) are
// reached via Nexus.
type Workflows struct {
	owner config.OwnerConfig
}

// NewWorkflows creates a new Workflows instance acting as the given owner.
func NewWorkflows(owner config.OwnerConfig) *Workflows {
	return &Workflows{owner: owner}
}

func (w *Workflows) CreateIdentity(ctx workflow.Context, input *pb.CreateIdentityWorkflowInput) (pb.CreateIdentityWorkflow, error) {
	return newCreateIdentityWorkflow(input, w.owner)
}

type createIdentityWorkflow struct {
	input *pb.CreateIdentityWorkflowInput
	owner config.OwnerConfig
}

// newCreateIdentityWorkflow creates and validates a new createIdentityWorkflow
func newCreateIdentityWorkflow(input *pb.CreateIdentityWorkflowInput, owner config.OwnerConfig) (*createIdentityWorkflow, error) {
	w := &createIdentityWorkflow{input: input, owner: owner}
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
	// The identity's signing certificate always comes from a locally-owned CA
	if !fabricconfig.OwnsCertificateAuthority(w.owner, req.GetOrgIssuerRef()) {
		return fmt.Errorf("%w: %q does not own CA %s-%s (org_issuer_ref)",
			ErrNotOwned, w.owner.OwnerID(), req.GetOrgIssuerRef().GetCertType(), req.GetOrgIssuerRef().GetId())
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
	identityName := req.GetIdentityName()
	orgIssuer := req.GetOrgIssuerRef()
	tlsIssuer := req.GetTlsIssuerRef()
	signingStorage := req.GetSigningStorage()
	tlsStorage := req.GetTlsStorage()

	// The org issuer is locally owned (validated above)
	orgIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, orgIssuer)

	// Create org identity cert (uses identity_type for proper OU classification)
	signingPurpose := fmt.Sprintf("identity-%s-signing", identityName)
	_, err := pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		IssuerRef: orgIssuer,
		Profile:   "", // Default profile
		Subject:   req.GetSubject(),
		Storage:   signingStorage,
		AdminIdentity: &fabricpb.IdentityRef{
			KeyRef:  orgIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: orgIssuerConfig.VaultAdminSigningCertRef(),
		},
		Purpose:      signingPurpose,
		Attributes:   req.GetAttributes(),
		IdentityType: req.GetIdentityType(), // Pass through for NodeOUs classification
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	// Create TLS client cert
	// TLS certs don't need identity_type since NodeOUs only applies to org/signing certs
	tlsPurpose := fmt.Sprintf("identity-%s-tls", identityName)
	tlsReq := &pb.CreateCertificateInput{
		IssuerRef: tlsIssuer,
		Profile:   req.GetTlsProfile(),
		Subject: &pb.CertificateSubject{
			CommonName:         req.GetSubject().GetCommonName(),
			Organization:       req.GetSubject().GetOrganization(),
			OrganizationalUnit: req.GetSubject().GetOrganizationalUnit(),
			DnsNames:           req.GetSubject().GetDnsNames(),
		},
		Storage: tlsStorage,
		Purpose: tlsPurpose,
	}

	if fabricconfig.OwnsCertificateAuthority(w.owner, tlsIssuer) {
		// Local TLS issuer: direct child, authenticated as the issuer's admin
		tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, tlsIssuer)
		tlsReq.AdminIdentity = &fabricpb.IdentityRef{
			KeyRef:  tlsIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: tlsIssuerConfig.VaultAdminSigningCertRef(),
		}
		_, err = pb.CreateCertificateChild(ctx, tlsReq)
	} else {
		// Foreign TLS issuer: route via Nexus to the issuer's worker
		_, err = NexusCreateCertificate(ctx, tlsReq)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS identity: %w", err)
	}

	return &pb.CreateIdentityOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}

func (w *Workflows) CreateCertificate(ctx workflow.Context, input *pb.CreateCertificateWorkflowInput) (pb.CreateCertificateWorkflow, error) {
	return newCreateCertificateWorkflow(input, w.owner)
}

type createCertificateWorkflow struct {
	input *pb.CreateCertificateWorkflowInput
	owner config.OwnerConfig
}

// newCreateCertificateWorkflow creates and validates a new createCertificateWorkflow
func newCreateCertificateWorkflow(input *pb.CreateCertificateWorkflowInput, owner config.OwnerConfig) (*createCertificateWorkflow, error) {
	w := &createCertificateWorkflow{input: input, owner: owner}
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
	if req.GetStorage() == nil {
		return fmt.Errorf("%w: storage is required", ErrInvalidInput)
	}
	if req.GetStorage().GetKeyRef() == nil {
		return fmt.Errorf("%w: storage.key_ref is required", ErrInvalidInput)
	}
	if req.GetStorage().GetCertRef() == nil {
		return fmt.Errorf("%w: storage.cert_ref is required", ErrInvalidInput)
	}
	certType := req.GetIssuerRef().GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: issuer cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	// The referenced issuer is always executed against locally - it must be
	// owned by this worker's owner. Foreign issuance arrives here via Nexus
	// on the issuer's own worker.
	if !fabricconfig.OwnsCertificateAuthority(w.owner, req.GetIssuerRef()) {
		return fmt.Errorf("%w: %q does not own CA %s-%s (issuer_ref)",
			ErrNotOwned, w.owner.OwnerID(), req.GetIssuerRef().GetCertType(), req.GetIssuerRef().GetId())
	}
	return nil
}

func (w *createCertificateWorkflow) Execute(ctx workflow.Context) (*pb.CreateCertificateOutput, error) {
	req := w.input.Req
	issuerRef := req.GetIssuerRef()
	storage := req.GetStorage()

	// The issuer is locally owned (validated above); all issuer paths derive
	// from the injected owner
	caConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, issuerRef)

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

	// Keys and certs are stored at separate paths via storage IdentityRef.
	// The issuer's certificate ref (for chain building) is derived from the
	// issuer's CA config rather than passed by the caller.
	var enrollOutput pb.EnrollIdentityOutput
	err = workflow.ExecuteActivity(activityCtx, pb.EnrollIdentityActivityName, &pb.EnrollIdentityInput{
		CaRef:            issuerRef,
		IdentityName:     identityName,
		EnrollmentSecret: registerOutput.GetEnrollmentSecret(),
		Storage:          storage, // IdentityRef with key and cert refs
		Subject:          subject,
		CaTlsCertRef:     caConfig.VaultTLSCertRef(),     // CA's TLS cert for mTLS
		IssuerCertRef:    caConfig.VaultSigningCertRef(), // Issuer's cert path for chain building
		Profile:          req.GetProfile(),
	}).Get(ctx, &enrollOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity: %w", err)
	}

	return &pb.CreateCertificateOutput{
		Identity: storage,
	}, nil
}
