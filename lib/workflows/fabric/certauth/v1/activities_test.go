package certauth_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-ca/api"
	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/certauth/v1"
	fabricpb "github.com/metacensus/workers-chain/gen/workflows/fabric/v1"
	secretspb "github.com/metacensus/workers-chain/gen/workflows/secrets/v1"
	workerconfig "github.com/metacensus/workers-chain/lib/workers/config"
	"github.com/metacensus/workers-chain/lib/workflows/common"
	certauth "github.com/metacensus/workers-chain/lib/workflows/fabric/certauth/v1"
	secrets "github.com/metacensus/workers-chain/lib/workflows/secrets/v1"
	certauthmocks "github.com/metacensus/workers-chain/mocks/workflows/fabric/certauth/v1"
	secretsmocks "github.com/metacensus/workers-chain/mocks/workflows/secrets/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

// The success paths of BootstrapEnrollCsr and SignCsr construct a
// BootstrapClient directly and speak HTTP to a live Fabric CA, so they are
// covered by integration tests rather than unit tests here. Revoke goes through
// the MultiClientProvider, which lets us inject a mocked Client.

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

// BootstrapEnrollCsr's success path requires a live Fabric CA (the bootstrap
// client is constructed directly, not injected), so these tests cover the
// validation and secret-handling error paths that precede client construction.
// It performs NO register step and takes a pre-built CSR, so there is no key
// fetch - only the issuer-chain and bootstrap-credential fetches.
func TestActivities_BootstrapEnrollCsr(t *testing.T) {
	credentialsRef := secrets.NewVaultRef("test/creds/bootstrap-admin")
	csrPEM := []byte("-----BEGIN CERTIFICATE REQUEST-----\nfake\n-----END CERTIFICATE REQUEST-----")

	tests := []struct {
		name            string
		input           *pb.BootstrapEnrollCsrInput
		setupSecretMock func(*secretsmocks.MockSecretProvider)
		errSubstr       string
	}{
		{
			name: "error - missing csr_pem",
			input: &pb.BootstrapEnrollCsrInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
			},
			errSubstr: "csr_pem is required",
		},
		{
			name: "error - issuer cert fetch fails",
			input: &pb.BootstrapEnrollCsrInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
				CsrPem:                  csrPEM,
				IssuerCertRef:           secrets.NewVaultRef("test/certs/issuer"),
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/certs/issuer")).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to fetch issuer cert",
		},
		{
			name: "error - bootstrap credentials fetch fails",
			input: &pb.BootstrapEnrollCsrInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
				CsrPem:                  csrPEM,
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/creds/bootstrap-admin")).
					Return(nil, errVaultUnavailable)
			},
			errSubstr: "failed to fetch bootstrap credentials",
		},
		{
			name: "error - bootstrap credentials missing username",
			input: &pb.BootstrapEnrollCsrInput{
				CaUrl:                   "https://ca.example.com:7054",
				BootstrapCredentialsRef: credentialsRef,
				CsrPem:                  csrPEM,
			},
			setupSecretMock: func(m *secretsmocks.MockSecretProvider) {
				m.EXPECT().
					FetchSecret(mock.Anything, refWithPath("test/creds/bootstrap-admin")).
					Return(secrets.Secret{"password": "only-password"}, nil)
			},
			errSubstr: "failed to parse credentials",
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
			env.RegisterActivity(activities.BootstrapEnrollCsr)

			_, err := env.ExecuteActivity(activities.BootstrapEnrollCsr, tt.input)

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Fatalf("expected error containing %q, got: %v", tt.errSubstr, err)
			}
		})
	}
}
