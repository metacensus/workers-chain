package fabric

import (
	"fmt"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (w *Workflows) BootstrapTlsCertificateAuthorities(ctx workflow.Context, input *pb.BootstrapTlsCertificateAuthoritiesWorkflowInput) (pb.BootstrapTlsCertificateAuthoritiesWorkflow, error) {
	return newBootstrapTlsCertificateAuthoritiesWorkflow(input)
}

type bootstrapTlsCertificateAuthoritiesWorkflow struct {
	input *pb.BootstrapTlsCertificateAuthoritiesWorkflowInput
}

func newBootstrapTlsCertificateAuthoritiesWorkflow(input *pb.BootstrapTlsCertificateAuthoritiesWorkflowInput) (*bootstrapTlsCertificateAuthoritiesWorkflow, error) {
	w := &bootstrapTlsCertificateAuthoritiesWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *bootstrapTlsCertificateAuthoritiesWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	return nil
}

func (w *bootstrapTlsCertificateAuthoritiesWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapTlsCertificateAuthoritiesOutput, error) {
	org := w.input.Req.GetOrg()

	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{
		Org: org,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	if _, err := pb.DeployTlsRootCertificateAuthorityChild(ctx, &pb.DeployTlsRootCertificateAuthorityInput{
		Org:              org,
		ValidityDuration: w.input.Req.GetTlsRootValidity(),
	}); err != nil {
		return nil, fmt.Errorf("deploy TLS Root CA: %w", err)
	}

	if _, err := pb.DeployTlsIntermediateCertificateAuthorityChild(ctx, &pb.DeployTlsIntermediateCertificateAuthorityInput{
		Org:              org,
		ValidityDuration: w.input.Req.GetTlsIntermediateValidity(),
	}); err != nil {
		return nil, fmt.Errorf("deploy TLS Intermediate CA: %w", err)
	}

	return &pb.BootstrapTlsCertificateAuthoritiesOutput{}, nil
}

func (w *Workflows) DeployTlsRootCertificateAuthority(ctx workflow.Context, input *pb.DeployTlsRootCertificateAuthorityWorkflowInput) (pb.DeployTlsRootCertificateAuthorityWorkflow, error) {
	return newDeployTlsRootCertificateAuthorityWorkflow(input)
}

type deployTlsRootCertificateAuthorityWorkflow struct {
	input *pb.DeployTlsRootCertificateAuthorityWorkflowInput
}

func newDeployTlsRootCertificateAuthorityWorkflow(input *pb.DeployTlsRootCertificateAuthorityWorkflowInput) (*deployTlsRootCertificateAuthorityWorkflow, error) {
	w := &deployTlsRootCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *deployTlsRootCertificateAuthorityWorkflow) validate() error {
	if w.input == nil || w.input.Req == nil {
		return fmt.Errorf("%w: input must not be nil", ErrInvalidInput)
	}
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org must not be empty", ErrInvalidInput)
	}
	return nil
}

func (w *deployTlsRootCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployTlsRootCertificateAuthorityOutput, error) {
	org := w.input.Req.GetOrg()
	operatorConfig := config.GetOperator()
	validityDuration := w.input.Req.GetValidityDuration().AsDuration()

	caRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, caRef)

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

	_, err := NexusSecretsCreateAccess(ctx, caConfig)
	if err != nil {
		return nil, err
	}

	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		Org:   org,
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: caConfig.VaultSigningKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing key pair: %w", err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
		KeyRef:             caConfig.VaultSigningKeyRef(),
		CertRef:            caConfig.VaultSigningCertRef(),
		CommonName:         caConfig.ComponentName(),
		Organization:       org,
		OrganizationalUnit: caConfig.OrganizationalUnit(),
		ValidityDuration:   durationpb.New(validityDuration),
		MaxPathLength:      1,
		DnsNames:           nil, // No SANs for CA signing certs
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed signing certificate: %w", err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateKeyPairActivityName, &secretspb.CreateKeyPairInput{
		SecretRef: caConfig.VaultTLSKeyRef(),
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS key pair: %w", err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreateSelfSignedCertificateActivityName, &secretspb.CreateSelfSignedCertificateInput{
		KeyRef:             caConfig.VaultTLSKeyRef(),
		CertRef:            caConfig.VaultTLSCertRef(),
		CommonName:         caConfig.ComponentName(),
		Organization:       org,
		OrganizationalUnit: caConfig.OrganizationalUnit(),
		ValidityDuration:   durationpb.New(validityDuration),
		MaxPathLength:      0,                      // Not a CA cert
		DnsNames:           caConfig.TLSDNSNames(), // SANs required for TLS
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed TLS certificate: %w", err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: caConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	_, err = pb.CreateBootstrapAdminChild(ctx, &pb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        caRef, // Admin TLS from self (TLS Root)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	_, err = pb.CreateAdminChild(ctx, &pb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef, // Admin signing from self
		TlsIssuer:        caRef, // Admin TLS from self
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	return &pb.DeployTlsRootCertificateAuthorityOutput{}, nil
}

