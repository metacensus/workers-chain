package temporalmtls

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	kubemocks "github.com/metacensus/library-temporal-kube/mocks/kube/v1"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
)

var errSecretGone = errors.New("secret not found")

func TestActivities_EnsureTemporalMTLSClientCert(t *testing.T) {
	caCertPEM, caKeyPEM := testCA(t)
	fabricData := map[string][]byte{"ca.crt": []byte("ca"), "tls.crt": []byte("crt"), "tls.key": []byte("key")}
	tests := []struct {
		name      string
		in        *EnsureTemporalMTLSInput
		setupMock func(m *kubemocks.MockClient, applied **corev1.Secret)
		wantErr   bool
		check     func(t *testing.T, applied *corev1.Secret)
	}{
		{
			name:    "error - namespaces are required",
			in:      &EnsureTemporalMTLSInput{TargetNamespace: "fabric"},
			wantErr: true,
		},
		{
			name: "success - fabric copies the issued client secret",
			in:   &EnsureTemporalMTLSInput{TargetNamespace: "fabric", TemporalNamespace: "fabric"},
			setupMock: func(m *kubemocks.MockClient, applied **corev1.Secret) {
				m.EXPECT().GetSecret(mock.Anything, "temporal", "temporal-mtls-client-fabric").Return(&corev1.Secret{Data: fabricData}, nil)
				m.EXPECT().ApplySecret(mock.Anything, mock.Anything, "workers-chain").RunAndReturn(func(_ context.Context, s *corev1.Secret, _ string) error {
					*applied = s
					return nil
				})
			},
			check: func(t *testing.T, applied *corev1.Secret) {
				if applied.Name != WorkerTemporalMTLSSecretName || applied.Namespace != "fabric" || string(applied.Data["tls.crt"]) != "crt" {
					t.Fatalf("unexpected secret: %v", applied.ObjectMeta)
				}
			},
		},
		{
			name: "success - a shard gets a leaf the client CA signed",
			in:   &EnsureTemporalMTLSInput{TargetNamespace: "fabric-shard-org1", TemporalNamespace: "fabric-shard-org1"},
			setupMock: func(m *kubemocks.MockClient, applied **corev1.Secret) {
				m.EXPECT().GetSecret(mock.Anything, "temporal", "temporal-mtls-client-ca").Return(&corev1.Secret{Data: map[string][]byte{"ca.crt": caCertPEM, "ca.key": caKeyPEM}}, nil)
				m.EXPECT().ApplySecret(mock.Anything, mock.Anything, "workers-chain").RunAndReturn(func(_ context.Context, s *corev1.Secret, _ string) error {
					*applied = s
					return nil
				})
			},
			check: func(t *testing.T, applied *corev1.Secret) {
				block, _ := pem.Decode(applied.Data["tls.crt"])
				leaf, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					t.Fatalf("parse leaf: %v", err)
				}
				pool := x509.NewCertPool()
				pool.AppendCertsFromPEM(applied.Data["ca.crt"])
				if _, err := leaf.Verify(x509.VerifyOptions{Roots: pool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}); err != nil {
					t.Fatalf("leaf does not verify against the CA: %v", err)
				}
				if leaf.Subject.CommonName != "fabric-shard-org1" {
					t.Fatalf("CN %q", leaf.Subject.CommonName)
				}
				if keyBlock, _ := pem.Decode(applied.Data["tls.key"]); keyBlock == nil {
					t.Fatal("tls.key is not PEM")
				}
			},
		},
		{
			name: "error - the CA secret lacks its key",
			in:   &EnsureTemporalMTLSInput{TargetNamespace: "fabric-shard-org1", TemporalNamespace: "fabric-shard-org1"},
			setupMock: func(m *kubemocks.MockClient, _ **corev1.Secret) {
				m.EXPECT().GetSecret(mock.Anything, "temporal", "temporal-mtls-client-ca").Return(&corev1.Secret{Data: map[string][]byte{"ca.crt": caCertPEM}}, nil)
			},
			wantErr: true,
		},
		{
			name: "error - the source secret cannot be read",
			in:   &EnsureTemporalMTLSInput{TargetNamespace: "fabric", TemporalNamespace: "fabric"},
			setupMock: func(m *kubemocks.MockClient, _ **corev1.Secret) {
				m.EXPECT().GetSecret(mock.Anything, "temporal", "temporal-mtls-client-fabric").Return(nil, errSecretGone)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := kubemocks.NewMockClient(t)
			var applied *corev1.Secret
			if tt.setupMock != nil {
				tt.setupMock(m, &applied)
			}
			_, err := NewActivities(m).EnsureTemporalMTLSClientCert(context.Background(), tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got %v", tt.wantErr, err)
			}
			if tt.check != nil {
				tt.check(t, applied)
			}
		})
	}
}

// testCA is a throwaway RSA CA, PEM-encoded as the cluster stores it.
func testCA(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "temporal-client-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}
