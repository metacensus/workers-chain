package certificate

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	BootstrapMetacensusOrgWorkflowName = "certificate.BootstrapMetacensusOrgWorkflow"
)

// BootstrapMetacensusOrgWorkflow bootstraps the metacensus organization
// This workflow creates the initial admin identity and organization structure
func BootstrapMetacensusOrgWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Metacensus organization bootstrap workflow started")

	// Create admin identity for the organization
	adminConfig := ClientConfig{
		ClientID: "admin",
		CustomAttributes: map[string]string{
			"hf.Registrar.Roles":      "admin",
			"hf.Registrar.Attributes": "*",
		},
		Namespace: "fabric",
	}

	err := workflow.ExecuteChildWorkflow(ctx, ProvisionClientWorkflow, adminConfig).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to create admin identity", "error", err)
		return fmt.Errorf("admin identity creation failed: %w", err)
	}

	// Create initial peer identities
	peerConfigs := []PeerConfig{
		{
			PeerID:       "peer0",
			EnrollmentCA: "http://ca0.fabric.svc.cluster.local:7054",
			TLSCA:        "http://ca0.fabric.svc.cluster.local:7054",
			OrgNodeOUs: NodeOUConfig{
				Enable: true,
				Client: map[string]string{
					"certificateAuthorities": "true",
					"clientOU":               "true",
				},
				Peer: map[string]string{
					"certificateAuthorities": "true",
					"peerOU":                 "true",
				},
				Admin: map[string]string{
					"certificateAuthorities": "true",
					"adminOU":                "true",
				},
				Orderer: map[string]string{
					"certificateAuthorities": "true",
					"ordererOU":              "true",
				},
			},
			Namespace: "fabric",
		},
		{
			PeerID:       "peer1",
			EnrollmentCA: "http://ca0.fabric.svc.cluster.local:7054",
			TLSCA:        "http://ca0.fabric.svc.cluster.local:7054",
			OrgNodeOUs: NodeOUConfig{
				Enable: true,
				Client: map[string]string{
					"certificateAuthorities": "true",
					"clientOU":               "true",
				},
				Peer: map[string]string{
					"certificateAuthorities": "true",
					"peerOU":                 "true",
				},
				Admin: map[string]string{
					"certificateAuthorities": "true",
					"adminOU":                "true",
				},
				Orderer: map[string]string{
					"certificateAuthorities": "true",
					"ordererOU":              "true",
				},
			},
			Namespace: "fabric",
		},
	}

	// Provision peers in parallel
	for _, peerConfig := range peerConfigs {
		config := peerConfig // Capture the variable
		workflow.Go(ctx, func(ctx workflow.Context) {
			workflow.ExecuteChildWorkflow(ctx, ProvisionPeerWorkflow, config).Get(ctx, nil)
		})
	}

	// Create initial orderer identities
	ordererConfigs := []OrdererConfig{
		{
			OrdererID:    "orderer0",
			EnrollmentCA: "http://ca0.fabric.svc.cluster.local:7054",
			TLSCA:        "http://ca0.fabric.svc.cluster.local:7054",
			OrgNodeOUs: NodeOUConfig{
				Enable: true,
				Client: map[string]string{
					"certificateAuthorities": "true",
					"clientOU":               "true",
				},
				Peer: map[string]string{
					"certificateAuthorities": "true",
					"peerOU":                 "true",
				},
				Admin: map[string]string{
					"certificateAuthorities": "true",
					"adminOU":                "true",
				},
				Orderer: map[string]string{
					"certificateAuthorities": "true",
					"ordererOU":              "true",
				},
			},
			Namespace: "fabric",
		},
		{
			OrdererID:    "orderer1",
			EnrollmentCA: "http://ca0.fabric.svc.cluster.local:7054",
			TLSCA:        "http://ca0.fabric.svc.cluster.local:7054",
			OrgNodeOUs: NodeOUConfig{
				Enable: true,
				Client: map[string]string{
					"certificateAuthorities": "true",
					"clientOU":               "true",
				},
				Peer: map[string]string{
					"certificateAuthorities": "true",
					"peerOU":                 "true",
				},
				Admin: map[string]string{
					"certificateAuthorities": "true",
					"adminOU":                "true",
				},
				Orderer: map[string]string{
					"certificateAuthorities": "true",
					"ordererOU":              "true",
				},
			},
			Namespace: "fabric",
		},
		{
			OrdererID:    "orderer2",
			EnrollmentCA: "http://ca0.fabric.svc.cluster.local:7054",
			TLSCA:        "http://ca0.fabric.svc.cluster.local:7054",
			OrgNodeOUs: NodeOUConfig{
				Enable: true,
				Client: map[string]string{
					"certificateAuthorities": "true",
					"clientOU":               "true",
				},
				Peer: map[string]string{
					"certificateAuthorities": "true",
					"peerOU":                 "true",
				},
				Admin: map[string]string{
					"certificateAuthorities": "true",
					"adminOU":                "true",
				},
				Orderer: map[string]string{
					"certificateAuthorities": "true",
					"ordererOU":              "true",
				},
			},
			Namespace: "fabric",
		},
	}

	// Provision orderers in parallel
	for _, ordererConfig := range ordererConfigs {
		config := ordererConfig // Capture the variable
		workflow.Go(ctx, func(ctx workflow.Context) {
			workflow.ExecuteChildWorkflow(ctx, ProvisionOrdererWorkflow, config).Get(ctx, nil)
		})
	}

	// Wait for all provisioning to complete
	workflow.Sleep(ctx, 5*time.Second)

	logger.Info("Metacensus organization bootstrap workflow completed successfully",
		"peers", len(peerConfigs),
		"orderers", len(ordererConfigs))

	return nil
}