func (w *Workflows) DeployTlsIntermediateCertificateAuthority(ctx workflow.Context, input *pb.DeployTlsIntermediateCertificateAuthorityWorkflowInput) (pb.DeployTlsIntermediateCertificateAuthorityWorkflow, error) {
	return newDeployTlsIntermediateCertificateAuthorityWorkflow(input)
}

type deployTlsIntermediateCertificateAuthorityWorkflow struct {
	input *pb.DeployTlsIntermediateCertificateAuthorityWorkflowInput
}

func newDeployTlsIntermediateCertificateAuthorityWorkflow(input *pb.DeployTlsIntermediateCertificateAuthorityWorkflowInput) (*deployTlsIntermediateCertificateAuthorityWorkflow, error) {
	w := &deployTlsIntermediateCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *deployTlsIntermediateCertificateAuthorityWorkflow) validate() error {
	if w.input == nil || w.input.Req == nil {
		return fmt.Errorf("%w: input must not be nil", ErrInvalidInput)
	}
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org must not be empty", ErrInvalidInput)
	}
	return nil
}

func (w *deployTlsIntermediateCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployTlsIntermediateCertificateAuthorityOutput, error) {
	org := w.input.Req.GetOrg()
	operatorConfig := config.GetOperator()

	caRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, caRef)

	// Parent is always TLS Root (operator-owned, direct call)

	tlsIssuerRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	signingIssuerRef := tlsIssuerRef
	signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, tlsIssuerRef)

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

	_, err := NexusSecretsCreateAccess(ctx, caConfig)
	if err != nil {
		return nil, err
	}

	if _, err := pb.CreateCertificateDatabaseChild(ctx, &pb.CreateCertificateDatabaseInput{
		Org:   org,
		CaRef: caRef,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateCertificateDatabase, err)
	}

	signingPurpose := fmt.Sprintf("tls-int-signing")
	_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		Org:       org,
		IssuerRef: signingIssuerRef,
		Profile:   "ca",
		Subject: &pb.CertificateSubject{
			CommonName:         caConfig.ComponentName(),
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultSigningKeyRef(),
			CertRef: caConfig.VaultSigningCertRef(),
		},
		AdminIdentity: &pb.IdentityRef{
			KeyRef:  signingIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: signingIssuerConfig.VaultAdminSigningCertRef(),
		},
		Purpose: signingPurpose,
		Attributes: map[string]string{
			"hf.IntermediateCA": "true",
		},
		IssuerCertRef: signingIssuerConfig.VaultSigningCertRef(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w (signing): %w", ErrCreateCertificate, err)
	}

	tlsPurpose := fmt.Sprintf("tls-int-tls")
	tlsCommonName := computeTLSCommonName(caConfig.ComponentName(), signingIssuerRef, tlsIssuerRef)
	_, err = pb.CreateCertificateChild(ctx, &pb.CreateCertificateInput{
		Org:       org,
		IssuerRef: tlsIssuerRef,
		Profile:   "tls",
		Subject: &pb.CertificateSubject{
			CommonName:         tlsCommonName,
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			DnsNames:           caConfig.TLSDNSNames(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultTLSKeyRef(),
			CertRef: caConfig.VaultTLSCertRef(),
		},
		AdminIdentity: &pb.IdentityRef{
			KeyRef:  signingIssuerConfig.VaultAdminSigningKeyRef(),
			CertRef: signingIssuerConfig.VaultAdminSigningCertRef(),
		},
		Purpose:       tlsPurpose,
		IssuerCertRef: signingIssuerConfig.VaultSigningCertRef(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w (tls): %w", ErrCreateCertificate, err)
	}

	err = workflow.ExecuteActivity(activityCtx, secretspb.CreatePasswordActivityName, &secretspb.CreatePasswordInput{
		SecretRef: caConfig.VaultBootstrapAdminCredentialsRef(),
		Username:  "bootstrap-admin",
	}).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate bootstrap admin credentials: %w", err)
	}

	_, err = clusterpb.CreateDeploymentChild(ctx, &clusterpb.CreateDeploymentInput{
		Org:       org,
		Component: caConfig.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deploy CA: %w", err)
	}

	_, err = pb.CreateBootstrapAdminChild(ctx, &pb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        caRef, // Admin TLS from self (TLS Int)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	_, err = pb.CreateAdminChild(ctx, &pb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef, // Admin signing from self
		TlsIssuer:        caRef, // Admin TLS from self
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	return &pb.DeployTlsIntermediateCertificateAuthorityOutput{}, nil
}
