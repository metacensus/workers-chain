package certauth

import (
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

// issuanceActivityOptions is the standard retry/timeout policy for the
// CSR-pipeline activities (EnsureKeyPair, GenerateCsr, SignCsr,
// BootstrapEnrollCsr, StoreCertificate).
func issuanceActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
}

// IssueCertificate issues a certificate through the unified CSR pipeline:
//
//  1. EnsureKeyPair (secrets, get-or-create) at storage.key_ref
//  2. GenerateCsr (certauth) from that key + subject
//  3. Sign the CSR: the local SignCsr activity when the issuer is owned by
//     `owner`, otherwise the SignCertificate workflow via Nexus on the issuer's
//     own worker
//  4. StoreCertificate (secrets) at storage.cert_ref
//
// The local/foreign branch is the OWNERSHIP.md P3 structural check
// OwnsCertificateAuthority(owner, issuerRef): the caller's resolved owner
// decides, not an input flag. It supersedes the former split between a local
// CreateCertificate child workflow and a foreign IssueForeignCertificate.
//
// The private key never leaves `owner`: only a public CSR is produced, and for
// a foreign issuer only that CSR crosses the boundary (the signed cert + chain
// come back). This is a plain function (not a shared workflow) per OWNERSHIP
// P3 - cross-owner reuse composes into each owner's own workflows rather than a
// shared workflow that would blur the ownership boundary.
//
// adminIdentity is the registrar for the LOCAL branch only (nil = the issuer's
// default permanent admin). It is intentionally NOT forwarded across Nexus: a
// foreign issuer always signs as its own permanent admin, so no caller-chosen
// registrar ever crosses the ownership boundary.
func IssueCertificate(
	ctx workflow.Context,
	owner config.OwnerConfig,
	issuerRef *fabricpb.CertificateAuthorityRef,
	commonName string,
	hosts []string,
	identityType string,
	attributes map[string]string,
	profile string,
	storage *fabricpb.IdentityRef,
	adminIdentity *fabricpb.IdentityRef,
) (*fabricpb.IdentityRef, error) {
	if issuerRef == nil {
		return nil, fmt.Errorf("%w: issuer ref is required", ErrInvalidInput)
	}
	if commonName == "" {
		return nil, fmt.Errorf("%w: common_name is required", ErrInvalidInput)
	}
	if storage == nil || storage.GetKeyRef() == nil || storage.GetCertRef() == nil {
		return nil, fmt.Errorf("%w: storage with key_ref and cert_ref is required", ErrInvalidInput)
	}

	activityCtx := workflow.WithActivityOptions(ctx, issuanceActivityOptions())

	csrPEM, err := ensureKeyAndCsr(ctx, activityCtx, storage.GetKeyRef(), commonName, hosts)
	if err != nil {
		return nil, err
	}

	// Default to "client" when unset (matches the OU classification Fabric CA
	// applies when no explicit type is registered).
	if identityType == "" {
		identityType = "client"
	}

	var certPEM, chainPEM []byte
	if fabricconfig.OwnsCertificateAuthority(owner, issuerRef) {
		// Local issuer (P3): sign inline with the SignCsr activity, using
		// adminIdentity as the registrar.
		signOutput, err := signCsrLocal(ctx, activityCtx, owner, issuerRef,
			commonName, identityType, attributes,
			csrPEM, hosts, profile, adminIdentity)
		if err != nil {
			return nil, err
		}
		certPEM, chainPEM = signOutput.GetCertificatePem(), signOutput.GetChainPem()
	} else {
		// Foreign issuer: cross the ownership boundary via Nexus. Only the
		// public CSR goes out; adminIdentity is deliberately not forwarded.
		signOutput, err := NexusSignCertificate(ctx, &pb.SignCertificateInput{
			IssuerRef:    issuerRef,
			Profile:      profile,
			CommonName:   commonName,
			IdentityType: identityType,
			Attributes:   attributes,
			CsrPem:       csrPEM,
			Hosts:        hosts,
		})
		if err != nil {
			return nil, err
		}
		certPEM, chainPEM = signOutput.GetCertificatePem(), signOutput.GetChainPem()
	}

	if err := storeCertificate(ctx, activityCtx, storage.GetCertRef(), certPEM, chainPEM); err != nil {
		return nil, err
	}
	return storage, nil
}

