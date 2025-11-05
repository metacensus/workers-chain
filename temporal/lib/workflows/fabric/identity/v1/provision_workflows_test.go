package identity

import (
	"context"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	fabricmocks "github.com/metacensus/infra/temporal/mocks/workflows/fabric/identity/v1"
	vaultmocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

var errActivityFailed = temporal.NewApplicationError("activity failed", "ActivityFailed")

func testProvisionPeerWorkflow(ctx workflow.Context, input *pb.ProvisionPeerWorkflowInput) (*pb.ProvisionPeerOutput, error) {
	w := &provisionPeerWorkflow{input: input}
	return w.Execute(ctx)
}

func testProvisionOrdererWorkflow(ctx workflow.Context, input *pb.ProvisionOrdererWorkflowInput) (*pb.ProvisionOrdererOutput, error) {
	w := &provisionOrdererWorkflow{input: input}
	return w.Execute(ctx)
}

func testProvisionClientWorkflow(ctx workflow.Context, input *pb.ProvisionClientWorkflowInput) (*pb.ProvisionClientOutput, error) {
	w := &provisionClientWorkflow{input: input}
	return w.Execute(ctx)
}

func TestProvisionPeerWorkflow(t *testing.T) {
	tests := []struct {
		name              string
		input             *pb.ProvisionPeerWorkflowInput
		generateSecretRes *pb.GenerateSecureSecretOutput
		generateSecretErr error
		registerRes       *pb.RegisterIdentityOutput
		registerErr       error
		enrollSigningRes  *pb.EnrollIdentityOutput
		enrollSigningErr  error
		enrollTlsRes      *pb.EnrollIdentityOutput
		enrollTlsErr      error
		expectErr         bool
	}{
		{
			name: "success - peer provisioned with all certificates",
			input: &pb.ProvisionPeerWorkflowInput{
				Req: &pb.ProvisionPeerInput{
					PeerId:       "peer0",
					Organization: "org1",
					Namespace:    "fabric",
					Hosts:        []string{"peer0.example.com"},
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "peer-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "peer0",
				Secret: "registered-secret",
			},
			enrollSigningRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("peer-signing-cert"),
				PrivateKey:  []byte("peer-signing-key"),
			},
			enrollTlsRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("peer-tls-cert"),
				PrivateKey:  []byte("peer-tls-key"),
			},
		},
		{
			name: "failure - secret generation fails",
			input: &pb.ProvisionPeerWorkflowInput{
				Req: &pb.ProvisionPeerInput{
					PeerId:       "peer0",
					Organization: "org1",
					Namespace:    "fabric",
				},
			},
			generateSecretErr: errActivityFailed,
			expectErr:         true,
		},
		{
			name: "failure - identity registration fails",
			input: &pb.ProvisionPeerWorkflowInput{
				Req: &pb.ProvisionPeerInput{
					PeerId:       "peer0",
					Organization: "org1",
					Namespace:    "fabric",
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "peer-secret-123"},
			registerErr:       errActivityFailed,
			expectErr:         true,
		},
		{
			name: "failure - signing certificate enrollment fails",
			input: &pb.ProvisionPeerWorkflowInput{
				Req: &pb.ProvisionPeerInput{
					PeerId:       "peer0",
					Organization: "org1",
					Namespace:    "fabric",
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "peer-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "peer0",
				Secret: "registered-secret",
			},
			enrollSigningErr: errActivityFailed,
			expectErr:        true,
		},
		{
			name: "failure - TLS certificate enrollment fails",
			input: &pb.ProvisionPeerWorkflowInput{
				Req: &pb.ProvisionPeerInput{
					PeerId:       "peer0",
					Organization: "org1",
					Namespace:    "fabric",
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "peer-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "peer0",
				Secret: "registered-secret",
			},
			enrollSigningRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("peer-signing-cert"),
				PrivateKey:  []byte("peer-signing-key"),
			},
			enrollTlsErr: errActivityFailed,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register the actual fabric identity workflows
			realWorkflows := NewWorkflows()
			pb.RegisterFabricIdentityWorkflows(env, realWorkflows)

			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)

			mockActivities.EXPECT().
				GenerateSecureSecret(mock.Anything, mock.MatchedBy(func(input *pb.GenerateSecureSecretInput) bool {
					return input.GetLength() == 32
				})).
				Return(tt.generateSecretRes, tt.generateSecretErr).
				Maybe()

			// Always mock RegisterIdentity activity, but vary return values based on test scenario
			mockActivities.EXPECT().
				RegisterIdentity(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, input *pb.RegisterIdentityInput) (*pb.RegisterIdentityOutput, error) {
					if tt.registerErr != nil {
						return nil, tt.registerErr
					}
					return tt.registerRes, nil
				}).
				Maybe()

			// Mock EnrollIdentity activities - can be called multiple times in success paths
			mockActivities.EXPECT().
				EnrollIdentity(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, input *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
					// Determine which error to return based on the profile
					if input.GetProfile() == "" && tt.enrollSigningErr != nil {
						return nil, tt.enrollSigningErr
					}
					if input.GetProfile() == "tls" && tt.enrollTlsErr != nil {
						return nil, tt.enrollTlsErr
					}
					// Return appropriate success response based on profile
					if input.GetProfile() == "" {
						return tt.enrollSigningRes, nil
					}
					return tt.enrollTlsRes, nil
				}).
				Maybe()

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Register vault activities mock - always register, but vary return values
			mockVaultActivities := vaultmocks.NewMockVaultActivities(t)

			// Always mock StoreSecret - it will be called in success paths
			// In failure paths, it won't be called, so set it to Maybe()
			if tt.generateSecretErr == nil && tt.registerErr == nil && tt.enrollSigningErr == nil && tt.enrollTlsErr == nil {
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.Anything).
					Return(&vaultpb.StoreSecretOutput{
						SecretRef: &vaultpb.SecretRef{
							Path:    "vault-secret-ref",
							Version: 1,
						},
					}, nil)
			} else {
				// For failure paths, still mock but use Maybe() to allow it to not be called
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.Anything).
					Return(&vaultpb.StoreSecretOutput{
						SecretRef: &vaultpb.SecretRef{
							Path:    "vault-secret-ref",
							Version: 1,
						},
					}, nil).
					Maybe()
			}
			vaultpb.RegisterVaultActivities(env, mockVaultActivities)

			env.ExecuteWorkflow(testProvisionPeerWorkflow, tt.input)

			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *pb.ProvisionPeerOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetIdentityName() != tt.registerRes.GetName() {
					t.Errorf("identity name mismatch: got %q, want %q",
						result.GetIdentityName(), tt.registerRes.GetName())
				}

				if result.GetOrganization() != tt.input.Req.GetOrganization() {
					t.Errorf("organization mismatch: got %q, want %q",
						result.GetOrganization(), tt.input.Req.GetOrganization())
				}

				if result.GetSigningSecret() == nil {
					t.Error("expected signing secret to be set")
				}

				if result.GetTlsSecret() == nil {
					t.Error("expected TLS secret to be set")
				}
			}
		})
	}
}

