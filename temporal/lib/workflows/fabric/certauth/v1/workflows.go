package certauth

import (
	"errors"
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/workflow"
)

// Sentinel errors for certificate authority workflows
var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrNotOwned          = errors.New("certificate authority not owned by this worker")
	ErrSignCertificate   = errors.New("sign certificate")
	ErrInvalidCertType   = errors.New("invalid cert type")
	ErrPrivilegedRequest = errors.New("privileged issuance is not available across the ownership boundary")
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
	if req.GetCommonName() == "" {
		return fmt.Errorf("%w: common_name is required", ErrInvalidInput)
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
	orgIssuer := req.GetOrgIssuerRef()
	tlsIssuer := req.GetTlsIssuerRef()
	signingStorage := req.GetSigningStorage()
	tlsStorage := req.GetTlsStorage()

	// Org signing cert: the org issuer is locally owned (validated above), so
	// IssueCertificate takes the local SignCsr branch. identity_type is passed
	// through for NodeOUs classification. adminIdentity is nil - the CA's
	// default permanent admin registers the identity.
	_, err := IssueCertificate(
		ctx,
		w.owner,
		orgIssuer,
		req.GetCommonName(),
		nil, // signing cert carries no SANs
		req.GetIdentityType(),
		req.GetAttributes(),
		"", // default profile
		signingStorage,
		nil, // default permanent admin
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signing identity: %w", err)
	}

	// TLS client cert: the TLS issuer may be local OR foreign (e.g. an org
	// obtaining a TLS cert from the operator's TLS chain). IssueCertificate
	// branches on ownership - local signs inline, foreign crosses via Nexus
	// with only the public CSR. TLS certs don't set identity_type (NodeOUs only
	// applies to org/signing certs).
	_, err = IssueCertificate(
		ctx,
		w.owner,
		tlsIssuer,
		req.GetCommonName(),
		req.GetHosts(),
		"",  // no identity type for TLS
		nil, // no attributes
		req.GetTlsProfile(),
		tlsStorage,
		nil, // default permanent admin (local branch only; not forwarded over Nexus)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS identity: %w", err)
	}

	return &pb.CreateIdentityOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}
