package kube

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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// EnsureTemporalMTLSClientCert writes TargetNamespace/temporal-mtls-client.
// fabric copies the TF-issued secret; shards get a CA-signed leaf. Material stays in the activity.
func (a *Activities) EnsureTemporalMTLSClientCert(ctx context.Context, in *EnsureTemporalMTLSInput) (*EnsureTemporalMTLSOutput, error) {
	if in == nil || in.TargetNamespace == "" || in.TemporalNamespace == "" {
		return nil, fmt.Errorf("target namespace and temporal namespace are required")
	}

	var data map[string][]byte
	var err error
	if in.TemporalNamespace == "fabric" {
		data, err = a.client.GetSecretData(ctx, temporalSystemNamespace, sourceFabricClientSecret)
	} else {
		data, err = a.issueClientCert(ctx, in.TemporalNamespace)
	}
	if err != nil {
		return nil, err
	}

	if err := a.client.UpsertOpaqueSecret(ctx, in.TargetNamespace, WorkerTemporalMTLSSecretName, data); err != nil {
		return nil, err
	}
	return &EnsureTemporalMTLSOutput{}, nil
}

func (a *Activities) issueClientCert(ctx context.Context, cn string) (map[string][]byte, error) {
	caSecret, err := a.client.GetSecretData(ctx, temporalSystemNamespace, sourceClientCASecret)
	if err != nil {
		return nil, fmt.Errorf("get client CA secret: %w", err)
	}
	caCertPEM := caSecret["ca.crt"]
	caKeyPEM := caSecret["ca.key"]
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

func (c *kubeClient) GetSecretData(ctx context.Context, namespace, name string) (map[string][]byte, error) {
	src, err := c.typedClient.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get secret %s/%s: %w", namespace, name, err)
	}
	out := make(map[string][]byte, len(src.Data))
	for k, v := range src.Data {
		cp := make([]byte, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out, nil
}

func (c *kubeClient) UpsertOpaqueSecret(ctx context.Context, namespace, name string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/component":  "temporal-mtls",
				"app.kubernetes.io/managed-by": "metacensus",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
	_, err := c.typedClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		_, err = c.typedClient.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	}
	if err != nil {
		return fmt.Errorf("upsert secret %s/%s: %w", namespace, name, err)
	}
	return nil
}
