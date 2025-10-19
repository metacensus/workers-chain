package identity

import (
	"errors"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	fabricmocks "github.com/metacensus/infra/temporal/mocks/workflows/fabric/identity/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// Mock implementations for vault workflows used in tests
type mockStoreSecretWorkflow struct {
	result *vaultpb.SecretRef
	err    error
}

func (m *mockStoreSecretWorkflow) Execute(ctx workflow.Context) (*vaultpb.SecretRef, error) {
	return m.result, m.err
}

// Test wrapper function for provisionPeerWorkflow
func testProvisionPeerWorkflow(ctx workflow.Context, input *pb.ProvisionPeerWorkflowInput) (*pb.ProvisionPeerOutput, error) {
	w := &provisionPeerWorkflow{input: input}
	return w.Execute(ctx)
}

func TestProvisionPeerWorkflow(t *testing.T) {
	tests := []struct {
		name                  string
		input                 *pb.ProvisionPeerInput
		vaultStoreSecretErr   error
		childWorkflowErr      error
		activityErr           error
		expectErr             bool
		validateVaultCalls    bool
		expectedVaultCallsNum int
	}{
		{
			name: "success - complete peer provisioning",
			input: &pb.ProvisionPeerInput{
				PeerId:       "peer0",
				Organization: "org1",
				Namespace:    "fabric-org1",
				Hosts:        []string{"peer0.org1.example.com"},
			},
			validateVaultCalls:    true,
			expectedVaultCallsNum: 2, // signing + TLS secrets
		},
		{
			name: "failure - vault store secret fails",
			input: &pb.ProvisionPeerInput{
				PeerId:       "peer0",
				Organization: "org1",
				Namespace:    "fabric-org1",
			},
			vaultStoreSecretErr: errors.New("vault unavailable"),
			expectErr:           true,
		},
		{
			name: "failure - enroll workflow fails",
			input: &pb.ProvisionPeerInput{
				PeerId:       "peer0",
				Organization: "org1",
				Namespace:    "fabric-org1",
			},
			childWorkflowErr: errors.New("enrollment failed"),
			expectErr:        true,
		},
		{
			name: "failure - generate secret activity fails",
			input: &pb.ProvisionPeerInput{
				PeerId:       "peer0",
				Organization: "org1",
				Namespace:    "fabric-org1",
			},
			activityErr: errors.New("crypto random failed"),
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Mock GenerateSecureSecret activity
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				GenerateSecureSecret(mock.Anything, mock.AnythingOfType("*fabricidentityv1.GenerateSecureSecretInput")).
				Return(&pb.GenerateSecureSecretOutput{Secret: "test-secret-12345"}, tt.activityErr).
				Maybe()

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Mock RegisterIdentity child workflow
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.RegisterIdentityWorkflowInput) (*pb.RegisterIdentityOutput, error) {
					if tt.childWorkflowErr != nil {
						return nil, tt.childWorkflowErr
					}
					return &pb.RegisterIdentityOutput{
						Name:   input.Req.GetName(),
						Secret: "enrollment-secret",
					}, nil
				},
				workflow.RegisterOptions{Name: pb.RegisterIdentityWorkflowName},
			)

			// Mock EnrollIdentity child workflow
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.EnrollIdentityWorkflowInput) (*pb.EnrollIdentityOutput, error) {
					if tt.childWorkflowErr != nil {
						return nil, tt.childWorkflowErr
					}
					return &pb.EnrollIdentityOutput{
						Certificate: []byte("test-cert-pem"),
						PrivateKey:  []byte("test-key-pem"),
					}, nil
				},
				workflow.RegisterOptions{Name: pb.EnrollIdentityWorkflowName},
			)

			// Mock Vault StoreSecret workflow using generated mocks
			mockVaultWorkflows := vaultmocks.NewMockVaultWorkflows(t)
			if tt.validateVaultCalls {
				// Expect exactly 2 calls (signing + TLS)
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(&mockStoreSecretWorkflow{
						result: &vaultpb.SecretRef{
							Name:      "test-secret",
							Namespace: tt.input.Namespace,
							Version:   1,
						},
					}, nil).
					Times(tt.expectedVaultCallsNum)
			} else if tt.vaultStoreSecretErr != nil {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(nil, tt.vaultStoreSecretErr).
					Maybe()
			} else {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(&mockStoreSecretWorkflow{
						result: &vaultpb.SecretRef{
							Name:      "test-secret",
							Namespace: tt.input.Namespace,
							Version:   1,
						},
					}, nil).
					Maybe()
			}

			// Register the vault workflows
			vaultpb.RegisterVaultWorkflows(env, mockVaultWorkflows)

			// Execute workflow
			env.ExecuteWorkflow(testProvisionPeerWorkflow, &pb.ProvisionPeerWorkflowInput{Req: tt.input})

			// Verify completion
			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected workflow error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				var result *pb.ProvisionPeerOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
				if result.SigningSecret == nil || result.TlsSecret == nil {
					t.Error("expected both signing and TLS secrets to be non-nil")
				}
				if result.Organization != tt.input.Organization {
					t.Errorf("organization mismatch: got %q, want %q", result.Organization, tt.input.Organization)
				}
			}

			// Verify mock expectations
			mockVaultWorkflows.AssertExpectations(t)
		})
	}
}

