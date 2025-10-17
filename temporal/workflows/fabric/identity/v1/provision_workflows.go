package v1

import (
	"fmt"
	"time"

	identitypb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ==================== Provision Workflows ====================
// These workflows orchestrate provisioning of nodes and organizations

// provisionPeerWorkflow implements identitypb.ProvisionPeerWorkflow
type provisionPeerWorkflow struct {
	input *identitypb.ProvisionPeerWorkflowInput
}

// Execute provisions a complete peer with all necessary certificates
func (w *provisionPeerWorkflow) Execute(ctx workflow.Context) (*identitypb.ProvisionPeerOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProvisionPeer workflow started",
		"peer_id", w.input.Req.GetPeerId(),
		"organization", w.input.Req.GetOrganization())

	// Set up activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Set up child workflow options
	wo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 20 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, wo)

	// Step 1: Generate secure secret for peer identity
	logger.Info("Generating secure secret for peer")
	var secretOutput *identitypb.GenerateSecureSecretOutput
	err := workflow.ExecuteActivity(activityCtx, identitypb.GenerateSecureSecretActivityName,
		&identitypb.GenerateSecureSecretInput{Length: 32}).Get(ctx, &secretOutput)
	if err != nil {
		logger.Error("Failed to generate peer secret", "error", err)
		return nil, fmt.Errorf("failed to generate peer secret: %w", err)
	}

	// Step 2: Register peer identity
	logger.Info("Registering peer identity")
	var regOutput *identitypb.RegisterIdentityOutput
	err = workflow.ExecuteChildWorkflow(ctx, identitypb.RegisterIdentityWorkflowName,
		&identitypb.RegisterIdentityWorkflowInput{
			Req: &identitypb.RegisterIdentityInput{
				Name:        w.input.Req.GetPeerId(),
				Type:        "peer",
				Secret:      secretOutput.GetSecret(),
				Affiliation: w.input.Req.GetOrganization(),
				Attributes: []*identitypb.Attribute{
					{Name: "hf.Registrar.Roles", Value: "peer", Ecert: true},
					{Name: "hf.Type", Value: "peer", Ecert: true},
				},
				MaxEnrollments: 1,
			},
		}).Get(ctx, &regOutput)
	if err != nil {
		logger.Error("Failed to register peer identity", "error", err)
		return nil, fmt.Errorf("failed to register peer identity: %w", err)
	}

	// Step 3: Enroll for signing and TLS certificates in parallel
	logger.Info("Enrolling peer for signing and TLS certificates")

	// Build hosts list for TLS cert
	hosts := []string{
		w.input.Req.GetPeerId(),
		fmt.Sprintf("%s.%s.svc.cluster.local", w.input.Req.GetPeerId(), w.input.Req.GetNamespace()),
	}
	hosts = append(hosts, w.input.Req.GetHosts()...)

	// Start both enrollments in parallel
	signingFuture := workflow.ExecuteChildWorkflow(ctx, identitypb.EnrollIdentityWorkflowName,
		&identitypb.EnrollIdentityWorkflowInput{
			Req: &identitypb.EnrollIdentityInput{
				Name:    regOutput.GetName(),
				Secret:  regOutput.GetSecret(),
				Profile: "", // Default profile for signing
				Csr: &identitypb.CSRInfo{
					Cn:    w.input.Req.GetPeerId(),
					Hosts: []string{w.input.Req.GetPeerId()},
				},
			},
		})

	tlsFuture := workflow.ExecuteChildWorkflow(ctx, identitypb.EnrollIdentityWorkflowName,
		&identitypb.EnrollIdentityWorkflowInput{
			Req: &identitypb.EnrollIdentityInput{
				Name:    regOutput.GetName(),
				Secret:  regOutput.GetSecret(),
				Profile: "tls",
				Csr: &identitypb.CSRInfo{
					Cn:    w.input.Req.GetPeerId(),
					Hosts: hosts,
				},
			},
		})

	// Wait for both enrollments
	var signingEnroll, tlsEnroll *identitypb.EnrollIdentityOutput
	err = signingFuture.Get(ctx, &signingEnroll)
	if err != nil {
		logger.Error("Failed to enroll for signing cert", "error", err)
		return nil, fmt.Errorf("failed to enroll for signing cert: %w", err)
	}

	err = tlsFuture.Get(ctx, &tlsEnroll)
	if err != nil {
		logger.Error("Failed to enroll for TLS cert", "error", err)
		return nil, fmt.Errorf("failed to enroll for TLS cert: %w", err)
	}

	// Step 4: Store signing secret in Vault
	logger.Info("Storing signing secret in Vault")
	var signingSecretRef *vaultpb.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
		&vaultpb.StoreSecretWorkflowInput{
			Req: &vaultpb.Secret{
				Name:      fmt.Sprintf("peer-%s-signing", w.input.Req.GetPeerId()),
				Namespace: w.input.Req.GetNamespace(),
				Data: map[string][]byte{
					"cert.pem": signingEnroll.GetCertificate(),
					"key.pem":  signingEnroll.GetPrivateKey(),
				},
			},
		}).Get(ctx, &signingSecretRef)
	if err != nil {
		logger.Error("Failed to store signing secret", "error", err)
		return nil, fmt.Errorf("failed to store signing secret: %w", err)
	}

	// Step 5: Store TLS secret in Vault
	logger.Info("Storing TLS secret in Vault")
	var tlsSecretRef *vaultpb.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
		&vaultpb.StoreSecretWorkflowInput{
			Req: &vaultpb.Secret{
				Name:      fmt.Sprintf("peer-%s-tls", w.input.Req.GetPeerId()),
				Namespace: w.input.Req.GetNamespace(),
				Data: map[string][]byte{
					"cert.pem": tlsEnroll.GetCertificate(),
					"key.pem":  tlsEnroll.GetPrivateKey(),
				},
			},
		}).Get(ctx, &tlsSecretRef)
	if err != nil {
		logger.Error("Failed to store TLS secret", "error", err)
		return nil, fmt.Errorf("failed to store TLS secret: %w", err)
	}

	logger.Info("ProvisionPeer workflow completed successfully",
		"peer_id", w.input.Req.GetPeerId(),
		"signing_secret", signingSecretRef.GetName(),
		"tls_secret", tlsSecretRef.GetName())

	return &identitypb.ProvisionPeerOutput{
		SigningSecret: signingSecretRef,
		TlsSecret:     tlsSecretRef,
		IdentityName:  regOutput.GetName(),
		Organization:  w.input.Req.GetOrganization(),
	}, nil
}

