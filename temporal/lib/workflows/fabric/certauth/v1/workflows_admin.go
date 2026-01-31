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

// computeTLSCommonName returns the CN for a TLS certificate, adding "-tls" suffix
// when both signing and TLS certs come from the same CA to avoid registration collision.
func computeTLSCommonName(baseName string, signingIssuerRef, tlsIssuerRef *fabricpb.CertificateAuthorityRef) string {
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
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	caRef := w.input.Req.GetCaRef()
	certType := caRef.GetCertType()

	// Use operator config for TLS CAs (shared), org config for Org CAs (per-org)
	var ownerConfig config.OwnerConfig
	if certType == "tls" {
		ownerConfig = operatorConfig
	} else {
		ownerConfig = orgConfig
	}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, caRef)

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
	if tlsIssuerRef == nil {
		// No explicit TLS issuer - use the CA itself (TLS Root CA case)
		tlsIssuerRef = caRef
	}
	tlsIssuerConfig = fabricconfig.NewCertificateAuthorityConfig(operatorConfig, tlsIssuerRef)

	// Compute TLS common name - signing identity always comes from this CA (bootstrap enrollment)
	// For collision detection, signing issuer is effectively this CA (caRef)
	tlsCommonName := computeTLSCommonName(
		fmt.Sprintf("bootstrap-admin@%s", caConfig.ComponentName()),
		caRef, // signing from this CA
		tlsIssuerRef,
	)

	tlsStorage := &fabricpb.IdentityRef{
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

	return &pb.CreateBootstrapAdminOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}

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
	caConfig := fabricconfig.NewCertificateAuthorityConfig(ownerConfig, caRef)

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
	signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(signingIssuerOwner, signingIssuerRef)
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
	tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(tlsIssuerOwner, tlsIssuerRef)
	tlsPurpose := fmt.Sprintf("%s-%s-admin-tls", certType, caID)

	// Compute TLS common name - add "-tls" suffix if same CA issues both to avoid collision
	tlsCommonName := computeTLSCommonName(caConfig.AdminCommonName(), signingIssuerRef, tlsIssuerRef)

	// Determine which credentials to use for registration with the TLS issuer:
	// - If TLS issuer is this CA itself → use this CA's bootstrap admin
	// - If TLS issuer is a different CA → use that CA's permanent admin
	var tlsAdminIdentity *fabricpb.IdentityRef
	if tlsIssuerRef.GetCertType() == caRef.GetCertType() && tlsIssuerRef.GetId() == caRef.GetId() {
		tlsAdminIdentity = &fabricpb.IdentityRef{
			KeyRef:  caConfig.VaultBootstrapAdminSigningKeyRef(),
			CertRef: caConfig.VaultBootstrapAdminSigningCertRef(),
		}
	} else {
		tlsAdminIdentity = &fabricpb.IdentityRef{
			KeyRef:  tlsIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: tlsIssuerConfig.VaultAdminSigningCertRef(),
		}
	}

	tlsStorage := &fabricpb.IdentityRef{
		KeyRef:  caConfig.VaultAdminTLSKeyRef(),
		CertRef: caConfig.VaultAdminTLSCertRef(),
	}

	// Determine if we need Nexus or direct call for TLS certificate:
	// - If TLS issuer is in a different namespace → use Nexus
	// - If TLS issuer is in the same namespace → use direct call
	// TLS CAs are in operator namespace, Org CAs are in org namespace
	needsNexusForTls := certType == "org" && tlsIssuerRef.GetCertType() == "tls"

	if needsNexusForTls {
		// Cross-namespace: Org CA admin getting TLS from operator TLS CA - use Nexus
		NexusCreateCertificate(ctx, &pb.CreateCertificateInput{
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
		})
	} else {
		// Same namespace: use direct CreateCertificate with appropriate admin identity
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

	return &pb.CreateAdminOutput{
		SigningIdentity: signingStorage,
		TlsIdentity:     tlsStorage,
	}, nil
}
