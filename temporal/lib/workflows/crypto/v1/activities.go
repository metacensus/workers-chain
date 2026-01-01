package crypto

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	pb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
)

// Sentinel errors for crypto activities
var (
	ErrGeneratePassword = errors.New("generate password")
	ErrGenerateKeyPair  = errors.New("generate key pair")
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
