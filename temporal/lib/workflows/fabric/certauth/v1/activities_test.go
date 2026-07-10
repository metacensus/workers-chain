package certauth_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-ca/api"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/metacensus/infra/temporal/lib/workflows/common"
	certauth "github.com/metacensus/infra/temporal/lib/workflows/fabric/certauth/v1"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	certauthmocks "github.com/metacensus/infra/temporal/mocks/workflows/fabric/certauth/v1"
	secretsmocks "github.com/metacensus/infra/temporal/mocks/workflows/secrets/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

// The success paths of EnrollIdentity and BootstrapEnrollIdentity construct a
// BootstrapClient directly and speak HTTP to a live Fabric CA, so they are
// covered by integration tests rather than unit tests here. Register/Revoke go
// through the MultiClientProvider, which lets us inject a mocked Client.

// Test identities mirror the constants in the internal test package; external
// test packages cannot reference internal test identifiers.
const externalTestOrg = "testorg"

// Activities under test act as a shard worker's org owner by default; every CA
// referenced in an activity input belongs to the executing worker's owner.
var (
	externalOrgOwner      = workerconfig.NewOrgConfig(externalTestOrg)
	externalOperatorOwner = workerconfig.NewOperatorConfig("metacensus")
)

var errVaultUnavailable = errors.New("vault unavailable")

// newTestActivities builds Activities whose client provider hydrates configs via
// the mocked secret provider and always constructs the given mocked CA client.
// Activities act as the given owner (the worker's injected identity).
func newTestActivities(secretProvider secrets.SecretProvider, client certauth.Client, owner workerconfig.OwnerConfig) *certauth.Activities {
	clientProvider := common.NewMultiClientProvider(
		secretProvider,
		func(certauth.ClientConfig) (certauth.Client, error) { return client, nil },
	)
	return certauth.NewActivities(secretProvider, clientProvider, owner)
}

// adminHydrationSecret satisfies both Secret.KeyPair() (identity_key ref) and
// Secret.Certificate() (identity_cert and server_tls refs), so a single stub
// covers all three refs fetched while hydrating the CA client config.
func adminHydrationSecret() secrets.Secret {
	return secrets.Secret{
		"private_key": "-----BEGIN PRIVATE KEY-----\nfake\n-----END PRIVATE KEY-----\n",
		"public_key":  "-----BEGIN PUBLIC KEY-----\nfake\n-----END PUBLIC KEY-----\n",
		"certificate": "-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----\n",
		"chain":       "-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----\n",
	}
}

// refWithPath matches a *secretspb.SecretRef whose vault path equals path.
func refWithPath(path string) interface{} {
	return mock.MatchedBy(func(ref *secretspb.SecretRef) bool {
		return ref.GetRef().GetFields()["path"].GetStringValue() == path
	})
}

