package certauth

import (
	"fmt"
	"time"

	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/certauth/v1"
	fabricpb "github.com/metacensus/workers-chain/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/workers-chain/gen/workflows/kube/v1"
	"github.com/metacensus/workers-chain/lib/workers/config"
	fabricconfig "github.com/metacensus/workers-chain/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// TLSEnrollmentCommonName returns the CN an identity presents when enrolling for
// its TLS certificate, disambiguating it against the CA's identity registry.
//
// Fabric CA keys its registry by enrollment ID (= CN here), so two identities that
// enroll at the same CA under the same CN clobber one record. Two collision domains
// need handling:
//
//   - A **foreign** (cross-owner) TLS issuer — the operator's shared TLS CA, at which
//     every org's identities enroll — is namespaced by appending the owner as the
//     final "@"-segment (certificate-authority-org-int@org1). This is the #4 fix.
//   - A **local** issuer that signs both this identity's signing and TLS certs gets
//     the historical "-tls" suffix so the two enrollments don't collide.
//
// Only the TLS enrollment CN is affected; the signing CN (the MSP identity) is built
// separately and is never routed through here. Exported so deployment orchestration
// (fabric/v1) derives the same names.
func TLSEnrollmentCommonName(owner config.OwnerConfig, baseName string, signingIssuerRef, tlsIssuerRef *fabricpb.CertificateAuthorityRef) string {
	// Self-signed: not registering with a CA, so no collision domain.
	if tlsIssuerRef == nil {
		return baseName
	}
	// Foreign TLS issuer (shared operator CA): owner-qualify to keep trust domains
	// that share that CA from colliding in its registry.
	if !fabricconfig.OwnsCertificateAuthority(owner, tlsIssuerRef) {
		return baseName + "@" + owner.OwnerID()
	}
	// Local issuer signing both certs: disambiguate the two enrollments.
	if signingIssuerRef != nil &&
		signingIssuerRef.GetCertType() == tlsIssuerRef.GetCertType() &&
		signingIssuerRef.GetId() == tlsIssuerRef.GetId() {
		return baseName + "-tls"
	}
	return baseName
}

func (w *Workflows) CreateBootstrapAdmin(ctx workflow.Context, input *pb.CreateBootstrapAdminWorkflowInput) (pb.CreateBootstrapAdminWorkflow, error) {
	return newCreateBootstrapAdminWorkflow(input, w.owner)
}

type createBootstrapAdminWorkflow struct {
	input *pb.CreateBootstrapAdminWorkflowInput
	owner config.OwnerConfig
}

// newCreateBootstrapAdminWorkflow creates and validates a new createBootstrapAdminWorkflow
func newCreateBootstrapAdminWorkflow(input *pb.CreateBootstrapAdminWorkflowInput, owner config.OwnerConfig) (*createBootstrapAdminWorkflow, error) {
	w := &createBootstrapAdminWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createBootstrapAdminWorkflow) validate() error {
	caRef := w.input.Req.GetCaRef()
	if caRef == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	// The bootstrap admin's CA is always owned by the executing worker
	if !fabricconfig.OwnsCertificateAuthority(w.owner, caRef) {
		return fmt.Errorf("%w: %q does not own CA %s-%s (ca_ref)",
			ErrNotOwned, w.owner.OwnerID(), caRef.GetCertType(), caRef.GetId())
	}
	return nil
}

func (w *createBootstrapAdminWorkflow) Execute(ctx workflow.Context) (*pb.CreateBootstrapAdminOutput, error) {
	caRef := w.input.Req.GetCaRef()
	caConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, caRef)

	// Wait for CA to be ready
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

	// Bootstrap-enroll the signing identity from this CA. There is no registrar
	// yet, so BootstrapIssueCertificate runs the register-less CSR pipeline
	// (EnsureKeyPair -> GenerateCsr -> BootstrapEnrollCsr -> StoreCertificate)
	// using the CA's preset bootstrap credentials. The private key stays local.
	signingStorage := &fabricpb.IdentityRef{
		KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
		CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
	}
	_, err = BootstrapIssueCertificate(
		ctx,
		caConfig.URL(),
		caConfig.VaultBootstrapAdminCredentialsRef(),
		// Register-less bootstrap enrollment authenticates as the CA's preset
		// bootstrap registrar, so the CSR CN must equal that enrollment ID (the
		// username) or Fabric CA rejects it (error 77).
		caConfig.BootstrapAdminUsername(),
		nil, // signing cert carries no SANs
		"",  // default profile for admin identity
		signingStorage,
		caConfig.VaultSigningCertRef(), // CA's signing cert for chain building
		caConfig.VaultTLSCertRef(),     // CA's TLS cert for mTLS
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap enroll signing identity: %w", err)
	}

	// Obtain the TLS client identity from the appropriate issuer.
	// Use explicit tls_issuer from input - if nil, use the CA itself (TLS Root CA case)
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	if tlsIssuerRef == nil {
		tlsIssuerRef = caRef
	}

	tlsStorage := &fabricpb.IdentityRef{
		KeyRef:  caConfig.VaultBootstrapAdminTLSKeyRef(),
		CertRef: caConfig.VaultBootstrapAdminTLSCertRef(),
	}
	// CommonName is set per-branch: the register-less local branch must use the
	// issuer's bootstrap enrollment ID, while the foreign branch is registered
	// under a normal per-identity CN.
	if fabricconfig.OwnsCertificateAuthority(w.owner, tlsIssuerRef) {
		// Local TLS issuer: bootstrap-enroll directly against it using its
		// bootstrap credentials (the issuer may not have a permanent admin
		// yet, e.g. a TLS CA enrolling its own bootstrap admin), so SignCsr
		// (which needs a registrar) is not usable here.
		tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, tlsIssuerRef)
		// Register-less bootstrap enrollment authenticates as the issuer's preset
		// bootstrap registrar, so the CSR CN must equal that enrollment ID (error
		// 77 otherwise). The -tls collision suffix does not apply: bootstrap
		// enroll registers nothing, and signing/TLS certs live at distinct Vault
		// paths.
		_, err = BootstrapIssueCertificate(
			ctx,
			tlsIssuerConfig.URL(),
			tlsIssuerConfig.VaultBootstrapAdminCredentialsRef(),
			tlsIssuerConfig.BootstrapAdminUsername(),
			nil,   // TLS SANs not set for the bootstrap admin client cert
			"tls", // TLS profile for client cert
			tlsStorage,
			tlsIssuerConfig.VaultSigningCertRef(), // TLS issuer's signing cert for chain building
			tlsIssuerConfig.VaultTLSCertRef(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to bootstrap enroll TLS identity: %w", err)
		}
	} else {
		// Foreign TLS issuer (org CA bootstrap admin getting TLS from the
		// operator's TLS chain): use the CSR flow. The issuer's permanent admin
		// registers the identity under this CN, so a normal per-identity CN is
		// used (with the -tls suffix to avoid a same-CA registration collision).
		// The private key is generated and kept in this worker's own subtree;
		// only a public CSR crosses to the issuer's worker, which signs with its
		// own permanent admin. This worker has no access to the foreign issuer's
		// Vault space.
		tlsCommonName := TLSEnrollmentCommonName(w.owner, caConfig.BootstrapAdminCommonName(), caRef, tlsIssuerRef)
		_, err = IssueCertificate(ctx, w.owner, tlsIssuerRef, tlsCommonName, nil, "", nil, "tls", tlsStorage, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS identity via CSR flow: %w", err)
		}
	}

	return &pb.CreateBootstrapAdminOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}

