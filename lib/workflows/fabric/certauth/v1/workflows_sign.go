package certauth

import (
	"fmt"
	"strings"

	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/certauth/v1"
	fabricpb "github.com/metacensus/workers-chain/gen/workflows/fabric/v1"
	"github.com/metacensus/workers-chain/lib/workers/config"
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
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
	// Intermediate CAs and registrars are minted against a locally-owned issuer, never via Nexus.
	if profile := req.GetProfile(); profile == profileCA {
		return fmt.Errorf("%w: profile %q", ErrPrivilegedRequest, profile)
	}
	for attribute := range req.GetAttributes() {
		if isPrivilegedFabricCAAttribute(attribute) {
			return fmt.Errorf("%w: attribute %q", ErrPrivilegedRequest, attribute)
		}
	}
	return validateSubjectScope(req.GetRequester(), req.GetCommonName(), req.GetHosts())
}

// validateSubjectScope keeps a caller's request inside its own trust domain: it
// may only ask for a subject that names itself.
//
// The requester is trustworthy because it is bound twice before we see it: the
// Temporal Authorizer refuses a request whose requester is not the org owning
// the per-org Nexus endpoint, and the endpoint allowlist admits only that org's
// namespace to that endpoint.
func validateSubjectScope(requester *fabricpb.Owner, commonName string, hosts []string) error {
	org := requester.GetOrgId()
	if org == "" {
		// Cross-owner issuance exists for orgs obtaining certificates from the
		// operator's shared chain; nothing else has a reason to arrive here.
		return fmt.Errorf("%w: requester.org_id is required", ErrInvalidInput)
	}
	orgCfg := config.NewOrgConfig(org)

	if subjectOrg(commonName) != org {
		return fmt.Errorf("%w: common_name %q is not qualified with requester org %q",
			ErrSubjectOutOfScope, commonName, org)
	}
	for _, host := range hosts {
		if !hostInNamespace(host, orgCfg.KubernetesNamespace()) {
			return fmt.Errorf("%w: host %q is outside namespace %q",
				ErrSubjectOutOfScope, host, orgCfg.KubernetesNamespace())
		}
	}
	return nil
}

// subjectOrg returns the owner a CN is qualified with: the final "@"-segment,
// as built by TLSEnrollmentCommonName and the component CommonName helpers
// ("peer0@org1", "certificate-authority-org-int@org1").
func subjectOrg(commonName string) string {
	idx := strings.LastIndex(commonName, "@")
	if idx < 0 {
		return ""
	}
	return commonName[idx+1:]
}

// hostInNamespace reports whether a requested SAN belongs to the given
// Kubernetes namespace. Service DNS names carry the namespace as their second
// label ("peer0.fabric-shard-org1[.svc[.cluster.local]]"); a single-label name
// resolves only inside the requester's own namespace, so it names nothing else.
func hostInNamespace(host, namespace string) bool {
	labels := strings.Split(host, ".")
	if len(labels) == 1 {
		return true
	}
	return labels[1] == namespace
}

const profileCA = "ca"

func isPrivilegedFabricCAAttribute(attribute string) bool {
	switch attribute {
	case "hf.IntermediateCA", "hf.Revoker", "hf.GenCRL", "hf.AffiliationMgr":
		return true
	}
	return strings.HasPrefix(attribute, "hf.Registrar.")
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