func TestProvisionOrdererWorkflow(t *testing.T) {
	tests := []struct {
		name              string
		input             *pb.ProvisionOrdererWorkflowInput
		generateSecretRes *pb.GenerateSecureSecretOutput
		generateSecretErr error
		registerRes       *pb.RegisterIdentityOutput
		registerErr       error
		enrollSigningRes  *pb.EnrollIdentityOutput
		enrollSigningErr  error
		enrollTlsRes      *pb.EnrollIdentityOutput
		enrollTlsErr      error
		expectErr         bool
	}{
		{
			name: "success - orderer provisioned with all certificates",
			input: &pb.ProvisionOrdererWorkflowInput{
				Req: &pb.ProvisionOrdererInput{
					OrdererId:    "orderer0",
					Organization: "org1",
					Namespace:    "fabric",
					Hosts:        []string{"orderer0.example.com"},
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "orderer-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "orderer0",
				Secret: "registered-secret",
			},
			enrollSigningRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("orderer-signing-cert"),
				PrivateKey:  []byte("orderer-signing-key"),
			},
			enrollTlsRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("orderer-tls-cert"),
				PrivateKey:  []byte("orderer-tls-key"),
			},
		},
		{
			name: "failure - secret generation fails",
			input: &pb.ProvisionOrdererWorkflowInput{
				Req: &pb.ProvisionOrdererInput{
					OrdererId:    "orderer0",
					Organization: "org1",
					Namespace:    "fabric",
				},
			},
			generateSecretErr: errActivityFailed,
			expectErr:         true,
		},
		{
			name: "failure - identity registration fails",
			input: &pb.ProvisionOrdererWorkflowInput{
				Req: &pb.ProvisionOrdererInput{
					OrdererId:    "orderer0",
					Organization: "org1",
					Namespace:    "fabric",
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "orderer-secret-123"},
			registerErr:       errActivityFailed,
			expectErr:         true,
		},
		{
			name: "failure - enrollment fails",
			input: &pb.ProvisionOrdererWorkflowInput{
				Req: &pb.ProvisionOrdererInput{
					OrdererId:    "orderer0",
					Organization: "org1",
					Namespace:    "fabric",
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "orderer-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "orderer0",
				Secret: "registered-secret",
			},
			enrollSigningRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("orderer-signing-cert"),
				PrivateKey:  []byte("orderer-signing-key"),
			},
			enrollTlsErr: errActivityFailed,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register the actual fabric identity workflows
			realWorkflows := NewWorkflows()
			pb.RegisterFabricIdentityWorkflows(env, realWorkflows)

			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)

			mockActivities.EXPECT().
				GenerateSecureSecret(mock.Anything, mock.MatchedBy(func(input *pb.GenerateSecureSecretInput) bool {
					return input.GetLength() == 32
				})).
				Return(tt.generateSecretRes, tt.generateSecretErr).
				Maybe()

			// Always mock RegisterIdentity activity, but vary return values based on test scenario
			mockActivities.EXPECT().
				RegisterIdentity(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, input *pb.RegisterIdentityInput) (*pb.RegisterIdentityOutput, error) {
					if tt.registerErr != nil {
						return nil, tt.registerErr
					}
					return tt.registerRes, nil
				}).
				Maybe()

			// Mock EnrollIdentity activities - can be called multiple times in success paths
			mockActivities.EXPECT().
				EnrollIdentity(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, input *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
					// Determine which error to return based on the profile
					if input.GetProfile() == "" && tt.enrollSigningErr != nil {
						return nil, tt.enrollSigningErr
					}
					if input.GetProfile() == "tls" && tt.enrollTlsErr != nil {
						return nil, tt.enrollTlsErr
					}
					// Return appropriate success response based on profile
					if input.GetProfile() == "" {
						return tt.enrollSigningRes, nil
					}
					return tt.enrollTlsRes, nil
				}).
				Maybe()

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Register vault activities mock - always register, but vary return values
			mockVaultActivities := vaultmocks.NewMockVaultActivities(t)

			// Always mock StoreSecret - it will be called in success paths
			// In failure paths, it won't be called, so set it to Maybe()
			if tt.generateSecretErr == nil && tt.registerErr == nil && tt.enrollSigningErr == nil && tt.enrollTlsErr == nil {
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.Anything).
					Return(&vaultpb.StoreSecretOutput{
						SecretRef: &vaultpb.SecretRef{
							Path:    "vault-secret-ref",
							Version: 1,
						},
					}, nil)
			} else {
				// For failure paths, still mock but use Maybe() to allow it to not be called
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.Anything).
					Return(&vaultpb.StoreSecretOutput{
						SecretRef: &vaultpb.SecretRef{
							Path:    "vault-secret-ref",
							Version: 1,
						},
					}, nil).
					Maybe()
			}
			vaultpb.RegisterVaultActivities(env, mockVaultActivities)

			env.ExecuteWorkflow(testProvisionOrdererWorkflow, tt.input)

			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *pb.ProvisionOrdererOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetIdentityName() != tt.registerRes.GetName() {
					t.Errorf("identity name mismatch: got %q, want %q",
						result.GetIdentityName(), tt.registerRes.GetName())
				}

				if result.GetOrganization() != tt.input.Req.GetOrganization() {
					t.Errorf("organization mismatch: got %q, want %q",
						result.GetOrganization(), tt.input.Req.GetOrganization())
				}
			}
		})
	}
}