// provisionOrdererWorkflow implements identitypb.ProvisionOrdererWorkflow
type provisionOrdererWorkflow struct {
	input *identitypb.ProvisionOrdererWorkflowInput
}

// Execute provisions a complete orderer with all necessary certificates
func (w *provisionOrdererWorkflow) Execute(ctx workflow.Context) (*identitypb.ProvisionOrdererOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProvisionOrderer workflow started",
		"orderer_id", w.input.Req.GetOrdererId(),
		"organization", w.input.Req.GetOrganization())

	// Orderer provisioning is almost identical to peer provisioning
	// The only differences are the identity type and naming

	// Set up activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Set up child workflow options
	wo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 20 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, wo)

	// Step 1: Generate secure secret for orderer identity
	logger.Info("Generating secure secret for orderer")
	var secretOutput *identitypb.GenerateSecureSecretOutput
	err := workflow.ExecuteActivity(activityCtx, identitypb.GenerateSecureSecretActivityName,
		&identitypb.GenerateSecureSecretInput{Length: 32}).Get(ctx, &secretOutput)
	if err != nil {
		logger.Error("Failed to generate orderer secret", "error", err)
		return nil, fmt.Errorf("failed to generate orderer secret: %w", err)
	}

	// Step 2: Register orderer identity
	logger.Info("Registering orderer identity")
	var regOutput *identitypb.RegisterIdentityOutput
	err = workflow.ExecuteChildWorkflow(ctx, identitypb.RegisterIdentityWorkflowName,
		&identitypb.RegisterIdentityWorkflowInput{
			Req: &identitypb.RegisterIdentityInput{
				Name:        w.input.Req.GetOrdererId(),
				Type:        "orderer",
				Secret:      secretOutput.GetSecret(),
				Affiliation: w.input.Req.GetOrganization(),
				Attributes: []*identitypb.Attribute{
					{Name: "hf.Registrar.Roles", Value: "orderer", Ecert: true},
					{Name: "hf.Type", Value: "orderer", Ecert: true},
				},
				MaxEnrollments: 1,
			},
		}).Get(ctx, &regOutput)
	if err != nil {
		logger.Error("Failed to register orderer identity", "error", err)
		return nil, fmt.Errorf("failed to register orderer identity: %w", err)
	}

	// Step 3: Enroll for signing and TLS certificates in parallel
	logger.Info("Enrolling orderer for signing and TLS certificates")

	// Build hosts list for TLS cert
	hosts := []string{
		w.input.Req.GetOrdererId(),
		fmt.Sprintf("%s.%s.svc.cluster.local", w.input.Req.GetOrdererId(), w.input.Req.GetNamespace()),
	}
	hosts = append(hosts, w.input.Req.GetHosts()...)

	// Start both enrollments in parallel
	signingFuture := workflow.ExecuteChildWorkflow(ctx, identitypb.EnrollIdentityWorkflowName,
		&identitypb.EnrollIdentityWorkflowInput{
			Req: &identitypb.EnrollIdentityInput{
				Name:    regOutput.GetName(),
				Secret:  regOutput.GetSecret(),
				Profile: "", // Default profile for signing
				Csr: &identitypb.CSRInfo{
					Cn:    w.input.Req.GetOrdererId(),
					Hosts: []string{w.input.Req.GetOrdererId()},
				},
			},
		})

	tlsFuture := workflow.ExecuteChildWorkflow(ctx, identitypb.EnrollIdentityWorkflowName,
		&identitypb.EnrollIdentityWorkflowInput{
			Req: &identitypb.EnrollIdentityInput{
				Name:    regOutput.GetName(),
				Secret:  regOutput.GetSecret(),
				Profile: "tls",
				Csr: &identitypb.CSRInfo{
					Cn:    w.input.Req.GetOrdererId(),
					Hosts: hosts,
				},
			},
		})

	// Wait for both enrollments
	var signingEnroll, tlsEnroll *identitypb.EnrollIdentityOutput
	err = signingFuture.Get(ctx, &signingEnroll)
	if err != nil {
		logger.Error("Failed to enroll for signing cert", "error", err)
		return nil, fmt.Errorf("failed to enroll for signing cert: %w", err)
	}

	err = tlsFuture.Get(ctx, &tlsEnroll)
	if err != nil {
		logger.Error("Failed to enroll for TLS cert", "error", err)
		return nil, fmt.Errorf("failed to enroll for TLS cert: %w", err)
	}

	// Step 4: Store signing secret in Vault
	logger.Info("Storing signing secret in Vault")
	var signingSecretRef *vaultpb.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
		&vaultpb.StoreSecretWorkflowInput{
			Req: &vaultpb.Secret{
				Name:      fmt.Sprintf("orderer-%s-signing", w.input.Req.GetOrdererId()),
				Namespace: w.input.Req.GetNamespace(),
				Data: map[string][]byte{
					"cert.pem": signingEnroll.GetCertificate(),
					"key.pem":  signingEnroll.GetPrivateKey(),
				},
			},
		}).Get(ctx, &signingSecretRef)
	if err != nil {
		logger.Error("Failed to store signing secret", "error", err)
		return nil, fmt.Errorf("failed to store signing secret: %w", err)
	}

	// Step 5: Store TLS secret in Vault
	logger.Info("Storing TLS secret in Vault")
	var tlsSecretRef *vaultpb.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
		&vaultpb.StoreSecretWorkflowInput{
			Req: &vaultpb.Secret{
				Name:      fmt.Sprintf("orderer-%s-tls", w.input.Req.GetOrdererId()),
				Namespace: w.input.Req.GetNamespace(),
				Data: map[string][]byte{
					"cert.pem": tlsEnroll.GetCertificate(),
					"key.pem":  tlsEnroll.GetPrivateKey(),
				},
			},
		}).Get(ctx, &tlsSecretRef)
	if err != nil {
		logger.Error("Failed to store TLS secret", "error", err)
		return nil, fmt.Errorf("failed to store TLS secret: %w", err)
	}

	logger.Info("ProvisionOrderer workflow completed successfully",
		"orderer_id", w.input.Req.GetOrdererId(),
		"signing_secret", signingSecretRef.GetName(),
		"tls_secret", tlsSecretRef.GetName())

	return &identitypb.ProvisionOrdererOutput{
		SigningSecret: signingSecretRef,
		TlsSecret:     tlsSecretRef,
		IdentityName:  regOutput.GetName(),
		Organization:  w.input.Req.GetOrganization(),
	}, nil
}

