package certauth

import (
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/workflow"
)

// SignCertificate is the operator-side (issuer) entrypoint of the cross-owner
// CSR flow. It runs on the worker that OWNS the issuing CA: a public CSR comes
// in via Nexus, the CA signs it, and only the public certificate + chain go
// back out. The issuer never sees the caller's private key and never touches
// the caller's Vault subtree.
func (w *Workflows) SignCertificate(ctx workflow.Context, input *pb.SignCertificateWorkflowInput) (pb.SignCertificateWorkflow, error) {
	return newSignCertificateWorkflow(input, w.owner)
}

type signCertificateWorkflow struct {
	input *pb.SignCertificateWorkflowInput
	owner config.OwnerConfig
}

// newSignCertificateWorkflow creates and validates a new signCertificateWorkflow
func newSignCertificateWorkflow(input *pb.SignCertificateWorkflowInput, owner config.OwnerConfig) (*signCertificateWorkflow, error) {
	w := &signCertificateWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid and that the issuing CA is
// owned by this worker's owner (this is the ownership boundary: a caller cannot
// coerce us into signing with a CA we do not own).
func (w *signCertificateWorkflow) validate() error {
	req := w.input.Req
	if req.GetIssuerRef() == nil {
		return fmt.Errorf("%w: issuer_ref is required", ErrInvalidInput)
	}
	if req.GetCommonName() == "" {
		return fmt.Errorf("%w: common_name is required", ErrInvalidInput)
	}
	if len(req.GetCsrPem()) == 0 {
		return fmt.Errorf("%w: csr_pem is required", ErrInvalidInput)
	}
	certType := req.GetIssuerRef().GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: issuer cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	// The issuer is always executed against locally - it must be owned by this
	// worker's owner. Foreign issuance arrives here via Nexus on the issuer's
	// own worker.
	if !fabricconfig.OwnsCertificateAuthority(w.owner, req.GetIssuerRef()) {
		return fmt.Errorf("%w: %q does not own CA %s-%s (issuer_ref)",
			ErrNotOwned, w.owner.OwnerID(), req.GetIssuerRef().GetCertType(), req.GetIssuerRef().GetId())
	}
	return nil
}

func (w *signCertificateWorkflow) Execute(ctx workflow.Context) (*pb.SignCertificateOutput, error) {
	req := w.input.Req
	issuerRef := req.GetIssuerRef()

	identityType := req.GetIdentityType()
	if identityType == "" {
		identityType = "client"
	}

	activityCtx := workflow.WithActivityOptions(ctx, issuanceActivityOptions())

	// The issuer is locally owned (validated above), so we sign inline via the
	// shared local-sign helper. adminIdentity is always nil here: a Nexus caller
	// never chooses the registrar, so the issuer signs as its own default
	// permanent admin.
	signOutput, err := signCsrLocal(ctx, activityCtx, w.owner, issuerRef,
		req.GetCommonName(), identityType, req.GetAttributes(),
		req.GetCsrPem(), req.GetHosts(), req.GetProfile(), nil)
	if err != nil {
		return nil, err
	}

	return &pb.SignCertificateOutput{
		CertificatePem: signOutput.GetCertificatePem(),
		ChainPem:       signOutput.GetChainPem(),
	}, nil
}
