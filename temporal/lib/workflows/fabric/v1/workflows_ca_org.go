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

func (w *Workflows) BootstrapOrgCertificateAuthorities(ctx workflow.Context, input *pb.BootstrapOrgCertificateAuthoritiesWorkflowInput) (pb.BootstrapOrgCertificateAuthoritiesWorkflow, error) {
	return newBootstrapOrgCertificateAuthoritiesWorkflow(input)
}

type bootstrapOrgCertificateAuthoritiesWorkflow struct {
	input *pb.BootstrapOrgCertificateAuthoritiesWorkflowInput
}

func newBootstrapOrgCertificateAuthoritiesWorkflow(input *pb.BootstrapOrgCertificateAuthoritiesWorkflowInput) (*bootstrapOrgCertificateAuthoritiesWorkflow, error) {
	w := &bootstrapOrgCertificateAuthoritiesWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *bootstrapOrgCertificateAuthoritiesWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	return nil
}

func (w *bootstrapOrgCertificateAuthoritiesWorkflow) Execute(ctx workflow.Context) (*pb.BootstrapOrgCertificateAuthoritiesOutput, error) {
	org := w.input.Req.GetOrg()

	if _, err := pb.DeployCertificateStoreChild(ctx, &pb.DeployCertificateStoreInput{
		Org: org,
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDeployCertificateStore, err)
	}

	if _, err := pb.DeployOrgRootCertificateAuthorityChild(ctx, &pb.DeployOrgRootCertificateAuthorityInput{
		Org:              org,
		ValidityDuration: w.input.Req.GetOrgRootValidity(),
		TlsIssuer:        &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
	}); err != nil {
		return nil, fmt.Errorf("deploy Org Root CA: %w", err)
	}

	if _, err := pb.DeployOrgIntermediateCertificateAuthorityChild(ctx, &pb.DeployOrgIntermediateCertificateAuthorityInput{
		Org:              org,
		ValidityDuration: w.input.Req.GetOrgIntermediateValidity(),
		TlsIssuer:        &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
	}); err != nil {
		return nil, fmt.Errorf("deploy Org Intermediate CA: %w", err)
	}

	return &pb.BootstrapOrgCertificateAuthoritiesOutput{}, nil
}

func (w *Workflows) DeployCertificateStore(ctx workflow.Context, input *pb.DeployCertificateStoreWorkflowInput) (pb.DeployCertificateStoreWorkflow, error) {
	return newDeployCertificateStoreWorkflow(input)
}

type deployCertificateStoreWorkflow struct {
	input *pb.DeployCertificateStoreWorkflowInput
}

// newDeployCertificateStoreWorkflow creates and validates a new deployCertificateStoreWorkflow
func newDeployCertificateStoreWorkflow(input *pb.DeployCertificateStoreWorkflowInput) (*deployCertificateStoreWorkflow, error) {
	w := &deployCertificateStoreWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// validate checks that the workflow input is valid
func (w *deployCertificateStoreWorkflow) validate() error {
	if w.input.Req.GetOrg() == "" {
		return fmt.Errorf("%w: org is required", ErrInvalidInput)
	}
	return nil
}

func (w *deployCertificateStoreWorkflow) Execute(ctx workflow.Context) (*pb.DeployCertificateStoreOutput, error) {
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	certStore := fabricconfig.NewCertificateStoreConfig(orgConfig)

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
		Org:       org,
		Component: certStore.ComponentName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate store deployment: %w", err)
	}

	return &pb.DeployCertificateStoreOutput{}, nil
}

func (w *Workflows) DeployOrgRootCertificateAuthority(ctx workflow.Context, input *pb.DeployOrgRootCertificateAuthorityWorkflowInput) (pb.DeployOrgRootCertificateAuthorityWorkflow, error) {
	return newDeployOrgRootCertificateAuthorityWorkflow(input)
}

type deployOrgRootCertificateAuthorityWorkflow struct {
	input *pb.DeployOrgRootCertificateAuthorityWorkflowInput
}

func newDeployOrgRootCertificateAuthorityWorkflow(input *pb.DeployOrgRootCertificateAuthorityWorkflowInput) (*deployOrgRootCertificateAuthorityWorkflow, error) {
	w := &deployOrgRootCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *deployOrgRootCertificateAuthorityWorkflow) validate() error {
	if w.input == nil || w.input.Req == nil {
		return fmt.Errorf("%w: input must not be nil", ErrInvalidInput)
	}
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org must not be empty", ErrInvalidInput)
	}
	if req.GetTlsIssuer() == nil {
		return fmt.Errorf("%w: tls_issuer must not be nil", ErrInvalidInput)
	}
	return nil
}