func TestActivities_RegisterIdentity(t *testing.T) {
	tests := []struct {
		name            string
		owner           workerconfig.OwnerConfig // nil = org owner (shard worker)
		input           *pb.RegisterIdentityInput
		setupSecretMock func(*secretsmocks.MockSecretProvider)
		setupClientMock func(*certauthmocks.MockClient)
		want            *pb.RegisterIdentityOutput
		errSubstr       string
	}{
		{
			name: "success - org CA",
			input: &pb.RegisterIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "peer0",
				IdentityType: "peer",
				Affiliation:  "org1.dept1",
				Attributes:   map[string]string{"hf.Registrar.Roles": "client"},
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(adminHydrationSecret(), nil)
			},
			setupClientMock: func(m *certauthmocks.MockClient) {
				m.EXPECT().
					Register(mock.Anything, mock.MatchedBy(func(req *api.RegistrationRequest) bool {
						return req.Name == "peer0" &&
							req.Type == "peer" &&
							req.Affiliation == "org1.dept1" &&
							len(req.Attributes) == 1
					})).
					Return(&api.RegistrationResponse{Secret: "enroll-secret"}, nil)
			},
			want: &pb.RegisterIdentityOutput{EnrollmentSecret: "enroll-secret"},
		},
		{
			name:  "success - tls CA on the operator worker",
			owner: externalOperatorOwner,
			input: &pb.RegisterIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				IdentityName: "peer0-tls",
				IdentityType: "peer",
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(adminHydrationSecret(), nil)
			},
			setupClientMock: func(m *certauthmocks.MockClient) {
				m.EXPECT().
					Register(mock.Anything, mock.MatchedBy(func(req *api.RegistrationRequest) bool {
						return req.Name == "peer0-tls"
					})).
					Return(&api.RegistrationResponse{Secret: "tls-enroll-secret"}, nil)
			},
			want: &pb.RegisterIdentityOutput{EnrollmentSecret: "tls-enroll-secret"},
		},
		{
			name: "success - explicit admin identity",
			input: &pb.RegisterIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "orderer0",
				IdentityType: "orderer",
				AdminIdentity: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/admin"),
					CertRef: secrets.NewVaultRef("test/certs/admin"),
				},
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(adminHydrationSecret(), nil)
			},
			setupClientMock: func(m *certauthmocks.MockClient) {
				m.EXPECT().
					Register(mock.Anything, mock.Anything).
					Return(&api.RegistrationResponse{Secret: "enroll-secret"}, nil)
			},
			want: &pb.RegisterIdentityOutput{EnrollmentSecret: "enroll-secret"},
		},
		{
			name: "error - missing ca_ref",
			input: &pb.RegisterIdentityInput{
				IdentityName: "peer0",
			},
			errSubstr: "ca_ref is required",
		},
		{
			name: "error - client hydration secret fetch fails",
			input: &pb.RegisterIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "peer0",
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to get CA client",
		},
		{
			name: "error - registration fails",
			input: &pb.RegisterIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "peer0",
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(adminHydrationSecret(), nil)
			},
			setupClientMock: func(m *certauthmocks.MockClient) {
				m.EXPECT().
					Register(mock.Anything, mock.Anything).
					Return(nil, errors.New("identity already registered"))
			},
			errSubstr: "registration failed",
		},
		{
			name: "error - registration returns empty secret",
			input: &pb.RegisterIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "peer0",
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(adminHydrationSecret(), nil)
			},
			setupClientMock: func(m *certauthmocks.MockClient) {
				m.EXPECT().
					Register(mock.Anything, mock.Anything).
					Return(&api.RegistrationResponse{}, nil)
			},
			errSubstr: "registration returned empty secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockSecretProvider := secretsmocks.NewMockSecretProvider(t)
			if tt.setupSecretMock != nil {
				tt.setupSecretMock(mockSecretProvider)
			}
			mockClient := certauthmocks.NewMockClient(t)
			if tt.setupClientMock != nil {
				tt.setupClientMock(mockClient)
			}

			owner := tt.owner
			if owner == nil {
				owner = externalOrgOwner
			}
			activities := newTestActivities(mockSecretProvider, mockClient, owner)
			env.RegisterActivity(activities.RegisterIdentity)

			val, err := env.ExecuteActivity(activities.RegisterIdentity, tt.input)

			if tt.errSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error from ExecuteActivity: %v", err)
			}
			var result *pb.RegisterIdentityOutput
			if err := val.Get(&result); err != nil {
				t.Fatalf("unexpected error from Get: %v", err)
			}
			if result.GetEnrollmentSecret() != tt.want.GetEnrollmentSecret() {
				t.Errorf("EnrollmentSecret = %q, want %q",
					result.GetEnrollmentSecret(), tt.want.GetEnrollmentSecret())
			}
		})
	}
}

