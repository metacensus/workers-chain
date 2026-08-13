package temporal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/stretchr/testify/require"
)

func TestLoadTLSConfig_PlaintextWhenUnset(t *testing.T) {
	t.Parallel()
	cfg, err := loadTLSConfig(&config.TemporalConfig{})
	require.NoError(t, err)
	require.Nil(t, cfg)
}

func TestLoadTLSConfig_RequiresAllPaths(t *testing.T) {
	t.Parallel()
	_, err := loadTLSConfig(&config.TemporalConfig{TLSCAPath: "x"})
	require.Error(t, err)
}

func TestLoadTLSConfig_LoadsFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	caCert, caKey := mustSelfSigned(t, "ca", true, nil, nil)
	leafCert, leafKey := mustSelfSigned(t, "infra", false, caCert, caKey)

	caPath := filepath.Join(dir, "ca.crt")
	certPath := filepath.Join(dir, "tls.crt")
	keyPath := filepath.Join(dir, "tls.key")
	require.NoError(t, os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw}), 0o600))
	require.NoError(t, os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafCert.Raw}), 0o600))
	require.NoError(t, os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(leafKey)}), 0o600))

	tlsCfg, err := loadTLSConfig(&config.TemporalConfig{
		TLSCAPath:     caPath,
		TLSCertPath:   certPath,
		TLSKeyPath:    keyPath,
		TLSServerName: "temporal-frontend",
	})
	require.NoError(t, err)
	require.NotNil(t, tlsCfg)
	require.Equal(t, "temporal-frontend", tlsCfg.ServerName)
	require.Len(t, tlsCfg.Certificates, 1)
}

func mustSelfSigned(t *testing.T, cn string, isCA bool, parent *x509.Certificate, parentKey *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}
	if isCA {
		tmpl.KeyUsage |= x509.KeyUsageCertSign
	}
	signerCert, signerKey := tmpl, key
	if parent != nil {
		signerCert, signerKey = parent, parentKey
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, signerCert, &key.PublicKey, signerKey)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert, key
}
