package temporal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/metacensus/infra/temporal/lib/workers/config"
	"go.temporal.io/sdk/client"
)

// NewClient creates a Temporal client from TemporalConfig
func NewClient(cfg *config.TemporalConfig) (client.Client, error) {
	opts, err := clientOptions(cfg)
	if err != nil {
		return nil, err
	}
	return client.Dial(opts)
}

// NewNamespaceClient creates a Temporal namespace client for namespace management operations
func NewNamespaceClient(cfg *config.TemporalConfig) (client.NamespaceClient, error) {
	opts, err := clientOptions(cfg)
	if err != nil {
		return nil, err
	}
	return client.NewNamespaceClient(opts)
}

func clientOptions(cfg *config.TemporalConfig) (client.Options, error) {
	opts := client.Options{
		HostPort:  cfg.Address,
		Namespace: cfg.Namespace,
	}
	tlsCfg, err := loadTLSConfig(cfg)
	if err != nil {
		return opts, err
	}
	if tlsCfg != nil {
		opts.ConnectionOptions = client.ConnectionOptions{TLS: tlsCfg}
	}
	return opts, nil
}

// loadTLSConfig returns nil when TLS paths are unset (plaintext dial).
func loadTLSConfig(cfg *config.TemporalConfig) (*tls.Config, error) {
	if cfg.TLSCertPath == "" && cfg.TLSKeyPath == "" && cfg.TLSCAPath == "" {
		return nil, nil
	}
	if cfg.TLSCertPath == "" || cfg.TLSKeyPath == "" || cfg.TLSCAPath == "" {
		return nil, fmt.Errorf("TEMPORAL_TLS_CA_PATH, TEMPORAL_TLS_CERT_PATH, and TEMPORAL_TLS_KEY_PATH must all be set together")
	}

	cert, err := tls.LoadX509KeyPair(cfg.TLSCertPath, cfg.TLSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load temporal client certificate: %w", err)
	}
	caPEM, err := os.ReadFile(cfg.TLSCAPath)
	if err != nil {
		return nil, fmt.Errorf("read temporal CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("parse temporal CA PEM")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   cfg.TLSServerName,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