func (w *deployOrgRootCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployOrgRootCertificateAuthorityOutput, error) {
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()
	validityDuration := w.input.Req.GetValidityDuration().AsDuration()

	caRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(orgConfig, caRef)

	// TLS issuer is always TLS Int (from input)
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, tlsIssuerRef)

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

	tlsPurpose := fmt.Sprintf("org-root-tls")
	tlsCommonName := certauth.ComputeTLSCommonName(caConfig.ComponentName(), nil, tlsIssuerRef)

	_, err = certauth.NexusCreateCertificate(ctx, &certauthpb.CreateCertificateInput{
		Org:       org,
		IssuerRef: tlsIssuerRef,
		Profile:   "tls",
		Subject: &certauthpb.CertificateSubject{
			CommonName:         tlsCommonName,
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			DnsNames:           caConfig.TLSDNSNames(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultTLSKeyRef(),
			CertRef: caConfig.VaultTLSCertRef(),
		},
		Purpose:       tlsPurpose,
		IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
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

	_, err = certauthpb.CreateBootstrapAdminChild(ctx, &certauthpb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        tlsIssuerRef, // Admin TLS from TLS Int (not self)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	_, err = certauthpb.CreateAdminChild(ctx, &certauthpb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef,        // Admin signing from self
		TlsIssuer:        tlsIssuerRef, // Admin TLS from TLS Int (not self)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	return &pb.DeployOrgRootCertificateAuthorityOutput{}, nil
}
func (w *Workflows) DeployOrgIntermediateCertificateAuthority(ctx workflow.Context, input *pb.DeployOrgIntermediateCertificateAuthorityWorkflowInput) (pb.DeployOrgIntermediateCertificateAuthorityWorkflow, error) {
	return newDeployOrgIntermediateCertificateAuthorityWorkflow(input)
}

type deployOrgIntermediateCertificateAuthorityWorkflow struct {
	input *pb.DeployOrgIntermediateCertificateAuthorityWorkflowInput
}

func newDeployOrgIntermediateCertificateAuthorityWorkflow(input *pb.DeployOrgIntermediateCertificateAuthorityWorkflowInput) (*deployOrgIntermediateCertificateAuthorityWorkflow, error) {
	w := &deployOrgIntermediateCertificateAuthorityWorkflow{input: input}
	if err := w.validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *deployOrgIntermediateCertificateAuthorityWorkflow) validate() error {
	if w.input == nil || w.input.Req == nil {
		return fmt.Errorf("%w: input must not be nil", ErrInvalidInput)
	}
	req := w.input.Req
	if req.GetOrg() == "" {
		return fmt.Errorf("%w: org must not be empty", ErrInvalidInput)
	}
	if req.GetTlsIssuer() == nil {
		return fmt.Errorf("%w: tls_issuer must not be nil", ErrInvalidInput)
	}
	return nil
}

func (w *deployOrgIntermediateCertificateAuthorityWorkflow) Execute(ctx workflow.Context) (*pb.DeployOrgIntermediateCertificateAuthorityOutput, error) {
	org := w.input.Req.GetOrg()
	orgConfig := config.NewOrgConfig(org)
	operatorConfig := config.GetOperator()

	caRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "int"}
	caConfig := fabricconfig.NewCertificateAuthorityConfig(orgConfig, caRef)

	// Signing issuer is always Org Root (org-owned, direct call)
	signingIssuerRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}
	signingIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(orgConfig, signingIssuerRef)

	// TLS issuer is always TLS Int (from input, via Nexus)
	tlsIssuerRef := w.input.Req.GetTlsIssuer()
	tlsIssuerConfig := fabricconfig.NewCertificateAuthorityConfig(operatorConfig, tlsIssuerRef)

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

	signingPurpose := fmt.Sprintf("org-int-signing")
	_, err = certauthpb.CreateCertificateChild(ctx, &certauthpb.CreateCertificateInput{
		Org:       org,
		IssuerRef: signingIssuerRef,
		Profile:   "ca",
		Subject: &certauthpb.CertificateSubject{
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

	tlsPurpose := fmt.Sprintf("org-int-tls")
	tlsCommonName := certauth.ComputeTLSCommonName(caConfig.ComponentName(), signingIssuerRef, tlsIssuerRef)

	_, err = certauth.NexusCreateCertificate(ctx, &certauthpb.CreateCertificateInput{
		Org:       org,
		IssuerRef: tlsIssuerRef,
		Profile:   "tls",
		Subject: &certauthpb.CertificateSubject{
			CommonName:         tlsCommonName,
			Organization:       org,
			OrganizationalUnit: caConfig.OrganizationalUnit(),
			DnsNames:           caConfig.TLSDNSNames(),
		},
		Storage: &pb.IdentityRef{
			KeyRef:  caConfig.VaultTLSKeyRef(),
			CertRef: caConfig.VaultTLSCertRef(),
		},
		Purpose:       tlsPurpose,
		IssuerCertRef: tlsIssuerConfig.VaultSigningCertRef(),
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

	_, err = certauthpb.CreateBootstrapAdminChild(ctx, &certauthpb.CreateBootstrapAdminInput{
		Org:              org,
		CaRef:            caRef,
		TlsIssuer:        tlsIssuerRef, // Admin TLS from TLS Int (not self)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateBootstrapAdmin, err)
	}

	_, err = certauthpb.CreateAdminChild(ctx, &certauthpb.CreateAdminInput{
		Org:              org,
		CaRef:            caRef,
		SigningIssuer:    caRef,        // Admin signing from self
		TlsIssuer:        tlsIssuerRef, // Admin TLS from TLS Int (not self)
		ValidityDuration: w.input.Req.GetValidityDuration(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateAdmin, err)
	}

	return &pb.DeployOrgIntermediateCertificateAuthorityOutput{}, nil
}
