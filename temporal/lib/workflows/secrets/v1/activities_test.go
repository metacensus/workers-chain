package secrets

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	errProviderConnectionRefused = errors.New("connection refused")
	errProviderForbidden         = errors.New("forbidden")
)

// mockSecretProvider is a local mock for SecretProvider to avoid import cycles
type mockSecretProvider struct {
	mock.Mock
}

func newMockSecretProvider(t *testing.T) *mockSecretProvider {
	m := &mockSecretProvider{}
	m.Mock.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *mockSecretProvider) FetchSecret(ctx context.Context, ref *pb.SecretRef) (Secret, error) {
	args := m.Called(ctx, ref)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(Secret), args.Error(1)
}

func (m *mockSecretProvider) StoreSecret(ctx context.Context, ref *pb.SecretRef, secret Secret) error {
	args := m.Called(ctx, ref, secret)
	return args.Error(0)
}

func (m *mockSecretProvider) UpdateSecret(ctx context.Context, ref *pb.SecretRef, partialSecret Secret) error {
	args := m.Called(ctx, ref, partialSecret)
	return args.Error(0)
}

// testSecretRef creates a valid SecretRef for testing
func testSecretRef(path string) *pb.SecretRef {
	ref, _ := structpb.NewStruct(map[string]interface{}{
		"path": path,
	})
	return &pb.SecretRef{
		Type: TypeVault,
		Ref:  ref,
	}
}

// validKeyPairSecret creates a valid key pair secret for testing CreateSelfSignedCertificate
func validKeyPairSecret() Secret {
	// Generate a real key pair so we have valid PEM data
	kp, _ := GenerateKeyPair()
	return FromKeyPair(kp)
}

func TestCreatePasswordActivity(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.CreatePasswordInput
		providerErr error
		expectErr   bool
	}{
		// Layer 1: Validation errors
		{
			name: "error - missing secret_ref",
			input: &pb.CreatePasswordInput{
				SecretRef: nil,
				Username:  "admin",
			},
			expectErr: true,
		},
		// Layer 2: Success cases
		{
			name: "success - creates password with username",
			input: &pb.CreatePasswordInput{
				SecretRef: testSecretRef("fabric/shard/org1/admin"),
				Username:  "admin",
			},
			expectErr: false,
		},
		{
			name: "success - creates password with empty username",
			input: &pb.CreatePasswordInput{
				SecretRef: testSecretRef("fabric/shard/org1/service"),
				Username:  "",
			},
			expectErr: false,
		},
		// Layer 3: Provider errors
		{
			name: "error - provider connection refused",
			input: &pb.CreatePasswordInput{
				SecretRef: testSecretRef("fabric/shard/org1/admin"),
				Username:  "admin",
			},
			providerErr: errProviderConnectionRefused,
			expectErr:   true,
		},
		{
			name: "error - provider forbidden",
			input: &pb.CreatePasswordInput{
				SecretRef: testSecretRef("fabric/shard/org1/admin"),
				Username:  "admin",
			},
			providerErr: errProviderForbidden,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockProvider := newMockSecretProvider(t)

			// Conditional mock: only expect provider call if input passes validation
			if tt.input.SecretRef != nil {
				mockProvider.On("UpdateSecret", mock.Anything, tt.input.SecretRef, mock.MatchedBy(func(secret Secret) bool {
					// Verify the secret contains credentials
					_, hasUsername := secret["username"]
					_, hasPassword := secret["password"]
					return hasUsername && hasPassword
				})).Return(tt.providerErr).Maybe()
			}

			activities := NewActivities(mockProvider)
			env.RegisterActivity(activities.CreatePassword)

			val, err := env.ExecuteActivity(activities.CreatePassword, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.CreatePasswordOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.SecretRef == nil {
					t.Fatal("result.SecretRef is nil")
				}
			}
		})
	}
}

