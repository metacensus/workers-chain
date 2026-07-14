package fabric

import (
	"fmt"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	certauthpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"github.com/metacensus/infra/temporal/lib/workflows/fabric/certauth/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

// ==================== DeployCertificateStore ====================

func (w *Workflows) DeployCertificateStore(ctx workflow.Context, input *pb.DeployCertificateStoreWorkflowInput) (pb.DeployCertificateStoreWorkflow, error) {
	return newDeployCertificateStoreWorkflow(input, w.owner)
}

type deployCertificateStoreWorkflow struct {
	input *pb.DeployCertificateStoreWorkflowInput
	owner config.OwnerConfig
}

// newDeployCertificateStoreWorkflow creates a new deployCertificateStoreWorkflow.
// The certificate store is fully derived from the worker's injected owner.
func newDeployCertificateStoreWorkflow(input *pb.DeployCertificateStoreWorkflowInput, owner config.OwnerConfig) (*deployCertificateStoreWorkflow, error) {
	return &deployCertificateStoreWorkflow{input: input, owner: owner}, nil
}

func (w *deployCertificateStoreWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateStoreOutput, error) {
	certStore := fabricconfig.NewCertificateStoreConfig(w.owner)

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

	_, err := NexusSecretsCreateAccess(ctx, certStore)
	if err != nil {
		return nil, err
	}

	credentialsRef := certStore.VaultAdminRef()
	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: credentialsRef,
		Username:  "admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate postgres credentials: %w", err)
	}

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Owner:     w.owner.OwnerProto(),
		Component: certStore.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate store deployment: %w", err)
	}

	return &pb.DeployCertificateStoreOutput{}, nil
}

// ==================== DeployCertificateAuthority ====================

func (w *Workflows) DeployCertificateAuthority(ctx workflow.Context, input *pb.DeployCertificateAuthorityWorkflowInput) (pb.DeployCertificateAuthorityWorkflow, error) {
	return newDeployCertificateAuthorityWorkflow(input, w.owner)
}

type deployCertificateAuthorityWorkflow struct {
	input *pb.DeployCertificateAuthorityWorkflowInput
	owner config.OwnerConfig
}