// provisionClientWorkflow implements identitypb.ProvisionClientWorkflow
type provisionClientWorkflow struct {
	input *identitypb.ProvisionClientWorkflowInput
}

// Execute provisions a client identity (simpler flow - optional TLS)
func (w *provisionClientWorkflow) Execute(ctx workflow.Context) (*identitypb.ProvisionClientOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProvisionClient workflow started",
		"client_id", w.input.Req.GetClientId(),
		"organization", w.input.Req.GetOrganization())

	// Set up activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, ao)

	// Set up child workflow options
	wo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 20 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, wo)

	// Step 1: Generate secure secret for client identity
	logger.Info("Generating secure secret for client")
	var secretOutput *identitypb.GenerateSecureSecretOutput
	err := workflow.ExecuteActivity(activityCtx, identitypb.GenerateSecureSecretActivityName,
		&identitypb.GenerateSecureSecretInput{Length: 32}).Get(ctx, &secretOutput)
	if err != nil {
		logger.Error("Failed to generate client secret", "error", err)
		return nil, fmt.Errorf("failed to generate client secret: %w", err)
	}

	// Step 2: Register client identity
	logger.Info("Registering client identity")
	var regOutput *identitypb.RegisterIdentityOutput
	err = workflow.ExecuteChildWorkflow(ctx, identitypb.RegisterIdentityWorkflowName,
		&identitypb.RegisterIdentityWorkflowInput{
			Req: &identitypb.RegisterIdentityInput{
				Name:           w.input.Req.GetClientId(),
				Type:           "client",
				Secret:         secretOutput.GetSecret(),
				Affiliation:    w.input.Req.GetOrganization(),
				Attributes:     w.input.Req.GetCustomAttributes(),
				MaxEnrollments: -1, // Unlimited for clients
			},
		}).Get(ctx, &regOutput)
	if err != nil {
		logger.Error("Failed to register client identity", "error", err)
		return nil, fmt.Errorf("failed to register client identity: %w", err)
	}

	// Step 3: Enroll for signing certificate
	logger.Info("Enrolling client for signing certificate")
	var signingEnroll *identitypb.EnrollIdentityOutput
	err = workflow.ExecuteChildWorkflow(ctx, identitypb.EnrollIdentityWorkflowName,
		&identitypb.EnrollIdentityWorkflowInput{
			Req: &identitypb.EnrollIdentityInput{
				Name:    regOutput.GetName(),
				Secret:  regOutput.GetSecret(),
				Profile: "", // Default profile for signing
				Csr: &identitypb.CSRInfo{
					Cn:    w.input.Req.GetClientId(),
					Hosts: []string{w.input.Req.GetClientId()},
				},
			},
		}).Get(ctx, &signingEnroll)
	if err != nil {
		logger.Error("Failed to enroll for signing cert", "error", err)
		return nil, fmt.Errorf("failed to enroll for signing cert: %w", err)
	}

	// Step 4: Store signing secret in Vault
	logger.Info("Storing signing secret in Vault")
	var signingSecretRef *vaultpb.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
		&vaultpb.StoreSecretWorkflowInput{
			Req: &vaultpb.Secret{
				Name:      fmt.Sprintf("client-%s-signing", w.input.Req.GetClientId()),
				Namespace: w.input.Req.GetNamespace(),
				Data: map[string][]byte{
					"cert.pem": signingEnroll.GetCertificate(),
					"key.pem":  signingEnroll.GetPrivateKey(),
				},
			},
		}).Get(ctx, &signingSecretRef)
	if err != nil {
		logger.Error("Failed to store signing secret", "error", err)
		return nil, fmt.Errorf("failed to store signing secret: %w", err)
	}

	// Step 5: Optionally enroll for TLS certificate
	var tlsSecretRef *vaultpb.SecretRef
	if w.input.Req.GetEnableTls() {
		logger.Info("Enrolling client for TLS certificate")

		// Build hosts list for TLS cert
		hosts := []string{w.input.Req.GetClientId()}
		hosts = append(hosts, w.input.Req.GetHosts()...)

		var tlsEnroll *identitypb.EnrollIdentityOutput
		err = workflow.ExecuteChildWorkflow(ctx, identitypb.EnrollIdentityWorkflowName,
			&identitypb.EnrollIdentityWorkflowInput{
				Req: &identitypb.EnrollIdentityInput{
					Name:    regOutput.GetName(),
					Secret:  regOutput.GetSecret(),
					Profile: "tls",
					Csr: &identitypb.CSRInfo{
						Cn:    w.input.Req.GetClientId(),
						Hosts: hosts,
					},
				},
			}).Get(ctx, &tlsEnroll)
		if err != nil {
			logger.Error("Failed to enroll for TLS cert", "error", err)
			return nil, fmt.Errorf("failed to enroll for TLS cert: %w", err)
		}

		// Store TLS secret
		logger.Info("Storing TLS secret in Vault")
		err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
			&vaultpb.StoreSecretWorkflowInput{
				Req: &vaultpb.Secret{
					Name:      fmt.Sprintf("client-%s-tls", w.input.Req.GetClientId()),
					Namespace: w.input.Req.GetNamespace(),
					Data: map[string][]byte{
						"cert.pem": tlsEnroll.GetCertificate(),
						"key.pem":  tlsEnroll.GetPrivateKey(),
					},
				},
			}).Get(ctx, &tlsSecretRef)
		if err != nil {
			logger.Error("Failed to store TLS secret", "error", err)
			return nil, fmt.Errorf("failed to store TLS secret: %w", err)
		}
	}

	logger.Info("ProvisionClient workflow completed successfully",
		"client_id", w.input.Req.GetClientId(),
		"signing_secret", signingSecretRef.GetName(),
		"tls_enabled", w.input.Req.GetEnableTls())

	return &identitypb.ProvisionClientOutput{
		SigningSecret: signingSecretRef,
		TlsSecret:     tlsSecretRef,
		IdentityName:  regOutput.GetName(),
		Organization:  w.input.Req.GetOrganization(),
	}, nil
}

