package crypto_test

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
	crypto "github.com/metacensus/infra/temporal/lib/workflows/crypto/v1"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestGeneratePasswordActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := crypto.NewActivities()
	env.RegisterActivity(activities.GeneratePassword)

	input := &pb.GeneratePasswordInput{}
	val, err := env.ExecuteActivity(activities.GeneratePassword, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result *pb.GeneratePasswordOutput
	if err := val.Get(&result); err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}

	// Verify password properties
	password := result.Password
	if len(password) != 32 {
		t.Errorf("password length: got %d, want 32", len(password))
	}

	// Verify it's valid base64 URL-safe encoding
	decoded, err := base64.RawURLEncoding.DecodeString(password)
	if err != nil {
		t.Errorf("password is not valid base64 URL-safe encoding: %v", err)
	}

	// Should decode to 24 bytes
	if len(decoded) != 24 {
		t.Errorf("decoded password length: got %d, want 24", len(decoded))
	}
}

func TestGeneratePasswordActivity_Uniqueness(t *testing.T) {
	// Test that multiple calls generate unique passwords
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := crypto.NewActivities()
	env.RegisterActivity(activities.GeneratePassword)

	passwords := make(map[string]bool)
	input := &pb.GeneratePasswordInput{}

	for i := 0; i < 10; i++ {
		val, err := env.ExecuteActivity(activities.GeneratePassword, input)
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}

		var result *pb.GeneratePasswordOutput
		if err := val.Get(&result); err != nil {
			t.Fatalf("unexpected error from Get on iteration %d: %v", i, err)
		}

		if passwords[result.Password] {
			t.Errorf("duplicate password generated: %s", result.Password)
		}
		passwords[result.Password] = true
	}

	// Should have generated 10 unique passwords
	if len(passwords) != 10 {
		t.Errorf("expected 10 unique passwords, got %d", len(passwords))
	}
}

func TestGenerateKeyPairActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := crypto.NewActivities()
	env.RegisterActivity(activities.GenerateKeyPair)

	input := &pb.GenerateKeyPairInput{}
	val, err := env.ExecuteActivity(activities.GenerateKeyPair, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result *pb.GenerateKeyPairOutput
	if err := val.Get(&result); err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}

	// Verify keys are PEM-encoded and valid
	if len(result.PrivateKey) == 0 {
		t.Fatal("private key is empty")
	}
	if len(result.PublicKey) == 0 {
		t.Fatal("public key is empty")
	}

	// Decode and verify private key
	privBlock, _ := pem.Decode(result.PrivateKey)
	if privBlock == nil {
		t.Fatal("failed to decode private key PEM")
	}
	if privBlock.Type != "PRIVATE KEY" {
		t.Errorf("private key PEM type: got %q, want %q", privBlock.Type, "PRIVATE KEY")
	}

	// Parse private key (PKCS8 format)
	parsedPrivKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	// Verify it's an ECDSA key
	ecdsaPrivKey, ok := parsedPrivKey.(*ecdsa.PrivateKey)
	if !ok {
		t.Fatalf("private key is not ECDSA, got type: %T", parsedPrivKey)
	}

	// Verify it uses P-256 curve (256-bit)
	if ecdsaPrivKey.Curve.Params().Name != "P-256" {
		t.Errorf("curve: got %s, want P-256", ecdsaPrivKey.Curve.Params().Name)
	}

	// Decode and verify public key
	pubBlock, _ := pem.Decode(result.PublicKey)
	if pubBlock == nil {
		t.Fatal("failed to decode public key PEM")
	}
	if pubBlock.Type != "PUBLIC KEY" {
		t.Errorf("public key PEM type: got %q, want %q", pubBlock.Type, "PUBLIC KEY")
	}

	// Parse public key (PKIX format)
	parsedPubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse public key: %v", err)
	}

	// Verify it's an ECDSA public key
	ecdsaPubKey, ok := parsedPubKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("public key is not ECDSA, got type: %T", parsedPubKey)
	}

	// Verify public key matches private key
	if !ecdsaPrivKey.PublicKey.Equal(ecdsaPubKey) {
		t.Error("public key does not match private key")
	}

	// Verify PEM format strings
	privateKeyStr := string(result.PrivateKey)
	publicKeyStr := string(result.PublicKey)

	if !strings.Contains(privateKeyStr, "BEGIN PRIVATE KEY") {
		t.Error("private key missing PEM header")
	}
	if !strings.Contains(privateKeyStr, "END PRIVATE KEY") {
		t.Error("private key missing PEM footer")
	}

	if !strings.Contains(publicKeyStr, "BEGIN PUBLIC KEY") {
		t.Error("public key missing PEM header")
	}
	if !strings.Contains(publicKeyStr, "END PUBLIC KEY") {
		t.Error("public key missing PEM footer")
	}
}

