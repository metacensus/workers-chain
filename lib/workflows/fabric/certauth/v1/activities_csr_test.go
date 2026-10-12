package certauth_test

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-ca/api"
	pb "github.com/metacensus/workers-chain/gen/workflows/fabric/certauth/v1"
	fabricpb "github.com/metacensus/workers-chain/gen/workflows/fabric/v1"
	certauth "github.com/metacensus/workers-chain/lib/workflows/fabric/certauth/v1"
	secrets "github.com/metacensus/workers-chain/lib/workflows/secrets/v1"
	certauthmocks "github.com/metacensus/workers-chain/mocks/workflows/fabric/certauth/v1"
	secretsmocks "github.com/metacensus/workers-chain/mocks/workflows/secrets/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

// keyPairSecret produces a real ECDSA key pair secret so GenerateCSR can run
// actual crypto against it.
func keyPairSecret(t *testing.T) secrets.Secret {
	t.Helper()
	kp, err := secrets.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}
	return secrets.FromKeyPair(kp)
}

// GenerateCsr reads the private key ensured up-front by secrets.EnsureKeyPair
// and builds a CSR from it. It never generates or overwrites the key (that is
// EnsureKeyPair's job), so these tests assert fetch-then-build behavior only.
func TestActivities_GenerateCsr(t *testing.T) {
	keyRef := secrets.NewVaultRef("fabric/shard/testorg/keys/ca/tls/int/tls")

	t.Run("existing key - builds and returns a CSR without writing", func(t *testing.T) {
		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestActivityEnvironment()

		m := secretsmocks.NewMockSecretProvider(t)
		m.EXPECT().FetchSecret(mock.Anything, refWithPath(keyRef.GetRef().GetFields()["path"].GetStringValue())).
			Return(keyPairSecret(t), nil)
		// No UpdateSecret expectation: GenerateCsr must never write the key.

		activities := newTestActivities(m, certauthmocks.NewMockClient(t), externalOrgOwner)
		env.RegisterActivity(activities.GenerateCsr)

		val, err := env.ExecuteActivity(activities.GenerateCsr, &pb.GenerateCsrInput{
			KeyRef:     keyRef,
			CommonName: "peer0@testorg",
			Hosts:      []string{"peer0.testorg"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var out *pb.GenerateCsrOutput
		if err := val.Get(&out); err != nil {
			t.Fatalf("get result: %v", err)
		}
		assertCsr(t, out.GetCsrPem(), "peer0@testorg", []string{"peer0.testorg"})
	})

	t.Run("fetch error - surfaces failure, does not write", func(t *testing.T) {
		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestActivityEnvironment()

		m := secretsmocks.NewMockSecretProvider(t)
		m.EXPECT().FetchSecret(mock.Anything, mock.Anything).Return(nil, errVaultUnavailable)

		activities := newTestActivities(m, certauthmocks.NewMockClient(t), externalOrgOwner)
		env.RegisterActivity(activities.GenerateCsr)

		_, err := env.ExecuteActivity(activities.GenerateCsr, &pb.GenerateCsrInput{
			KeyRef:     keyRef,
			CommonName: "peer0@testorg",
		})
		if err == nil || !strings.Contains(err.Error(), "failed to fetch key") {
			t.Fatalf("expected fetch-key error, got: %v", err)
		}
	})

	t.Run("validation - missing key_ref", func(t *testing.T) {
		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestActivityEnvironment()
		activities := newTestActivities(secretsmocks.NewMockSecretProvider(t), certauthmocks.NewMockClient(t), externalOrgOwner)
		env.RegisterActivity(activities.GenerateCsr)
		_, err := env.ExecuteActivity(activities.GenerateCsr, &pb.GenerateCsrInput{
			CommonName: "peer0@testorg",
		})
		if err == nil || !strings.Contains(err.Error(), "key_ref is required") {
			t.Fatalf("expected key_ref validation error, got: %v", err)
		}
	})

	t.Run("validation - missing common name", func(t *testing.T) {
		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestActivityEnvironment()
		activities := newTestActivities(secretsmocks.NewMockSecretProvider(t), certauthmocks.NewMockClient(t), externalOrgOwner)
		env.RegisterActivity(activities.GenerateCsr)
		_, err := env.ExecuteActivity(activities.GenerateCsr, &pb.GenerateCsrInput{
			KeyRef: keyRef,
		})
		if err == nil || !strings.Contains(err.Error(), "common_name is required") {
			t.Fatalf("expected common_name validation error, got: %v", err)
		}
	})
}

// SignCsr's enrollment step constructs a BootstrapClient directly and speaks HTTP
// to a live Fabric CA, so the full success path is covered by integration tests.
// These unit tests cover validation and the register / 409->ModifyIdentity logic
// that precedes enrollment.
func TestActivities_SignCsr(t *testing.T) {
	caRef := &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"}
	validInput := func() *pb.SignCsrInput {
		return &pb.SignCsrInput{
			CaRef:         caRef,
			IdentityName:  "peer0@testorg",
			IdentityType:  "peer",
			CsrPem:        []byte("-----BEGIN CERTIFICATE REQUEST-----\nfake\n-----END CERTIFICATE REQUEST-----"),
			IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			CaTlsCertRef:  secrets.NewVaultRef("test/certs/ca-tls"),
		}
	}

	t.Run("validation - missing ca_ref", func(t *testing.T) {
		err := runSignCsr(t, &pb.SignCsrInput{IdentityName: "x", CsrPem: []byte("y")}, nil, nil)
		if err == nil || !strings.Contains(err.Error(), "ca_ref is required") {
			t.Fatalf("expected ca_ref error, got: %v", err)
		}
	})

	t.Run("validation - missing identity_name", func(t *testing.T) {
		err := runSignCsr(t, &pb.SignCsrInput{CaRef: caRef, CsrPem: []byte("y")}, nil, nil)
		if err == nil || !strings.Contains(err.Error(), "identity_name is required") {
			t.Fatalf("expected identity_name error, got: %v", err)
		}
	})

	t.Run("validation - missing csr_pem", func(t *testing.T) {
		err := runSignCsr(t, &pb.SignCsrInput{CaRef: caRef, IdentityName: "x"}, nil, nil)
		if err == nil || !strings.Contains(err.Error(), "csr_pem is required") {
			t.Fatalf("expected csr_pem error, got: %v", err)
		}
	})

	t.Run("register non-409 error - fails without ModifyIdentity", func(t *testing.T) {
		setupSecret := func(m *secretsmocks.MockSecretProvider) {
			m.EXPECT().FetchSecret(mock.Anything, mock.Anything).Return(adminHydrationSecret(), nil)
		}
		setupClient := func(m *certauthmocks.MockClient) {
			m.EXPECT().Register(mock.Anything, mock.Anything).Return(nil, errors.New("boom"))
			// ModifyIdentity must NOT be called for a non-duplicate error.
		}
		err := runSignCsr(t, validInput(), setupSecret, setupClient)
		if err == nil || !strings.Contains(err.Error(), "registration failed") {
			t.Fatalf("expected registration error, got: %v", err)
		}
	})

	t.Run("register 409 - resets secret via ModifyIdentity", func(t *testing.T) {
		setupSecret := func(m *secretsmocks.MockSecretProvider) {
			m.EXPECT().FetchSecret(mock.Anything, mock.Anything).Return(adminHydrationSecret(), nil)
		}
		setupClient := func(m *certauthmocks.MockClient) {
			m.EXPECT().Register(mock.Anything, mock.Anything).
				Return(nil, errors.New("Response from server: Error Code: 74 - Identity 'peer0@testorg' is already registered"))
			// The 409 path MUST reset the enrollment secret before enrolling.
			m.EXPECT().ModifyIdentity(mock.Anything, mock.MatchedBy(func(req *api.ModifyIdentityRequest) bool {
				return req.ID == "peer0@testorg" && req.Secret != ""
			})).Return(&api.IdentityResponse{}, nil)
		}
		// Enrollment then runs against a non-existent CA and fails; the point of
		// this test is that ModifyIdentity was invoked (asserted by the mock).
		err := runSignCsr(t, validInput(), setupSecret, setupClient)
		if err == nil {
			t.Fatal("expected enrollment to fail against a nonexistent CA, got nil")
		}
	})
}

// runSignCsr executes the SignCsr activity with the given mocks and returns the error.
func runSignCsr(
	t *testing.T,
	input *pb.SignCsrInput,
	setupSecret func(*secretsmocks.MockSecretProvider),
	setupClient func(*certauthmocks.MockClient),
) error {
	t.Helper()
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	m := secretsmocks.NewMockSecretProvider(t)
	if setupSecret != nil {
		setupSecret(m)
	}
	client := certauthmocks.NewMockClient(t)
	if setupClient != nil {
		setupClient(client)
	}

	activities := newTestActivities(m, client, externalOperatorOwner)
	env.RegisterActivity(activities.SignCsr)

	_, err := env.ExecuteActivity(activities.SignCsr, input)
	return err
}

// assertCsr parses a PEM CSR and checks its subject CN and DNS SANs.
func assertCsr(t *testing.T, csrPEM []byte, wantCN string, wantDNS []string) {
	t.Helper()
	if len(csrPEM) == 0 {
		t.Fatal("empty CSR")
	}
	block, _ := pem.Decode(csrPEM)
	if block == nil {
		t.Fatal("CSR is not valid PEM")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse CSR: %v", err)
	}
	if csr.Subject.CommonName != wantCN {
		t.Errorf("CSR CN = %q, want %q", csr.Subject.CommonName, wantCN)
	}
	if len(wantDNS) > 0 && strings.Join(csr.DNSNames, ",") != strings.Join(wantDNS, ",") {
		t.Errorf("CSR DNSNames = %v, want %v", csr.DNSNames, wantDNS)
	}
}

// TestGenerateCSR covers the standalone key-owner CSR builder used by
// GenerateCsr: it must produce a valid CSR bound to the supplied key.
func TestGenerateCSR(t *testing.T) {
	kp, err := secrets.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	t.Run("success - CN and hosts", func(t *testing.T) {
		csrPEM, err := certauth.GenerateCSR(kp.PrivateKey, &api.CSRInfo{
			CN:    "peer0@testorg",
			Hosts: []string{"peer0.testorg", "localhost"},
		})
		if err != nil {
			t.Fatalf("GenerateCSR failed: %v", err)
		}
		assertCsr(t, csrPEM, "peer0@testorg", []string{"peer0.testorg", "localhost"})
	})

	t.Run("error - empty key", func(t *testing.T) {
		if _, err := certauth.GenerateCSR(nil, &api.CSRInfo{CN: "x"}); err == nil {
			t.Fatal("expected error for empty key, got nil")
		}
	})

	t.Run("error - nil csr info", func(t *testing.T) {
		if _, err := certauth.GenerateCSR(kp.PrivateKey, nil); err == nil {
			t.Fatal("expected error for nil csr info, got nil")
		}
	})
}
