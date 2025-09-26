package certificate

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	vault "github.com/metacensus/infra/temporal/workflows/vault"
)

const (
	ProvisionPeerWorkflowName    = "certificate.ProvisionPeerWorkflow"
	ProvisionOrdererWorkflowName = "certificate.ProvisionOrdererWorkflow"
	ProvisionClientWorkflowName  = "certificate.ProvisionClientWorkflow"
)

// ProvisionPeerWorkflow provisions a complete peer with all necessary certificates
func ProvisionPeerWorkflow(ctx workflow.Context, config PeerConfig) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Peer provisioning workflow started",
		"peer_id", config.PeerID,
		"namespace", config.Namespace)

	// Register peer identity
	var reg Registration
	err := workflow.ExecuteChildWorkflow(ctx, RegisterIdentityWorkflow, IdentityParams{
		Type:       "peer",
		Name:       config.PeerID,
		CAEndpoint: config.EnrollmentCA,
		Attributes: map[string]string{
			"hf.Registrar.Roles": "peer",
			"hf.Type":            "peer",
		},
		Affiliation: "org1.department1",
		Secret:      "peerpw", // TODO: Generate secure password
	}).Get(ctx, &reg)
	if err != nil {
		logger.Error("Failed to register peer identity", "error", err)
		return fmt.Errorf("peer registration failed: %w", err)
	}

	// Parallel enrollment for signing and TLS
	var signingEnroll, tlsEnroll Enrollment

	// Signing enrollment
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.ExecuteChildWorkflow(ctx, EnrollIdentityWorkflow,
			EnrollmentParams{
				Registration: reg,
				Profile:      "sign",
				CAEndpoint:   config.EnrollmentCA,
				CSRConfig: CSRConfig{
					CommonName:   config.PeerID,
					Organization: "Metacensus",
					Country:      "US",
					Province:     "CA",
					Locality:     "San Francisco",
				},
			}).Get(ctx, &signingEnroll)
	})

	// TLS enrollment
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.ExecuteChildWorkflow(ctx, EnrollIdentityWorkflow,
			EnrollmentParams{
				Registration: reg,
				Profile:      "tls",
				CAEndpoint:   config.TLSCA,
				CSRConfig: CSRConfig{
					CommonName:   config.PeerID,
					Organization: "Metacensus",
					Country:      "US",
					Province:     "CA",
					Locality:     "San Francisco",
					Hosts: []string{
						config.PeerID,
						fmt.Sprintf("%s.%s.svc.cluster.local", config.PeerID, config.Namespace),
					},
				},
			}).Get(ctx, &tlsEnroll)
	})

	// Wait for both enrollments to complete
	workflow.Sleep(ctx, time.Second) // Give time for parallel execution

	// Assemble complete MSP
	var msp MSPBundle
	err = workflow.ExecuteChildWorkflow(ctx, AssembleMSPWorkflow, MSPParams{
		SigningEnrollment: signingEnroll,
		TLSEnrollment:     &tlsEnroll,
		NodeOUConfig:      config.OrgNodeOUs,
		CACerts:           [][]byte{}, // TODO: Get CA certificates
	}).Get(ctx, &msp)
	if err != nil {
		logger.Error("Failed to assemble MSP", "error", err)
		return fmt.Errorf("MSP assembly failed: %w", err)
	}

	// Store signing certificates
	// Create vault client and workflows for this operation
	vaultConfig := vault.GetVaultConfig()
	vaultClient := vault.NewVaultClient(vaultConfig)
	vaultWorkflows := vault.NewWorkflows(vaultClient)

	var signingSecret vault.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultWorkflows.StoreSecretWorkflow,
		vault.SecretParams{
			Name: fmt.Sprintf("peer-%s-signing", config.PeerID),
			Data: map[string][]byte{
				"cert.pem": msp.SigningCert,
				"key.pem":  msp.SigningKey,
			},
			Labels: map[string]string{
				"fabric-component": "peer",
				"organization":     "metacensus",
				"managed-by":       "temporal",
				"cert-type":        "signing",
			},
			Version: time.Now().Unix(),
		}).Get(ctx, &signingSecret)
	if err != nil {
		logger.Error("Failed to store signing secret", "error", err)
		return fmt.Errorf("signing secret storage failed: %w", err)
	}

	// Store TLS certificates
	var tlsSecret vault.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultWorkflows.StoreSecretWorkflow,
		vault.SecretParams{
			Name: fmt.Sprintf("peer-%s-tls", config.PeerID),
			Data: map[string][]byte{
				"tls.crt": msp.TLSCert,
				"tls.key": msp.TLSKey,
			},
			Labels: map[string]string{
				"fabric-component": "peer",
				"organization":     "metacensus",
				"managed-by":       "temporal",
				"cert-type":        "tls",
			},
			Version: time.Now().Unix(),
		}).Get(ctx, &tlsSecret)
	if err != nil {
		logger.Error("Failed to store TLS secret", "error", err)
		return fmt.Errorf("TLS secret storage failed: %w", err)
	}

	logger.Info("Peer provisioning workflow completed successfully",
		"peer_id", config.PeerID,
		"signing_secret", signingSecret.Name,
		"tls_secret", tlsSecret.Name)

	return nil
}

