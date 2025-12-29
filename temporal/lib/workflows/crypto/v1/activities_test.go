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