func TestActivities_RevokeIdentity(t *testing.T) {
	tests := []struct {
		name            string
		owner           workerconfig.OwnerConfig // nil = org owner (shard worker)
		input           *pb.RevokeIdentityInput
		setupSecretMock func(*secretsmocks.MockSecretProvider)
		setupClientMock func(*certauthmocks.MockClient)
		errSubstr       string
	}{
		{
			name: "success - org CA",
			input: &pb.RevokeIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "bootstrap-admin",
				Reason:       "superseded",
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(adminHydrationSecret(), nil)
			},
			setupClientMock: func(m *certauthmocks.MockClient) {
				m.EXPECT().
					Revoke(mock.Anything, mock.MatchedBy(func(req *api.RevocationRequest) bool {
						return req.Name == "bootstrap-admin" && req.Reason == "superseded"
					})).
					Return(&api.RevocationResponse{}, nil)
			},
		},
		{
			name:  "success - tls CA on the operator worker",
			owner: externalOperatorOwner,
			input: &pb.RevokeIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				IdentityName: "bootstrap-admin-tls",
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(adminHydrationSecret(), nil)
			},
			setupClientMock: func(m *certauthmocks.MockClient) {
				m.EXPECT().
					Revoke(mock.Anything, mock.MatchedBy(func(req *api.RevocationRequest) bool {
						return req.Name == "bootstrap-admin-tls"
					})).
					Return(&api.RevocationResponse{}, nil)
			},
		},
		{
			name: "error - missing ca_ref",
			input: &pb.RevokeIdentityInput{
				IdentityName: "bootstrap-admin",
			},
			errSubstr: "ca_ref is required",
		},
		{
			name: "error - client hydration secret fetch fails",
			input: &pb.RevokeIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "bootstrap-admin",
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to get CA client",
		},
		{
			name: "error - revocation fails",
			input: &pb.RevokeIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "bootstrap-admin",
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, mock.Anything).
					Return(adminHydrationSecret(), nil)
			},
			setupClientMock: func(m *certauthmocks.MockClient) {
				m.EXPECT().
					Revoke(mock.Anything, mock.Anything).
					Return(nil, errors.New("identity not found"))
			},
			errSubstr: "revocation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockSecretProvider := secretsmocks.NewMockSecretProvider(t)
			if tt.setupSecretMock != nil {
				tt.setupSecretMock(mockSecretProvider)
			}
			mockClient := certauthmocks.NewMockClient(t)
			if tt.setupClientMock != nil {
				tt.setupClientMock(mockClient)
			}

			owner := tt.owner
			if owner == nil {
				owner = externalOrgOwner
			}
			activities := newTestActivities(mockSecretProvider, mockClient, owner)
			env.RegisterActivity(activities.RevokeIdentity)

			val, err := env.ExecuteActivity(activities.RevokeIdentity, tt.input)

			if tt.errSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error from ExecuteActivity: %v", err)
			}
			var result *pb.RevokeIdentityOutput
			if err := val.Get(&result); err != nil {
				t.Fatalf("unexpected error from Get: %v", err)
			}
			if !result.GetRevoked() {
				t.Error("Revoked = false, want true")
			}
		})
	}
}

// EnrollIdentity's success path requires a live Fabric CA (the bootstrap client
// is constructed directly, not injected), so these tests cover the validation
// and secret-handling error paths that precede client construction.
func TestActivities_EnrollIdentity(t *testing.T) {
	storage := &fabricpb.IdentityRef{
		KeyRef:  secrets.NewVaultRef("test/keys/identity"),
		CertRef: secrets.NewVaultRef("test/certs/identity"),
	}

	tests := []struct {
		name            string
		input           *pb.EnrollIdentityInput
		setupSecretMock func(*secretsmocks.MockSecretProvider)
		errSubstr       string
	}{
		{
			name: "error - missing ca_ref",
			input: &pb.EnrollIdentityInput{
				IdentityName: "peer0",
				Storage:      storage,
			},
			errSubstr: "ca_ref is required",
		},
		{
			name: "error - missing storage",
			input: &pb.EnrollIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "peer0",
			},
			errSubstr: "storage is required",
		},
		{
			name: "error - key secret fetch fails",
			input: &pb.EnrollIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "peer0",
				Storage:      storage,
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/keys/identity")).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to fetch secret",
		},
		{
			name: "error - key secret missing private key",
			input: &pb.EnrollIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "peer0",
				Storage:      storage,
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/keys/identity")).
					Return(secrets.Secret{"public_key": "only-public"}, nil)
			},
			errSubstr: "failed to parse key pair",
		},
		{
			name: "error - CA TLS cert fetch fails",
			input: &pb.EnrollIdentityInput{
				CaRef:        &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				IdentityName: "peer0",
				Storage:      storage,
				CaTlsCertRef: secrets.NewVaultRef("test/certs/ca-tls"),
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/keys/identity")).
					Return(adminHydrationSecret(), nil)
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/certs/ca-tls")).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to fetch TLS certificate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockSecretProvider := secretsmocks.NewMockSecretProvider(t)
			if tt.setupSecretMock != nil {
				tt.setupSecretMock(mockSecretProvider)
			}

			activities := newTestActivities(mockSecretProvider, certauthmocks.NewMockClient(t), externalOrgOwner)
			env.RegisterActivity(activities.EnrollIdentity)

			_, err := env.ExecuteActivity(activities.EnrollIdentity, tt.input)

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Fatalf("expected error containing %q, got: %v", tt.errSubstr, err)
			}
		})
	}
}

