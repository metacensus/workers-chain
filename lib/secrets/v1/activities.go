package secrets

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	pb "github.com/metacensus/workers-chain/gen/secrets/v1"
)

// Sentinel errors
var (
	ErrEnsurePassword              = errors.New("ensure password")
	ErrEnsureKeyPair               = errors.New("ensure key pair")
	ErrCheckCertificate            = errors.New("check certificate")
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

// CheckCertificate reports whether the certificate at cert_ref can be kept: it
// certifies key_ref's key, names the common name and every host, and has at
// least min_remaining left. An absent or unparseable certificate is not
// current; a failure to fetch either secret is an error, so a transient fault
// never reads as "issue a new one".
func (a *Activities) CheckCertificate(ctx context.Context, req *pb.CheckCertificateInput) (*pb.CheckCertificateOutput, error) {
	if req.GetKeyRef() == nil || req.GetCertRef() == nil {
		return nil, fmt.Errorf("%w: key_ref and cert_ref are required", ErrCheckCertificate)
	}

	certSecret, err := a.provider.FetchSecret(ctx, req.GetCertRef())
	if errors.Is(err, ErrSecretNotFound) {
		return &pb.CheckCertificateOutput{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: fetch certificate: %w", ErrCheckCertificate, err)
	}
	keySecret, err := a.provider.FetchSecret(ctx, req.GetKeyRef())
	if err != nil {
		return nil, fmt.Errorf("%w: fetch key: %w", ErrCheckCertificate, err)
	}

	stored, err := certSecret.Certificate()
	if err != nil {
		return &pb.CheckCertificateOutput{}, nil
	}
	cert, err := parseCertificatePEM(stored.Certificate)
	if err != nil {
		return &pb.CheckCertificateOutput{}, nil
	}
	kp, err := keySecret.KeyPair()
	if err != nil {
		return nil, fmt.Errorf("%w: parse key: %w", ErrCheckCertificate, err)
	}
	publicKey, err := parsePublicKeyPEM(kp.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("%w: parse key: %w", ErrCheckCertificate, err)
	}

	current := publicKey.Equal(cert.PublicKey) &&
		cert.Subject.CommonName == req.GetCommonName() &&
		namesHosts(cert, req.GetHosts()) &&
		time.Until(cert.NotAfter) >= req.GetMinRemaining().AsDuration()
	return &pb.CheckCertificateOutput{Current: current}, nil
}

func parseCertificatePEM(data []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}

func parsePublicKeyPEM(data []byte) (interface{ Equal(crypto.PublicKey) bool }, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM block")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	equaler, ok := key.(interface{ Equal(crypto.PublicKey) bool })
	if !ok {
		return nil, fmt.Errorf("unsupported public key type %T", key)
	}
	return equaler, nil
}

// namesHosts reports whether every host is among the certificate's DNS or IP
// SANs.
func namesHosts(cert *x509.Certificate, hosts []string) bool {
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			if !slices.ContainsFunc(cert.IPAddresses, ip.Equal) {
				return false
			}
			continue
		}
		if !slices.Contains(cert.DNSNames, host) {
			return false
		}
	}
	return true
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