func TestCreateKeyPairActivity(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.CreateKeyPairInput
		providerErr error
		expectErr   bool
	}{
		// Layer 1: Validation errors
		{
			name: "error - missing secret_ref",
			input: &pb.CreateKeyPairInput{
				SecretRef: nil,
			},
			expectErr: true,
		},
		// Layer 2: Success cases
		{
			name: "success - creates key pair",
			input: &pb.CreateKeyPairInput{
				SecretRef: testSecretRef("fabric/shard/org1/ca/key"),
			},
			expectErr: false,
		},
		// Layer 3: Provider errors
		{
			name: "error - provider connection refused",
			input: &pb.CreateKeyPairInput{
				SecretRef: testSecretRef("fabric/shard/org1/ca/key"),
			},
			providerErr: errProviderConnectionRefused,
			expectErr:   true,
		},
		{
			name: "error - provider forbidden",
			input: &pb.CreateKeyPairInput{
				SecretRef: testSecretRef("fabric/shard/org1/ca/key"),
			},
			providerErr: errProviderForbidden,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockProvider := newMockSecretProvider(t)

			// Conditional mock: only expect provider call if input passes validation
			if tt.input.SecretRef != nil {
				mockProvider.On("UpdateSecret", mock.Anything, tt.input.SecretRef, mock.MatchedBy(func(secret Secret) bool {
					// Verify the secret contains a key pair
					_, hasPrivateKey := secret["private_key"]
					_, hasPublicKey := secret["public_key"]
					return hasPrivateKey && hasPublicKey
				})).Return(tt.providerErr).Maybe()
			}

			activities := NewActivities(mockProvider)
			env.RegisterActivity(activities.CreateKeyPair)

			val, err := env.ExecuteActivity(activities.CreateKeyPair, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.CreateKeyPairOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.SecretRef == nil {
					t.Fatal("result.SecretRef is nil")
				}
			}
		})
	}
}

func TestCreateSelfSignedCertificateActivity(t *testing.T) {
	validKeySecret := validKeyPairSecret()

	tests := []struct {
		name        string
		input       *pb.CreateSelfSignedCertificateInput
		fetchSecret Secret
		fetchErr    error
		storeErr    error
		expectErr   bool
		expectFetch bool
		expectStore bool
	}{
		// Layer 1: Validation errors
		{
			name: "error - missing key_ref",
			input: &pb.CreateSelfSignedCertificateInput{
				KeyRef:           nil,
				CertRef:          testSecretRef("fabric/shard/org1/ca/cert"),
				CommonName:       "org1-ca",
				Organization:     "org1",
				ValidityDuration: durationpb.New(365 * 24 * time.Hour),
			},
			expectErr:   true,
			expectFetch: false,
			expectStore: false,
		},
		{
			name: "error - missing cert_ref",
			input: &pb.CreateSelfSignedCertificateInput{
				KeyRef:           testSecretRef("fabric/shard/org1/ca/key"),
				CertRef:          nil,
				CommonName:       "org1-ca",
				Organization:     "org1",
				ValidityDuration: durationpb.New(365 * 24 * time.Hour),
			},
			expectErr:   true,
			expectFetch: false,
			expectStore: false,
		},
		// Layer 2: Success cases
		{
			name: "success - creates certificate with all params",
			input: &pb.CreateSelfSignedCertificateInput{
				KeyRef:             testSecretRef("fabric/shard/org1/ca/key"),
				CertRef:            testSecretRef("fabric/shard/org1/ca/cert"),
				CommonName:         "org1-ca",
				Organization:       "org1",
				OrganizationalUnit: "fabric",
				ValidityDuration:   durationpb.New(365 * 24 * time.Hour),
				MaxPathLength:      1,
				DnsNames:           []string{"ca.org1.example.com"},
			},
			fetchSecret: validKeySecret,
			expectErr:   false,
			expectFetch: true,
			expectStore: true,
		},
		{
			name: "success - creates certificate with minimal params",
			input: &pb.CreateSelfSignedCertificateInput{
				KeyRef:           testSecretRef("fabric/shard/org1/ca/key"),
				CertRef:          testSecretRef("fabric/shard/org1/ca/cert"),
				CommonName:       "org1-ca",
				Organization:     "org1",
				ValidityDuration: durationpb.New(30 * 24 * time.Hour),
			},
			fetchSecret: validKeySecret,
			expectErr:   false,
			expectFetch: true,
			expectStore: true,
		},
		// Layer 3: Provider errors
		{
			name: "error - provider fetch fails",
			input: &pb.CreateSelfSignedCertificateInput{
				KeyRef:           testSecretRef("fabric/shard/org1/ca/key"),
				CertRef:          testSecretRef("fabric/shard/org1/ca/cert"),
				CommonName:       "org1-ca",
				Organization:     "org1",
				ValidityDuration: durationpb.New(365 * 24 * time.Hour),
			},
			fetchErr:    errProviderConnectionRefused,
			expectErr:   true,
			expectFetch: true,
			expectStore: false,
		},
		{
			name: "error - invalid key pair in secret",
			input: &pb.CreateSelfSignedCertificateInput{
				KeyRef:           testSecretRef("fabric/shard/org1/ca/key"),
				CertRef:          testSecretRef("fabric/shard/org1/ca/cert"),
				CommonName:       "org1-ca",
				Organization:     "org1",
				ValidityDuration: durationpb.New(365 * 24 * time.Hour),
			},
			fetchSecret: Secret{"invalid": "data"}, // Missing private_key and public_key
			expectErr:   true,
			expectFetch: true,
			expectStore: false,
		},
		{
			name: "error - provider store fails",
			input: &pb.CreateSelfSignedCertificateInput{
				KeyRef:           testSecretRef("fabric/shard/org1/ca/key"),
				CertRef:          testSecretRef("fabric/shard/org1/ca/cert"),
				CommonName:       "org1-ca",
				Organization:     "org1",
				ValidityDuration: durationpb.New(365 * 24 * time.Hour),
			},
			fetchSecret: validKeySecret,
			storeErr:    errProviderForbidden,
			expectErr:   true,
			expectFetch: true,
			expectStore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockProvider := newMockSecretProvider(t)

			// Conditional mock: FetchSecret
			if tt.expectFetch {
				mockProvider.On("FetchSecret", mock.Anything, tt.input.KeyRef).
					Return(tt.fetchSecret, tt.fetchErr).Maybe()
			}

			// Conditional mock: UpdateSecret (only if fetch succeeds and key parses)
			if tt.expectStore {
				mockProvider.On("UpdateSecret", mock.Anything, tt.input.CertRef, mock.MatchedBy(func(secret Secret) bool {
					_, hasCert := secret["certificate"]
					return hasCert
				})).Return(tt.storeErr).Maybe()
			}

			activities := NewActivities(mockProvider)
			env.RegisterActivity(activities.CreateSelfSignedCertificate)

			val, err := env.ExecuteActivity(activities.CreateSelfSignedCertificate, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.CreateSelfSignedCertificateOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.CertRef == nil {
					t.Fatal("result.CertRef is nil")
				}
			}
		})
	}
}