func TestGenerateKeyPairActivity_Uniqueness(t *testing.T) {
	// Test that multiple calls generate unique key pairs
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := crypto.NewActivities()
	env.RegisterActivity(activities.GenerateKeyPair)

	privateKeys := make(map[string]bool)
	input := &pb.GenerateKeyPairInput{}

	for i := 0; i < 5; i++ {
		val, err := env.ExecuteActivity(activities.GenerateKeyPair, input)
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}

		var result *pb.GenerateKeyPairOutput
		if err := val.Get(&result); err != nil {
			t.Fatalf("unexpected error from Get on iteration %d: %v", i, err)
		}

		keyStr := string(result.PrivateKey)
		if privateKeys[keyStr] {
			t.Errorf("duplicate private key generated at iteration %d", i)
		}
		privateKeys[keyStr] = true
	}

	// Should have generated 5 unique key pairs
	if len(privateKeys) != 5 {
		t.Errorf("expected 5 unique key pairs, got %d", len(privateKeys))
	}
}

func TestGenerateCertificateActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := crypto.NewActivities()
	env.RegisterActivity(activities.GenerateKeyPair)
	env.RegisterActivity(activities.GenerateCertificate)

	// First generate a key pair to use for signing
	keyPairInput := &pb.GenerateKeyPairInput{}
	keyPairVal, err := env.ExecuteActivity(activities.GenerateKeyPair, keyPairInput)
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	var keyPairResult *pb.GenerateKeyPairOutput
	if err := keyPairVal.Get(&keyPairResult); err != nil {
		t.Fatalf("failed to get key pair result: %v", err)
	}

	// Generate certificate
	input := &pb.GenerateCertificateInput{
		PrivateKey:         keyPairResult.PrivateKey,
		CommonName:         "org1-tls-root-ca",
		Organization:       "org1",
		OrganizationalUnit: "root-ca",
		ValidityDuration:   durationpb.New(365 * 24 * 60 * 60 * 1e9), // 1 year in nanoseconds
		MaxPathLength:      1,
	}

	val, err := env.ExecuteActivity(activities.GenerateCertificate, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result *pb.GenerateCertificateOutput
	if err := val.Get(&result); err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}

	// Verify certificate is PEM-encoded
	if len(result.Certificate) == 0 {
		t.Fatal("certificate is empty")
	}

	certStr := string(result.Certificate)
	if !strings.Contains(certStr, "BEGIN CERTIFICATE") {
		t.Error("certificate missing PEM header")
	}
	if !strings.Contains(certStr, "END CERTIFICATE") {
		t.Error("certificate missing PEM footer")
	}

	// Decode and verify certificate
	certBlock, _ := pem.Decode(result.Certificate)
	if certBlock == nil {
		t.Fatal("failed to decode certificate PEM")
	}
	if certBlock.Type != "CERTIFICATE" {
		t.Errorf("certificate PEM type: got %q, want %q", certBlock.Type, "CERTIFICATE")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	// Verify subject information
	if cert.Subject.CommonName != "org1-tls-root-ca" {
		t.Errorf("CommonName: got %q, want %q", cert.Subject.CommonName, "org1-tls-root-ca")
	}
	if len(cert.Subject.Organization) != 1 || cert.Subject.Organization[0] != "org1" {
		t.Errorf("Organization: got %v, want [org1]", cert.Subject.Organization)
	}
	if len(cert.Subject.OrganizationalUnit) != 1 || cert.Subject.OrganizationalUnit[0] != "root-ca" {
		t.Errorf("OrganizationalUnit: got %v, want [root-ca]", cert.Subject.OrganizationalUnit)
	}

	// Verify CA properties
	if !cert.IsCA {
		t.Error("certificate IsCA should be true")
	}
	if cert.MaxPathLen != 1 {
		t.Errorf("MaxPathLen: got %d, want 1", cert.MaxPathLen)
	}
	if !cert.BasicConstraintsValid {
		t.Error("BasicConstraintsValid should be true")
	}

	// Verify key usage
	if cert.KeyUsage&x509.KeyUsageCertSign == 0 {
		t.Error("certificate missing KeyUsageCertSign")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign == 0 {
		t.Error("certificate missing KeyUsageCRLSign")
	}

	// Verify it's self-signed (issuer == subject)
	if cert.Issuer.CommonName != cert.Subject.CommonName {
		t.Errorf("certificate is not self-signed: issuer %q != subject %q",
			cert.Issuer.CommonName, cert.Subject.CommonName)
	}

	// Verify serial number exists and is non-zero
	if cert.SerialNumber.Sign() == 0 {
		t.Error("certificate has zero serial number")
	}

	// Verify validity period
	if cert.NotBefore.IsZero() {
		t.Error("NotBefore is zero")
	}
	if cert.NotAfter.IsZero() {
		t.Error("NotAfter is zero")
	}
	if !cert.NotAfter.After(cert.NotBefore) {
		t.Error("NotAfter should be after NotBefore")
	}

	// Verify SubjectKeyId is set
	if len(cert.SubjectKeyId) == 0 {
		t.Error("SubjectKeyId is empty")
	}

	// Verify signature - self-signed cert should verify against itself
	if err := cert.CheckSignatureFrom(cert); err != nil {
		t.Errorf("certificate signature verification failed: %v", err)
	}
}

func TestGenerateCertificateActivity_InvalidPrivateKey(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := crypto.NewActivities()
	env.RegisterActivity(activities.GenerateCertificate)

	tests := []struct {
		name       string
		privateKey []byte
	}{
		{
			name:       "invalid PEM data",
			privateKey: []byte("not a valid PEM"),
		},
		{
			name:       "empty private key",
			privateKey: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &pb.GenerateCertificateInput{
				PrivateKey:         tt.privateKey,
				CommonName:         "test-ca",
				Organization:       "test",
				OrganizationalUnit: "test",
				ValidityDuration:   durationpb.New(365 * 24 * 60 * 60 * 1e9), // 1 year in nanoseconds
				MaxPathLength:      1,
			}

			_, err := env.ExecuteActivity(activities.GenerateCertificate, input)
			if err == nil {
				t.Error("expected error for invalid private key, got nil")
			}
		})
	}
}

