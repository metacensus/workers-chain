package fabric_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nexus-rpc/sdk-go/nexus"
	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	cryptopb "github.com/metacensus/infra/temporal/gen/workflows/crypto/v1"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	postgrespb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	vaultpb "github.com/metacensus/infra/temporal/gen/workflows/vault/v1"
	vaultv1nexus "github.com/metacensus/infra/temporal/gen/workflows/vault/v1/vaultv1nexus"
	"github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	fabric "github.com/metacensus/infra/temporal/lib/workflows/fabric/v1"
	postgresv1mocks "github.com/metacensus/infra/temporal/mocks/workflows/postgres/v1"
	vaultv1mocks "github.com/metacensus/infra/temporal/mocks/workflows/vault/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

const testOrg = "testorg"

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

	register(cryptopb.GenerateAndStorePasswordWorkflowName,
		func(workflow.Context, *cryptopb.GenerateAndStorePasswordInput) (*cryptopb.GenerateAndStorePasswordOutput, error) {
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

		// Mock Vault CreateAccess Nexus operation
		vaultNexusSvc, nexusErr := vaultv1nexus.NewVaultNexusService(&vaultv1nexus.UnimplementedVaultNexusHandler{})
		if nexusErr != nil {
			t.Fatalf("failed to create vault nexus service: %v", nexusErr)
		}
		env.RegisterNexusService(vaultNexusSvc)
		env.OnNexusOperation(vaultNexusSvc, vaultv1nexus.VaultCreateAccessOperationName, mock.Anything, mock.Anything).
			Return(&nexus.HandlerStartOperationResultSync[*vaultpb.CreateAccessOutput]{Value: &vaultpb.CreateAccessOutput{}}, nil)

		// Mock GenerateAndStorePassword child workflow
		mockChildWorkflow[cryptopb.GenerateAndStorePasswordOutput](env, cryptopb.GenerateAndStorePasswordWorkflowName, nil)

		// Register and mock GetStatus activity
		env.RegisterActivityWithOptions(
			func(ctx context.Context, input *kubepb.GetStatusInput) (*kubepb.GetStatusOutput, error) {
				return &kubepb.GetStatusOutput{}, nil
			},
			activity.RegisterOptions{Name: kubepb.GetStatusActivityName},
		)
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
				workflows := fabric.NewWorkflows(orgConfig, fabricConfig)
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
		fetchSecretErr         error
		fetchSecretData        map[string][]byte
		createUserErr          error
		grantPrivilegesErr     error
		expectedErr            error // nil = success
		validateCredentialsRef bool
	}{
		{
			name:     "success - certificate database created for tls root CA",
			org:      testOrg,
			certType: "tls",
			caID:     "root",
			fetchSecretData: map[string][]byte{
				"username": []byte("postgres_tls_root"),
				"password": []byte("secure-password-123"),
			},
			expectedErr:            nil,
			validateCredentialsRef: true,
		},
		{
			name:     "success - certificate database created for org intermediate CA",
			org:      testOrg,
			certType: "org",
			caID:     "v1",
			fetchSecretData: map[string][]byte{
				"username": []byte("postgres_org_v1"),
				"password": []byte("another-secure-pass"),
			},
			expectedErr:            nil,
			validateCredentialsRef: true,
		},
		{
			name:        "failure - invalid cert type",
			org:         testOrg,
			certType:    "invalid",
			caID:        "root",
			expectedErr: fabric.ErrInvalidCertType,
		},
		{
			name:        "failure - empty certificate authority ID",
			org:         testOrg,
			certType:    "tls",
			caID:        "",
			expectedErr: errors.New("certificate_authority_id cannot be empty"),
		},
		{
			name:              "failure - create database activity fails",
			org:               testOrg,
			certType:          "tls",
			caID:              "root",
			createDatabaseErr: errors.New("connection to postgres failed"),
			expectedErr:       errors.New("failed to create database"),
		},
		{
			name:                "failure - generate password workflow fails",
			org:                 testOrg,
			certType:            "org",
			caID:                "v2",
			generatePasswordErr: errors.New("crypto error"),
			expectedErr:         errors.New("failed to generate credentials"),
		},
		{
			name:           "failure - fetch secret activity fails",
			org:            testOrg,
			certType:       "tls",
			caID:           "root",
			fetchSecretErr: errors.New("vault connection timeout"),
			expectedErr:    errors.New("failed to fetch credentials from Vault"),
		},
		{
			name:     "failure - username not found in vault secret",
			org:      testOrg,
			certType: "tls",
			caID:     "root",
			fetchSecretData: map[string][]byte{
				"password": []byte("secure-password"),
			},
			expectedErr: errors.New("username not found in Vault secret"),
		},
		{
			name:     "failure - password not found in vault secret",
			org:      testOrg,
			certType: "org",
			caID:     "v1",
			fetchSecretData: map[string][]byte{
				"username": []byte("postgres_org_v1"),
			},
			expectedErr: errors.New("password not found in Vault secret"),
		},
		{
			name:     "failure - create user activity fails",
			org:      testOrg,
			certType: "tls",
			caID:     "root",
			fetchSecretData: map[string][]byte{
				"username": []byte("postgres_tls_root"),
				"password": []byte("secure-password"),
			},
			createUserErr: errors.New("user creation failed"),
			expectedErr:   errors.New("failed to create user"),
		},
		{
			name:     "failure - grant privileges activity fails",
			org:      testOrg,
			certType: "org",
			caID:     "v1",
			fetchSecretData: map[string][]byte{
				"username": []byte("postgres_org_v1"),
				"password": []byte("secure-password"),
			},
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

			// Mock GenerateAndStorePassword child workflow
			if tt.generatePasswordErr != nil {
				env.OnWorkflow(cryptopb.GenerateAndStorePasswordWorkflowName, mock.Anything, mock.Anything).
					Return(nil, tt.generatePasswordErr)
			} else if tt.certType != "" && tt.caID != "" {
				// Only mock if we're past validation
				shouldMock := true
				if tt.expectedErr != nil {
					if errors.Is(tt.expectedErr, fabric.ErrInvalidCertType) {
						shouldMock = false
					} else if strings.Contains(tt.expectedErr.Error(), "certificate_authority_id cannot be empty") {
						shouldMock = false
					}
				}

				if shouldMock {
					env.OnWorkflow(cryptopb.GenerateAndStorePasswordWorkflowName, mock.Anything, mock.Anything).
						Return(&cryptopb.GenerateAndStorePasswordOutput{
							SecretRef: &vaultpb.SecretRef{
								Path:    "fabric/" + tt.org + "/ca/" + tt.certType + "/" + tt.caID + "/db/credentials",
								Version: 1,
							},
						}, nil)
				}
			}

			// Mock activities only if we pass validation
			if tt.certType == "tls" || tt.certType == "org" {
				if tt.caID != "" {
					// Mock PostgreSQL activities
					mockPostgresActivities := postgresv1mocks.NewMockPostgresActivities(t)

					// Mock CreateDatabase
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

					// Mock CreateUser
					// Call CreateUser if we're past CreateDatabase and FetchSecret
					if tt.createUserErr != nil {
						mockPostgresActivities.EXPECT().
							CreateUser(mock.Anything, mock.Anything).
							Return(nil, tt.createUserErr).
							Maybe()
					} else if tt.fetchSecretData != nil &&
						tt.fetchSecretData["username"] != nil &&
						tt.fetchSecretData["password"] != nil &&
						tt.fetchSecretErr == nil {
						// CreateUser should succeed if we're testing GrantPrivileges or success cases
						mockPostgresActivities.EXPECT().
							CreateUser(mock.Anything, mock.Anything).
							Return(&postgrespb.CreateUserOutput{}, nil).
							Maybe()
					}

					// Mock GrantPrivileges
					// Call GrantPrivileges if we're past CreateUser
					if tt.grantPrivilegesErr != nil {
						mockPostgresActivities.EXPECT().
							GrantPrivileges(mock.Anything, mock.Anything).
							Return(nil, tt.grantPrivilegesErr).
							Maybe()
					} else if tt.fetchSecretData != nil &&
						tt.fetchSecretData["username"] != nil &&
						tt.fetchSecretData["password"] != nil &&
						tt.fetchSecretErr == nil &&
						tt.createUserErr == nil {
						// GrantPrivileges should succeed for success cases
						mockPostgresActivities.EXPECT().
							GrantPrivileges(mock.Anything, mock.Anything).
							Return(&postgrespb.GrantPrivilegesOutput{}, nil).
							Maybe()
					}

					postgrespb.RegisterPostgresActivities(env, mockPostgresActivities)

					// Mock Vault activities
					mockVaultActivities := vaultv1mocks.NewMockVaultActivities(t)

					if tt.fetchSecretErr != nil {
						mockVaultActivities.EXPECT().
							FetchSecret(mock.Anything, mock.Anything).
							Return(nil, tt.fetchSecretErr).
							Maybe()
					} else if tt.fetchSecretData != nil {
						mockVaultActivities.EXPECT().
							FetchSecret(mock.Anything, mock.Anything).
							Return(&vaultpb.FetchSecretOutput{
								Data: tt.fetchSecretData,
							}, nil).
							Maybe()
					}

					vaultpb.RegisterVaultActivities(env, mockVaultActivities)
				}
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateCertificateDatabaseInput) (*pb.CreateCertificateDatabaseOutput, error) {
				fabricConfig := fabricconfig.NewFabricConfig(orgConfig)
				workflows := fabric.NewWorkflows(orgConfig, fabricConfig)
				wf, err := workflows.CreateCertificateDatabase(ctx, &pb.CreateCertificateDatabaseWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateCertificateDatabaseInput{
				CertType:               tt.certType,
				CertificateAuthorityId: tt.caID,
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
					if result.CredentialsRef.Path == "" {
						t.Fatal("credentials_ref.path is empty")
					}
					if result.CredentialsRef.Version == 0 {
						t.Fatal("credentials_ref.version is 0")
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