func TestStoreCertificateActivity(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.StoreCertificateInput
		providerErr error
		expectErr   bool
	}{
		// Layer 1: Validation errors
		{
			name: "error - missing cert_ref",
			input: &pb.StoreCertificateInput{
				CertRef:        nil,
				CertificatePem: []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"),
			},
			expectErr: true,
		},
		// Layer 2: Success cases
		{
			name: "success - stores certificate",
			input: &pb.StoreCertificateInput{
				CertRef:        testSecretRef("fabric/shard/org1/peer/tls-cert"),
				CertificatePem: []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"),
			},
			expectErr: false,
		},
		{
			name: "success - stores certificate with empty PEM",
			input: &pb.StoreCertificateInput{
				CertRef:        testSecretRef("fabric/shard/org1/peer/tls-cert"),
				CertificatePem: []byte{},
			},
			expectErr: false,
		},
		// Layer 3: Provider errors
		{
			name: "error - provider connection refused",
			input: &pb.StoreCertificateInput{
				CertRef:        testSecretRef("fabric/shard/org1/peer/tls-cert"),
				CertificatePem: []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"),
			},
			providerErr: errProviderConnectionRefused,
			expectErr:   true,
		},
		{
			name: "error - provider forbidden",
			input: &pb.StoreCertificateInput{
				CertRef:        testSecretRef("fabric/shard/org1/peer/tls-cert"),
				CertificatePem: []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"),
			},
			providerErr: errProviderForbidden,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockProvider := newMockSecretProvider(t)

			// Conditional mock: only expect provider call if input passes validation
			if tt.input.CertRef != nil {
				mockProvider.On("UpdateSecret", mock.Anything, tt.input.CertRef, mock.MatchedBy(func(secret Secret) bool {
					_, hasCert := secret["certificate"]
					return hasCert
				})).Return(tt.providerErr).Maybe()
			}

			activities := NewActivities(mockProvider)
			env.RegisterActivity(activities.StoreCertificate)

			val, err := env.ExecuteActivity(activities.StoreCertificate, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				var result *pb.StoreCertificateOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("error getting result: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
				if result.CertRef == nil {
					t.Fatal("result.CertRef is nil")
				}
			}
		})
	}
}