// ProvisionOrdererWorkflow provisions a complete orderer with all necessary certificates
func ProvisionOrdererWorkflow(ctx workflow.Context, config OrdererConfig) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Orderer provisioning workflow started",
		"orderer_id", config.OrdererID,
		"namespace", config.Namespace)

	// Register orderer identity
	var reg Registration
	err := workflow.ExecuteChildWorkflow(ctx, RegisterIdentityWorkflow, IdentityParams{
		Type:       "orderer",
		Name:       config.OrdererID,
		CAEndpoint: config.EnrollmentCA,
		Attributes: map[string]string{
			"hf.Registrar.Roles": "orderer",
			"hf.Type":            "orderer",
		},
		Affiliation: "org1.department1",
		Secret:      "ordererpw", // TODO: Generate secure password
	}).Get(ctx, &reg)
	if err != nil {
		logger.Error("Failed to register orderer identity", "error", err)
		return fmt.Errorf("orderer registration failed: %w", err)
	}

	// Parallel enrollment for signing and TLS
	var signingEnroll, tlsEnroll Enrollment

	// Signing enrollment
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.ExecuteChildWorkflow(ctx, EnrollIdentityWorkflow,
			EnrollmentParams{
				Registration: reg,
				Profile:      "sign",
				CAEndpoint:   config.EnrollmentCA,
				CSRConfig: CSRConfig{
					CommonName:   config.OrdererID,
					Organization: "Metacensus",
					Country:      "US",
					Province:     "CA",
					Locality:     "San Francisco",
				},
			}).Get(ctx, &signingEnroll)
	})

	// TLS enrollment
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.ExecuteChildWorkflow(ctx, EnrollIdentityWorkflow,
			EnrollmentParams{
				Registration: reg,
				Profile:      "tls",
				CAEndpoint:   config.TLSCA,
				CSRConfig: CSRConfig{
					CommonName:   config.OrdererID,
					Organization: "Metacensus",
					Country:      "US",
					Province:     "CA",
					Locality:     "San Francisco",
					Hosts: []string{
						config.OrdererID,
						fmt.Sprintf("%s.%s.svc.cluster.local", config.OrdererID, config.Namespace),
					},
				},
			}).Get(ctx, &tlsEnroll)
	})

	// Wait for both enrollments to complete
	workflow.Sleep(ctx, time.Second) // Give time for parallel execution

	// Assemble complete MSP
	var msp MSPBundle
	err = workflow.ExecuteChildWorkflow(ctx, AssembleMSPWorkflow, MSPParams{
		SigningEnrollment: signingEnroll,
		TLSEnrollment:     &tlsEnroll,
		NodeOUConfig:      config.OrgNodeOUs,
		CACerts:           [][]byte{}, // TODO: Get CA certificates
	}).Get(ctx, &msp)
	if err != nil {
		logger.Error("Failed to assemble MSP", "error", err)
		return fmt.Errorf("MSP assembly failed: %w", err)
	}

	// Store signing certificates
	// Create vault client and workflows for this operation
	vaultConfig := vault.GetVaultConfig()
	vaultClient := vault.NewVaultClient(vaultConfig)
	vaultWorkflows := vault.NewWorkflows(vaultClient)

	var signingSecret vault.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultWorkflows.StoreSecretWorkflow,
		vault.SecretParams{
			Name: fmt.Sprintf("orderer-%s-signing", config.OrdererID),
			Data: map[string][]byte{
				"cert.pem": msp.SigningCert,
				"key.pem":  msp.SigningKey,
			},
			Labels: map[string]string{
				"fabric-component": "orderer",
				"organization":     "metacensus",
				"managed-by":       "temporal",
				"cert-type":        "signing",
			},
			Version: time.Now().Unix(),
		}).Get(ctx, &signingSecret)
	if err != nil {
		logger.Error("Failed to store signing secret", "error", err)
		return fmt.Errorf("signing secret storage failed: %w", err)
	}

	// Store TLS certificates
	var tlsSecret vault.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultWorkflows.StoreSecretWorkflow,
		vault.SecretParams{
			Name: fmt.Sprintf("orderer-%s-tls", config.OrdererID),
			Data: map[string][]byte{
				"tls.crt": msp.TLSCert,
				"tls.key": msp.TLSKey,
			},
			Labels: map[string]string{
				"fabric-component": "orderer",
				"organization":     "metacensus",
				"managed-by":       "temporal",
				"cert-type":        "tls",
			},
			Version: time.Now().Unix(),
		}).Get(ctx, &tlsSecret)
	if err != nil {
		logger.Error("Failed to store TLS secret", "error", err)
		return fmt.Errorf("TLS secret storage failed: %w", err)
	}

	logger.Info("Orderer provisioning workflow completed successfully",
		"orderer_id", config.OrdererID,
		"signing_secret", signingSecret.Name,
		"tls_secret", tlsSecret.Name)

	return nil
}