// provisionOrgWorkflow implements identitypb.ProvisionOrgWorkflow
type provisionOrgWorkflow struct {
	input *identitypb.ProvisionOrgWorkflowInput
}

// Execute provisions a complete organization with admin identity, peers, and orderers
func (w *provisionOrgWorkflow) Execute(ctx workflow.Context) (*identitypb.ProvisionOrgOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProvisionOrg workflow started",
		"organization", w.input.Req.GetOrganization(),
		"namespace", w.input.Req.GetNamespace(),
		"peer_count", len(w.input.Req.GetPeerIds()),
		"orderer_count", len(w.input.Req.GetOrdererIds()))

	// Set up child workflow options
	wo := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, wo)

	// Step 1: Bootstrap admin identity
	logger.Info("Bootstrapping admin identity")
	var adminOutput *identitypb.BootstrapAdminOutput
	err := workflow.ExecuteChildWorkflow(ctx, identitypb.BootstrapAdminWorkflowName,
		&identitypb.BootstrapAdminWorkflowInput{
			Req: w.input.Req.GetAdmin(),
		}).Get(ctx, &adminOutput)
	if err != nil {
		logger.Error("Failed to bootstrap admin", "error", err)
		return nil, fmt.Errorf("failed to bootstrap admin: %w", err)
	}

	// Step 2: Store admin secret in Vault
	logger.Info("Storing admin secret in Vault")
	var adminSecretRef *vaultpb.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
		&vaultpb.StoreSecretWorkflowInput{
			Req: &vaultpb.Secret{
				Name:      fmt.Sprintf("%s-admin", w.input.Req.GetOrganization()),
				Namespace: w.input.Req.GetNamespace(),
				Data: map[string][]byte{
					"cert.pem":    adminOutput.GetAdminIdentity().GetCertificate(),
					"key.pem":     adminOutput.GetAdminIdentity().GetPrivateKey(),
					"ca-cert.pem": adminOutput.GetAdminIdentity().GetCaCert(),
				},
			},
		}).Get(ctx, &adminSecretRef)
	if err != nil {
		logger.Error("Failed to store admin secret", "error", err)
		return nil, fmt.Errorf("failed to store admin secret: %w", err)
	}

	// Step 3: Get CA certificates (we have them from admin enrollment)
	// Store CA certificates as org-wide secret
	logger.Info("Storing org-wide CA certificates")
	var caCertRef *vaultpb.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
		&vaultpb.StoreSecretWorkflowInput{
			Req: &vaultpb.Secret{
				Name:      fmt.Sprintf("%s-ca-certificates", w.input.Req.GetOrganization()),
				Namespace: w.input.Req.GetNamespace(),
				Data: map[string][]byte{
					"ca-cert.pem": adminOutput.GetAdminIdentity().GetCaCert(),
				},
			},
		}).Get(ctx, &caCertRef)
	if err != nil {
		logger.Error("Failed to store CA certificates", "error", err)
		return nil, fmt.Errorf("failed to store CA certificates: %w", err)
	}

	// Step 4: Store NodeOU configuration as org-wide secret
	logger.Info("Storing org-wide NodeOU configuration")
	nodeOuBytes, err := serializeNodeOUConfig(w.input.Req.GetOrgNodeOus())
	if err != nil {
		logger.Error("Failed to serialize NodeOU config", "error", err)
		return nil, fmt.Errorf("failed to serialize NodeOU config: %w", err)
	}

	var nodeOuConfigRef *vaultpb.SecretRef
	err = workflow.ExecuteChildWorkflow(ctx, vaultpb.StoreSecretWorkflowName,
		&vaultpb.StoreSecretWorkflowInput{
			Req: &vaultpb.Secret{
				Name:      fmt.Sprintf("%s-node-ou-config", w.input.Req.GetOrganization()),
				Namespace: w.input.Req.GetNamespace(),
				Data: map[string][]byte{
					"config.yaml": nodeOuBytes,
				},
			},
		}).Get(ctx, &nodeOuConfigRef)
	if err != nil {
		logger.Error("Failed to store NodeOU config", "error", err)
		return nil, fmt.Errorf("failed to store NodeOU config: %w", err)
	}

	// Step 5: Provision all peers in parallel
	logger.Info("Provisioning peers", "count", len(w.input.Req.GetPeerIds()))
	peerFutures := make([]workflow.ChildWorkflowFuture, 0, len(w.input.Req.GetPeerIds()))
	for _, peerID := range w.input.Req.GetPeerIds() {
		future := workflow.ExecuteChildWorkflow(ctx, identitypb.ProvisionPeerWorkflowName,
			&identitypb.ProvisionPeerWorkflowInput{
				Req: &identitypb.ProvisionPeerInput{
					PeerId:       peerID,
					Organization: w.input.Req.GetOrganization(),
					OrgNodeOus:   w.input.Req.GetOrgNodeOus(),
					Namespace:    w.input.Req.GetNamespace(),
					Hosts:        w.input.Req.GetAdditionalHosts(),
				},
			})
		peerFutures = append(peerFutures, future)
	}

	// Wait for all peers to complete
	peers := make([]*identitypb.ProvisionPeerOutput, 0, len(peerFutures))
	for i, future := range peerFutures {
		var peerOutput *identitypb.ProvisionPeerOutput
		err := future.Get(ctx, &peerOutput)
		if err != nil {
			logger.Error("Failed to provision peer", "peer_index", i, "error", err)
			return nil, fmt.Errorf("failed to provision peer %d: %w", i, err)
		}
		peers = append(peers, peerOutput)
	}

	// Step 6: Provision all orderers in parallel
	logger.Info("Provisioning orderers", "count", len(w.input.Req.GetOrdererIds()))
	ordererFutures := make([]workflow.ChildWorkflowFuture, 0, len(w.input.Req.GetOrdererIds()))
	for _, ordererID := range w.input.Req.GetOrdererIds() {
		future := workflow.ExecuteChildWorkflow(ctx, identitypb.ProvisionOrdererWorkflowName,
			&identitypb.ProvisionOrdererWorkflowInput{
				Req: &identitypb.ProvisionOrdererInput{
					OrdererId:    ordererID,
					Organization: w.input.Req.GetOrganization(),
					OrgNodeOus:   w.input.Req.GetOrgNodeOus(),
					Namespace:    w.input.Req.GetNamespace(),
					Hosts:        w.input.Req.GetAdditionalHosts(),
				},
			})
		ordererFutures = append(ordererFutures, future)
	}

	// Wait for all orderers to complete
	orderers := make([]*identitypb.ProvisionOrdererOutput, 0, len(ordererFutures))
	for i, future := range ordererFutures {
		var ordererOutput *identitypb.ProvisionOrdererOutput
		err := future.Get(ctx, &ordererOutput)
		if err != nil {
			logger.Error("Failed to provision orderer", "orderer_index", i, "error", err)
			return nil, fmt.Errorf("failed to provision orderer %d: %w", i, err)
		}
		orderers = append(orderers, ordererOutput)
	}

	logger.Info("ProvisionOrg workflow completed successfully",
		"organization", w.input.Req.GetOrganization(),
		"peers_provisioned", len(peers),
		"orderers_provisioned", len(orderers))

	return &identitypb.ProvisionOrgOutput{
		Organization:   w.input.Req.GetOrganization(),
		AdminSecret:    adminSecretRef,
		CaCertificates: caCertRef,
		NodeOuConfig:   nodeOuConfigRef,
		Peers:          peers,
		Orderers:       orderers,
	}, nil
}

