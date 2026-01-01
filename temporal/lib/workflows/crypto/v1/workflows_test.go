package crypto_test

import (
	"testing"

	crypto "github.com/metacensus/infra/temporal/lib/workflows/crypto/v1"

	cryptopb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	cryptomocks "github.com/metacensus/infra/temporal/mocks/workflows/crypto/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"google.golang.org/protobuf/types/known/durationpb"
)

var errActivityFailed = temporal.NewApplicationError("activity failed", "ActivityFailed")

func TestGenerateAndStorePasswordWorkflow(t *testing.T) {
	tests := []struct {
		name                 string
		input                *cryptopb.GenerateAndStorePasswordWorkflowInput
		generatePasswordRes  *cryptopb.GeneratePasswordOutput
		generatePasswordErr  error
		storeSecretRes       *vaultpb.StoreSecretOutput
		storeSecretErr       error
		expectErr            bool
		validateStoreRequest func(*testing.T, *vaultpb.StoreSecretInput)
	}{
		{
			name: "success - password generated and stored",
			input: &cryptopb.GenerateAndStorePasswordWorkflowInput{
				Req: &cryptopb.GenerateAndStorePasswordInput{
					VaultPath: "fabric/org1/ca/tls/root/admin",
					Username:  "postgres_admin",
				},
			},
			generatePasswordRes: &cryptopb.GeneratePasswordOutput{
				Password: "test-password-abc123XYZ",
			},
			storeSecretRes: &vaultpb.StoreSecretOutput{
				SecretRef: &vaultpb.SecretRef{
					Path:    "fabric/org1/ca/tls/root/admin",
					Version: 1,
				},
			},
			validateStoreRequest: func(t *testing.T, input *vaultpb.StoreSecretInput) {
				if input.GetPath() != "fabric/org1/ca/tls/root/admin" {
					t.Errorf("path: got %q, want %q", input.GetPath(), "fabric/org1/ca/tls/root/admin")
				}
				if len(input.GetData()) != 2 {
					t.Errorf("data length: got %d, want 2", len(input.GetData()))
				}
				if string(input.GetData()["username"]) != "postgres_admin" {
					t.Errorf("username: got %q, want %q", input.GetData()["username"], "postgres_admin")
				}
				if string(input.GetData()["password"]) != "test-password-abc123XYZ" {
					t.Errorf("password: got %q, want %q", input.GetData()["password"], "test-password-abc123XYZ")
				}
			},
		},
		{
			name: "failure - GeneratePassword activity fails",
			input: &cryptopb.GenerateAndStorePasswordWorkflowInput{
				Req: &cryptopb.GenerateAndStorePasswordInput{
					VaultPath: "fabric/org1/ca/tls/root/admin",
					Username:  "postgres_admin",
				},
			},
			generatePasswordErr: errActivityFailed,
			expectErr:           true,
		},
		{
			name: "failure - StoreSecret activity fails",
			input: &cryptopb.GenerateAndStorePasswordWorkflowInput{
				Req: &cryptopb.GenerateAndStorePasswordInput{
					VaultPath: "fabric/org1/ca/tls/root/admin",
					Username:  "postgres_admin",
				},
			},
			generatePasswordRes: &cryptopb.GeneratePasswordOutput{
				Password: "test-password-abc123XYZ",
			},
			storeSecretErr: errActivityFailed,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register the actual crypto workflows
			realWorkflows := crypto.NewWorkflows()
			cryptopb.RegisterCryptoWorkflows(env, realWorkflows)

			// Mock crypto activities
			mockCryptoActivities := cryptomocks.NewMockCryptoActivities(t)
			mockCryptoActivities.EXPECT().
				GeneratePassword(mock.Anything, mock.Anything).
				Return(tt.generatePasswordRes, tt.generatePasswordErr).
				Maybe()
			cryptopb.RegisterCryptoActivities(env, mockCryptoActivities)

			// Mock vault activities - only called if GeneratePassword succeeds
			mockVaultActivities := vaultmocks.NewMockVaultActivities(t)
			if tt.generatePasswordErr == nil {
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.MatchedBy(func(input *vaultpb.StoreSecretInput) bool {
						if tt.validateStoreRequest != nil {
							tt.validateStoreRequest(t, input)
						}
						return true
					})).
					Return(tt.storeSecretRes, tt.storeSecretErr).
					Maybe()
			}
			vaultpb.RegisterVaultActivities(env, mockVaultActivities)

			// Execute workflow using the test helper function
			env.ExecuteWorkflow(crypto.TestGenerateAndStorePasswordWorkflow, tt.input)

			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *cryptopb.GenerateAndStorePasswordOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetSecretRef().GetPath() != tt.storeSecretRes.GetSecretRef().GetPath() {
					t.Errorf("path mismatch: got %q, want %q",
						result.GetSecretRef().GetPath(), tt.storeSecretRes.GetSecretRef().GetPath())
				}

				if result.GetSecretRef().GetVersion() != tt.storeSecretRes.GetSecretRef().GetVersion() {
					t.Errorf("version mismatch: got %d, want %d",
						result.GetSecretRef().GetVersion(), tt.storeSecretRes.GetSecretRef().GetVersion())
				}
			}
		})
	}
}

