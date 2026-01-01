package crypto

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
)

// Sentinel errors for crypto activities
var (
	ErrGeneratePassword     = errors.New("generate password")
	ErrGenerateKeyPair      = errors.New("generate key pair")
	ErrGenerateCertificate  = errors.New("generate certificate")
)

// Activities implements the CryptoServiceActivities interface from generated code
type Activities struct{}

// NewActivities creates a new Activities instance
func NewActivities() *Activities {
	return &Activities{}
}

// GeneratePassword generates a secure random password
// Generates 32 random bytes and base64-encodes them (URL-safe)
func (a *Activities) GeneratePassword(ctx context.Context, req *pb.GeneratePasswordInput) (*pb.GeneratePasswordOutput, error) {
	// Generate 24 random bytes (will become 32 base64 characters)
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("%w: failed to generate random bytes: %w", ErrGeneratePassword, err)
	}

	// Encode as URL-safe base64 (no padding to get exactly 32 chars)
	password := base64.RawURLEncoding.EncodeToString(randomBytes)

	return &pb.GeneratePasswordOutput{
		Password: password,
	}, nil
}

// GenerateKeyPair generates a 256-bit ECDSA key pair
// Returns PEM-encoded private key and public key
func (a *Activities) GenerateKeyPair(ctx context.Context, req *pb.GenerateKeyPairInput) (*pb.GenerateKeyPairOutput, error) {
	// Generate ECDSA private key using P-256 curve (256-bit)
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate key pair: %w", ErrGenerateKeyPair, err)
	}

	// Marshal private key to PKCS8 format
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal private key: %w", ErrGenerateKeyPair, err)
	}

	// Encode private key to PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Marshal public key to PKIX format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal public key: %w", ErrGenerateKeyPair, err)
	}

	// Encode public key to PEM
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return &pb.GenerateKeyPairOutput{
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
	}, nil
}

// GenerateCertificate generates a self-signed X.509 certificate
// Used for root CA certificates. Certificate is signed using the provided private key.
func (a *Activities) GenerateCertificate(ctx context.Context, req *pb.GenerateCertificateInput) (*pb.GenerateCertificateOutput, error) {
	// Parse the PEM-encoded private key
	block, _ := pem.Decode(req.GetPrivateKey())
	if block == nil {
		return nil, fmt.Errorf("%w: failed to decode PEM block", ErrGenerateCertificate)
	}

	// Parse the PKCS8 private key
	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse private key: %w", ErrGenerateCertificate, err)
	}

	// Assert that it's an ECDSA private key
	privateKey, ok := privateKeyInterface.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w: private key is not ECDSA", ErrGenerateCertificate)
	}

	// Generate a random serial number (128-bit)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate serial number: %w", ErrGenerateCertificate, err)
	}

	// Calculate validity period
	notBefore := time.Now()
	validityDuration := req.GetValidityDuration().AsDuration()
	notAfter := notBefore.Add(validityDuration)

	// Calculate Subject Key Identifier (hash of public key)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal public key: %w", ErrGenerateCertificate, err)
	}
	subjectKeyID := sha256.Sum256(publicKeyBytes)

	// Create certificate template
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         req.GetCommonName(),
			Organization:       []string{req.GetOrganization()},
			OrganizationalUnit: []string{req.GetOrganizationalUnit()},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            int(req.GetMaxPathLength()),
		MaxPathLenZero:        req.GetMaxPathLength() == 0,
		SubjectKeyId:          subjectKeyID[:],
	}

	// Create the self-signed certificate (parent = template)
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create certificate: %w", ErrGenerateCertificate, err)
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	return &pb.GenerateCertificateOutput{
		Certificate: certPEM,
	}, nil
}
