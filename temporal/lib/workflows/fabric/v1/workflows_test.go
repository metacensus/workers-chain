package fabric

import (
	"context"
	"errors"
	"strings"
	"testing"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	postgrespb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/secrets/v1/secretsv1nexus"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	postgresv1mocks "github.com/metacensus/infra/temporal/mocks/workflows/postgres/v1"
	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const testOrg = "testorg"

// =============================================================================
// Constructor / Validation Tests
// =============================================================================
// These tests verify input validation without needing a workflow environment.
// They're fast and focused on the validation logic.

func TestNewDeployCertificateAuthorityWorkflow(t *testing.T) {
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.DeployCertificateAuthorityWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - tls root CA",
			input: &pb.DeployCertificateAuthorityWorkflowInput{Req: &pb.DeployCertificateAuthorityInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr: false,
		},
		{
			name: "success - org intermediate CA",
			input: &pb.DeployCertificateAuthorityWorkflowInput{Req: &pb.DeployCertificateAuthorityInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
			}},
			wantErr: false,
		},
		{
			name: "error - nil ca_ref",
			input: &pb.DeployCertificateAuthorityWorkflowInput{Req: &pb.DeployCertificateAuthorityInput{
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
		{
			name: "error - invalid cert type",
			input: &pb.DeployCertificateAuthorityWorkflowInput{Req: &pb.DeployCertificateAuthorityInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "invalid", Id: "root"},
			}},
			wantErr:   true,
			errSubstr: "cert_type must be 'tls' or 'org'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployCertificateAuthorityWorkflow(tt.input, fabricCfg)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
				if wf != nil {
					t.Error("expected nil workflow on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if wf == nil {
					t.Fatal("expected non-nil workflow")
				}
			}
		})
	}
}

func TestNewCreateCertificateDatabaseWorkflow(t *testing.T) {
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.CreateCertificateDatabaseWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - tls root CA",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr: false,
		},
		{
			name: "success - org intermediate CA",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
			}},
			wantErr: false,
		},
		{
			name: "error - nil ca_ref",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
		{
			name: "error - invalid cert type",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "invalid", Id: "root"},
			}},
			wantErr:   true,
			errSubstr: "cert_type must be 'tls' or 'org'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateCertificateDatabaseWorkflow(tt.input, fabricCfg)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
				if wf != nil {
					t.Error("expected nil workflow on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if wf == nil {
					t.Fatal("expected non-nil workflow")
				}
			}
		})
	}
}

func TestNewCreateIdentityWorkflow(t *testing.T) {
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.CreateIdentityWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - all required fields",
			input: &pb.CreateIdentityWorkflowInput{Req: &pb.CreateIdentityInput{
				IdentityName:      "test-identity",
				OrgIssuerRef:      &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
				TlsIssuerRef:      &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				StorageSigningRef: secrets.NewVaultRef("test/signing"),
				StorageTlsRef:     secrets.NewVaultRef("test/tls"),
			}},
			wantErr: false,
		},
		{
			name: "error - missing identity name",
			input: &pb.CreateIdentityWorkflowInput{Req: &pb.CreateIdentityInput{
				IdentityName:      "",
				OrgIssuerRef:      &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
				TlsIssuerRef:      &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				StorageSigningRef: secrets.NewVaultRef("test/signing"),
				StorageTlsRef:     secrets.NewVaultRef("test/tls"),
			}},
			wantErr:   true,
			errSubstr: "identity_name is required",
		},
		{
			name: "error - missing org_issuer_ref",
			input: &pb.CreateIdentityWorkflowInput{Req: &pb.CreateIdentityInput{
				IdentityName:      "test-identity",
				OrgIssuerRef:      nil,
				TlsIssuerRef:      &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				StorageSigningRef: secrets.NewVaultRef("test/signing"),
				StorageTlsRef:     secrets.NewVaultRef("test/tls"),
			}},
			wantErr:   true,
			errSubstr: "org_issuer_ref is required",
		},
		{
			name: "error - missing tls_issuer_ref",
			input: &pb.CreateIdentityWorkflowInput{Req: &pb.CreateIdentityInput{
				IdentityName:      "test-identity",
				OrgIssuerRef:      &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
				TlsIssuerRef:      nil,
				StorageSigningRef: secrets.NewVaultRef("test/signing"),
				StorageTlsRef:     secrets.NewVaultRef("test/tls"),
			}},
			wantErr:   true,
			errSubstr: "tls_issuer_ref is required",
		},
		{
			name: "error - missing storage_signing_ref",
			input: &pb.CreateIdentityWorkflowInput{Req: &pb.CreateIdentityInput{
				IdentityName:      "test-identity",
				OrgIssuerRef:      &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
				TlsIssuerRef:      &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				StorageSigningRef: nil,
				StorageTlsRef:     secrets.NewVaultRef("test/tls"),
			}},
			wantErr:   true,
			errSubstr: "storage_signing_ref is required",
		},
		{
			name: "error - missing storage_tls_ref",
			input: &pb.CreateIdentityWorkflowInput{Req: &pb.CreateIdentityInput{
				IdentityName:      "test-identity",
				OrgIssuerRef:      &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
				TlsIssuerRef:      &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				StorageSigningRef: secrets.NewVaultRef("test/signing"),
				StorageTlsRef:     nil,
			}},
			wantErr:   true,
			errSubstr: "storage_tls_ref is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateIdentityWorkflow(tt.input, fabricCfg)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tt.errSubstr, err)
				}
				if wf != nil {
					t.Error("expected nil workflow on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if wf == nil {
					t.Fatal("expected non-nil workflow")
				}
			}
		})
	}
}