// BootstrapIssueCertificate issues a certificate during CA initialization,
// before any registrar (permanent admin) exists. It runs the same
// key/CSR/store pipeline as IssueCertificate but signs via the register-less
// BootstrapEnrollCsr activity, which authenticates with the CA's preset
// bootstrap credentials. The CA is always owned by the executing worker
// (bootstrap enrollment never crosses the ownership boundary), so there is no
// Nexus branch. The private key never leaves this worker; only the public CSR
// is submitted.
func BootstrapIssueCertificate(
	ctx workflow.Context,
	caURL string,
	bootstrapCredentialsRef *secretspb.SecretRef,
	commonName string,
	hosts []string,
	profile string,
	storage *fabricpb.IdentityRef,
	issuerCertRef *secretspb.SecretRef,
	caTLSCertRef *secretspb.SecretRef,
) (*fabricpb.IdentityRef, error) {
	if commonName == "" {
		return nil, fmt.Errorf("%w: common_name is required", ErrInvalidInput)
	}
	if storage == nil || storage.GetKeyRef() == nil || storage.GetCertRef() == nil {
		return nil, fmt.Errorf("%w: storage with key_ref and cert_ref is required", ErrInvalidInput)
	}

	activityCtx := workflow.WithActivityOptions(ctx, issuanceActivityOptions())

	csrPEM, err := ensureKeyAndCsr(ctx, activityCtx, storage.GetKeyRef(), commonName, hosts)
	if err != nil {
		return nil, err
	}

	var enrollOutput pb.BootstrapEnrollCsrOutput
	err = workflow.ExecuteActivity(activityCtx, pb.BootstrapEnrollCsrActivityName, &pb.BootstrapEnrollCsrInput{
		CaUrl:                   caURL,
		BootstrapCredentialsRef: bootstrapCredentialsRef,
		CsrPem:                  csrPEM,
		Hosts:                   hosts,
		Profile:                 profile,
		CaTlsCertRef:            caTLSCertRef,
		IssuerCertRef:           issuerCertRef,
	}).Get(ctx, &enrollOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap enroll CSR: %w", err)
	}

	if err := storeCertificate(ctx, activityCtx, storage.GetCertRef(), enrollOutput.GetCertificatePem(), enrollOutput.GetChainPem()); err != nil {
		return nil, err
	}
	return storage, nil
}

// signCsrLocal signs a CSR with a locally-owned issuer CA via the SignCsr
// activity. issuer_cert_ref and ca_tls_cert_ref are derived from the issuer's
// own CA config, so they point only into the issuer owner's own subtree.
// adminIdentity selects the registrar (nil = the issuer's default permanent
// admin). Shared by IssueCertificate's local branch and the issuer-side
// SignCertificate workflow so both build the SignCsr request identically.
func signCsrLocal(
	ctx workflow.Context,
	activityCtx workflow.Context,
	owner config.OwnerConfig,
	issuerRef *fabricpb.CertificateAuthorityRef,
	commonName string,
	identityType string,
	attributes map[string]string,
	csrPEM []byte,
	hosts []string,
	profile string,
	adminIdentity *fabricpb.IdentityRef,
) (*pb.SignCsrOutput, error) {
	caConfig := fabricconfig.NewCertificateAuthorityConfig(owner, issuerRef)
	var signOutput pb.SignCsrOutput
	err := workflow.ExecuteActivity(activityCtx, pb.SignCsrActivityName, &pb.SignCsrInput{
		CaRef:         issuerRef,
		IdentityName:  commonName,
		IdentityType:  identityType,
		Attributes:    attributes,
		CsrPem:        csrPEM,
		Hosts:         hosts,
		Profile:       profile,
		IssuerCertRef: caConfig.VaultSigningCertRef(), // issuer's own cert path for chain building
		CaTlsCertRef:  caConfig.VaultTLSCertRef(),     // CA's TLS cert for mTLS to the CA server
		AdminIdentity: adminIdentity,
	}).Get(ctx, &signOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to sign CSR: %w", err)
	}
	return &signOutput, nil
}

// ensureKeyAndCsr runs the caller-side head of the CSR pipeline: get-or-create
// the private key locally, then produce a public CSR bound to it. The private
// key never leaves the executing worker's subtree.
func ensureKeyAndCsr(
	ctx workflow.Context,
	activityCtx workflow.Context,
	keyRef *secretspb.SecretRef,
	commonName string,
	hosts []string,
) ([]byte, error) {
	if err := workflow.ExecuteActivity(activityCtx, secretspb.EnsureKeyPairActivityName, &secretspb.EnsureKeyPairInput{
		SecretRef: keyRef,
	}).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ensure key pair: %w", err)
	}

	var genOutput pb.GenerateCsrOutput
	if err := workflow.ExecuteActivity(activityCtx, pb.GenerateCsrActivityName, &pb.GenerateCsrInput{
		KeyRef:     keyRef,
		CommonName: commonName,
		Hosts:      hosts,
	}).Get(ctx, &genOutput); err != nil {
		return nil, fmt.Errorf("failed to generate CSR: %w", err)
	}
	return genOutput.GetCsrPem(), nil
}

// storeCertificate persists the returned leaf certificate and full chain in the
// caller's own certs/ subtree.
func storeCertificate(
	ctx workflow.Context,
	activityCtx workflow.Context,
	certRef *secretspb.SecretRef,
	certPEM []byte,
	chainPEM []byte,
) error {
	if err := workflow.ExecuteActivity(activityCtx, secretspb.StoreCertificateActivityName, &secretspb.StoreCertificateInput{
		CertRef:        certRef,
		CertificatePem: certPEM,
		CaChainPem:     chainPEM,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("failed to store certificate: %w", err)
	}
	return nil
}
