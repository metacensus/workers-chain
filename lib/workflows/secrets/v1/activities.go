package secrets

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/metacensus/workers-chain/gen/workflows/secrets/v1"
)

// Sentinel errors
var (
	ErrEnsurePassword              = errors.New("ensure password")
	ErrEnsureKeyPair               = errors.New("ensure key pair")
	ErrCreateSelfSignedCertificate = errors.New("create self-signed certificate")
	ErrStoreGenesisBlock           = errors.New("store genesis block")
)

// Activities implements the SecretsActivities interface
type Activities struct {
	provider SecretProvider
}

// NewActivities creates a new Activities instance
func NewActivities(provider SecretProvider) *Activities {
	return &Activities{provider: provider}
}

// EnsurePassword get-or-creates credentials at secret_ref.
//
// It fetches the existing secret first and reuses it when present; a new
// password is generated and stored ONLY when the backend explicitly reports the
// secret is absent (ErrSecretNotFound). A transient fetch failure is fatal: a
// possibly-valid password must never be overwritten, because rotating the stored
// value would desync it from a consumer already provisioned from the old one
// (e.g. a PostgreSQL role whose CreateUser is a no-op when the role exists).
func (a *Activities) EnsurePassword(ctx context.Context, req *pb.EnsurePasswordInput) (*pb.EnsurePasswordOutput, error) {
	ref := req.GetSecretRef()
	if ref == nil {
		return nil, fmt.Errorf("%w: secret_ref is required", ErrEnsurePassword)
	}

	_, err := a.provider.FetchSecret(ctx, ref)
	switch {
	case err == nil:
		// Password already exists: reuse it, do not overwrite.
		return &pb.EnsurePasswordOutput{SecretRef: ref}, nil
	case errors.Is(err, ErrSecretNotFound):
		// Absent: generate and persist fresh credentials.
		password, err := GeneratePassword()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrEnsurePassword, err)
		}
		secret := FromCredentials(&Credentials{
			Username: req.GetUsername(),
			Password: password,
		})
		if err := a.provider.UpdateSecret(ctx, ref, secret); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrEnsurePassword, err)
		}
		return &pb.EnsurePasswordOutput{SecretRef: ref}, nil
	default:
		// Transient/permission error: do NOT touch a possibly-valid password.
		return nil, fmt.Errorf("%w: fetch existing password: %w", ErrEnsurePassword, err)
	}
}

// EnsureKeyPair get-or-creates an ECDSA key pair at secret_ref.
//
// It fetches the existing key first and reuses it when present; a new key pair
// is generated and stored ONLY when the backend explicitly reports the secret
// is absent (ErrSecretNotFound). A transient fetch failure is fatal: a
// possibly-valid key must never be overwritten, because regenerating it would
// orphan an already-signed certificate and reopen the key/cert-mismatch window
// that get-or-create is meant to close.
func (a *Activities) EnsureKeyPair(ctx context.Context, req *pb.EnsureKeyPairInput) (*pb.EnsureKeyPairOutput, error) {
	ref := req.GetSecretRef()
	if ref == nil {
		return nil, fmt.Errorf("%w: secret_ref is required", ErrEnsureKeyPair)
	}

	_, err := a.provider.FetchSecret(ctx, ref)
	switch {
	case err == nil:
		// Key already exists: reuse it, do not overwrite.
		return &pb.EnsureKeyPairOutput{SecretRef: ref}, nil
	case errors.Is(err, ErrSecretNotFound):
		// Absent: generate and persist a fresh key pair.
		kp, err := GenerateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrEnsureKeyPair, err)
		}
		if err := a.provider.UpdateSecret(ctx, ref, FromKeyPair(kp)); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrEnsureKeyPair, err)
		}
		return &pb.EnsureKeyPairOutput{SecretRef: ref}, nil
	default:
		// Transient/permission error: do NOT touch a possibly-valid key.
		return nil, fmt.Errorf("%w: fetch existing key: %w", ErrEnsureKeyPair, err)
	}
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

	secret, err := a.provider.FetchSecret(ctx, keyRef)
	if err != nil {
		return nil, fmt.Errorf("%w: fetch secret: %w", ErrCreateSelfSignedCertificate, err)
	}

	kp, err := secret.KeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: parse key pair: %w", ErrCreateSelfSignedCertificate, err)
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

	certSecret := FromCertificate(cert)
	if err := a.provider.UpdateSecret(ctx, certRef, certSecret); err != nil {
		return nil, fmt.Errorf("%w: store certificate: %w", ErrCreateSelfSignedCertificate, err)
	}

	return &pb.CreateSelfSignedCertificateOutput{CertRef: certRef}, nil
}

// StoreGenesisBlock stores a Fabric channel genesis block in the secret backend.
// Stub: fails loudly until the storage logic lands.
func (a *Activities) StoreGenesisBlock(_ context.Context, _ *pb.StoreGenesisBlockInput) (*pb.StoreGenesisBlockOutput, error) {
	return nil, fmt.Errorf("store genesis block: %w", ErrStoreGenesisBlock)
}

// StoreCertificate stores certificate data at the specified location.
// Persists both the leaf certificate and the full chain (ca_chain_pem), mirroring
// how EnrollIdentity stores {certificate, chain} so downstream consumers (e.g. mTLS
// server.crt) get the complete chain. Falls back to the leaf when no chain is given.
func (a *Activities) StoreCertificate(ctx context.Context, req *pb.StoreCertificateInput) (*pb.StoreCertificateOutput, error) {
	certRef := req.GetCertRef()
	if certRef == nil {
		return nil, fmt.Errorf("cert_ref is required")
	}

	chain := req.GetCaChainPem()
	if len(chain) == 0 {
		// No explicit chain: store the leaf as the chain (root/self-contained case)
		chain = req.GetCertificatePem()
	}

	cert := &Certificate{
		Certificate: req.GetCertificatePem(),
		Chain:       chain,
	}

	secret := FromCertificate(cert)
	if err := a.provider.UpdateSecret(ctx, certRef, secret); err != nil {
		return nil, fmt.Errorf("failed to store certificate: %w", err)
	}

	return &pb.StoreCertificateOutput{CertRef: certRef}, nil
}
