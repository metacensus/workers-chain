package certificate

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	GenerateCertificateWorkflowName = "certificate.GenerateCertificateWorkflow"
	GenerateCertificateActivityName = "certificate.GenerateCertificateActivity"
)

// CertificateRequest represents a request to generate a certificate
type CertificateRequest struct {
	Name         string `json:"name"`         // Name of the certificate (e.g., "orderer1", "peer2")
	Namespace    string `json:"namespace"`    // Kubernetes namespace
	Type         string `json:"type"`         // Type: "orderer", "peer", "ca"
	SecretName   string `json:"secret_name"`  // Kubernetes secret name
	CommonName   string `json:"common_name"`  // Certificate common name
	Organization string `json:"organization"` // Organization name
	Country      string `json:"country"`      // Country code
	Province     string `json:"province"`     // Province/State
	Locality     string `json:"locality"`     // City/Locality
	ValidDays    int    `json:"valid_days"`   // Certificate validity in days
}

// CertificateResult represents the result of certificate generation
type CertificateResult struct {
	SecretName string `json:"secret_name"`
	Namespace  string `json:"namespace"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// GenerateCertificateWorkflow generates a single certificate and stores it as a Kubernetes secret
func GenerateCertificateWorkflow(ctx workflow.Context, req CertificateRequest) (*CertificateResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Certificate generation workflow started",
		"name", req.Name,
		"type", req.Type,
		"namespace", req.Namespace)

	var result CertificateResult
	err := workflow.ExecuteActivity(ctx, GenerateCertificateActivity, req).Get(ctx, &result)
	if err != nil {
		logger.Error("Certificate generation activity failed", "error", err)
		return nil, fmt.Errorf("certificate generation failed: %w", err)
	}

	logger.Info("Certificate generation workflow completed",
		"secret_name", result.SecretName,
		"status", result.Status)

	return &result, nil
}

// GenerateCertificateActivity generates a certificate and creates a Kubernetes secret
func GenerateCertificateActivity(ctx context.Context, req CertificateRequest) (*CertificateResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting certificate generation",
		"name", req.Name,
		"type", req.Type,
		"namespace", req.Namespace)

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   req.CommonName,
			Organization: []string{req.Organization},
			Country:      []string{req.Country},
			Province:     []string{req.Province},
			Locality:     []string{req.Locality},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, req.ValidDays),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{req.CommonName},
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate and private key to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create Kubernetes secret
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create secret data
	secretData := map[string][]byte{
		"tls.crt": certPEM,
		"tls.key": privKeyPEM,
		"ca.crt":  certPEM, // For self-signed certificates, use the same cert as CA
	}

	// Create or update the secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.SecretName,
			Namespace: req.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: secretData,
	}

	_, err = clientset.CoreV1().Secrets(req.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		// If secret already exists, update it
		if strings.Contains(err.Error(), "already exists") {
			_, err = clientset.CoreV1().Secrets(req.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to update existing secret: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to create secret: %w", err)
		}
	}

	logger.Info("Certificate generated and secret created successfully",
		"secret_name", req.SecretName,
		"namespace", req.Namespace)

	return &CertificateResult{
		SecretName: req.SecretName,
		Namespace:  req.Namespace,
		Status:     "success",
		Message:    "Certificate generated and secret created successfully",
	}, nil
}