// serializeNodeOUConfig converts NodeOUConfig to YAML bytes for storage
// This is a simple serialization - in production you might want to use a YAML library
func serializeNodeOUConfig(config *identitypb.NodeOUConfig) ([]byte, error) {
	if config == nil {
		return []byte("NodeOUs:\n  Enable: false\n"), nil
	}

	yaml := fmt.Sprintf("NodeOUs:\n  Enable: %t\n", config.GetEnable())

	if config.GetClient() != nil {
		yaml += fmt.Sprintf("  ClientOUIdentifier:\n    Certificate: %s\n    OrganizationalUnitIdentifier: %s\n",
			config.GetClient().GetCertificate(),
			config.GetClient().GetOrganizationalUnitIdentifier())
	}

	if config.GetPeer() != nil {
		yaml += fmt.Sprintf("  PeerOUIdentifier:\n    Certificate: %s\n    OrganizationalUnitIdentifier: %s\n",
			config.GetPeer().GetCertificate(),
			config.GetPeer().GetOrganizationalUnitIdentifier())
	}

	if config.GetAdmin() != nil {
		yaml += fmt.Sprintf("  AdminOUIdentifier:\n    Certificate: %s\n    OrganizationalUnitIdentifier: %s\n",
			config.GetAdmin().GetCertificate(),
			config.GetAdmin().GetOrganizationalUnitIdentifier())
	}

	if config.GetOrderer() != nil {
		yaml += fmt.Sprintf("  OrdererOUIdentifier:\n    Certificate: %s\n    OrganizationalUnitIdentifier: %s\n",
			config.GetOrderer().GetCertificate(),
			config.GetOrderer().GetOrganizationalUnitIdentifier())
	}

	return []byte(yaml), nil
}