func TestGenerateCertificateActivity_Uniqueness(t *testing.T) {
	// Test that multiple calls generate unique certificates (different serial numbers, timestamps)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	activities := crypto.NewActivities()
	env.RegisterActivity(activities.GenerateKeyPair)
	env.RegisterActivity(activities.GenerateCertificate)

	// Generate one key pair to reuse
	keyPairVal, err := env.ExecuteActivity(activities.GenerateKeyPair, &pb.GenerateKeyPairInput{})
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	var keyPairResult *pb.GenerateKeyPairOutput
	if err := keyPairVal.Get(&keyPairResult); err != nil {
		t.Fatalf("failed to get key pair result: %v", err)
	}

	certificates := make(map[string]bool)
	input := &pb.GenerateCertificateInput{
		PrivateKey:         keyPairResult.PrivateKey,
		CommonName:         "test-ca",
		Organization:       "test",
		OrganizationalUnit: "test",
		ValidityDuration:   durationpb.New(365 * 24 * 60 * 60 * 1e9), // 1 year in nanoseconds
		MaxPathLength:      1,
	}

	for i := 0; i < 5; i++ {
		val, err := env.ExecuteActivity(activities.GenerateCertificate, input)
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}

		var result *pb.GenerateCertificateOutput
		if err := val.Get(&result); err != nil {
			t.Fatalf("unexpected error from Get on iteration %d: %v", i, err)
		}

		certStr := string(result.Certificate)
		if certificates[certStr] {
			t.Errorf("duplicate certificate generated at iteration %d", i)
		}
		certificates[certStr] = true
	}

	// Should have generated 5 unique certificates
	if len(certificates) != 5 {
		t.Errorf("expected 5 unique certificates, got %d", len(certificates))
	}
}