// Test wrapper function for provisionOrdererWorkflow
func testProvisionOrdererWorkflow(ctx workflow.Context, input *pb.ProvisionOrdererWorkflowInput) (*pb.ProvisionOrdererOutput, error) {
	w := &provisionOrdererWorkflow{input: input}
	return w.Execute(ctx)
}

func TestProvisionOrdererWorkflow(t *testing.T) {
	tests := []struct {
		name                  string
		input                 *pb.ProvisionOrdererInput
		vaultStoreSecretErr   error
		childWorkflowErr      error
		activityErr           error
		expectErr             bool
		validateVaultCalls    bool
		expectedVaultCallsNum int
	}{
		{
			name: "success - complete orderer provisioning",
			input: &pb.ProvisionOrdererInput{
				OrdererId:    "orderer0",
				Organization: "org1",
				Namespace:    "fabric-org1",
				Hosts:        []string{"orderer0.org1.example.com"},
			},
			validateVaultCalls:    true,
			expectedVaultCallsNum: 2, // signing + TLS secrets
		},
		{
			name: "failure - vault store secret fails",
			input: &pb.ProvisionOrdererInput{
				OrdererId:    "orderer0",
				Organization: "org1",
				Namespace:    "fabric-org1",
			},
			vaultStoreSecretErr: errors.New("vault unavailable"),
			expectErr:           true,
		},
		{
			name: "failure - register workflow fails",
			input: &pb.ProvisionOrdererInput{
				OrdererId:    "orderer0",
				Organization: "org1",
				Namespace:    "fabric-org1",
			},
			childWorkflowErr: errors.New("registration failed"),
			expectErr:        true,
		},
		{
			name: "failure - generate secret activity fails",
			input: &pb.ProvisionOrdererInput{
				OrdererId:    "orderer0",
				Organization: "org1",
				Namespace:    "fabric-org1",
			},
			activityErr: errors.New("crypto random failed"),
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Mock GenerateSecureSecret activity
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				GenerateSecureSecret(mock.Anything, mock.AnythingOfType("*fabricidentityv1.GenerateSecureSecretInput")).
				Return(&pb.GenerateSecureSecretOutput{Secret: "test-secret-12345"}, tt.activityErr).
				Maybe()

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Mock RegisterIdentity child workflow
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.RegisterIdentityWorkflowInput) (*pb.RegisterIdentityOutput, error) {
					if tt.childWorkflowErr != nil {
						return nil, tt.childWorkflowErr
					}
					return &pb.RegisterIdentityOutput{
						Name:   input.Req.GetName(),
						Secret: "enrollment-secret",
					}, nil
				},
				workflow.RegisterOptions{Name: pb.RegisterIdentityWorkflowName},
			)

			// Mock EnrollIdentity child workflow
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.EnrollIdentityWorkflowInput) (*pb.EnrollIdentityOutput, error) {
					if tt.childWorkflowErr != nil {
						return nil, tt.childWorkflowErr
					}
					return &pb.EnrollIdentityOutput{
						Certificate: []byte("test-cert-pem"),
						PrivateKey:  []byte("test-key-pem"),
					}, nil
				},
				workflow.RegisterOptions{Name: pb.EnrollIdentityWorkflowName},
			)

			// Mock Vault StoreSecret workflow using generated mocks
			mockVaultWorkflows := vaultmocks.NewMockVaultWorkflows(t)
			if tt.validateVaultCalls {
				// Expect exactly 2 calls (signing + TLS)
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(&mockStoreSecretWorkflow{
						result: &vaultpb.SecretRef{
							Name:      "test-secret",
							Namespace: tt.input.Namespace,
							Version:   1,
						},
					}, nil).
					Times(tt.expectedVaultCallsNum)
			} else if tt.vaultStoreSecretErr != nil {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(nil, tt.vaultStoreSecretErr).
					Maybe()
			} else {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(&mockStoreSecretWorkflow{
						result: &vaultpb.SecretRef{
							Name:      "test-secret",
							Namespace: tt.input.Namespace,
							Version:   1,
						},
					}, nil).
					Maybe()
			}

			// Register the vault workflows
			vaultpb.RegisterVaultWorkflows(env, mockVaultWorkflows)

			// Execute workflow
			env.ExecuteWorkflow(testProvisionOrdererWorkflow, &pb.ProvisionOrdererWorkflowInput{Req: tt.input})

			// Verify completion
			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected workflow error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				var result *pb.ProvisionOrdererOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
				if result.SigningSecret == nil || result.TlsSecret == nil {
					t.Error("expected both signing and TLS secrets to be non-nil")
				}
				if result.Organization != tt.input.Organization {
					t.Errorf("organization mismatch: got %q, want %q", result.Organization, tt.input.Organization)
				}
			}

			// Verify mock expectations
			mockVaultWorkflows.AssertExpectations(t)
		})
	}
}

