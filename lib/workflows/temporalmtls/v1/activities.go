// Package temporalmtls provisions the Temporal mTLS client secret that the
// worker charts mount. Material moves between Secrets inside the cluster,
// through kube.Client, and never enters an activity payload: that is why this
// is a plain activity over the client and not a proto one.
package temporalmtls

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/metacensus/library-temporal-kube/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/metacensus/workers-chain/lib/workers/config"
)

const (
	WorkerTemporalMTLSSecretName = "temporal-mtls-client" // pragma: allowlist secret

	temporalSystemNamespace  = "temporal"
	sourceFabricClientSecret = "temporal-mtls-client-fabric" // pragma: allowlist secret
	sourceClientCASecret     = "temporal-mtls-client-ca"     // pragma: allowlist secret
)

type EnsureTemporalMTLSInput struct {
	TargetNamespace   string
	TemporalNamespace string
}

type EnsureTemporalMTLSOutput struct{}

// Activities holds the cluster client of the worker that hosts it.
type Activities struct {
	client kube.Client
}

func NewActivities(client kube.Client) *Activities {
	return &Activities{client: client}
}

// EnsureTemporalMTLSClientCert writes TargetNamespace/temporal-mtls-client.
// fabric copies the TF-issued secret; shards get a CA-signed leaf.
func (a *Activities) EnsureTemporalMTLSClientCert(ctx context.Context, in *EnsureTemporalMTLSInput) (*EnsureTemporalMTLSOutput, error) {
	if in == nil || in.TargetNamespace == "" || in.TemporalNamespace == "" {
		return nil, fmt.Errorf("target namespace and temporal namespace are required")
	}

	var data map[string][]byte
	if in.TemporalNamespace == "fabric" {
		src, err := a.client.GetSecret(ctx, temporalSystemNamespace, sourceFabricClientSecret)
		if err != nil {
			return nil, fmt.Errorf("get secret %s/%s: %w", temporalSystemNamespace, sourceFabricClientSecret, err)
		}
		data = src.Data
	} else {
		var err error
		if data, err = a.issueClientCert(ctx, in.TemporalNamespace); err != nil {
			return nil, err
		}
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      WorkerTemporalMTLSSecretName,
			Namespace: in.TargetNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "temporal-mtls",
				"app.kubernetes.io/managed-by": "metacensus",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
	if err := a.client.ApplySecret(ctx, secret, config.FieldManager); err != nil {
		return nil, fmt.Errorf("apply secret %s/%s: %w", in.TargetNamespace, WorkerTemporalMTLSSecretName, err)
	}
	return &EnsureTemporalMTLSOutput{}, nil
}

func (a *Activities) issueClientCert(ctx context.Context, cn string) (map[string][]byte, error) {
	caSecret, err := a.client.GetSecret(ctx, temporalSystemNamespace, sourceClientCASecret)
	if err != nil {
		return nil, fmt.Errorf("get client CA secret: %w", err)
	}
	caCertPEM := caSecret.Data["ca.crt"]
	caKeyPEM := caSecret.Data["ca.key"]
	if len(caCertPEM) == 0 || len(caKeyPEM) == 0 {
		return nil, fmt.Errorf("client CA secret missing ca.crt or ca.key")
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return nil, fmt.Errorf("decode CA cert PEM")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse CA cert: %w", err)
	}
	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return nil, fmt.Errorf("decode CA key PEM")
	}
	var caKey *rsa.PrivateKey
	if k, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes); err == nil {
		caKey = k
	} else {
		parsed, err2 := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse CA key: %w", err)
		}
		var ok bool
		caKey, ok = parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("CA key is not RSA")
		}
	}

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate leaf key: %w", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:   cn,
			Organization: []string{"MetaCensus"},
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("sign leaf cert: %w", err)
	}

	return map[string][]byte{
		"ca.crt":  caCertPEM,
		"tls.crt": pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		"tls.key": pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(leafKey)}),
	}, nil
}