// BootstrapEnrollIdentity's success path requires a live Fabric CA (the
// bootstrap client is constructed directly, not injected), so these tests cover
// the validation and secret-handling error paths that precede client construction.
func TestActivities_BootstrapEnrollIdentity(t *testing.T) {
	storage := &fabricpb.IdentityRef{
		KeyRef:  secrets.NewVaultRef("test/keys/bootstrap-admin"),
		CertRef: secrets.NewVaultRef("test/certs/bootstrap-admin"),
	}
	credentialsRef := secrets.NewVaultRef("test/creds/bootstrap-admin")

	tests := []struct {
		name            string
		input           *pb.BootstrapEnrollIdentityInput
		setupSecretMock func(*secretsmocks.MockSecretProvider)
		errSubstr       string
	}{
		{
			name: "error - missing storage",
			input: &pb.BootstrapEnrollIdentityInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
			},
			errSubstr: "storage is required",
		},
		{
			name: "error - key secret fetch fails",
			input: &pb.BootstrapEnrollIdentityInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
				Storage:                 storage,
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/keys/bootstrap-admin")).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to fetch secret",
		},
		{
			name: "error - key secret missing private key",
			input: &pb.BootstrapEnrollIdentityInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
				Storage:                 storage,
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/keys/bootstrap-admin")).
					Return(secrets.Secret{"public_key": "only-public"}, nil)
			},
			errSubstr: "failed to parse key pair",
		},
		{
			name: "error - bootstrap credentials fetch fails",
			input: &pb.BootstrapEnrollIdentityInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
				Storage:                 storage,
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/keys/bootstrap-admin")).
					Return(adminHydrationSecret(), nil)
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/creds/bootstrap-admin")).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to fetch bootstrap credentials",
		},
		{
			name: "error - bootstrap credentials missing username",
			input: &pb.BootstrapEnrollIdentityInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
				Storage:                 storage,
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/keys/bootstrap-admin")).
					Return(adminHydrationSecret(), nil)
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/creds/bootstrap-admin")).
					Return(secrets.Secret{"password": "only-password"}, nil)
			},
			errSubstr: "failed to parse credentials",
		},
		{
			name: "error - issuer cert fetch fails",
			input: &pb.BootstrapEnrollIdentityInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
				Storage:                 storage,
				IssuerCertRef:           secrets.NewVaultRef("test/certs/issuer"),
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/keys/bootstrap-admin")).
					Return(adminHydrationSecret(), nil)
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/certs/issuer")).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to fetch issuer cert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockSecretProvider := secretsmocks.NewMockSecretProvider(t)
			if tt.setupSecretMock != nil {
				tt.setupSecretMock(mockSecretProvider)
			}

			activities := newTestActivities(mockSecretProvider, certauthmocks.NewMockClient(t), externalOrgOwner)
			env.RegisterActivity(activities.BootstrapEnrollIdentity)

			_, err := env.ExecuteActivity(activities.BootstrapEnrollIdentity, tt.input)

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Fatalf("expected error containing %q, got: %v", tt.errSubstr, err)
			}
		})
	}
}