// Test wrapper function for provisionClientWorkflow
func testProvisionClientWorkflow(ctx workflow.Context, input *pb.ProvisionClientWorkflowInput) (*pb.ProvisionClientOutput, error) {
	w := &provisionClientWorkflow{input: input}
	return w.Execute(ctx)
}

func TestProvisionClientWorkflow(t *testing.T) {
	tests := []struct {
		name                  string
		input                 *pb.ProvisionClientInput
		vaultStoreSecretErr   error
		childWorkflowErr      error
		activityErr           error
		expectErr             bool
		validateVaultCalls    bool
		expectedVaultCallsNum int
	}{
		{
			name: "success - client without TLS",
			input: &pb.ProvisionClientInput{
				ClientId:     "client1",
				Organization: "org1",
				Namespace:    "fabric-org1",
				EnableTls:    false,
			},
			validateVaultCalls:    true,
			expectedVaultCallsNum: 1, // signing secret only
		},
		{
			name: "success - client with TLS",
			input: &pb.ProvisionClientInput{
				ClientId:     "client1",
				Organization: "org1",
				Namespace:    "fabric-org1",
				EnableTls:    true,
				Hosts:        []string{"client1.org1.example.com"},
			},
			validateVaultCalls:    true,
			expectedVaultCallsNum: 2, // signing + TLS secrets
		},
		{
			name: "failure - vault store secret fails",
			input: &pb.ProvisionClientInput{
				ClientId:     "client1",
				Organization: "org1",
				Namespace:    "fabric-org1",
				EnableTls:    false,
			},
			vaultStoreSecretErr: errors.New("vault unavailable"),
			expectErr:           true,
		},
		{
			name: "failure - enroll workflow fails",
			input: &pb.ProvisionClientInput{
				ClientId:     "client1",
				Organization: "org1",
				Namespace:    "fabric-org1",
				EnableTls:    false,
			},
			childWorkflowErr: errors.New("enrollment failed"),
			expectErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Mock GenerateSecureSecret activity
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				GenerateSecureSecret(mock.Anything, mock.AnythingOfType("*fabricidentityv1.GenerateSecureSecretInput")).
				Return(&pb.GenerateSecureSecretOutput{Secret: "test-secret-12345"}, tt.activityErr).
				Maybe()

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Mock RegisterIdentity child workflow
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.RegisterIdentityWorkflowInput) (*pb.RegisterIdentityOutput, error) {
					if tt.childWorkflowErr != nil {
						return nil, tt.childWorkflowErr
					}
					return &pb.RegisterIdentityOutput{
						Name:   input.Req.GetName(),
						Secret: "enrollment-secret",
					}, nil
				},
				workflow.RegisterOptions{Name: pb.RegisterIdentityWorkflowName},
			)

			// Mock EnrollIdentity child workflow
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.EnrollIdentityWorkflowInput) (*pb.EnrollIdentityOutput, error) {
					if tt.childWorkflowErr != nil {
						return nil, tt.childWorkflowErr
					}
					return &pb.EnrollIdentityOutput{
						Certificate: []byte("test-cert-pem"),
						PrivateKey:  []byte("test-key-pem"),
					}, nil
				},
				workflow.RegisterOptions{Name: pb.EnrollIdentityWorkflowName},
			)

			// Mock Vault StoreSecret workflow
			mockVaultWorkflows := vaultmocks.NewMockVaultWorkflows(t)
			if tt.validateVaultCalls {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(&mockStoreSecretWorkflow{
						result: &vaultpb.SecretRef{
							Name:      "test-secret",
							Namespace: tt.input.Namespace,
							Version:   1,
						},
					}, nil).
					Times(tt.expectedVaultCallsNum)
			} else if tt.vaultStoreSecretErr != nil {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(nil, tt.vaultStoreSecretErr).
					Maybe()
			} else {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(&mockStoreSecretWorkflow{
						result: &vaultpb.SecretRef{
							Name:      "test-secret",
							Namespace: tt.input.Namespace,
							Version:   1,
						},
					}, nil).
					Maybe()
			}

			vaultpb.RegisterVaultWorkflows(env, mockVaultWorkflows)

			// Execute workflow
			env.ExecuteWorkflow(testProvisionClientWorkflow, &pb.ProvisionClientWorkflowInput{Req: tt.input})

			// Verify completion
			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected workflow error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				var result *pb.ProvisionClientOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
				if result.SigningSecret == nil {
					t.Error("expected signing secret to be non-nil")
				}
				if tt.input.EnableTls && result.TlsSecret == nil {
					t.Error("expected TLS secret to be non-nil when TLS is enabled")
				}
				if result.Organization != tt.input.Organization {
					t.Errorf("organization mismatch: got %q, want %q", result.Organization, tt.input.Organization)
				}
			}

			// Verify mock expectations
			mockVaultWorkflows.AssertExpectations(t)
		})
	}
}

