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

var (
	// Test error sentinels for activity errors
	errActivityFailed = errors.New("activity failed")
)

// Test wrapper functions that accept workflow.Context
func testRegisterIdentityWorkflow(ctx workflow.Context, input *pb.RegisterIdentityWorkflowInput) (*pb.RegisterIdentityOutput, error) {
	w := &registerIdentityWorkflow{input: input}
	return w.Execute(ctx)
}

func testEnrollIdentityWorkflow(ctx workflow.Context, input *pb.EnrollIdentityWorkflowInput) (*pb.EnrollIdentityOutput, error) {
	w := &enrollIdentityWorkflow{input: input}
	return w.Execute(ctx)
}

func testAssembleMspWorkflow(ctx workflow.Context, input *pb.AssembleMspWorkflowInput) (*pb.AssembleMspOutput, error) {
	w := &assembleMspWorkflow{input: input}
	return w.Execute(ctx)
}

func testBootstrapAdminWorkflow(ctx workflow.Context, input *pb.BootstrapAdminWorkflowInput) (*pb.BootstrapAdminOutput, error) {
	w := &bootstrapAdminWorkflow{input: input}
	return w.Execute(ctx)
}

func TestRegisterIdentityWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		input         *pb.RegisterIdentityWorkflowInput
		activityRes   *pb.RegisterIdentityOutput
		activityErr   error
		expectErr     bool
		expectErrType error
	}{
		{
			name: "success - peer registration",
			input: &pb.RegisterIdentityWorkflowInput{
				Req: &pb.RegisterIdentityInput{
					Name:        "peer0",
					Type:        "peer",
					Affiliation: "org1",
				},
			},
			activityRes: &pb.RegisterIdentityOutput{
				Name:   "peer0",
				Type:   "peer",
				Secret: "generated-secret",
			},
		},
		{
			name: "success - orderer registration",
			input: &pb.RegisterIdentityWorkflowInput{
				Req: &pb.RegisterIdentityInput{
					Name:        "orderer0",
					Type:        "orderer",
					Affiliation: "org1",
				},
			},
			activityRes: &pb.RegisterIdentityOutput{
				Name:   "orderer0",
				Type:   "orderer",
				Secret: "generated-secret",
			},
		},
		{
			name: "activity failure propagated",
			input: &pb.RegisterIdentityWorkflowInput{
				Req: &pb.RegisterIdentityInput{
					Name: "peer1",
					Type: "peer",
				},
			},
			activityErr: errActivityFailed,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				RegisterIdentity(mock.Anything, mock.MatchedBy(func(req *pb.RegisterIdentityInput) bool {
					if req.GetName() != tt.input.Req.GetName() {
						t.Errorf("name mismatch: got %q, want %q", req.GetName(), tt.input.Req.GetName())
						return false
					}
					return true
				})).
				Return(tt.activityRes, tt.activityErr)

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testRegisterIdentityWorkflow, tt.input)

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

				var result *pb.RegisterIdentityOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetName() != tt.activityRes.GetName() {
					t.Errorf("name mismatch: got %q, want %q", result.GetName(), tt.activityRes.GetName())
				}
			}
		})
	}
}

func TestEnrollIdentityWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.EnrollIdentityWorkflowInput
		activityRes *pb.EnrollIdentityOutput
		activityErr error
		expectErr   bool
	}{
		{
			name: "success - enroll with default profile",
			input: &pb.EnrollIdentityWorkflowInput{
				Req: &pb.EnrollIdentityInput{
					Name:   "peer0",
					Secret: "secret123",
				},
			},
			activityRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("cert-data"),
				PrivateKey:  []byte("key-data"),
				CaCert:      []byte("ca-cert-data"),
			},
		},
		{
			name: "success - enroll with TLS profile",
			input: &pb.EnrollIdentityWorkflowInput{
				Req: &pb.EnrollIdentityInput{
					Name:    "peer0",
					Secret:  "secret123",
					Profile: "tls",
				},
			},
			activityRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("tls-cert-data"),
				PrivateKey:  []byte("tls-key-data"),
				CaCert:      []byte("ca-cert-data"),
				Profile:     "tls",
			},
		},
		{
			name: "activity failure propagated",
			input: &pb.EnrollIdentityWorkflowInput{
				Req: &pb.EnrollIdentityInput{
					Name:   "peer1",
					Secret: "invalid",
				},
			},
			activityErr: errActivityFailed,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				EnrollIdentity(mock.Anything, mock.MatchedBy(func(req *pb.EnrollIdentityInput) bool {
					if req.GetName() != tt.input.Req.GetName() {
						t.Errorf("name mismatch: got %q, want %q", req.GetName(), tt.input.Req.GetName())
						return false
					}
					return true
				})).
				Return(tt.activityRes, tt.activityErr)

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testEnrollIdentityWorkflow, tt.input)

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

				var result *pb.EnrollIdentityOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if len(result.GetCertificate()) == 0 || len(result.GetPrivateKey()) == 0 {
					t.Error("expected certificate and private key in result")
				}
			}
		})
	}
}

func TestAssembleMspWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.AssembleMspWorkflowInput
		activityRes *pb.AssembleMspOutput
		activityErr error
		expectErr   bool
	}{
		{
			name: "success - assemble MSP",
			input: &pb.AssembleMspWorkflowInput{
				Req: &pb.AssembleMspInput{
					SigningEnrollment: &pb.EnrollIdentityOutput{
						Certificate: []byte("signing-cert"),
						PrivateKey:  []byte("signing-key"),
					},
					TlsEnrollment: &pb.EnrollIdentityOutput{
						Certificate: []byte("tls-cert"),
						PrivateKey:  []byte("tls-key"),
					},
					CaCerts:      [][]byte{[]byte("ca-cert-1"), []byte("ca-cert-2")},
					NodeOuConfig: &pb.NodeOUConfig{Enable: true},
				},
			},
			activityRes: &pb.AssembleMspOutput{
				SigningCert:  []byte("signing-cert"),
				SigningKey:   []byte("signing-key"),
				TlsCert:      []byte("tls-cert"),
				TlsKey:       []byte("tls-key"),
				CaCerts:      [][]byte{[]byte("ca-cert-1"), []byte("ca-cert-2")},
				NodeOuConfig: &pb.NodeOUConfig{Enable: true},
			},
		},
		{
			name: "activity failure propagated",
			input: &pb.AssembleMspWorkflowInput{
				Req: &pb.AssembleMspInput{
					SigningEnrollment: &pb.EnrollIdentityOutput{},
				},
			},
			activityErr: errActivityFailed,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)
			mockActivities.EXPECT().
				AssembleMsp(mock.Anything, mock.AnythingOfType("*fabricidentityv1.AssembleMspInput")).
				Return(tt.activityRes, tt.activityErr)

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testAssembleMspWorkflow, tt.input)

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

				var result *pb.AssembleMspOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if len(result.GetSigningCert()) == 0 || len(result.GetTlsCert()) == 0 {
					t.Error("expected signing and TLS certificates in result")
				}
			}
		})
	}
}

