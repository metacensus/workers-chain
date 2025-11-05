package identity

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
	err = workflow.ExecuteActivity(activityCtx, identitypb.RegisterIdentityActivityName,
		&identitypb.RegisterIdentityInput{
			Name:        w.input.Req.GetPeerId(),
			Type:        "peer",
			Secret:      secretOutput.GetSecret(),
			Affiliation: w.input.Req.GetOrganization(),
			Attributes: []*identitypb.Attribute{
				{Name: "hf.Registrar.Roles", Value: "peer", Ecert: true},
				{Name: "hf.Type", Value: "peer", Ecert: true},
			},
			MaxEnrollments: 1,
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
	signingFuture := workflow.ExecuteActivity(activityCtx, identitypb.EnrollIdentityActivityName,
		&identitypb.EnrollIdentityInput{
			Name:    regOutput.GetName(),
			Secret:  regOutput.GetSecret(),
			Profile: "", // Default profile for signing
			Csr: &identitypb.CSRInfo{
				Cn:    w.input.Req.GetPeerId(),
				Hosts: hosts,
			},
		})

	tlsFuture := workflow.ExecuteActivity(activityCtx, identitypb.EnrollIdentityActivityName,
		&identitypb.EnrollIdentityInput{
			Name:    regOutput.GetName(),
			Secret:  regOutput.GetSecret(),
			Profile: "tls",
			Csr: &identitypb.CSRInfo{
				Cn:    w.input.Req.GetPeerId(),
				Hosts: hosts,
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

	// Step 4: Store peer certificates in Vault using MSP directory structure
	logger.Info("Storing peer certificates in Vault")
	var storeOutput *vaultpb.StoreSecretOutput
	vaultPaths := NewVaultPaths(w.input.Req.GetOrganization())
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: vaultPaths.PeerMSP(w.input.Req.GetPeerId()),
			Data: map[string][]byte{
				fmt.Sprintf("signcerts/peer-%s.pem", w.input.Req.GetPeerId()):    signingEnroll.GetCertificate(),
				fmt.Sprintf("keystore/peer-%s-key.pem", w.input.Req.GetPeerId()): signingEnroll.GetPrivateKey(),
				fmt.Sprintf("tlscerts/peer-%s-tls.pem", w.input.Req.GetPeerId()): tlsEnroll.GetCertificate(),
			},
		}).Get(ctx, &storeOutput)
	if err != nil {
		logger.Error("Failed to store peer certificates", "error", err)
		return nil, fmt.Errorf("failed to store peer certificates: %w", err)
	}

	logger.Info("ProvisionPeer workflow completed successfully",
		"peer_id", w.input.Req.GetPeerId(),
		"vault_path", vaultPaths.PeerMSP(w.input.Req.GetPeerId()))

	return &identitypb.ProvisionPeerOutput{
		SigningSecret: storeOutput.SecretRef,
		TlsSecret:     storeOutput.SecretRef, // Both are in the same secret now
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
	err = workflow.ExecuteActivity(activityCtx, identitypb.RegisterIdentityActivityName,
		&identitypb.RegisterIdentityInput{
			Name:        w.input.Req.GetOrdererId(),
			Type:        "orderer",
			Secret:      secretOutput.GetSecret(),
			Affiliation: w.input.Req.GetOrganization(),
			Attributes: []*identitypb.Attribute{
				{Name: "hf.Registrar.Roles", Value: "orderer", Ecert: true},
				{Name: "hf.Type", Value: "orderer", Ecert: true},
			},
			MaxEnrollments: 1,
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
	signingFuture := workflow.ExecuteActivity(activityCtx, identitypb.EnrollIdentityActivityName,
		&identitypb.EnrollIdentityInput{
			Name:    regOutput.GetName(),
			Secret:  regOutput.GetSecret(),
			Profile: "", // Default profile for signing
			Csr: &identitypb.CSRInfo{
				Cn:    w.input.Req.GetOrdererId(),
				Hosts: hosts,
			},
		})

	tlsFuture := workflow.ExecuteActivity(activityCtx, identitypb.EnrollIdentityActivityName,
		&identitypb.EnrollIdentityInput{
			Name:    regOutput.GetName(),
			Secret:  regOutput.GetSecret(),
			Profile: "tls",
			Csr: &identitypb.CSRInfo{
				Cn:    w.input.Req.GetOrdererId(),
				Hosts: hosts,
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

	// Step 4: Store orderer certificates in Vault using MSP directory structure
	logger.Info("Storing orderer certificates in Vault")
	var storeOutput *vaultpb.StoreSecretOutput
	vaultPaths := NewVaultPaths(w.input.Req.GetOrganization())
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: vaultPaths.OrdererMSP(w.input.Req.GetOrdererId()),
			Data: map[string][]byte{
				fmt.Sprintf("signcerts/orderer-%s.pem", w.input.Req.GetOrdererId()):    signingEnroll.GetCertificate(),
				fmt.Sprintf("keystore/orderer-%s-key.pem", w.input.Req.GetOrdererId()): signingEnroll.GetPrivateKey(),
				fmt.Sprintf("tlscerts/orderer-%s-tls.pem", w.input.Req.GetOrdererId()): tlsEnroll.GetCertificate(),
			},
		}).Get(ctx, &storeOutput)
	if err != nil {
		logger.Error("Failed to store orderer certificates", "error", err)
		return nil, fmt.Errorf("failed to store orderer certificates: %w", err)
	}

	logger.Info("ProvisionOrderer workflow completed successfully",
		"orderer_id", w.input.Req.GetOrdererId(),
		"vault_path", vaultPaths.OrdererMSP(w.input.Req.GetOrdererId()))

	return &identitypb.ProvisionOrdererOutput{
		SigningSecret: storeOutput.SecretRef,
		TlsSecret:     storeOutput.SecretRef, // Both are in the same secret now
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
	err = workflow.ExecuteActivity(activityCtx, identitypb.RegisterIdentityActivityName,
		&identitypb.RegisterIdentityInput{
			Name:        w.input.Req.GetClientId(),
			Type:        "client",
			Secret:      secretOutput.GetSecret(),
			Affiliation: w.input.Req.GetOrganization(),
			Attributes: []*identitypb.Attribute{
				{Name: "hf.Registrar.Roles", Value: "user", Ecert: true},
			},
			MaxEnrollments: 1,
		}).Get(ctx, &regOutput)
	if err != nil {
		logger.Error("Failed to register client identity", "error", err)
		return nil, fmt.Errorf("failed to register client identity: %w", err)
	}

	// Step 3: Enroll for signing certificate
	logger.Info("Enrolling client for signing certificate")
	var signingEnroll *identitypb.EnrollIdentityOutput
	err = workflow.ExecuteActivity(activityCtx, identitypb.EnrollIdentityActivityName,
		&identitypb.EnrollIdentityInput{
			Name:    regOutput.GetName(),
			Secret:  regOutput.GetSecret(),
			Profile: "", // Default profile for signing
			Csr: &identitypb.CSRInfo{
				Cn:    w.input.Req.GetClientId(),
				Hosts: []string{w.input.Req.GetClientId()},
			},
		}).Get(ctx, &signingEnroll)
	if err != nil {
		logger.Error("Failed to enroll for signing cert", "error", err)
		return nil, fmt.Errorf("failed to enroll for signing cert: %w", err)
	}

	// Step 4: Store signing certificate in Vault
	logger.Info("Storing client signing certificate in Vault")
	var signingStoreOutput *vaultpb.StoreSecretOutput
	vaultPaths := NewVaultPaths(w.input.Req.GetOrganization())
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: vaultPaths.ClientMSP(w.input.Req.GetClientId()),
			Data: map[string][]byte{
				fmt.Sprintf("signcerts/client-%s.pem", w.input.Req.GetClientId()):    signingEnroll.GetCertificate(),
				fmt.Sprintf("keystore/client-%s-key.pem", w.input.Req.GetClientId()): signingEnroll.GetPrivateKey(),
			},
		}).Get(ctx, &signingStoreOutput)
	if err != nil {
		logger.Error("Failed to store client signing certificate", "error", err)
		return nil, fmt.Errorf("failed to store client signing certificate: %w", err)
	}

	// Step 5: Optionally enroll for TLS certificate
	var tlsSecretRef *vaultpb.SecretRef
	if w.input.Req.GetEnableTls() {
		logger.Info("Enrolling client for TLS certificate")

		// Build hosts list for TLS cert
		hosts := []string{w.input.Req.GetClientId()}
		hosts = append(hosts, w.input.Req.GetHosts()...)

		var tlsEnroll *identitypb.EnrollIdentityOutput
		err = workflow.ExecuteActivity(activityCtx, identitypb.EnrollIdentityActivityName,
			&identitypb.EnrollIdentityInput{
				Name:    regOutput.GetName(),
				Secret:  regOutput.GetSecret(),
				Profile: "tls",
				Csr: &identitypb.CSRInfo{
					Cn:    w.input.Req.GetClientId(),
					Hosts: hosts,
				},
			}).Get(ctx, &tlsEnroll)
		if err != nil {
			logger.Error("Failed to enroll for TLS cert", "error", err)
			return nil, fmt.Errorf("failed to enroll for TLS cert: %w", err)
		}

		// Store TLS certificate
		logger.Info("Storing client TLS certificate in Vault")
		var tlsStoreOutput *vaultpb.StoreSecretOutput
		err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
			&vaultpb.StoreSecretInput{
				Path: vaultPaths.ClientTLSMSP(w.input.Req.GetClientId()),
				Data: map[string][]byte{
					fmt.Sprintf("tlscerts/client-%s-tls.pem", w.input.Req.GetClientId()):     tlsEnroll.GetCertificate(),
					fmt.Sprintf("keystore/client-%s-tls-key.pem", w.input.Req.GetClientId()): tlsEnroll.GetPrivateKey(),
				},
			}).Get(ctx, &tlsStoreOutput)
		if err != nil {
			logger.Error("Failed to store client TLS certificate", "error", err)
			return nil, fmt.Errorf("failed to store client TLS certificate: %w", err)
		}
		tlsSecretRef = tlsStoreOutput.SecretRef
	}

	logger.Info("ProvisionClient workflow completed successfully",
		"client_id", w.input.Req.GetClientId(),
		"tls_enabled", w.input.Req.GetEnableTls())

	return &identitypb.ProvisionClientOutput{
		SigningSecret: signingStoreOutput.SecretRef,
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

	// Set up activity options for vault operations
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

	// Step 2: Store admin identity in Vault using MSP directory structure
	logger.Info("Storing admin identity in Vault")
	var adminStoreOutput *vaultpb.StoreSecretOutput
	vaultPaths := NewVaultPaths(w.input.Req.GetOrganization())
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: vaultPaths.AdminMSP(),
			Data: map[string][]byte{
				"signcerts/admin.pem":    adminOutput.GetAdminIdentity().GetCertificate(),
				"keystore/admin-key.pem": adminOutput.GetAdminIdentity().GetPrivateKey(),
				"ca-certs/org-ca.pem":    adminOutput.GetAdminIdentity().GetCaCert(),
			},
		}).Get(ctx, &adminStoreOutput)
	if err != nil {
		logger.Error("Failed to store admin identity", "error", err)
		return nil, fmt.Errorf("failed to store admin identity: %w", err)
	}
	adminSecretRef := adminStoreOutput.SecretRef

	// Step 3: Store CA certificates in Vault
	logger.Info("Storing CA certificates in Vault")
	var caCertStoreOutput *vaultpb.StoreSecretOutput
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: vaultPaths.CACerts(),
			Data: map[string][]byte{
				"org-ca.pem": adminOutput.GetAdminIdentity().GetCaCert(),
			},
		}).Get(ctx, &caCertStoreOutput)
	if err != nil {
		logger.Error("Failed to store CA certificates", "error", err)
		return nil, fmt.Errorf("failed to store CA certificates: %w", err)
	}
	caCertRef := caCertStoreOutput.SecretRef

	// Step 4: Store NodeOU configuration in Vault
	logger.Info("Storing NodeOU configuration in Vault")
	nodeOuBytes, err := serializeNodeOUConfig(w.input.Req.GetOrgNodeOus())
	if err != nil {
		logger.Error("Failed to serialize NodeOU config", "error", err)
		return nil, fmt.Errorf("failed to serialize NodeOU config: %w", err)
	}

	var nodeOuConfigRef *vaultpb.SecretRef
	var nodeOuConfigStoreOutput *vaultpb.StoreSecretOutput
	err = workflow.ExecuteActivity(activityCtx, vaultpb.StoreSecretActivityName,
		&vaultpb.StoreSecretInput{
			Path: vaultPaths.Config(),
			Data: map[string][]byte{
				"node-ou-config.yaml": nodeOuBytes,
			},
		}).Get(ctx, &nodeOuConfigStoreOutput)
	if err != nil {
		logger.Error("Failed to store NodeOU config", "error", err)
		return nil, fmt.Errorf("failed to store NodeOU config: %w", err)
	}
	nodeOuConfigRef = nodeOuConfigStoreOutput.SecretRef

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