func TestGenerateAndStoreKeyPairWorkflow(t *testing.T) {
	tests := []struct {
		name                 string
		input                *cryptopb.GenerateAndStoreKeyPairWorkflowInput
		generateKeyPairRes   *cryptopb.GenerateKeyPairOutput
		generateKeyPairErr   error
		storeSecretRes       *vaultpb.StoreSecretOutput
		storeSecretErr       error
		expectErr            bool
		validateStoreRequest func(*testing.T, *vaultpb.StoreSecretInput)
	}{
		{
			name: "success - key pair generated and stored",
			input: &cryptopb.GenerateAndStoreKeyPairWorkflowInput{
				Req: &cryptopb.GenerateAndStoreKeyPairInput{
					VaultPath: "fabric/org1/ca/tls/root/keys",
				},
			},
			generateKeyPairRes: &cryptopb.GenerateKeyPairOutput{
				PrivateKey: []byte("-----BEGIN PRIVATE KEY-----\ntest-private-key\n-----END PRIVATE KEY-----"),
				PublicKey:  []byte("-----BEGIN PUBLIC KEY-----\ntest-public-key\n-----END PUBLIC KEY-----"),
			},
			storeSecretRes: &vaultpb.StoreSecretOutput{
				SecretRef: &vaultpb.SecretRef{
					Path:    "fabric/org1/ca/tls/root/keys",
					Version: 1,
				},
			},
			validateStoreRequest: func(t *testing.T, input *vaultpb.StoreSecretInput) {
				if input.GetPath() != "fabric/org1/ca/tls/root/keys" {
					t.Errorf("path: got %q, want %q", input.GetPath(), "fabric/org1/ca/tls/root/keys")
				}
				if len(input.GetData()) != 2 {
					t.Errorf("data length: got %d, want 2", len(input.GetData()))
				}
				if _, ok := input.GetData()["private"]; !ok {
					t.Error("missing 'private' key in data")
				}
				if _, ok := input.GetData()["public"]; !ok {
					t.Error("missing 'public' key in data")
				}
			},
		},
		{
			name: "failure - GenerateKeyPair activity fails",
			input: &cryptopb.GenerateAndStoreKeyPairWorkflowInput{
				Req: &cryptopb.GenerateAndStoreKeyPairInput{
					VaultPath: "fabric/org1/ca/tls/root/keys",
				},
			},
			generateKeyPairErr: errActivityFailed,
			expectErr:          true,
		},
		{
			name: "failure - StoreSecret activity fails",
			input: &cryptopb.GenerateAndStoreKeyPairWorkflowInput{
				Req: &cryptopb.GenerateAndStoreKeyPairInput{
					VaultPath: "fabric/org1/ca/tls/root/keys",
				},
			},
			generateKeyPairRes: &cryptopb.GenerateKeyPairOutput{
				PrivateKey: []byte("-----BEGIN PRIVATE KEY-----\ntest-private-key\n-----END PRIVATE KEY-----"),
				PublicKey:  []byte("-----BEGIN PUBLIC KEY-----\ntest-public-key\n-----END PUBLIC KEY-----"),
			},
			storeSecretErr: errActivityFailed,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register the actual crypto workflows
			realWorkflows := crypto.NewWorkflows()
			cryptopb.RegisterCryptoWorkflows(env, realWorkflows)

			// Mock crypto activities
			mockCryptoActivities := cryptomocks.NewMockCryptoActivities(t)
			mockCryptoActivities.EXPECT().
				GenerateKeyPair(mock.Anything, mock.Anything).
				Return(tt.generateKeyPairRes, tt.generateKeyPairErr).
				Maybe()
			cryptopb.RegisterCryptoActivities(env, mockCryptoActivities)

			// Mock vault activities - only called if GenerateKeyPair succeeds
			mockVaultActivities := vaultmocks.NewMockVaultActivities(t)
			if tt.generateKeyPairErr == nil {
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.MatchedBy(func(input *vaultpb.StoreSecretInput) bool {
						if tt.validateStoreRequest != nil {
							tt.validateStoreRequest(t, input)
						}
						return true
					})).
					Return(tt.storeSecretRes, tt.storeSecretErr).
					Maybe()
			}
			vaultpb.RegisterVaultActivities(env, mockVaultActivities)

			// Execute workflow using the test helper function
			env.ExecuteWorkflow(crypto.TestGenerateAndStoreKeyPairWorkflow, tt.input)

			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *cryptopb.GenerateAndStoreKeyPairOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetSecretRef().GetPath() != tt.storeSecretRes.GetSecretRef().GetPath() {
					t.Errorf("path mismatch: got %q, want %q",
						result.GetSecretRef().GetPath(), tt.storeSecretRes.GetSecretRef().GetPath())
				}

				if result.GetSecretRef().GetVersion() != tt.storeSecretRes.GetSecretRef().GetVersion() {
					t.Errorf("version mismatch: got %d, want %d",
						result.GetSecretRef().GetVersion(), tt.storeSecretRes.GetSecretRef().GetVersion())
				}
			}
		})
	}
}

