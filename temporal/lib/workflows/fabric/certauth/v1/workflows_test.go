package certauth

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1/certauthv1nexus"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const testOrg = "testorg"
const testOperatorID = "metacensus"

func init() {
	// Initialize global OperatorConfig for tests
	os.Setenv("OPERATOR_ID", testOperatorID)
	viper.AutomaticEnv()
	if err := workerconfig.InitializeOperator(); err != nil {
		panic("failed to initialize operator config for tests: " + err.Error())
	}
}

// =============================================================================
// Constructor / Validation Tests
// =============================================================================

func TestNewCreateCertificateWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateCertificateWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - all required fields with tls issuer",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				Org:       testOrg,
				IssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				Storage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/storage"),
					CertRef: secrets.NewVaultRef("test/certs/storage"),
				},
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			}},
			wantErr: false,
		},
		{
			name: "success - all required fields with org issuer",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				Org:       testOrg,
				IssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				Storage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/storage"),
					CertRef: secrets.NewVaultRef("test/certs/storage"),
				},
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			}},
			wantErr: false,
		},
		{
			name: "error - missing org",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				IssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				Storage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/storage"),
					CertRef: secrets.NewVaultRef("test/certs/storage"),
				},
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			}},
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name: "error - nil issuer_ref",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				Org:       testOrg,
				IssuerRef: nil,
				Storage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/storage"),
					CertRef: secrets.NewVaultRef("test/certs/storage"),
				},
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			}},
			wantErr:   true,
			errSubstr: "issuer_ref is required",
		},
		{
			name: "error - nil storage",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				Org:           testOrg,
				IssuerRef:     &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				Storage:       nil,
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			}},
			wantErr:   true,
			errSubstr: "storage is required",
		},
		{
			name: "error - nil storage.key_ref",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				Org:       testOrg,
				IssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				Storage: &fabricpb.IdentityRef{
					KeyRef:  nil,
					CertRef: secrets.NewVaultRef("test/certs/storage"),
				},
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			}},
			wantErr:   true,
			errSubstr: "storage.key_ref is required",
		},
		{
			name: "error - nil storage.cert_ref",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				Org:       testOrg,
				IssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				Storage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/storage"),
					CertRef: nil,
				},
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			}},
			wantErr:   true,
			errSubstr: "storage.cert_ref is required",
		},
		{
			name: "error - nil issuer_cert_ref",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				Org:       testOrg,
				IssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				Storage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/storage"),
					CertRef: secrets.NewVaultRef("test/certs/storage"),
				},
				IssuerCertRef: nil,
			}},
			wantErr:   true,
			errSubstr: "issuer_cert_ref is required",
		},
		{
			name: "error - invalid cert type",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				Org:       testOrg,
				IssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "invalid", Id: "int"},
				Storage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/storage"),
					CertRef: secrets.NewVaultRef("test/certs/storage"),
				},
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
			}},
			wantErr:   true,
			errSubstr: "issuer cert_type must be 'tls' or 'org'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateCertificateWorkflow(tt.input)

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

func TestNewCreateBootstrapAdminWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateBootstrapAdminWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - valid ca_ref for tls CA",
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				Org:   testOrg,
				CaRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr: false,
		},
		{
			name: "success - valid ca_ref for org CA",
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				Org:   testOrg,
				CaRef: &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
			}},
			wantErr: false,
		},
		{
			name: "error - missing org",
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				CaRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name: "error - nil ca_ref",
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				Org:   testOrg,
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateBootstrapAdminWorkflow(tt.input)

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

func TestNewCreateAdminWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateAdminWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - valid ca_ref for tls CA",
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				Org:   testOrg,
				CaRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr: false,
		},
		{
			name: "success - valid ca_ref for org CA",
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				Org:   testOrg,
				CaRef: &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
			}},
			wantErr: false,
		},
		{
			name: "error - missing org",
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name: "error - nil ca_ref",
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				Org:   testOrg,
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateAdminWorkflow(tt.input)

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
	validInput := func() *pb.CreateIdentityInput {
		return &pb.CreateIdentityInput{
			Org:          testOrg,
			IdentityName: "test-identity",
			OrgIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
			TlsIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			SigningStorage: &fabricpb.IdentityRef{
				KeyRef:  secrets.NewVaultRef("test/keys/signing"),
				CertRef: secrets.NewVaultRef("test/certs/signing"),
			},
			TlsStorage: &fabricpb.IdentityRef{
				KeyRef:  secrets.NewVaultRef("test/keys/tls"),
				CertRef: secrets.NewVaultRef("test/certs/tls"),
			},
		}
	}

	tests := []struct {
		name      string
		mutate    func(*pb.CreateIdentityInput)
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - all required fields",
			mutate:  func(in *pb.CreateIdentityInput) {},
			wantErr: false,
		},
		{
			name:      "error - missing org",
			mutate:    func(in *pb.CreateIdentityInput) { in.Org = "" },
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name:      "error - missing identity name",
			mutate:    func(in *pb.CreateIdentityInput) { in.IdentityName = "" },
			wantErr:   true,
			errSubstr: "identity_name is required",
		},
		{
			name:      "error - missing org_issuer_ref",
			mutate:    func(in *pb.CreateIdentityInput) { in.OrgIssuerRef = nil },
			wantErr:   true,
			errSubstr: "org_issuer_ref is required",
		},
		{
			name:      "error - missing tls_issuer_ref",
			mutate:    func(in *pb.CreateIdentityInput) { in.TlsIssuerRef = nil },
			wantErr:   true,
			errSubstr: "tls_issuer_ref is required",
		},
		{
			name:      "error - missing signing_storage",
			mutate:    func(in *pb.CreateIdentityInput) { in.SigningStorage = nil },
			wantErr:   true,
			errSubstr: "signing_storage is required",
		},
		{
			name:      "error - missing tls_storage",
			mutate:    func(in *pb.CreateIdentityInput) { in.TlsStorage = nil },
			wantErr:   true,
			errSubstr: "tls_storage is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validInput()
			tt.mutate(input)
			wf, err := newCreateIdentityWorkflow(&pb.CreateIdentityWorkflowInput{Req: input})

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
// Helper Function Tests
// =============================================================================

func TestComputeTLSCommonName(t *testing.T) {
	tests := []struct {
		name             string
		baseName         string
		signingIssuerRef *fabricpb.CertificateAuthorityRef
		tlsIssuerRef     *fabricpb.CertificateAuthorityRef
		expected         string
	}{
		{
			name:             "both nil - self-signed - returns base name",
			baseName:         "ca-tls-root",
			signingIssuerRef: nil,
			tlsIssuerRef:     nil,
			expected:         "ca-tls-root",
		},
		{
			name:             "signing nil - self-signed signing - returns base name",
			baseName:         "ca-tls-root",
			signingIssuerRef: nil,
			tlsIssuerRef:     &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			expected:         "ca-tls-root",
		},
		{
			name:             "tls nil - self-signed tls - returns base name",
			baseName:         "ca-tls-root",
			signingIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuerRef:     nil,
			expected:         "ca-tls-root",
		},
		{
			name:             "same CA issues both - collision - returns base name with -tls suffix",
			baseName:         "ca-tls-int",
			signingIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuerRef:     &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			expected:         "ca-tls-int-tls",
		},
		{
			name:             "different CAs - no collision - returns base name",
			baseName:         "ca-org-int",
			signingIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "root"},
			tlsIssuerRef:     &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			expected:         "ca-org-int",
		},
		{
			name:             "different ID same type - no collision - returns base name",
			baseName:         "ca-tls-v2",
			signingIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuerRef:     &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			expected:         "ca-tls-v2",
		},
		{
			name:             "admin collision case - same CA issues both admin certs",
			baseName:         "admin@ca-tls-root",
			signingIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuerRef:     &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			expected:         "admin@ca-tls-root-tls",
		},
		{
			name:             "admin no collision - different CAs for admin certs",
			baseName:         "admin@ca-org-int",
			signingIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
			tlsIssuerRef:     &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			expected:         "admin@ca-org-int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeTLSCommonName(tt.baseName, tt.signingIssuerRef, tt.tlsIssuerRef)
			if result != tt.expected {
				t.Errorf("ComputeTLSCommonName(%q, %v, %v) = %q, want %q",
					tt.baseName, tt.signingIssuerRef, tt.tlsIssuerRef, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Nexus Service Registration Test
// =============================================================================

// fakeNexusRegistry captures Nexus services registered through RegisterNexusService.
// It embeds worker.Worker (left nil) so it satisfies the worker.Worker interface
// that RegisterNexusService requires; only RegisterNexusService is exercised here.
type fakeNexusRegistry struct {
	worker.Worker
	services []*nexus.Service
}

func (r *fakeNexusRegistry) RegisterNexusService(svc *nexus.Service) {
	r.services = append(r.services, svc)
}

// TestRegisterNexusService is a regression test for the fx.go miswiring where the
// Fabric Nexus service was registered instead of the CertificateAuthority service.
// It asserts the registered service is fabric.certauth.v1.CertificateAuthority and
// that it handles the CreateCertificate operation.
func TestRegisterNexusService(t *testing.T) {
	registry := &fakeNexusRegistry{}
	if err := RegisterNexusService(registry); err != nil {
		t.Fatalf("RegisterNexusService failed: %v", err)
	}
	if len(registry.services) != 1 {
		t.Fatalf("expected 1 registered service, got %d", len(registry.services))
	}
	svc := registry.services[0]
	if svc.Name != certauthv1nexus.CertificateAuthorityServiceName {
		t.Errorf("service name = %q, want %q", svc.Name, certauthv1nexus.CertificateAuthorityServiceName)
	}
	if op := svc.Operation(certauthv1nexus.CertificateAuthorityCreateCertificateOperationName); op == nil {
		t.Errorf("service does not handle operation %q", certauthv1nexus.CertificateAuthorityCreateCertificateOperationName)
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
	env.RegisterWorkflowWithOptions(
		func(workflow.Context, *pb.CreateCertificateInput) (*pb.CreateCertificateOutput, error) {
			return nil, nil
		},
		workflow.RegisterOptions{Name: pb.CreateCertificateWorkflowName},
	)
}

// registerCertAuthNexusMock registers the CertificateAuthority Nexus service and
// mocks the CreateCertificate operation (used for cross-namespace TLS issuance).
func registerCertAuthNexusMock(t *testing.T, env *testsuite.TestWorkflowEnvironment, opErr error) {
	t.Helper()
	nexusSvc, err := certauthv1nexus.NewCertificateAuthorityNexusService(&certauthv1nexus.UnimplementedCertificateAuthorityNexusHandler{})
	if err != nil {
		t.Fatalf("failed to create certauth nexus service: %v", err)
	}
	env.RegisterNexusService(nexusSvc)
	if opErr != nil {
		env.OnNexusOperation(nexusSvc, certauthv1nexus.CertificateAuthorityCreateCertificateOperationName, mock.Anything, mock.Anything).
			Return(nil, opErr)
	} else {
		env.OnNexusOperation(nexusSvc, certauthv1nexus.CertificateAuthorityCreateCertificateOperationName, mock.Anything, mock.Anything).
			Return(&nexus.HandlerStartOperationResultSync[*pb.CreateCertificateOutput]{
				Value: &pb.CreateCertificateOutput{
					Identity: &fabricpb.IdentityRef{
						KeyRef:  secrets.NewVaultRef("test/keys"),
						CertRef: secrets.NewVaultRef("test/certs"),
					},
				},
			}, nil)
	}
}

func TestCreateCertificateWorkflow(t *testing.T) {
	// Note: Validation tests for CreateCertificate are in TestNewCreateCertificateWorkflow above.
	// This test covers execution/orchestration with mocked activities.

	tests := []struct {
		name             string
		org              string
		certType         string
		caID             string
		createKeyPairErr error
		registerErr      error
		enrollErr        error
		expectedErr      error
		validateOutput   bool
	}{
		{
			name:           "success - certificate created",
			org:            testOrg,
			certType:       "tls",
			caID:           "int",
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:             "error - keypair generation fails",
			org:              testOrg,
			certType:         "org",
			caID:             "int",
			createKeyPairErr: errors.New("vault connection failed"),
			expectedErr:      errors.New("failed to generate key pair"),
		},
		{
			name:        "error - identity registration fails",
			org:         testOrg,
			certType:    "tls",
			caID:        "int",
			registerErr: errors.New("identity already exists"),
			expectedErr: errors.New("failed to register identity"),
		},
		{
			name:        "error - enrollment fails",
			org:         testOrg,
			certType:    "org",
			caID:        "int",
			enrollErr:   errors.New("CA rejected enrollment"),
			expectedErr: errors.New("failed to enroll identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()

			// Register activities before mocking
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *secretspb.CreateKeyPairInput) (*secretspb.CreateKeyPairOutput, error) {
					return &secretspb.CreateKeyPairOutput{SecretRef: input.SecretRef}, nil
				},
				activity.RegisterOptions{Name: secretspb.CreateKeyPairActivityName},
			)
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *pb.RegisterIdentityInput) (*pb.RegisterIdentityOutput, error) {
					return &pb.RegisterIdentityOutput{EnrollmentSecret: "test-secret"}, nil
				},
				activity.RegisterOptions{Name: pb.RegisterIdentityActivityName},
			)
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *pb.EnrollIdentityInput) (*pb.EnrollIdentityOutput, error) {
					return &pb.EnrollIdentityOutput{}, nil
				},
				activity.RegisterOptions{Name: pb.EnrollIdentityActivityName},
			)

			// Mock CreateKeyPair activity
			if tt.createKeyPairErr != nil {
				env.OnActivity(secretspb.CreateKeyPairActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.createKeyPairErr)
			} else {
				env.OnActivity(secretspb.CreateKeyPairActivityName, mock.Anything, mock.Anything).
					Return(&secretspb.CreateKeyPairOutput{SecretRef: secrets.NewVaultRef("test/key")}, nil)
			}

			// Mock RegisterIdentity activity
			if tt.registerErr != nil {
				env.OnActivity(pb.RegisterIdentityActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.registerErr)
			} else {
				env.OnActivity(pb.RegisterIdentityActivityName, mock.Anything, mock.Anything).
					Return(&pb.RegisterIdentityOutput{EnrollmentSecret: "test-enrollment-secret"}, nil)
			}

			// Mock EnrollIdentity activity
			if tt.enrollErr != nil {
				env.OnActivity(pb.EnrollIdentityActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.enrollErr)
			} else {
				env.OnActivity(pb.EnrollIdentityActivityName, mock.Anything, mock.Anything).
					Return(&pb.EnrollIdentityOutput{}, nil)
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateCertificateInput) (*pb.CreateCertificateOutput, error) {
				wf, err := workflows.CreateCertificate(ctx, &pb.CreateCertificateWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateCertificateInput{
				Org:       tt.org,
				IssuerRef: &fabricpb.CertificateAuthorityRef{CertType: tt.certType, Id: tt.caID},
				Storage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/storage"),
					CertRef: secrets.NewVaultRef("test/certs/storage"),
				},
				IssuerCertRef: secrets.NewVaultRef("test/certs/issuer"),
				Subject: &pb.CertificateSubject{
					CommonName:   "test-identity",
					Organization: tt.org,
				},
			})

			var result *pb.CreateCertificateOutput
			err := env.GetWorkflowResult(&result)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				if tt.validateOutput {
					if result == nil {
						t.Fatal("result is nil")
					}
					if result.Identity == nil {
						t.Error("identity is nil")
					}
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestCreateBootstrapAdminWorkflow(t *testing.T) {
	// Note: Validation tests for CreateBootstrapAdmin are in TestNewCreateBootstrapAdminWorkflow above.
	// This test covers execution/orchestration with mocked activities.

	tests := []struct {
		name                    string
		org                     string
		certType                string
		caID                    string
		getStatusErr            error
		createSigningKeyPairErr error
		enrollSigningErr        error
		createTLSKeyPairErr     error
		enrollTLSErr            error
		expectedErr             error
		validateOutput          bool
	}{
		{
			name:           "success - bootstrap admin created for tls root CA",
			org:            testOrg,
			certType:       "tls",
			caID:           "root",
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "success - bootstrap admin created for org intermediate CA",
			org:            testOrg,
			certType:       "org",
			caID:           "int",
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:         "error - CA never ready",
			org:          testOrg,
			certType:     "tls",
			caID:         "root",
			getStatusErr: errors.New("pod not ready"),
			expectedErr:  errors.New("CA did not become ready"),
		},
		{
			name:                    "error - signing keypair generation fails",
			org:                     testOrg,
			certType:                "tls",
			caID:                    "root",
			createSigningKeyPairErr: errors.New("vault unavailable"),
			expectedErr:             errors.New("failed to generate signing key pair"),
		},
		{
			name:             "error - signing enrollment fails",
			org:              testOrg,
			certType:         "tls",
			caID:             "root",
			enrollSigningErr: errors.New("CA enrollment rejected"),
			expectedErr:      errors.New("failed to bootstrap enroll signing identity"),
		},
		{
			name:                "error - TLS keypair generation fails",
			org:                 testOrg,
			certType:            "org",
			caID:                "int",
			createTLSKeyPairErr: errors.New("vault quota exceeded"),
			expectedErr:         errors.New("failed to generate TLS key pair"),
		},
		{
			name:         "error - TLS enrollment fails",
			org:          testOrg,
			certType:     "org",
			caID:         "int",
			enrollTLSErr: errors.New("TLS CA rejected enrollment"),
			expectedErr:  errors.New("failed to bootstrap enroll TLS identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()

			// Register activities before mocking
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *kubepb.GetStatusInput) (*kubepb.GetStatusOutput, error) {
					return &kubepb.GetStatusOutput{}, nil
				},
				activity.RegisterOptions{Name: kubepb.GetStatusActivityName},
			)
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *secretspb.CreateKeyPairInput) (*secretspb.CreateKeyPairOutput, error) {
					return &secretspb.CreateKeyPairOutput{SecretRef: input.SecretRef}, nil
				},
				activity.RegisterOptions{Name: secretspb.CreateKeyPairActivityName},
			)
			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *pb.BootstrapEnrollIdentityInput) (*pb.BootstrapEnrollIdentityOutput, error) {
					return &pb.BootstrapEnrollIdentityOutput{}, nil
				},
				activity.RegisterOptions{Name: pb.BootstrapEnrollIdentityActivityName},
			)

			// Mock GetStatus activity
			if tt.getStatusErr != nil {
				env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.getStatusErr)
			} else {
				env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).
					Return(&kubepb.GetStatusOutput{}, nil)
			}

			// Mock CreateKeyPair - distinguish signing vs TLS by input ref path
			// Signing refs contain "/signing", TLS refs contain "/tls"
			env.OnActivity(secretspb.CreateKeyPairActivityName, mock.Anything, mock.Anything).
				Return(func(ctx context.Context, input *secretspb.CreateKeyPairInput) (*secretspb.CreateKeyPairOutput, error) {
					ref := input.GetSecretRef()
					path := ""
					if ref != nil && ref.GetRef() != nil {
						if pathVal := ref.GetRef().GetFields()["path"]; pathVal != nil {
							path = pathVal.GetStringValue()
						}
					}
					isSigningRef := strings.Contains(path, "/signing")
					if isSigningRef {
						if tt.createSigningKeyPairErr != nil {
							return nil, tt.createSigningKeyPairErr
						}
					} else {
						if tt.createTLSKeyPairErr != nil {
							return nil, tt.createTLSKeyPairErr
						}
					}
					return &secretspb.CreateKeyPairOutput{SecretRef: input.SecretRef}, nil
				})

			// Mock BootstrapEnrollIdentity - distinguish signing vs TLS by profile
			// Signing uses "" (empty) profile, TLS uses "tls" profile
			env.OnActivity(pb.BootstrapEnrollIdentityActivityName, mock.Anything, mock.Anything).
				Return(func(ctx context.Context, input *pb.BootstrapEnrollIdentityInput) (*pb.BootstrapEnrollIdentityOutput, error) {
					isTLSProfile := input.GetProfile() == "tls"
					if isTLSProfile {
						if tt.enrollTLSErr != nil {
							return nil, tt.enrollTLSErr
						}
					} else {
						if tt.enrollSigningErr != nil {
							return nil, tt.enrollSigningErr
						}
					}
					return &pb.BootstrapEnrollIdentityOutput{}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateBootstrapAdminInput) (*pb.CreateBootstrapAdminOutput, error) {
				wf, err := workflows.CreateBootstrapAdmin(ctx, &pb.CreateBootstrapAdminWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateBootstrapAdminInput{
				Org: tt.org,
				CaRef: &fabricpb.CertificateAuthorityRef{
					CertType: tt.certType,
					Id:       tt.caID,
				},
			})

			var result *pb.CreateBootstrapAdminOutput
			err := env.GetWorkflowResult(&result)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				if tt.validateOutput {
					if result == nil {
						t.Fatal("result is nil")
					}
					if result.SigningIdentity == nil {
						t.Error("signing_identity is nil")
					}
					if result.TlsIdentity == nil {
						t.Error("tls_identity is nil")
					}
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestCreateAdminWorkflow(t *testing.T) {
	// Note: Validation tests for CreateAdmin are in TestNewCreateAdminWorkflow above.
	// This test covers execution/orchestration with mocked child workflows and Nexus operations.
	//
	// CreateAdmin issues the signing cert via a direct CreateCertificate child workflow.
	// The TLS cert goes via Nexus when the CA is org-owned but the TLS issuer is a shared
	// TLS CA (cross-namespace); otherwise it is another direct child workflow call.

	tests := []struct {
		name           string
		org            string
		caRef          *fabricpb.CertificateAuthorityRef
		tlsIssuer      *fabricpb.CertificateAuthorityRef
		signingCertErr error // CreateCertificate child call (signing)
		tlsCertErr     error // CreateCertificate child or Nexus call (TLS)
		expectedErr    error
		validateOutput bool
	}{
		{
			name:           "success - admin created for tls root CA (direct TLS)",
			org:            testOrg,
			caRef:          &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "success - admin created for org intermediate CA (TLS via Nexus)",
			org:            testOrg,
			caRef:          &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
			tlsIssuer:      &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "error - signing identity creation fails",
			org:            testOrg,
			caRef:          &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			signingCertErr: errors.New("CA enrollment failed"),
			expectedErr:    errors.New("failed to create admin signing identity"),
		},
		{
			name:        "error - TLS identity creation fails (direct)",
			org:         testOrg,
			caRef:       &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			tlsCertErr:  errors.New("TLS CA unavailable"),
			expectedErr: errors.New("failed to create admin TLS identity"),
		},
		{
			name:        "error - TLS identity creation fails (Nexus)",
			org:         testOrg,
			caRef:       &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
			tlsIssuer:   &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			tlsCertErr:  errors.New("TLS CA unavailable"),
			expectedErr: errors.New("failed to create admin TLS identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()

			registerChildWorkflowStubs(env)

			// Cross-namespace TLS issuance goes via the CertificateAuthority Nexus service
			registerCertAuthNexusMock(t, env, tt.tlsCertErr)

			// Track call count to differentiate between the signing and TLS child calls
			createCertCallCount := 0
			env.OnWorkflow(pb.CreateCertificateWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.CreateCertificateInput) (*pb.CreateCertificateOutput, error) {
					createCertCallCount++
					if createCertCallCount == 1 {
						if tt.signingCertErr != nil {
							return nil, tt.signingCertErr
						}
						return &pb.CreateCertificateOutput{Identity: input.Storage}, nil
					}
					if tt.tlsCertErr != nil {
						return nil, tt.tlsCertErr
					}
					return &pb.CreateCertificateOutput{Identity: input.Storage}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateAdminInput) (*pb.CreateAdminOutput, error) {
				wf, err := workflows.CreateAdmin(ctx, &pb.CreateAdminWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateAdminInput{
				Org:       tt.org,
				CaRef:     tt.caRef,
				TlsIssuer: tt.tlsIssuer,
			})

			var result *pb.CreateAdminOutput
			err := env.GetWorkflowResult(&result)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				if tt.validateOutput {
					if result == nil {
						t.Fatal("result is nil")
					}
					if result.SigningIdentity == nil {
						t.Error("signing_identity is nil")
					}
					if result.TlsIdentity == nil {
						t.Error("tls_identity is nil")
					}
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestCreateIdentityWorkflow(t *testing.T) {
	// Note: Validation tests for CreateIdentity are in TestNewCreateIdentityWorkflow above.
	// This test covers execution/orchestration with mocked child workflows and Nexus operations.
	//
	// CreateIdentity issues the signing cert via a direct CreateCertificate child workflow
	// and the TLS cert via the CertificateAuthority Nexus service (shared TLS CAs live in
	// the operator namespace).

	tests := []struct {
		name           string
		org            string
		signingCertErr error // CreateCertificate child call (signing)
		tlsCertErr     error // Nexus CreateCertificate call (TLS)
		expectedErr    error
		validateOutput bool
	}{
		{
			name:           "success - identity created with signing and TLS certs",
			org:            testOrg,
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "error - signing certificate creation fails",
			org:            testOrg,
			signingCertErr: errors.New("CA connection refused"),
			expectedErr:    errors.New("failed to create signing identity"),
		},
		{
			name:        "error - TLS certificate creation fails",
			org:         testOrg,
			tlsCertErr:  errors.New("TLS enrollment failed"),
			expectedErr: errors.New("failed to create TLS identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()

			registerChildWorkflowStubs(env)
			registerCertAuthNexusMock(t, env, tt.tlsCertErr)

			// Mock the signing CreateCertificate child workflow
			env.OnWorkflow(pb.CreateCertificateWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.CreateCertificateInput) (*pb.CreateCertificateOutput, error) {
					if tt.signingCertErr != nil {
						return nil, tt.signingCertErr
					}
					return &pb.CreateCertificateOutput{Identity: input.Storage}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateIdentityInput) (*pb.CreateIdentityOutput, error) {
				wf, err := workflows.CreateIdentity(ctx, &pb.CreateIdentityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateIdentityInput{
				Org:          tt.org,
				IdentityName: "test-peer",
				OrgIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"},
				TlsIssuerRef: &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				SigningStorage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/signing"),
					CertRef: secrets.NewVaultRef("test/certs/signing"),
				},
				TlsStorage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/tls"),
					CertRef: secrets.NewVaultRef("test/certs/tls"),
				},
				Subject: &pb.CertificateSubject{
					CommonName:   "test-peer",
					Organization: tt.org,
				},
				IdentityType: "peer",
			})

			var result *pb.CreateIdentityOutput
			err := env.GetWorkflowResult(&result)

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				if tt.validateOutput {
					if result == nil {
						t.Fatal("result is nil")
					}
					if result.SigningIdentity == nil {
						t.Error("signing_identity is nil")
					}
					if result.TlsIdentity == nil {
						t.Error("tls_identity is nil")
					}
				}
			} else {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}
