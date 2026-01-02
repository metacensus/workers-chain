package fabric_test

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
	fabric "github.com/metacensus/infra/temporal/lib/workflows/fabric/v1"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	postgresv1mocks "github.com/metacensus/infra/temporal/mocks/workflows/postgres/v1"
	"github.com/nexus-rpc/sdk-go/nexus"
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

	register(pb.CreateIdentityWorkflowName,
		func(workflow.Context, *pb.CreateIdentityInput) (*pb.CreateIdentityOutput, error) {
			return nil, nil
		})

	register(pb.BootstrapCertificateAuthorityAdminWorkflowName,
		func(workflow.Context, *pb.BootstrapCertificateAuthorityAdminInput) (*pb.BootstrapCertificateAuthorityAdminOutput, error) {
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
				workflows := fabric.NewWorkflows(fabricConfig)
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

					postgrespb.RegisterPostgresActivities(env, mockPostgresActivities)

					// Mock CreatePassword activity (must come after all RegisterActivity calls)
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
				workflows := fabric.NewWorkflows(fabricConfig)
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

func TestBootstrapCertificateAuthorityAdminWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		org         string
		certType    string
		caID        string
		expectedErr error // nil = success
	}{
		{
			name:        "success - bootstrap admin for tls root CA",
			org:         testOrg,
			certType:    "tls",
			caID:        "root",
			expectedErr: nil,
		},
		{
			name:        "success - bootstrap admin for org intermediate CA",
			org:         testOrg,
			certType:    "org",
			caID:        "v1",
			expectedErr: nil,
		},
		{
			name:        "failure - invalid cert type",
			org:         testOrg,
			certType:    "invalid",
			caID:        "root",
			expectedErr: fabric.ErrInvalidCertType,
		},
		// TODO: Add more test cases during implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Create OrgConfig for testing
			orgConfig := config.OrgConfig{OrgID: tt.org}

			// TODO: Mock CreateIdentity child workflow
			// TODO: Mock RevokeIdentity activity

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.BootstrapCertificateAuthorityAdminInput) (*pb.BootstrapCertificateAuthorityAdminOutput, error) {
				fabricConfig := fabricconfig.NewFabricConfig(orgConfig)
				workflows := fabric.NewWorkflows(fabricConfig)
				wf, err := workflows.BootstrapCertificateAuthorityAdmin(ctx, &pb.BootstrapCertificateAuthorityAdminWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.BootstrapCertificateAuthorityAdminInput{
				CertType:               tt.certType,
				CertificateAuthorityId: tt.caID,
			})

			var result *pb.BootstrapCertificateAuthorityAdminOutput
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

func TestCreateIdentityWorkflow(t *testing.T) {
	tests := []struct {
		name         string
		org          string
		certType     string
		identityName string
		expectedErr  error // nil = success
	}{
		{
			name:         "success - create tls admin identity",
			org:          testOrg,
			certType:     "tls",
			identityName: "admin",
			expectedErr:  nil,
		},
		{
			name:         "success - create org peer0 identity",
			org:          testOrg,
			certType:     "org",
			identityName: "peer0",
			expectedErr:  nil,
		},
		{
			name:         "failure - invalid cert type",
			org:          testOrg,
			certType:     "invalid",
			identityName: "admin",
			expectedErr:  fabric.ErrInvalidCertType,
		},
		// TODO: Add more test cases during implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Create OrgConfig for testing
			orgConfig := config.OrgConfig{OrgID: tt.org}

			// TODO: Mock CreateKeyPair activity
			// TODO: Mock RegisterIdentity activity
			// TODO: Mock EnrollIdentity activity
			// TODO: Mock StoreCertificate activity

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateIdentityInput) (*pb.CreateIdentityOutput, error) {
				fabricConfig := fabricconfig.NewFabricConfig(orgConfig)
				workflows := fabric.NewWorkflows(fabricConfig)
				wf, err := workflows.CreateIdentity(ctx, &pb.CreateIdentityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateIdentityInput{
				CertType:     tt.certType,
				IdentityName: tt.identityName,
			})

			var result *pb.CreateIdentityOutput
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
