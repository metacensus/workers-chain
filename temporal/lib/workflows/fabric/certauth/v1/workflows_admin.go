package certauth

import (
	"fmt"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ComputeTLSCommonName returns the CN for a TLS certificate, adding "-tls" suffix
// when both signing and TLS certs come from the same CA to avoid registration collision.
// Exported so deployment orchestration (fabric/v1) can derive the same identity names.
func ComputeTLSCommonName(baseName string, signingIssuerRef, tlsIssuerRef *fabricpb.CertificateAuthorityRef) string {
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
	signingStorage := &fabricpb.IdentityRef{
		KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
		CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
	}
	err = workflow.ExecuteActivity(activityCtx, pb.BootstrapEnrollIdentityActivityName, &pb.BootstrapEnrollIdentityInput{
		CaUrl:                   caConfig.URL(),
		BootstrapCredentialsRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Storage:                 signingStorage,
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.BootstrapAdminCommonName(),
			Organization:       w.owner.OwnerID(),
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		CaTlsCertRef:  caConfig.VaultTLSCertRef(),
		IssuerCertRef: caConfig.VaultSigningCertRef(), // CA's signing cert for chain building
		Profile:       "",                             // Default profile for admin identity
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap enroll signing identity: %w", err)
	}

	// Obtain the TLS client identity from the appropriate issuer.
	// Use explicit tls_issuer from input - if nil, use the CA itself (TLS Root CA case)
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	if tlsIssuerRef == nil {
		tlsIssuerRef = caRef
	}

	// Compute TLS common name - signing identity always comes from this CA (bootstrap enrollment)
	// For collision detection, signing issuer is effectively this CA (caRef)
	tlsCommonName := ComputeTLSCommonName(
		caConfig.BootstrapAdminCommonName(),
		caRef, // signing from this CA
		tlsIssuerRef,
	)

	tlsStorage := &fabricpb.IdentityRef{
		KeyRef:  caConfig.VaultBootstrapAdminTLSKeyRef(),
		CertRef: caConfig.VaultBootstrapAdminTLSCertRef(),
	}

	if fabricconfig.OwnsCertificateAuthority(w.owner, tlsIssuerRef) {
		// Local TLS issuer: bootstrap-enroll directly against it using its
		// bootstrap credentials (the issuer may not have a permanent admin
		// yet, e.g. a TLS CA enrolling its own bootstrap admin).
		tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, tlsIssuerRef)

		// Create keypair for TLS client cert (keys/ path)
		err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
			SecretRef: caConfig.VaultBootstrapAdminTLSKeyRef(),
		}).Get(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to generate TLS key pair: %w", err)
		}

		err = workflow.ExecuteActivity(activityCtx, pb.BootstrapEnrollIdentityActivityName, &pb.BootstrapEnrollIdentityInput{
			CaUrl:                   tlsIssuerConfig.URL(),
			BootstrapCredentialsRef: tlsIssuerConfig.VaultBootstrapAdminCredentialsRef(),
			Storage:                 tlsStorage,
			Subject: &pb.CertificateSubject{
				CommonName:         tlsCommonName,
				Organization:       w.owner.OwnerID(),
				OrganizationalUnit: caConfig.OrganizationalUnit(),
			},
			CaTlsCertRef:  tlsIssuerConfig.VaultTLSCertRef(),
			IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(), // TLS issuer's signing cert for chain building
			Profile:       "tls",                                 // TLS profile for client cert
		}).Get(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to bootstrap enroll TLS identity: %w", err)
		}
	} else {
		// Foreign TLS issuer (org CA bootstrap admin getting TLS from the
		// operator's TLS chain): route via Nexus so issuance runs on the
		// issuer's worker with its own credentials. This worker has no
		// access to the foreign issuer's bootstrap credentials.
		_, err = NexusCreateCertificate(ctx, &pb.CreateCertificateInput{
			IssuerRef: tlsIssuerRef,
			Profile:   "tls",
			Subject: &pb.CertificateSubject{
				CommonName:         tlsCommonName,
				Organization:       w.owner.OwnerID(),
				OrganizationalUnit: caConfig.OrganizationalUnit(),
			},
			Storage: tlsStorage,
			Purpose: fmt.Sprintf("%s-%s-bootstrap-admin-tls", caRef.GetCertType(), caRef.GetId()),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS identity via nexus: %w", err)
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
	certType := caRef.GetCertType()
	caID := caRef.GetId()
	caConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, caRef)

	// Use explicit signing issuer from input - if nil, use the CA itself
	// (Per Fabric CA constraints, admin must be in CA's database)
	signingIssuerRef := w.input.Req.GetSigningIssuer()
	if signingIssuerRef == nil || signingIssuerRef.GetCertType() == "" {
		signingIssuerRef = caRef // Default: signing from this CA itself
	}
	// Signing issuance never crosses the ownership boundary (validated above)
	signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, signingIssuerRef)
	signingPurpose := fmt.Sprintf("%s-%s-admin-signing", certType, caID)

	// Determine which credentials to use for registration with the signing issuer:
	// - If signing issuer is this CA itself → use this CA's bootstrap admin
	// - If signing issuer is a different CA → use that CA's permanent admin
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
	_, err := pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		IssuerRef: signingIssuerRef,
		Profile:   "", // Default profile for admin identity
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.AdminCommonName(),
			Organization:       w.owner.OwnerID(),
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		Storage:       signingStorage,
		AdminIdentity: signingAdminIdentity,
		Purpose:       signingPurpose,
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
	tlsPurpose := fmt.Sprintf("%s-%s-admin-tls", certType, caID)

	// Compute TLS common name - add "-tls" suffix if same CA issues both to avoid collision
	tlsCommonName := ComputeTLSCommonName(caConfig.AdminCommonName(), signingIssuerRef, tlsIssuerRef)

	tlsStorage := &fabricpb.IdentityRef{
		KeyRef:  caConfig.VaultAdminTLSKeyRef(),
		CertRef: caConfig.VaultAdminTLSCertRef(),
	}

	tlsReq := &pb.CreateCertificateInput{
		IssuerRef: tlsIssuerRef,
		Profile:   "tls",
		Subject: &pb.CertificateSubject{
			CommonName:         tlsCommonName, // Different name if same CA issues both
			Organization:       w.owner.OwnerID(),
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		Storage: tlsStorage,
		Purpose: tlsPurpose,
	}

	if fabricconfig.OwnsCertificateAuthority(w.owner, tlsIssuerRef) {
		// Local TLS issuer: direct child with the appropriate admin identity:
		// - If TLS issuer is this CA itself → use this CA's bootstrap admin
		// - If TLS issuer is a different local CA → use that CA's permanent admin
		tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, tlsIssuerRef)
		if tlsIssuerRef.GetCertType() == caRef.GetCertType() && tlsIssuerRef.GetId() == caRef.GetId() {
			tlsReq.AdminIdentity = &fabricpb.IdentityRef{
				KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
				CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
			}
		} else {
			tlsReq.AdminIdentity = &fabricpb.IdentityRef{
				KeyRef:  tlsIssuerConfig.VaultAdminSigningKeyRef(),
				CertRef: tlsIssuerConfig.VaultAdminSigningCertRef(),
			}
		}
		_, err = pb.CreateCertificateChild(ctx, tlsReq)
	} else {
		// Foreign TLS issuer (org CA admin getting TLS from the operator's
		// TLS chain): route via Nexus to the issuer's worker
		_, err = NexusCreateCertificate(ctx, tlsReq)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create admin TLS identity: %w", err)
	}

	return &pb.CreateAdminOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}