func (w *Workflows) CreateAdmin(ctx workflow.Context, input *pb.CreateAdminWorkflowInput) (pb.CreateAdminWorkflow, error) {
	return newCreateAdminWorkflow(input, w.owner)
}

type createAdminWorkflow struct {
	input *pb.CreateAdminWorkflowInput
	owner config.OwnerConfig
}

// newCreateAdminWorkflow creates and validates a new createAdminWorkflow
func newCreateAdminWorkflow(input *pb.CreateAdminWorkflowInput, owner config.OwnerConfig) (*createAdminWorkflow, error) {
	w := &createAdminWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *createAdminWorkflow) validate() error {
	caRef := w.input.Req.GetCaRef()
	if caRef == nil {
		return fmt.Errorf("%w: ca_ref is required", ErrInvalidInput)
	}
	// The admin's CA is always owned by the executing worker
	if !fabricconfig.OwnsCertificateAuthority(w.owner, caRef) {
		return fmt.Errorf("%w: %q does not own CA %s-%s (ca_ref)",
			ErrNotOwned, w.owner.OwnerID(), caRef.GetCertType(), caRef.GetId())
	}
	// Signing issuance never crosses the ownership boundary
	if signingIssuer := w.input.Req.GetSigningIssuer(); signingIssuer != nil && signingIssuer.GetCertType() != "" {
		if !fabricconfig.OwnsCertificateAuthority(w.owner, signingIssuer) {
			return fmt.Errorf("%w: %q does not own CA %s-%s (signing_issuer)",
				ErrNotOwned, w.owner.OwnerID(), signingIssuer.GetCertType(), signingIssuer.GetId())
		}
	}
	return nil
}