// =============================================================================
// Workflow Execution Tests
// =============================================================================
// These tests verify workflow orchestration with mocked activities.
// All inputs are valid - validation is tested separately above.

// registerChildWorkflowStubs registers child workflow stubs for mocking.
// Stubs return zero values - they're never executed when properly mocked.
func registerChildWorkflowStubs(env *testsuite.TestWorkflowEnvironment) {
	register := func(name string, fn any) {
		env.RegisterWorkflowWithOptions(fn, workflow.RegisterOptions{Name: name})
	}

	register(clusterpb.CreateDeploymentWorkflowName,
		func(workflow.Context, *clusterpb.CreateDeploymentInput) (*clusterpb.CreateDeploymentOutput, error) {
			return nil, nil
		})

	register(pb.DeployCertificateStoreWorkflowName,
		func(workflow.Context, *pb.DeployCertificateStoreInput) (*pb.DeployCertificateStoreOutput, error) {
			return nil, nil
		})

	register(pb.BootstrapCertificateAuthoritiesWorkflowName,
		func(workflow.Context, *pb.BootstrapCertificateAuthoritiesInput) (*pb.BootstrapCertificateAuthoritiesOutput, error) {
			return nil, nil
		})

	register(pb.CreateCertificateDatabaseWorkflowName,
		func(workflow.Context, *pb.CreateCertificateDatabaseInput) (*pb.CreateCertificateDatabaseOutput, error) {
			return nil, nil
		})

	register(pb.CreateIdentityWorkflowName,
		func(workflow.Context, *pb.CreateIdentityInput) (*pb.CreateIdentityOutput, error) {
			return nil, nil
		})

	register(pb.CreateBootstrapAdminWorkflowName,
		func(workflow.Context, *pb.CreateBootstrapAdminInput) (*pb.CreateBootstrapAdminOutput, error) {
			return nil, nil
		})

	register(pb.CreateAdminWorkflowName,
		func(workflow.Context, *pb.CreateAdminInput) (*pb.CreateAdminOutput, error) {
			return nil, nil
		})

	register(pb.CreateCertificateWorkflowName,
		func(workflow.Context, *pb.CreateCertificateInput) (*pb.CreateCertificateOutput, error) {
			return nil, nil
		})
}

// mockChildWorkflow sets up a mock for a child workflow, returning success or the given error.
func mockChildWorkflow[O any](env *testsuite.TestWorkflowEnvironment, name string, err error) {
	if err != nil {
		env.OnWorkflow(name, mock.Anything, mock.Anything).Return(nil, err)
	} else {
		env.OnWorkflow(name, mock.Anything, mock.Anything).Return(new(O), nil)
	}
}