func TestProvisionClientWorkflow(t *testing.T) {
	tests := []struct {
		name              string
		input             *pb.ProvisionClientWorkflowInput
		generateSecretRes *pb.GenerateSecureSecretOutput
		generateSecretErr error
		registerRes       *pb.RegisterIdentityOutput
		registerErr       error
		enrollSigningRes  *pb.EnrollIdentityOutput
		enrollSigningErr  error
		enrollTlsRes      *pb.EnrollIdentityOutput
		enrollTlsErr      error
		expectErr         bool
	}{
		{
			name: "success - client provisioned with signing only",
			input: &pb.ProvisionClientWorkflowInput{
				Req: &pb.ProvisionClientInput{
					ClientId:     "client1",
					Organization: "org1",
					EnableTls:    false,
					Hosts:        []string{},
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "client-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "client1",
				Secret: "registered-secret",
			},
			enrollSigningRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("client-signing-cert"),
				PrivateKey:  []byte("client-signing-key"),
			},
		},
		{
			name: "success - client provisioned with signing and TLS",
			input: &pb.ProvisionClientWorkflowInput{
				Req: &pb.ProvisionClientInput{
					ClientId:     "client1",
					Organization: "org1",
					EnableTls:    true,
					Hosts:        []string{"client1.example.com"},
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "client-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "client1",
				Secret: "registered-secret",
			},
			enrollSigningRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("client-signing-cert"),
				PrivateKey:  []byte("client-signing-key"),
			},
			enrollTlsRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("client-tls-cert"),
				PrivateKey:  []byte("client-tls-key"),
			},
		},
		{
			name: "failure - secret generation fails",
			input: &pb.ProvisionClientWorkflowInput{
				Req: &pb.ProvisionClientInput{
					ClientId:     "client1",
					Organization: "org1",
					EnableTls:    false,
				},
			},
			generateSecretErr: errActivityFailed,
			expectErr:         true,
		},
		{
			name: "failure - client registration fails",
			input: &pb.ProvisionClientWorkflowInput{
				Req: &pb.ProvisionClientInput{
					ClientId:     "client1",
					Organization: "org1",
					EnableTls:    false,
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "client-secret-123"},
			registerErr:       errActivityFailed,
			expectErr:         true,
		},
		{
			name: "failure - signing enrollment fails",
			input: &pb.ProvisionClientWorkflowInput{
				Req: &pb.ProvisionClientInput{
					ClientId:     "client1",
					Organization: "org1",
					EnableTls:    false,
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "client-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "client1",
				Secret: "registered-secret",
			},
			enrollSigningErr: errActivityFailed,
			expectErr:        true,
		},
		{
			name: "failure - TLS enrollment fails when TLS enabled",
			input: &pb.ProvisionClientWorkflowInput{
				Req: &pb.ProvisionClientInput{
					ClientId:     "client1",
					Organization: "org1",
					EnableTls:    true,
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{Secret: "client-secret-123"},
			registerRes: &pb.RegisterIdentityOutput{
				Name:   "client1",
				Secret: "registered-secret",
			},
			enrollSigningRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("client-signing-cert"),
				PrivateKey:  []byte("client-signing-key"),
			},
			enrollTlsErr: errActivityFailed,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register the actual fabric identity workflows
			realWorkflows := NewWorkflows()
			pb.RegisterFabricIdentityWorkflows(env, realWorkflows)

			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)

			mockActivities.EXPECT().
				GenerateSecureSecret(mock.Anything, mock.MatchedBy(func(input *pb.GenerateSecureSecretInput) bool {
					return input.GetLength() == 32
				})).
				Return(tt.generateSecretRes, tt.generateSecretErr).
				Maybe()

			// Always mock RegisterIdentity activity, but vary return values based on test scenario
			mockActivities.EXPECT().
				RegisterIdentity(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, input *pb.RegisterIdentityInput) (*pb.RegisterIdentityOutput, error) {
					if tt.registerErr != nil {
						return nil, tt.registerErr
					}
					return tt.registerRes, nil
				}).
				Maybe()

			// Mock EnrollIdentity activities - can be called multiple times in success paths
			mockActivities.EXPECT().
				EnrollIdentity(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, input *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
					// Determine which error to return based on the profile
					if input.GetProfile() == "" && tt.enrollSigningErr != nil {
						return nil, tt.enrollSigningErr
					}
					if input.GetProfile() == "tls" && tt.enrollTlsErr != nil {
						return nil, tt.enrollTlsErr
					}
					// Return appropriate success response based on profile
					if input.GetProfile() == "" {
						return tt.enrollSigningRes, nil
					}
					return tt.enrollTlsRes, nil
				}).
				Maybe()

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Register vault activities mock - always register, but vary return values
			mockVaultActivities := vaultmocks.NewMockVaultActivities(t)

			// Always mock StoreSecret - it will be called in success paths
			// In failure paths, it won't be called, so set it to Maybe()
			if tt.generateSecretErr == nil && tt.registerErr == nil && tt.enrollSigningErr == nil && (tt.enrollTlsErr == nil || !tt.input.Req.GetEnableTls()) {
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.Anything).
					Return(&vaultpb.StoreSecretOutput{
						SecretRef: &vaultpb.SecretRef{
							Path:    "vault-secret-ref",
							Version: 1,
						},
					}, nil)
			} else {
				// For failure paths, still mock but use Maybe() to allow it to not be called
				mockVaultActivities.EXPECT().
					StoreSecret(mock.Anything, mock.Anything).
					Return(&vaultpb.StoreSecretOutput{
						SecretRef: &vaultpb.SecretRef{
							Path:    "vault-secret-ref",
							Version: 1,
						},
					}, nil).
					Maybe()
			}
			vaultpb.RegisterVaultActivities(env, mockVaultActivities)

			env.ExecuteWorkflow(testProvisionClientWorkflow, tt.input)

			if !env.IsWorkflowCompleted() {
				t.Fatal("expected workflow to complete")
			}

			if tt.expectErr {
				if env.GetWorkflowError() == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}

				var result *pb.ProvisionClientOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetIdentityName() != tt.registerRes.GetName() {
					t.Errorf("identity name mismatch: got %q, want %q",
						result.GetIdentityName(), tt.registerRes.GetName())
				}

				if result.GetSigningSecret() == nil {
					t.Error("expected signing secret to be set")
				}

				if tt.input.Req.GetEnableTls() {
					if result.GetTlsSecret() == nil {
						t.Error("expected TLS secret to be set when TLS enabled")
					}
				} else {
					if result.GetTlsSecret() != nil {
						t.Error("expected TLS secret to be nil when TLS disabled")
					}
				}
			}
		})
	}
}