// ProvisionClientWorkflow provisions a client identity (simpler flow)
func ProvisionClientWorkflow(ctx workflow.Context, config ClientConfig) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Client provisioning workflow started",
		"client_id", config.ClientID,
		"namespace", config.Namespace)

	// Register client identity
	var reg Registration
	err := workflow.ExecuteChildWorkflow(ctx, RegisterIdentityWorkflow, IdentityParams{
		Type:        "client",
		Name:        config.ClientID,
		CAEndpoint:  "http://ca.fabric.svc.cluster.local:7054", // TODO: Make configurable
		Attributes:  config.CustomAttributes,
		Affiliation: "org1.department1",
		Secret:      "clientpw", // TODO: Generate secure password
	}).Get(ctx, &reg)
	if err != nil {
		logger.Error("Failed to register client identity", "error", err)
		return fmt.Errorf("client registration failed: %w", err)
	}

	// Enroll client identity
	var enrollment Enrollment
	err = workflow.ExecuteChildWorkflow(ctx, EnrollIdentityWorkflow,
		EnrollmentParams{
			Registration: reg,
			Profile:      "sign",
			CAEndpoint:   "http://ca.fabric.svc.cluster.local:7054", // TODO: Make configurable
			CSRConfig: CSRConfig{
				CommonName:   config.ClientID,
				Organization: "Metacensus",
				Country:      "US",
				Province:     "CA",
				Locality:     "San Francisco",
			},
		}).Get(ctx, &enrollment)
	if err != nil {
		logger.Error("Failed to enroll client identity", "error", err)
		return fmt.Errorf("client enrollment failed: %w", err)
	}

	// Store client certificates
	// Create vault client and workflows for this operation
	vaultConfig := vault.GetVaultConfig()
	vaultClient := vault.NewVaultClient(vaultConfig)
	vaultWorkflows := vault.NewWorkflows(vaultClient)

	var secret vault.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultWorkflows.StoreSecretWorkflow,
		vault.SecretParams{
			Name: fmt.Sprintf("client-%s", config.ClientID),
			Data: enrollment.ToSecretData(),
			Labels: map[string]string{
				"fabric-component": "client",
				"organization":     "metacensus",
				"managed-by":       "temporal",
				"cert-type":        "signing",
			},
			Version: time.Now().Unix(),
		}).Get(ctx, &secret)
	if err != nil {
		logger.Error("Failed to store client secret", "error", err)
		return fmt.Errorf("client secret storage failed: %w", err)
	}

	logger.Info("Client provisioning workflow completed successfully",
		"client_id", config.ClientID,
		"secret", secret.Name)

	return nil
}
