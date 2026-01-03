package secrets

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
)

// Sentinel errors
var (
	ErrCreatePassword              = errors.New("create password")
	ErrCreateKeyPair               = errors.New("create key pair")
	ErrCreateSelfSignedCertificate = errors.New("create self-signed certificate")
)

// Activities implements the SecretsActivities interface
type Activities struct {
	provider SecretProvider
}

// NewActivities creates a new Activities instance
func NewActivities(provider SecretProvider) *Activities {
	return &Activities{provider: provider}
}

// CreatePassword generates a password and stores it with the given username
func (a *Activities) CreatePassword(ctx context.Context, req *pb.CreatePasswordInput) (*pb.CreatePasswordOutput, error) {
	ref := req.GetSecretRef()
	if ref == nil {
		return nil, fmt.Errorf("%w: secret_ref is required", ErrCreatePassword)
	}

	password, err := GeneratePassword()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreatePassword, err)
	}

	creds := &Credentials{
		Username: req.GetUsername(),
		Password: password,
	}
	if err := a.provider.StoreCredentials(ctx, ref, creds); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreatePassword, err)
	}

	return &pb.CreatePasswordOutput{SecretRef: ref}, nil
}

// CreateKeyPair generates an ECDSA key pair and stores it
func (a *Activities) CreateKeyPair(ctx context.Context, req *pb.CreateKeyPairInput) (*pb.CreateKeyPairOutput, error) {
	ref := req.GetSecretRef()
	if ref == nil {
		return nil, fmt.Errorf("%w: secret_ref is required", ErrCreateKeyPair)
	}

	kp, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateKeyPair, err)
	}

	if err := a.provider.StoreKeyPair(ctx, ref, kp); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateKeyPair, err)
	}

	return &pb.CreateKeyPairOutput{SecretRef: ref}, nil
}

// CreateSelfSignedCertificate fetches a key pair, generates a self-signed certificate, and stores it
func (a *Activities) CreateSelfSignedCertificate(ctx context.Context, req *pb.CreateSelfSignedCertificateInput) (*pb.CreateSelfSignedCertificateOutput, error) {
	keyRef := req.GetKeyRef()
	certRef := req.GetCertRef()

	if keyRef == nil {
		return nil, fmt.Errorf("%w: key_ref is required", ErrCreateSelfSignedCertificate)
	}
	if certRef == nil {
		return nil, fmt.Errorf("%w: cert_ref is required", ErrCreateSelfSignedCertificate)
	}

	kp, err := a.provider.FetchKeyPair(ctx, keyRef)
	if err != nil {
		return nil, fmt.Errorf("%w: fetch key pair: %w", ErrCreateSelfSignedCertificate, err)
	}

	params := &CertificateParams{
		CommonName:         req.GetCommonName(),
		Organization:       req.GetOrganization(),
		OrganizationalUnit: req.GetOrganizationalUnit(),
		ValidityDuration:   req.GetValidityDuration().AsDuration(),
		MaxPathLength:      req.GetMaxPathLength(),
		DNSNames:           req.GetDnsNames(),
	}
	cert, err := GenerateSelfSignedCertificate(kp.PrivateKey, params)
	if err != nil {
		return nil, fmt.Errorf("%w: generate certificate: %w", ErrCreateSelfSignedCertificate, err)
	}

	if err := a.provider.StoreCertificate(ctx, certRef, cert); err != nil {
		return nil, fmt.Errorf("%w: store certificate: %w", ErrCreateSelfSignedCertificate, err)
	}

	return &pb.CreateSelfSignedCertificateOutput{CertRef: certRef}, nil
}

// StoreCertificate stores certificate data at the specified location
func (a *Activities) StoreCertificate(ctx context.Context, req *pb.StoreCertificateInput) (*pb.StoreCertificateOutput, error) {
	certRef := req.GetCertRef()
	if certRef == nil {
		return nil, fmt.Errorf("cert_ref is required")
	}

	cert := &Certificate{
		Certificate: req.GetCertificatePem(),
	}

	if err := a.provider.StoreCertificate(ctx, certRef, cert); err != nil {
		return nil, fmt.Errorf("failed to store certificate: %w", err)
	}

	return &pb.StoreCertificateOutput{CertRef: certRef}, nil
}

// StoreIdentityCredentials stores keypair+certificate together at one location
// This is used for storing admin identity credentials where the CA client needs both
func (a *Activities) StoreIdentityCredentials(ctx context.Context, req *pb.StoreIdentityCredentialsInput) (*pb.StoreIdentityCredentialsOutput, error) {
	credentialsRef := req.GetCredentialsRef()
	if credentialsRef == nil {
		return nil, fmt.Errorf("credentials_ref is required")
	}

	// Store as IdentityCredentials which includes both keypair and certificate
	creds := &IdentityCredentials{
		PrivateKey:  req.GetPrivateKeyPem(),
		PublicKey:   req.GetPublicKeyPem(),
		Certificate: req.GetCertificatePem(),
		CAChain:     req.GetCaChainPem(),
	}

	if err := a.provider.StoreIdentityCredentials(ctx, credentialsRef, creds); err != nil {
		return nil, fmt.Errorf("failed to store identity credentials: %w", err)
	}

	return &pb.StoreIdentityCredentialsOutput{CredentialsRef: credentialsRef}, nil
}
