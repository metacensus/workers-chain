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

// TestEnsurePasswordActivity verifies the get-or-create semantics: an existing
// password is reused (never overwritten, so it can't desync from a consumer
// already provisioned from it), an absent secret (ErrSecretNotFound) triggers
// generation + store, and any other fetch error is fatal.
func TestEnsurePasswordActivity(t *testing.T) {
	existingPassword := FromCredentials(&Credentials{Username: "admin", Password: "existing-secret"})

	tests := []struct {
		name        string
		input       *pb.EnsurePasswordInput
		fetchSecret Secret
		fetchErr    error
		storeErr    error
		expectFetch bool
		expectStore bool
		expectErr   bool
	}{
		// Layer 1: Validation errors
		{
			name:      "error - missing secret_ref",
			input:     &pb.EnsurePasswordInput{SecretRef: nil, Username: "admin"},
			expectErr: true,
		},
		// Layer 2: get-or-create behavior
		{
			name:        "success - password absent, generates and stores",
			input:       &pb.EnsurePasswordInput{SecretRef: testSecretRef("fabric/shard/org1/admin"), Username: "admin"},
			fetchErr:    ErrSecretNotFound,
			expectFetch: true,
			expectStore: true,
			expectErr:   false,
		},
		{
			name:        "success - password exists, reuses without storing",
			input:       &pb.EnsurePasswordInput{SecretRef: testSecretRef("fabric/shard/org1/admin"), Username: "admin"},
			fetchSecret: existingPassword,
			expectFetch: true,
			expectStore: false,
			expectErr:   false,
		},
		// Layer 3: Provider errors
		{
			name:        "error - transient fetch failure is fatal (no overwrite)",
			input:       &pb.EnsurePasswordInput{SecretRef: testSecretRef("fabric/shard/org1/admin"), Username: "admin"},
			fetchErr:    errProviderConnectionRefused,
			expectFetch: true,
			expectStore: false,
			expectErr:   true,
		},
		{
			name:        "error - store fails after generation",
			input:       &pb.EnsurePasswordInput{SecretRef: testSecretRef("fabric/shard/org1/admin"), Username: "admin"},
			fetchErr:    ErrSecretNotFound,
			storeErr:    errProviderForbidden,
			expectFetch: true,
			expectStore: true,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockProvider := newMockSecretProvider(t)

			if tt.expectFetch {
				mockProvider.On("FetchSecret", mock.Anything, tt.input.SecretRef).
					Return(tt.fetchSecret, tt.fetchErr).Maybe()
			}
			if tt.expectStore {
				mockProvider.On("UpdateSecret", mock.Anything, tt.input.SecretRef, mock.MatchedBy(func(secret Secret) bool {
					_, hasUsername := secret["username"]
					_, hasPassword := secret["password"]
					return hasUsername && hasPassword
				})).Return(tt.storeErr).Maybe()
			}

			activities := NewActivities(mockProvider)
			env.RegisterActivity(activities.EnsurePassword)

			val, err := env.ExecuteActivity(activities.EnsurePassword, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var result *pb.EnsurePasswordOutput
			if err := val.Get(&result); err != nil {
				t.Fatalf("error getting result: %v", err)
			}
			if result == nil || result.SecretRef == nil {
				t.Fatal("result or result.SecretRef is nil")
			}
			// When the password already exists it must be reused, not overwritten.
			if !tt.expectStore {
				mockProvider.AssertNotCalled(t, "UpdateSecret", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

// TestEnsureKeyPairActivity verifies the get-or-create semantics: an existing
// key is reused (never overwritten), an absent key (ErrSecretNotFound) triggers
// generation + store, and any other fetch error is fatal so a possibly-valid
// key is never clobbered.
func TestEnsureKeyPairActivity(t *testing.T) {
	existingKey := validKeyPairSecret()

	tests := []struct {
		name        string
		input       *pb.EnsureKeyPairInput
		fetchSecret Secret
		fetchErr    error
		storeErr    error
		expectFetch bool
		expectStore bool
		expectErr   bool
	}{
		// Layer 1: Validation errors
		{
			name:      "error - missing secret_ref",
			input:     &pb.EnsureKeyPairInput{SecretRef: nil},
			expectErr: true,
		},
		// Layer 2: get-or-create behavior
		{
			name:        "success - key absent, generates and stores",
			input:       &pb.EnsureKeyPairInput{SecretRef: testSecretRef("fabric/shard/org1/ca/key")},
			fetchErr:    ErrSecretNotFound,
			expectFetch: true,
			expectStore: true,
			expectErr:   false,
		},
		{
			name:        "success - key exists, reuses without storing",
			input:       &pb.EnsureKeyPairInput{SecretRef: testSecretRef("fabric/shard/org1/ca/key")},
			fetchSecret: existingKey,
			expectFetch: true,
			expectStore: false,
			expectErr:   false,
		},
		// Layer 3: Provider errors
		{
			name:        "error - transient fetch failure is fatal (no overwrite)",
			input:       &pb.EnsureKeyPairInput{SecretRef: testSecretRef("fabric/shard/org1/ca/key")},
			fetchErr:    errProviderConnectionRefused,
			expectFetch: true,
			expectStore: false,
			expectErr:   true,
		},
		{
			name:        "error - store fails after generation",
			input:       &pb.EnsureKeyPairInput{SecretRef: testSecretRef("fabric/shard/org1/ca/key")},
			fetchErr:    ErrSecretNotFound,
			storeErr:    errProviderForbidden,
			expectFetch: true,
			expectStore: true,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockProvider := newMockSecretProvider(t)

			if tt.expectFetch {
				mockProvider.On("FetchSecret", mock.Anything, tt.input.SecretRef).
					Return(tt.fetchSecret, tt.fetchErr).Maybe()
			}
			if tt.expectStore {
				mockProvider.On("UpdateSecret", mock.Anything, tt.input.SecretRef, mock.MatchedBy(func(secret Secret) bool {
					_, hasPrivateKey := secret["private_key"]
					_, hasPublicKey := secret["public_key"]
					return hasPrivateKey && hasPublicKey
				})).Return(tt.storeErr).Maybe()
			}

			activities := NewActivities(mockProvider)
			env.RegisterActivity(activities.EnsureKeyPair)

			val, err := env.ExecuteActivity(activities.EnsureKeyPair, tt.input)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var result *pb.EnsureKeyPairOutput
			if err := val.Get(&result); err != nil {
				t.Fatalf("error getting result: %v", err)
			}
			if result == nil || result.SecretRef == nil {
				t.Fatal("result or result.SecretRef is nil")
			}
			// When the key already exists it must be reused, not overwritten.
			if !tt.expectStore {
				mockProvider.AssertNotCalled(t, "UpdateSecret", mock.Anything, mock.Anything, mock.Anything)
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
		{
			// Zero/omitted validity would yield NotAfter == NotBefore; the
			// generator rejects it after the key is fetched, before storing.
			name: "error - non-positive validity duration",
			input: &pb.CreateSelfSignedCertificateInput{
				KeyRef:       testSecretRef("fabric/shard/org1/ca/key"),
				CertRef:      testSecretRef("fabric/shard/org1/ca/cert"),
				CommonName:   "org1-ca",
				Organization: "org1",
				// ValidityDuration omitted → zero
			},
			fetchSecret: validKeySecret,
			expectErr:   true,
			expectFetch: true,
			expectStore: false,
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

// TestStoreCertificateActivity_PersistsChain verifies the fix that StoreCertificate
// persists BOTH the leaf certificate and the full chain: when ca_chain_pem is
// given the stored "chain" equals it; when omitted the leaf is stored as the
// chain (root/self-contained case). Downstream mTLS consumers rely on the full
// chain being present.
func TestStoreCertificateActivity_PersistsChain(t *testing.T) {
	leaf := []byte("-----BEGIN CERTIFICATE-----\nleaf\n-----END CERTIFICATE-----")
	chain := []byte("-----BEGIN CERTIFICATE-----\nleaf\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nissuer\n-----END CERTIFICATE-----")

	tests := []struct {
		name      string
		input     *pb.StoreCertificateInput
		wantCert  string
		wantChain string
	}{
		{
			name: "with explicit chain",
			input: &pb.StoreCertificateInput{
				CertRef:        testSecretRef("fabric/shard/org1/peer/tls-cert"),
				CertificatePem: leaf,
				CaChainPem:     chain,
			},
			wantCert:  string(leaf),
			wantChain: string(chain),
		},
		{
			name: "no chain falls back to leaf",
			input: &pb.StoreCertificateInput{
				CertRef:        testSecretRef("fabric/shard/org1/peer/tls-cert"),
				CertificatePem: leaf,
			},
			wantCert:  string(leaf),
			wantChain: string(leaf),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockProvider := newMockSecretProvider(t)
			var stored Secret
			mockProvider.On("UpdateSecret", mock.Anything, tt.input.CertRef, mock.Anything).
				Run(func(args mock.Arguments) {
					stored = args.Get(2).(Secret)
				}).
				Return(nil)

			activities := NewActivities(mockProvider)
			env.RegisterActivity(activities.StoreCertificate)

			if _, err := env.ExecuteActivity(activities.StoreCertificate, tt.input); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got, _ := stored["certificate"].(string); got != tt.wantCert {
				t.Errorf("certificate = %q, want %q", got, tt.wantCert)
			}
			if got, _ := stored["chain"].(string); got != tt.wantChain {
				t.Errorf("chain = %q, want %q", got, tt.wantChain)
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