func TestBootstrapAdminWorkflow(t *testing.T) {
	tests := []struct {
		name                     string
		input                    *pb.BootstrapAdminWorkflowInput
		generateSecretRes        *pb.GenerateSecureSecretOutput
		generateSecretErr        error
		createAdminRes           *pb.CreateAdminIdentityOutput
		createAdminErr           error
		expectErr                bool
		expectGenerateSecretCall bool
	}{
		{
			name: "success - with provided secret",
			input: &pb.BootstrapAdminWorkflowInput{
				Req: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "provided-secret",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
					VaultNamespace: "org1",
				},
			},
			createAdminRes: &pb.CreateAdminIdentityOutput{
				AdminIdentity: &pb.AdminIdentity{
					Id:          "admin",
					Secret:      "provided-secret",
					Certificate: []byte("admin-cert"),
					PrivateKey:  []byte("admin-key"),
				},
			},
			expectGenerateSecretCall: false,
		},
		{
			name: "success - generate secret automatically",
			input: &pb.BootstrapAdminWorkflowInput{
				Req: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "", // Empty - should trigger generation
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
					VaultNamespace: "org1",
				},
			},
			generateSecretRes: &pb.GenerateSecureSecretOutput{
				Secret: "generated-secret-abc123",
			},
			createAdminRes: &pb.CreateAdminIdentityOutput{
				AdminIdentity: &pb.AdminIdentity{
					Id:          "admin",
					Secret:      "generated-secret-abc123",
					Certificate: []byte("admin-cert"),
					PrivateKey:  []byte("admin-key"),
				},
			},
			expectGenerateSecretCall: true,
		},
		{
			name: "failure - secret generation fails",
			input: &pb.BootstrapAdminWorkflowInput{
				Req: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
			},
			generateSecretErr:        errActivityFailed,
			expectErr:                true,
			expectGenerateSecretCall: true,
		},
		{
			name: "failure - admin creation fails",
			input: &pb.BootstrapAdminWorkflowInput{
				Req: &pb.BootstrapAdminInput{
					AdminId:        "admin",
					AdminSecret:    "secret",
					DefaultAdminId: "admin",
					DefaultSecret:  "adminpw",
				},
			},
			createAdminErr:           errActivityFailed,
			expectErr:                true,
			expectGenerateSecretCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register mock activities
			mockActivities := fabricmocks.NewMockFabricIdentityActivities(t)

			// Register GenerateSecureSecret activity
			if tt.expectGenerateSecretCall {
				mockActivities.EXPECT().
					GenerateSecureSecret(mock.Anything, mock.MatchedBy(func(req *pb.GenerateSecureSecretInput) bool {
						if req.GetLength() != 32 {
							t.Errorf("expected length 32, got %d", req.GetLength())
							return false
						}
						return true
					})).
					Return(tt.generateSecretRes, tt.generateSecretErr)
			}

			// Register CreateAdminIdentity activity (only if secret generation doesn't fail)
			if tt.generateSecretErr == nil {
				mockActivities.EXPECT().
					CreateAdminIdentity(mock.Anything, mock.MatchedBy(func(req *pb.CreateAdminIdentityInput) bool {
						if req.GetAdminId() != tt.input.Req.GetAdminId() {
							t.Errorf("admin_id mismatch: got %q, want %q", req.GetAdminId(), tt.input.Req.GetAdminId())
							return false
						}
						return true
					})).
					Return(tt.createAdminRes, tt.createAdminErr)
			}

			pb.RegisterFabricIdentityActivities(env, mockActivities)

			// Execute workflow
			env.ExecuteWorkflow(testBootstrapAdminWorkflow, tt.input)

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

				var result *pb.BootstrapAdminOutput
				if err := env.GetWorkflowResult(&result); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}

				if result.GetAdminIdentity().GetId() != tt.createAdminRes.GetAdminIdentity().GetId() {
					t.Errorf("admin_id mismatch: got %q, want %q",
						result.GetAdminIdentity().GetId(),
						tt.createAdminRes.GetAdminIdentity().GetId())
				}

				if !result.GetSuccess() {
					t.Error("expected success to be true")
				}
			}
		})
	}
}

// Mock implementations for vault workflows used in tests
type mockStoreSecretWorkflow struct {
	result *vaultpb.SecretRef
	err    error
}

func (m *mockStoreSecretWorkflow) Execute(ctx workflow.Context) (*vaultpb.SecretRef, error) {
	return m.result, m.err
}

// Test wrapper for provisionPeerWorkflow
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