func TestDeployCertificateStoreWorkflow(t *testing.T) {
	tests := []struct {
		name                    string
		org                     string
		createDeploymentErr     error
		expectedErr             error // nil = success
		validateDeploymentInput bool  // if true, verify the deployment input
	}{
		{
			name:                "failure - create deployment fails",
			org:                 testOrg,
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to create certificate store deployment"),
		},
		{
			name:                    "success - certificate store deployed",
			org:                     testOrg,
			createDeploymentErr:     nil,
			expectedErr:             nil,
			validateDeploymentInput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Create OrgConfig for testing
			orgConfig := config.OrgConfig{OrgID: tt.org}

			// Register activities before mocking
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *secretspb.CreatePasswordInput) (*secretspb.CreatePasswordOutput, error) {
					return &secretspb.CreatePasswordOutput{SecretRef: input.SecretRef}, nil
				},
				activity.RegisterOptions{Name: secretspb.CreatePasswordActivityName},
			)
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *kubepb.GetStatusInput) (*kubepb.GetStatusOutput, error) {
					return &kubepb.GetStatusOutput{}, nil
				},
				activity.RegisterOptions{Name: kubepb.GetStatusActivityName},
			)

			// Mock Secrets CreateAccess Nexus operation
			secretsNexusSvc, nexusErr := secretsv1nexus.NewSecretsNexusService(&secretsv1nexus.UnimplementedSecretsNexusHandler{})
			if nexusErr != nil {
				t.Fatalf("failed to create secrets nexus service: %v", nexusErr)
			}
			env.RegisterNexusService(secretsNexusSvc)
			env.OnNexusOperation(secretsNexusSvc, secretsv1nexus.SecretsCreateAccessOperationName, mock.Anything, mock.Anything).
				Return(&nexus.HandlerStartOperationResultSync[*secretspb.CreateAccessOutput]{Value: &secretspb.CreateAccessOutput{}}, nil)

			// Mock activities
			env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreatePasswordOutput{SecretRef: secrets.NewVaultRef("test/path")}, nil)
			env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).Return(&kubepb.GetStatusOutput{}, nil)

			// Mock CreateDeployment child workflow
			if tt.validateDeploymentInput {
				// Validate that correct component is passed
				env.OnWorkflow(clusterpb.CreateDeploymentWorkflowName, mock.Anything,
					mock.MatchedBy(func(input *clusterpb.CreateDeploymentInput) bool {
						return input.Org == tt.org && input.Component == "certificate-store"
					})).Return(&clusterpb.CreateDeploymentOutput{
					DeploymentName: "certificate-store",
				}, tt.createDeploymentErr)
			} else {
				mockChildWorkflow[clusterpb.CreateDeploymentOutput](env, clusterpb.CreateDeploymentWorkflowName, tt.createDeploymentErr)
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployCertificateStoreInput) (*pb.DeployCertificateStoreOutput, error) {
				fabricConfig := fabricconfig.NewFabricConfig(orgConfig)
				workflows := NewWorkflows(fabricConfig)
				wf, err := workflows.DeployCertificateStore(ctx, &pb.DeployCertificateStoreWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployCertificateStoreInput{})

			var result *pb.DeployCertificateStoreOutput
			err := env.GetWorkflowResult(&result)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				// Temporal serializes workflow errors, breaking Go's error chain, so we check the message.
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestCreateCertificateDatabaseWorkflow(t *testing.T) {
	tests := []struct {
		name                   string
		org                    string
		certType               string
		caID                   string
		createDatabaseErr      error
		generatePasswordErr    error
		createUserErr          error
		grantPrivilegesErr     error
		expectedErr            error // nil = success
		validateCredentialsRef bool
	}{
		{
			name:                   "success - certificate database created for tls root CA",
			org:                    testOrg,
			certType:               "tls",
			caID:                   "root",
			expectedErr:            nil,
			validateCredentialsRef: true,
		},
		{
			name:                   "success - certificate database created for org intermediate CA",
			org:                    testOrg,
			certType:               "org",
			caID:                   "v1",
			expectedErr:            nil,
			validateCredentialsRef: true,
		},
		// Note: "invalid cert type" validation is tested in TestNewCreateCertificateDatabaseWorkflow above
		{
			name:              "failure - create database activity fails",
			org:               testOrg,
			certType:          "tls",
			caID:              "root",
			createDatabaseErr: errors.New("connection to postgres failed"),
			expectedErr:       errors.New("failed to create database"),
		},
		{
			name:                "failure - generate password activity fails",
			org:                 testOrg,
			certType:            "org",
			caID:                "v2",
			generatePasswordErr: errors.New("crypto error"),
			expectedErr:         errors.New("failed to generate credentials"),
		},
		{
			name:          "failure - create user activity fails",
			org:           testOrg,
			certType:      "tls",
			caID:          "root",
			createUserErr: errors.New("user creation failed"),
			expectedErr:   errors.New("failed to create user"),
		},
		{
			name:               "failure - grant privileges activity fails",
			org:                testOrg,
			certType:           "org",
			caID:               "v1",
			grantPrivilegesErr: errors.New("grant operation failed"),
			expectedErr:        errors.New("failed to grant privileges"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Create OrgConfig for testing
			orgConfig := config.OrgConfig{OrgID: tt.org}

			// Register activities before mocking
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *secretspb.CreatePasswordInput) (*secretspb.CreatePasswordOutput, error) {
					return &secretspb.CreatePasswordOutput{SecretRef: input.SecretRef}, nil
				},
				activity.RegisterOptions{Name: secretspb.CreatePasswordActivityName},
			)
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *kubepb.GetStatusInput) (*kubepb.GetStatusOutput, error) {
					return &kubepb.GetStatusOutput{}, nil
				},
				activity.RegisterOptions{Name: kubepb.GetStatusActivityName},
			)

			// Register and mock activities only if we pass validation
			if tt.certType == "tls" || tt.certType == "org" {
				if tt.caID != "" {
					// Mock PostgreSQL activities - must register BEFORE OnActivity calls
					mockPostgresActivities := postgresv1mocks.NewMockPostgresActivities(t)

					// Set up mock expectations
					if tt.createDatabaseErr != nil {
						mockPostgresActivities.EXPECT().
							CreateDatabase(mock.Anything, mock.Anything).
							Return(nil, tt.createDatabaseErr).
							Maybe()
					} else if tt.expectedErr == nil ||
						!strings.Contains(tt.expectedErr.Error(), "failed to create database") {
						mockPostgresActivities.EXPECT().
							CreateDatabase(mock.Anything, mock.Anything).
							Return(&postgrespb.CreateDatabaseOutput{}, nil).
							Maybe()
					}

					// Mock CreateUser (now uses CredentialsRef internally)
					if tt.createUserErr != nil {
						mockPostgresActivities.EXPECT().
							CreateUser(mock.Anything, mock.Anything).
							Return(nil, tt.createUserErr).
							Maybe()
					} else if tt.generatePasswordErr == nil && tt.createDatabaseErr == nil {
						mockPostgresActivities.EXPECT().
							CreateUser(mock.Anything, mock.Anything).
							Return(&postgrespb.CreateUserOutput{}, nil).
							Maybe()
					}

					// Mock GrantPrivileges
					if tt.grantPrivilegesErr != nil {
						mockPostgresActivities.EXPECT().
							GrantPrivileges(mock.Anything, mock.Anything).
							Return(nil, tt.grantPrivilegesErr).
							Maybe()
					} else if tt.generatePasswordErr == nil && tt.createDatabaseErr == nil && tt.createUserErr == nil {
						mockPostgresActivities.EXPECT().
							GrantPrivileges(mock.Anything, mock.Anything).
							Return(&postgrespb.GrantPrivilegesOutput{}, nil).
							Maybe()
					}

					// Register activities - must be before OnActivity calls
					postgrespb.RegisterPostgresActivities(env, mockPostgresActivities)
				}
			}

			// Mock activities using OnActivity (must come after all RegisterActivity calls)
			env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).
				Return(&kubepb.GetStatusOutput{}, nil)

			if tt.certType == "tls" || tt.certType == "org" {
				if tt.caID != "" {
					if tt.generatePasswordErr != nil {
						env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
							Return(nil, tt.generatePasswordErr).Maybe()
					} else {
						env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
							Return(&secretspb.CreatePasswordOutput{
								SecretRef: secrets.NewVaultRef("fabric/" + tt.org + "/ca/" + tt.certType + "/" + tt.caID + "/db/credentials"),
							}, nil).Maybe()
					}
				}
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateCertificateDatabaseInput) (*pb.CreateCertificateDatabaseOutput, error) {
				fabricConfig := fabricconfig.NewFabricConfig(orgConfig)
				workflows := NewWorkflows(fabricConfig)
				wf, err := workflows.CreateCertificateDatabase(ctx, &pb.CreateCertificateDatabaseWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
		}, &pb.CreateCertificateDatabaseInput{
			CaRef: &pb.CertificateAuthorityRef{
				CertType: tt.certType,
				Id:       tt.caID,
			},
		})

			var result *pb.CreateCertificateDatabaseOutput
			err := env.GetWorkflowResult(&result)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				if result == nil {
					t.Fatal("result is nil")
				}

				// Validate credentials ref if requested
				if tt.validateCredentialsRef {
					if result.CredentialsRef == nil {
						t.Fatal("credentials_ref is nil")
					}
					if result.CredentialsRef.Type == "" {
						t.Fatal("credentials_ref.type is empty")
					}
					if result.CredentialsRef.Ref == nil {
						t.Fatal("credentials_ref.ref is nil")
					}
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				// Temporal serializes workflow errors, breaking Go's error chain, so we check the message.
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestCreateBootstrapAdminWorkflow(t *testing.T) {
	// This workflow requires mocking GetStatus, CreateKeyPair, and BootstrapEnrollIdentity activities
	// which is complex. For now, we skip this test with a note to add proper tests later.
	// The workflow logic is covered by integration tests.
	t.Skip("TODO: Add tests with proper activity mocking for CreateBootstrapAdmin workflow")
}

func TestCreateIdentityWorkflow(t *testing.T) {
	// Note: Validation tests for CreateIdentity are in TestNewCreateIdentityWorkflow above.
	// This test is reserved for execution/orchestration tests.
	// TODO: Add execution test cases - requires mocking CreateCertificate child workflow
	t.Skip("TODO: Add execution tests with proper child workflow mocking for CreateIdentity workflow")
}

func TestDeployCertificateAuthorityWorkflow(t *testing.T) {
	tests := []struct {
		name                                  string
		org                                   string
		certType                              string
		caID                                  string
		createAccessErr              error
		createCertificateDatabaseErr error
		createCertificateErr         error
		createPasswordErr            error
		createDeploymentErr          error
		createBootstrapAdminErr      error
		expectedErr                  error // nil = success
	}{
		{
			name:        "success - intermediate CA deployed with certificate from root",
			org:         testOrg,
			certType:    "tls",
			caID:        "v1",
			expectedErr: nil,
		},
		{
			name:        "success - intermediate org CA deployed",
			org:         testOrg,
			certType:    "org",
			caID:        "v1",
			expectedErr: nil,
		},
		// Note: "invalid cert type" validation is tested in TestNewDeployCertificateAuthorityWorkflow above
		{
			name:            "failure - create access fails",
			org:             testOrg,
			certType:        "tls",
			caID:            "v1",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:                         "failure - create certificate database fails",
			org:                          testOrg,
			certType:                     "tls",
			caID:                         "v1",
			createCertificateDatabaseErr: errors.New("postgres connection failed"),
			expectedErr:                  ErrCreateCertificateDatabase,
		},
		{
			name:                 "failure - create certificate fails",
			org:                  testOrg,
			certType:             "tls",
			caID:                 "v1",
			createCertificateErr: errors.New("CA connection failed"),
			expectedErr:          ErrCreateCertificate,
		},
		{
			name:              "failure - create password fails",
			org:               testOrg,
			certType:          "tls",
			caID:              "v1",
			createPasswordErr: errors.New("random generation failed"),
			expectedErr:       errors.New("failed to generate bootstrap admin credentials"),
		},
		{
			name:                "failure - create deployment fails",
			org:                 testOrg,
			certType:            "tls",
			caID:                "v1",
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to deploy CA"),
		},
		{
			name:                    "failure - create bootstrap admin fails",
			org:                     testOrg,
			certType:                "tls",
			caID:                    "v1",
			createBootstrapAdminErr: errors.New("admin enrollment failed"),
			expectedErr:             ErrCreateBootstrapAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			registerChildWorkflowStubs(env)

			// Register activities before mocking
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *secretspb.CreatePasswordInput) (*secretspb.CreatePasswordOutput, error) {
					return &secretspb.CreatePasswordOutput{SecretRef: input.SecretRef}, nil
				},
				activity.RegisterOptions{Name: secretspb.CreatePasswordActivityName},
			)

			// Mock Nexus call to secrets.v1.CreateAccess
			secretsNexusSvc, nexusErr := secretsv1nexus.NewSecretsNexusService(&secretsv1nexus.UnimplementedSecretsNexusHandler{})
			if nexusErr != nil {
				t.Fatalf("failed to create secrets nexus service: %v", nexusErr)
			}
			env.RegisterNexusService(secretsNexusSvc)
			if tt.createAccessErr != nil {
				env.OnNexusOperation(secretsNexusSvc, secretsv1nexus.SecretsCreateAccessOperationName, mock.Anything, mock.Anything).
					Return(nil, tt.createAccessErr)
			} else {
				env.OnNexusOperation(secretsNexusSvc, secretsv1nexus.SecretsCreateAccessOperationName, mock.Anything, mock.Anything).
					Return(&nexus.HandlerStartOperationResultSync[*secretspb.CreateAccessOutput]{Value: &secretspb.CreateAccessOutput{}}, nil)
			}

			// Mock CreateCertificateDatabase child workflow
			mockChildWorkflow[pb.CreateCertificateDatabaseOutput](
				env,
				pb.CreateCertificateDatabaseWorkflowName,
				tt.createCertificateDatabaseErr,
			)

		// Mock CreateCertificate child workflow (for signing and TLS certificates)
		mockChildWorkflow[pb.CreateCertificateOutput](
			env,
			pb.CreateCertificateWorkflowName,
			tt.createCertificateErr,
		)

			// Mock CreatePassword activity
			env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreatePasswordOutput{}, tt.createPasswordErr)

			// Mock CreateDeployment child workflow
			mockChildWorkflow[clusterpb.CreateDeploymentOutput](
				env,
				clusterpb.CreateDeploymentWorkflowName,
				tt.createDeploymentErr,
			)

		// Mock CreateBootstrapAdmin child workflow
		mockChildWorkflow[pb.CreateBootstrapAdminOutput](
			env,
			pb.CreateBootstrapAdminWorkflowName,
			tt.createBootstrapAdminErr,
		)

		// Mock CreateAdmin child workflow (always succeeds in these tests)
		mockChildWorkflow[pb.CreateAdminOutput](
			env,
			pb.CreateAdminWorkflowName,
			nil,
		)

		// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployCertificateAuthorityInput) (*pb.DeployCertificateAuthorityOutput, error) {
				wf, err := workflows.DeployCertificateAuthority(ctx, &pb.DeployCertificateAuthorityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
		}, &pb.DeployCertificateAuthorityInput{
			CaRef: &pb.CertificateAuthorityRef{
				CertType: tt.certType,
				Id:       tt.caID,
			},
		})

			// Verify results
			if tt.expectedErr == nil {
				if !env.IsWorkflowCompleted() {
					t.Fatal("expected workflow to complete")
				}
				if err := env.GetWorkflowError(); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				var output pb.DeployCertificateAuthorityOutput
				if err := env.GetWorkflowResult(&output); err != nil {
					t.Fatalf("failed to get workflow result: %v", err)
				}
			} else {
				if !env.IsWorkflowCompleted() {
					t.Fatal("expected workflow to complete with error")
				}

				err := env.GetWorkflowError()
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				// Temporal serializes workflow errors, breaking Go's error chain, so we check the message.
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}