// Test wrapper function for provisionOrgWorkflow
func testProvisionOrgWorkflow(ctx workflow.Context, input *pb.ProvisionOrgWorkflowInput) (*pb.ProvisionOrgOutput, error) {
	w := &provisionOrgWorkflow{input: input}
	return w.Execute(ctx)
}

func TestProvisionOrgWorkflow(t *testing.T) {
	tests := []struct {
		name                 string
		input                *pb.ProvisionOrgInput
		bootstrapAdminErr    error
		vaultStoreSecretErr  error
		provisionPeerErr     error
		provisionOrdererErr  error
		expectErr            bool
		validateVaultCalls   bool
		expectedVaultCalls   int
		expectedPeerCalls    int
		expectedOrdererCalls int
	}{
		{
			name: "success - org with peers and orderers",
			input: &pb.ProvisionOrgInput{
				Organization: "org1",
				Namespace:    "fabric-org1",
				Admin: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "adminpw",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
				PeerIds:    []string{"peer0", "peer1"},
				OrdererIds: []string{"orderer0"},
				OrgNodeOus: &pb.NodeOUConfig{Enable: true},
			},
			validateVaultCalls:   true,
			expectedVaultCalls:   3, // admin + CA certs + NodeOU config
			expectedPeerCalls:    2,
			expectedOrdererCalls: 1,
		},
		{
			name: "success - org without orderers",
			input: &pb.ProvisionOrgInput{
				Organization: "org1",
				Namespace:    "fabric-org1",
				Admin: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
				PeerIds:    []string{"peer0"},
				OrdererIds: []string{},
				OrgNodeOus: &pb.NodeOUConfig{Enable: true},
			},
			validateVaultCalls:   true,
			expectedVaultCalls:   3,
			expectedPeerCalls:    1,
			expectedOrdererCalls: 0,
		},
		{
			name: "failure - bootstrap admin fails",
			input: &pb.ProvisionOrgInput{
				Organization: "org1",
				Namespace:    "fabric-org1",
				Admin: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
				PeerIds: []string{"peer0"},
			},
			bootstrapAdminErr: errors.New("admin bootstrap failed"),
			expectErr:         true,
		},
		{
			name: "failure - vault store fails",
			input: &pb.ProvisionOrgInput{
				Organization: "org1",
				Namespace:    "fabric-org1",
				Admin: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
				PeerIds: []string{"peer0"},
			},
			vaultStoreSecretErr: errors.New("vault unavailable"),
			expectErr:           true,
		},
		{
			name: "failure - peer provisioning fails",
			input: &pb.ProvisionOrgInput{
				Organization: "org1",
				Namespace:    "fabric-org1",
				Admin: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
				PeerIds: []string{"peer0"},
			},
			provisionPeerErr: errors.New("peer provisioning failed"),
			expectErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Mock BootstrapAdmin child workflow
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.BootstrapAdminWorkflowInput) (*pb.BootstrapAdminOutput, error) {
					if tt.bootstrapAdminErr != nil {
						return nil, tt.bootstrapAdminErr
					}
					return &pb.BootstrapAdminOutput{
						AdminIdentity: &pb.AdminIdentity{
							Id:          input.Req.GetAdminId(),
							Certificate: []byte("admin-cert"),
							PrivateKey:  []byte("admin-key"),
							CaCert:      []byte("ca-cert"),
						},
						Success: true,
					}, nil
				},
				workflow.RegisterOptions{Name: pb.BootstrapAdminWorkflowName},
			)

			// Mock ProvisionPeer child workflow
			peerCallCount := 0
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.ProvisionPeerWorkflowInput) (*pb.ProvisionPeerOutput, error) {
					peerCallCount++
					if tt.provisionPeerErr != nil {
						return nil, tt.provisionPeerErr
					}
					return &pb.ProvisionPeerOutput{
						SigningSecret: &vaultpb.SecretRef{Name: "peer-signing"},
						TlsSecret:     &vaultpb.SecretRef{Name: "peer-tls"},
						IdentityName:  input.Req.GetPeerId(),
						Organization:  input.Req.GetOrganization(),
					}, nil
				},
				workflow.RegisterOptions{Name: pb.ProvisionPeerWorkflowName},
			)

			// Mock ProvisionOrderer child workflow
			ordererCallCount := 0
			env.RegisterWorkflowWithOptions(
				func(ctx workflow.Context, input *pb.ProvisionOrdererWorkflowInput) (*pb.ProvisionOrdererOutput, error) {
					ordererCallCount++
					if tt.provisionOrdererErr != nil {
						return nil, tt.provisionOrdererErr
					}
					return &pb.ProvisionOrdererOutput{
						SigningSecret: &vaultpb.SecretRef{Name: "orderer-signing"},
						TlsSecret:     &vaultpb.SecretRef{Name: "orderer-tls"},
						IdentityName:  input.Req.GetOrdererId(),
						Organization:  input.Req.GetOrganization(),
					}, nil
				},
				workflow.RegisterOptions{Name: pb.ProvisionOrdererWorkflowName},
			)

			// Mock Vault StoreSecret workflow
			mockVaultWorkflows := vaultmocks.NewMockVaultWorkflows(t)
			if tt.validateVaultCalls {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(&mockStoreSecretWorkflow{
						result: &vaultpb.SecretRef{
							Name:      "test-secret",
							Namespace: tt.input.Namespace,
							Version:   1,
						},
					}, nil).
					Times(tt.expectedVaultCalls)
			} else if tt.vaultStoreSecretErr != nil {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(nil, tt.vaultStoreSecretErr).
					Maybe()
			} else {
				mockVaultWorkflows.EXPECT().
					StoreSecret(
						mock.AnythingOfType("*internal.valueCtx"),
						mock.AnythingOfType("*vaultv1.StoreSecretWorkflowInput"),
					).
					Return(&mockStoreSecretWorkflow{
						result: &vaultpb.SecretRef{
							Name:      "test-secret",
							Namespace: tt.input.Namespace,
							Version:   1,
						},
					}, nil).
					Maybe()
			}

			vaultpb.RegisterVaultWorkflows(env, mockVaultWorkflows)

			// Execute workflow
			env.ExecuteWorkflow(testProvisionOrgWorkflow, &pb.ProvisionOrgWorkflowInput{Req: tt.input})

			// Verify completion
			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected workflow error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				var result *pb.ProvisionOrgOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
				if result.Organization != tt.input.Organization {
					t.Errorf("organization mismatch: got %q, want %q", result.Organization, tt.input.Organization)
				}
				if result.AdminSecret == nil {
					t.Error("expected admin secret to be non-nil")
				}
				if len(result.Peers) != tt.expectedPeerCalls {
					t.Errorf("peer count mismatch: got %d, want %d", len(result.Peers), tt.expectedPeerCalls)
				}
				if len(result.Orderers) != tt.expectedOrdererCalls {
					t.Errorf("orderer count mismatch: got %d, want %d", len(result.Orderers), tt.expectedOrdererCalls)
				}

				// Verify the actual call counts
				if peerCallCount != tt.expectedPeerCalls {
					t.Errorf("peer workflow call count: got %d, want %d", peerCallCount, tt.expectedPeerCalls)
				}
				if ordererCallCount != tt.expectedOrdererCalls {
					t.Errorf("orderer workflow call count: got %d, want %d", ordererCallCount, tt.expectedOrdererCalls)
				}
			}

			// Verify mock expectations
			mockVaultWorkflows.AssertExpectations(t)
		})
	}
}