// newDeployCertificateAuthorityWorkflow creates and validates a new deployCertificateAuthorityWorkflow
func newDeployCertificateAuthorityWorkflow(input *pb.DeployCertificateAuthorityWorkflowInput, owner config.OwnerConfig) (*deployCertificateAuthorityWorkflow, error) {
	w := &deployCertificateAuthorityWorkflow{input: input, owner: owner}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid and that the spec respects
// the ownership rules:
//   - spec.self MUST be owned by this worker's owner
//   - spec.signing_issuer (if set) MUST be owned by this worker's owner
//     (signing issuance never crosses the ownership boundary)
//   - spec.tls_issuer may be foreign (TLS certs for org CAs come from the
//     operator's TLS chain via Nexus)
func (w *deployCertificateAuthorityWorkflow) validate() error {
	spec := w.input.Req.GetSpec()
	if spec == nil {
		return fmt.Errorf("%w: spec is required", ErrInvalidInput)
	}
	self := spec.GetSelf()
	if self == nil {
		return fmt.Errorf("%w: spec.self is required", ErrInvalidInput)
	}
	if err := validateCARef(self); err != nil {
		return err
	}
	if !fabricconfig.OwnsCertificateAuthority(w.owner, self) {
		return fmt.Errorf("%w: %q does not own CA %s-%s (spec.self)",
			ErrNotOwned, w.owner.OwnerID(), self.GetCertType(), self.GetId())
	}
	if signingIssuer := spec.GetSigningIssuer(); signingIssuer != nil {
		if err := validateCARef(signingIssuer); err != nil {
			return err
		}
		if !fabricconfig.OwnsCertificateAuthority(w.owner, signingIssuer) {
			return fmt.Errorf("%w: %q does not own CA %s-%s (spec.signing_issuer; signing issuance never crosses the ownership boundary)",
				ErrNotOwned, w.owner.OwnerID(), signingIssuer.GetCertType(), signingIssuer.GetId())
		}
	}
	if tlsIssuer := spec.GetTlsIssuer(); tlsIssuer != nil {
		if err := validateCARef(tlsIssuer); err != nil {
			return err
		}
	}
	return nil
}

func validateCARef(ref *pb.CertificateAuthorityRef) error {
	certType := ref.GetCertType()
	if certType != "tls" && certType != "org" {
		return fmt.Errorf("%w: cert_type must be 'tls' or 'org', got '%s'", ErrInvalidCertType, certType)
	}
	if ref.GetId() == "" {
		return fmt.Errorf("%w: ca id is required", ErrInvalidInput)
	}
	return nil
}

// Execute deploys one certificate authority per the spec. All four variants
// (tls/root, tls/int, org/root, org/int) share the same spine:
//
//  1. Provision Vault access for the CA service account
//  2. Create the CA's certificate database in the certificate store
//  3. Obtain the CA's signing identity (self-signed or from spec.signing_issuer)
//  4. Obtain the CA's TLS identity (self-signed, local, or foreign via Nexus)
//  5. Generate bootstrap admin credentials
//  6. Deploy the CA server
//  7. Enroll the bootstrap admin
//  8. Create the permanent admin
func (w *deployCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateAuthorityOutput, error) {
	spec := w.input.Req.GetSpec()
	caRef := spec.GetSelf()
	caConfig := fabricconfig.NewCertificateAuthorityConfig(w.owner, caRef)

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

	// Step 1: Vault access for the CA service account
	if _, err := NexusSecretsCreateAccess(ctx, caConfig); err != nil {
		return nil, err
	}

	// Step 2: dedicated certificate database
	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	// Step 3: signing identity
	if err := w.ensureSigningIdentity(ctx, activityCtx, caConfig, spec); err != nil {
		return nil, fmt.Errorf("%w (signing): %w", ErrCreateCertificate, err)
	}

	// Step 4: TLS identity
	if err := w.ensureTLSIdentity(ctx, activityCtx, caConfig, spec); err != nil {
		return nil, fmt.Errorf("%w (tls): %w", ErrCreateCertificate, err)
	}

	// Step 5: bootstrap admin credentials
	if err := workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Username:  caConfig.BootstrapAdminUsername(),
	}).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	// Step 6: deploy the CA server
	if _, err := clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Owner:     w.owner.OwnerProto(),
		Component: caConfig.ComponentName(),
	}); err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	// Admin TLS certificates come from this CA itself when the spec's TLS
	// issuer is nil or locally owned (the operator's TLS CAs issue their own
	// admins' TLS certs); org CA admins get TLS from the foreign issuer.
	adminTLSIssuer := caRef
	if tlsIssuer := spec.GetTlsIssuer(); tlsIssuer != nil && !fabricconfig.OwnsCertificateAuthority(w.owner, tlsIssuer) {
		adminTLSIssuer = tlsIssuer
	}

	// Step 7: bootstrap admin enrollment
	if _, err := certauthpb.CreateBootstrapAdminChild(ctx, &certauthpb.CreateBootstrapAdminInput{
		CaRef:            caRef,
		TlsIssuer:        adminTLSIssuer,
		ValidityDuration: spec.GetValidityDuration(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	// Step 8: permanent admin
	if _, err := certauthpb.CreateAdminChild(ctx, &certauthpb.CreateAdminInput{
		CaRef:            caRef,
		SigningIssuer:    caRef, // Admin signing from self (admin must be in this CA's database)
		TlsIssuer:        adminTLSIssuer,
		ValidityDuration: spec.GetValidityDuration(),
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	return &pb.DeployCertificateAuthorityOutput{}, nil
}

// ensureSigningIdentity obtains the CA's signing key and certificate.
// SigningIssuer == nil: root CA, ensure a keypair and self-sign.
// SigningIssuer != nil: intermediate CA, issued by a local (same-owner) CA
// through the unified CSR pipeline with the hf.IntermediateCA attribute.
func (w *deployCertificateAuthorityWorkflow) ensureSigningIdentity(
	ctx workflow.Context,
	activityCtx workflow.Context,
	caConfig fabricconfig.CertificateAuthorityConfig,
	spec *pb.CertificateAuthoritySpec,
) error {
	signingIssuerRef := spec.GetSigningIssuer()

	if signingIssuerRef == nil {
		// Root CA: self-signed signing certificate. EnsureKeyPair is
		// get-or-create so re-runs reuse the existing private key.
		if err := workflow.ExecuteActivity(activityCtx, secretspb.EnsureKeyPairActivityName, &secretspb.EnsureKeyPairInput{
			SecretRef: caConfig.VaultSigningKeyRef(),
		}).Get(ctx, nil); err != nil {
			return fmt.Errorf("failed to ensure signing key pair: %w", err)
		}

		if err := workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
			KeyRef:             caConfig.VaultSigningKeyRef(),
			CertRef:            caConfig.VaultSigningCertRef(),
			CommonName:         caConfig.ComponentName(),
			Organization:       w.owner.OwnerID(),
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			ValidityDuration:   durationpb.New(spec.GetValidityDuration().AsDuration()),
			MaxPathLength:      1,
			DnsNames:           nil, // No SANs for CA signing certs
		}).Get(ctx, nil); err != nil {
			return fmt.Errorf("failed to generate self-signed signing certificate: %w", err)
		}
		return nil
	}

	// Intermediate CA: signing certificate issued by the (local) signing issuer
	// through the unified CSR pipeline. The signing issuer is locally owned
	// (validated), so IssueCertificate signs inline authenticating as the
	// issuer's default permanent admin (nil). The "ca" profile plus the
	// hf.IntermediateCA attribute mark this as an intermediate CA certificate.
	if _, err := certauth.IssueCertificate(
		ctx,
		w.owner,
		signingIssuerRef,
		caConfig.ComponentName(),
		nil, // CA signing cert carries no SANs
		"",  // no identity type
		map[string]string{
			"hf.IntermediateCA": "true",
		},
		"ca",
		&pb.IdentityRef{
			KeyRef:  caConfig.VaultSigningKeyRef(),
			CertRef: caConfig.VaultSigningCertRef(),
		},
		nil, // default permanent admin of the signing issuer
	); err != nil {
		return err
	}
	return nil
}

// ensureTLSIdentity obtains the CA's TLS key and certificate.
// TlsIssuer == nil: self-signed (TLS Root CA only).
// TlsIssuer owned by this worker's owner: local SignCsr via IssueCertificate.
// TlsIssuer foreign: SignCertificate via Nexus to the issuer's worker
// (the operator's TLS chain issues TLS certs for org CAs).
func (w *deployCertificateAuthorityWorkflow) ensureTLSIdentity(
	ctx workflow.Context,
	activityCtx workflow.Context,
	caConfig fabricconfig.CertificateAuthorityConfig,
	spec *pb.CertificateAuthoritySpec,
) error {
	tlsIssuerRef := spec.GetTlsIssuer()

	if tlsIssuerRef == nil {
		// Self-signed TLS certificate (TLS Root CA). EnsureKeyPair is
		// get-or-create so re-runs reuse the existing private key.
		if err := workflow.ExecuteActivity(activityCtx, secretspb.EnsureKeyPairActivityName, &secretspb.EnsureKeyPairInput{
			SecretRef: caConfig.VaultTLSKeyRef(),
		}).Get(ctx, nil); err != nil {
			return fmt.Errorf("failed to ensure TLS key pair: %w", err)
		}

		if err := workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
			KeyRef:             caConfig.VaultTLSKeyRef(),
			CertRef:            caConfig.VaultTLSCertRef(),
			CommonName:         caConfig.ComponentName(),
			Organization:       w.owner.OwnerID(),
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			ValidityDuration:   durationpb.New(spec.GetValidityDuration().AsDuration()),
			MaxPathLength:      0,                      // Not a CA cert
			DnsNames:           caConfig.TLSDNSNames(), // SANs required for TLS
		}).Get(ctx, nil); err != nil {
			return fmt.Errorf("failed to generate self-signed TLS certificate: %w", err)
		}
		return nil
	}

	// Non-self-signed TLS certificate via the unified CSR pipeline.
	// IssueCertificate branches on ownership: a local TLS issuer signs inline
	// (authenticating as its default permanent admin, nil); a foreign TLS issuer
	// (the operator's TLS chain issuing for an org CA) is reached via Nexus with
	// only the public CSR crossing the boundary. Either way the CA's TLS private
	// key is generated and kept in this worker's own subtree.
	tlsCommonName := certauth.TLSEnrollmentCommonName(w.owner, caConfig.ComponentName(), spec.GetSigningIssuer(), tlsIssuerRef)
	if _, err := certauth.IssueCertificate(
		ctx,
		w.owner,
		tlsIssuerRef,
		tlsCommonName,
		caConfig.TLSDNSNames(),
		"",  // no identity type
		nil, // no attributes
		"tls",
		&pb.IdentityRef{
			KeyRef:  caConfig.VaultTLSKeyRef(),
			CertRef: caConfig.VaultTLSCertRef(),
		},
		nil, // default permanent admin (local branch only; not forwarded over Nexus)
	); err != nil {
		return err
	}
	return nil
}