func TestCreateSelfSignedCertificateWorkflow(t *testing.T) {
	tests := []struct {
		name                    string
		input                   *cryptopb.CreateSelfSignedCertificateWorkflowInput
		fetchSecretRes          *vaultpb.FetchSecretOutput
		fetchSecretErr          error
		generateCertificateRes  *cryptopb.GenerateCertificateOutput
		generateCertificateErr  error
		storeSecretRes          *vaultpb.StoreSecretOutput
		storeSecretErr          error
		expectErr               bool
		validateFetchRequest    func(*testing.T, *vaultpb.FetchSecretInput)
		validateGenerateRequest func(*testing.T, *cryptopb.GenerateCertificateInput)
		validateStoreRequest    func(*testing.T, *vaultpb.StoreSecretInput)
	}{
		{
			name: "success - certificate generated and stored",
			input: &cryptopb.CreateSelfSignedCertificateWorkflowInput{
				Req: &cryptopb.CreateSelfSignedCertificateInput{
					VaultKeyPath:         "fabric/org1/ca/tls/root/keys",
					VaultCertificatePath: "fabric/org1/ca/tls/root/cert",
					CommonName:           "org1-tls-root-ca",
					Organization:         "org1",
					OrganizationalUnit:   "root-ca",
					ValidityDuration:     durationpb.New(365 * 24 * 60 * 60 * 1e9),
					MaxPathLength:        1,
				},
			},
			fetchSecretRes: &vaultpb.FetchSecretOutput{
				Data: map[string][]byte{
					"private": []byte("-----BEGIN PRIVATE KEY-----\ntest-private-key\n-----END PRIVATE KEY-----"),
					"public":  []byte("-----BEGIN PUBLIC KEY-----\ntest-public-key\n-----END PUBLIC KEY-----"),
				},
			},
			generateCertificateRes: &cryptopb.GenerateCertificateOutput{
				Certificate: []byte("-----BEGIN CERTIFICATE-----\ntest-certificate\n-----END CERTIFICATE-----"),
			},
			storeSecretRes: &vaultpb.StoreSecretOutput{
				SecretRef: &vaultpb.SecretRef{
					Path:    "fabric/org1/ca/tls/root/cert",
					Version: 1,
				},
			},
			validateFetchRequest: func(t *testing.T, input *vaultpb.FetchSecretInput) {
				if input.GetPath() != "fabric/org1/ca/tls/root/keys" {
					t.Errorf("fetch path: got %q, want %q", input.GetPath(), "fabric/org1/ca/tls/root/keys")
				}
			},
			validateGenerateRequest: func(t *testing.T, input *cryptopb.GenerateCertificateInput) {
				if input.GetCommonName() != "org1-tls-root-ca" {
					t.Errorf("CommonName: got %q, want %q", input.GetCommonName(), "org1-tls-root-ca")
				}
				if input.GetOrganization() != "org1" {
					t.Errorf("Organization: got %q, want %q", input.GetOrganization(), "org1")
				}
				if input.GetOrganizationalUnit() != "root-ca" {
					t.Errorf("OrganizationalUnit: got %q, want %q", input.GetOrganizationalUnit(), "root-ca")
				}
				if input.GetMaxPathLength() != 1 {
					t.Errorf("MaxPathLength: got %d, want 1", input.GetMaxPathLength())
				}
			},
			validateStoreRequest: func(t *testing.T, input *vaultpb.StoreSecretInput) {
				if input.GetPath() != "fabric/org1/ca/tls/root/cert" {
					t.Errorf("store path: got %q, want %q", input.GetPath(), "fabric/org1/ca/tls/root/cert")
				}
				if len(input.GetData()) != 1 {
					t.Errorf("data length: got %d, want 1", len(input.GetData()))
				}
				if _, ok := input.GetData()["certificate"]; !ok {
					t.Error("missing 'certificate' key in data")
				}
			},
		},
		{
			name: "failure - FetchSecret activity fails",
			input: &cryptopb.CreateSelfSignedCertificateWorkflowInput{
				Req: &cryptopb.CreateSelfSignedCertificateInput{
					VaultKeyPath:         "fabric/org1/ca/tls/root/keys",
					VaultCertificatePath: "fabric/org1/ca/tls/root/cert",
					CommonName:           "org1-tls-root-ca",
					Organization:         "org1",
					OrganizationalUnit:   "root-ca",
					ValidityDuration:     durationpb.New(365 * 24 * 60 * 60 * 1e9),
					MaxPathLength:        1,
				},
			},
			fetchSecretErr: errActivityFailed,
			expectErr:      true,
		},
		{
			name: "failure - private key not found in vault secret",
			input: &cryptopb.CreateSelfSignedCertificateWorkflowInput{
				Req: &cryptopb.CreateSelfSignedCertificateInput{
					VaultKeyPath:         "fabric/org1/ca/tls/root/keys",
					VaultCertificatePath: "fabric/org1/ca/tls/root/cert",
					CommonName:           "org1-tls-root-ca",
					Organization:         "org1",
					OrganizationalUnit:   "root-ca",
					ValidityDuration:     durationpb.New(365 * 24 * 60 * 60 * 1e9),
					MaxPathLength:        1,
				},
			},
			fetchSecretRes: &vaultpb.FetchSecretOutput{
				Data: map[string][]byte{
					"public": []byte("-----BEGIN PUBLIC KEY-----\ntest-public-key\n-----END PUBLIC KEY-----"),
				},
			},
			expectErr: true,
		},
		{
			name: "failure - GenerateCertificate activity fails",
			input: &cryptopb.CreateSelfSignedCertificateWorkflowInput{
				Req: &cryptopb.CreateSelfSignedCertificateInput{
					VaultKeyPath:         "fabric/org1/ca/tls/root/keys",
					VaultCertificatePath: "fabric/org1/ca/tls/root/cert",
					CommonName:           "org1-tls-root-ca",
					Organization:         "org1",
					OrganizationalUnit:   "root-ca",
					ValidityDuration:     durationpb.New(365 * 24 * 60 * 60 * 1e9),
					MaxPathLength:        1,
				},
			},
			fetchSecretRes: &vaultpb.FetchSecretOutput{
				Data: map[string][]byte{
					"private": []byte("-----BEGIN PRIVATE KEY-----\ntest-private-key\n-----END PRIVATE KEY-----"),
					"public":  []byte("-----BEGIN PUBLIC KEY-----\ntest-public-key\n-----END PUBLIC KEY-----"),
				},
			},
			generateCertificateErr: errActivityFailed,
			expectErr:              true,
		},
		{
			name: "failure - StoreSecret activity fails",
			input: &cryptopb.CreateSelfSignedCertificateWorkflowInput{
				Req: &cryptopb.CreateSelfSignedCertificateInput{
					VaultKeyPath:         "fabric/org1/ca/tls/root/keys",
					VaultCertificatePath: "fabric/org1/ca/tls/root/cert",
					CommonName:           "org1-tls-root-ca",
					Organization:         "org1",
					OrganizationalUnit:   "root-ca",
					ValidityDuration:     durationpb.New(365 * 24 * 60 * 60 * 1e9),
					MaxPathLength:        1,
				},
			},
			fetchSecretRes: &vaultpb.FetchSecretOutput{
				Data: map[string][]byte{
					"private": []byte("-----BEGIN PRIVATE KEY-----\ntest-private-key\n-----END PRIVATE KEY-----"),
					"public":  []byte("-----BEGIN PUBLIC KEY-----\ntest-public-key\n-----END PUBLIC KEY-----"),
				},
			},
			generateCertificateRes: &cryptopb.GenerateCertificateOutput{
				Certificate: []byte("-----BEGIN CERTIFICATE-----\ntest-certificate\n-----END CERTIFICATE-----"),
			},
			storeSecretErr: errActivityFailed,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			realWorkflows := crypto.NewWorkflows()
			cryptopb.RegisterCryptoWorkflows(env, realWorkflows)

			mockCryptoActivities := cryptomocks.NewMockCryptoActivities(t)
			if tt.fetchSecretErr == nil && tt.fetchSecretRes != nil && tt.fetchSecretRes.Data["private"] != nil {
				mockCryptoActivities.EXPECT().
					GenerateCertificate(mock.Anything, mock.MatchedBy(func(input *cryptopb.GenerateCertificateInput) bool {
						if tt.validateGenerateRequest != nil {
							tt.validateGenerateRequest(t, input)
						}
						return true
					})).
					Return(tt.generateCertificateRes, tt.generateCertificateErr).
					Maybe()
			}
			cryptopb.RegisterCryptoActivities(env, mockCryptoActivities)

			mockVaultActivities := vaultmocks.NewMockVaultActivities(t)
			mockVaultActivities.EXPECT().
				FetchSecret(mock.Anything, mock.MatchedBy(func(input *vaultpb.FetchSecretInput) bool {
					if tt.validateFetchRequest != nil {
						tt.validateFetchRequest(t, input)
					}
					return true
				})).
				Return(tt.fetchSecretRes, tt.fetchSecretErr).
				Maybe()

			if tt.fetchSecretErr == nil && tt.fetchSecretRes != nil && tt.fetchSecretRes.Data["private"] != nil && tt.generateCertificateErr == nil {
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.MatchedBy(func(input *vaultpb.StoreSecretInput) bool {
						if tt.validateStoreRequest != nil {
							tt.validateStoreRequest(t, input)
						}
						return true
					})).
					Return(tt.storeSecretRes, tt.storeSecretErr).
					Maybe()
			}
			vaultpb.RegisterVaultActivities(env, mockVaultActivities)

			env.ExecuteWorkflow(crypto.TestCreateSelfSignedCertificateWorkflow, tt.input)

			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *cryptopb.CreateSelfSignedCertificateOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetSecretRef().GetPath() != tt.storeSecretRes.GetSecretRef().GetPath() {
					t.Errorf("path mismatch: got %q, want %q",
						result.GetSecretRef().GetPath(), tt.storeSecretRes.GetSecretRef().GetPath())
				}

				if result.GetSecretRef().GetVersion() != tt.storeSecretRes.GetSecretRef().GetVersion() {
					t.Errorf("version mismatch: got %d, want %d",
						result.GetSecretRef().GetVersion(), tt.storeSecretRes.GetSecretRef().GetVersion())
				}
			}
		})
	}
}