func (w *createAdminWorkflow) Execute(ctx workflow.Context) (*pb.CreateAdminOutput, error) {
	caRef := w.input.Req.GetCaRef()
	caConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, caRef)

	// Use explicit signing issuer from input - if nil, use the CA itself
	// (Per Fabric CA constraints, admin must be in CA's database)
	signingIssuerRef := w.input.Req.GetSigningIssuer()
	if signingIssuerRef == nil || signingIssuerRef.GetCertType() == "" {
		signingIssuerRef = caRef // Default: signing from this CA itself
	}
	// Signing issuance never crosses the ownership boundary (validated above)
	signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, signingIssuerRef)

	// Determine which registrar signs the CSR with the signing issuer:
	// - If signing issuer is this CA itself → its bootstrap admin (the only
	//   registrar that exists yet; this is how the bootstrap admin registers
	//   the permanent admin).
	// - If signing issuer is a different CA → that CA's permanent admin.
	var signingAdminIdentity *fabricpb.IdentityRef
	if signingIssuerRef.GetCertType() == caRef.GetCertType() && signingIssuerRef.GetId() == caRef.GetId() {
		signingAdminIdentity = &fabricpb.IdentityRef{
			KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
			CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
		}
	} else {
		signingAdminIdentity = &fabricpb.IdentityRef{
			KeyRef:  signingIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: signingIssuerConfig.VaultAdminSigningCertRef(),
		}
	}

	signingStorage := &fabricpb.IdentityRef{
		KeyRef:  caConfig.VaultAdminSigningKeyRef(),
		CertRef: caConfig.VaultAdminSigningCertRef(),
	}
	// The signing issuer is locally owned (validated above), so IssueCertificate
	// takes the local SignCsr branch, authenticating as signingAdminIdentity.
	_, err := IssueCertificate(
		ctx,
		w.owner,
		signingIssuerRef,
		caConfig.AdminCommonName(),
		nil, // signing cert carries no SANs
		"",  // default identity type (admin is a client-type identity)
		map[string]string{
			"hf.Registrar.Roles":         "*",
			"hf.Registrar.DelegateRoles": "*",
			"hf.Registrar.Attributes":    "*",
			"hf.Revoker":                 "true",
			"hf.IntermediateCA":          "true",
			"hf.GenCRL":                  "true",
			"hf.AffiliationMgr":          "true",
		},
		"", // default profile for admin identity
		signingStorage,
		signingAdminIdentity,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin signing identity: %w", err)
	}

	// Use explicit TLS issuer from input - if nil, use the CA itself
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	if tlsIssuerRef == nil || tlsIssuerRef.GetCertType() == "" {
		tlsIssuerRef = caRef // Default: TLS from this CA itself
	}

	// Compute the TLS enrollment CN: owner-qualified at a foreign (shared) issuer,
	// "-tls"-suffixed when one local CA signs both certs.
	tlsCommonName := TLSEnrollmentCommonName(w.owner, caConfig.AdminCommonName(), signingIssuerRef, tlsIssuerRef)

	tlsStorage := &fabricpb.IdentityRef{
		KeyRef:  caConfig.VaultAdminTLSKeyRef(),
		CertRef: caConfig.VaultAdminTLSCertRef(),
	}

	// Determine the registrar for a LOCAL TLS issuer. Ignored for a foreign
	// issuer (which signs as its own permanent admin and never receives a
	// caller-chosen registrar), so it stays nil in that case.
	var tlsAdminIdentity *fabricpb.IdentityRef
	if fabricconfig.OwnsCertificateAuthority(w.owner, tlsIssuerRef) {
		tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, tlsIssuerRef)
		if tlsIssuerRef.GetCertType() == caRef.GetCertType() && tlsIssuerRef.GetId() == caRef.GetId() {
			// TLS issuer is this CA itself → register with its bootstrap admin.
			tlsAdminIdentity = &fabricpb.IdentityRef{
				KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
				CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
			}
		} else {
			// Different local CA → register with its permanent admin.
			tlsAdminIdentity = &fabricpb.IdentityRef{
				KeyRef:  tlsIssuerConfig.VaultAdminSigningKeyRef(),
				CertRef: tlsIssuerConfig.VaultAdminSigningCertRef(),
			}
		}
	}

	// IssueCertificate branches on ownership: local signs inline with
	// tlsAdminIdentity as the registrar; foreign crosses via Nexus with only
	// the public CSR (tlsAdminIdentity is not forwarded).
	_, err = IssueCertificate(ctx, w.owner, tlsIssuerRef, tlsCommonName, nil, "", nil, "tls", tlsStorage, tlsAdminIdentity)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin TLS identity: %w", err)
	}

	return &pb.CreateAdminOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}
