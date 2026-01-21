package fabric

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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
	"google.golang.org/protobuf/types/known/durationpb"
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
				Spec: &pb.CertificateAuthoritySpec{Self: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}},
			}},
			wantErr: false,
		},
		{
			name: "success - org intermediate CA",
			input: &pb.DeployCertificateAuthorityWorkflowInput{Req: &pb.DeployCertificateAuthorityInput{
				Spec: &pb.CertificateAuthoritySpec{Self: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"}},
			}},
			wantErr: false,
		},
		{
			name: "error - nil spec",
			input: &pb.DeployCertificateAuthorityWorkflowInput{Req: &pb.DeployCertificateAuthorityInput{
				Spec: nil,
			}},
			wantErr:   true,
			errSubstr: "spec is required",
		},
		{
			name: "error - nil spec.self",
			input: &pb.DeployCertificateAuthorityWorkflowInput{Req: &pb.DeployCertificateAuthorityInput{
				Spec: &pb.CertificateAuthoritySpec{Self: nil},
			}},
			wantErr:   true,
			errSubstr: "spec.self is required",
		},
		{
			name: "error - invalid cert type",
			input: &pb.DeployCertificateAuthorityWorkflowInput{Req: &pb.DeployCertificateAuthorityInput{
				Spec: &pb.CertificateAuthoritySpec{Self: &pb.CertificateAuthorityRef{CertType: "invalid", Id: "root"}},
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

func TestNewCreateCertificateWorkflow(t *testing.T) {
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.CreateCertificateWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - all required fields with tls issuer",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				IssuerRef:  &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				StorageRef: secrets.NewVaultRef("test/storage"),
			}},
			wantErr: false,
		},
		{
			name: "success - all required fields with org issuer",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				IssuerRef:  &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
				StorageRef: secrets.NewVaultRef("test/storage"),
			}},
			wantErr: false,
		},
		{
			name: "error - nil issuer_ref",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				IssuerRef:  nil,
				StorageRef: secrets.NewVaultRef("test/storage"),
			}},
			wantErr:   true,
			errSubstr: "issuer_ref is required",
		},
		{
			name: "error - nil storage_ref",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				IssuerRef:  &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				StorageRef: nil,
			}},
			wantErr:   true,
			errSubstr: "storage_ref is required",
		},
		{
			name: "error - invalid cert type",
			input: &pb.CreateCertificateWorkflowInput{Req: &pb.CreateCertificateInput{
				IssuerRef:  &pb.CertificateAuthorityRef{CertType: "invalid", Id: "v1"},
				StorageRef: secrets.NewVaultRef("test/storage"),
			}},
			wantErr:   true,
			errSubstr: "issuer cert_type must be 'tls' or 'org'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateCertificateWorkflow(tt.input, fabricCfg)

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
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.CreateBootstrapAdminWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - valid ca_ref for tls CA",
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr: false,
		},
		{
			name: "success - valid ca_ref for org CA",
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
			}},
			wantErr: false,
		},
		{
			name: "error - nil ca_ref",
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateBootstrapAdminWorkflow(tt.input, fabricCfg)

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
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.CreateAdminWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - valid ca_ref for tls CA",
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr: false,
		},
		{
			name: "success - valid ca_ref for org CA",
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
			}},
			wantErr: false,
		},
		{
			name: "error - nil ca_ref",
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateAdminWorkflow(tt.input, fabricCfg)

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
// Helper Function Tests
// =============================================================================

func TestComputeTLSCommonName(t *testing.T) {
	tests := []struct {
		name             string
		baseName         string
		signingIssuerRef *pb.CertificateAuthorityRef
		tlsIssuerRef     *pb.CertificateAuthorityRef
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
			tlsIssuerRef:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			expected:         "ca-tls-root",
		},
		{
			name:             "tls nil - self-signed tls - returns base name",
			baseName:         "ca-tls-root",
			signingIssuerRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuerRef:     nil,
			expected:         "ca-tls-root",
		},
		{
			name:             "same CA issues both - collision - returns base name with -tls suffix",
			baseName:         "ca-tls-v1",
			signingIssuerRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuerRef:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			expected:         "ca-tls-v1-tls",
		},
		{
			name:             "different CAs - no collision - returns base name",
			baseName:         "ca-org-v1",
			signingIssuerRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "root"},
			tlsIssuerRef:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
			expected:         "ca-org-v1",
		},
		{
			name:             "different ID same type - no collision - returns base name",
			baseName:         "ca-tls-v2",
			signingIssuerRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuerRef:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
			expected:         "ca-tls-v2",
		},
		{
			name:             "admin collision case - same CA issues both admin certs",
			baseName:         "admin@ca-tls-root",
			signingIssuerRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuerRef:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			expected:         "admin@ca-tls-root-tls",
		},
		{
			name:             "admin no collision - different CAs for admin certs",
			baseName:         "admin@ca-org-v1",
			signingIssuerRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
			tlsIssuerRef:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
			expected:         "admin@ca-org-v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeTLSCommonName(tt.baseName, tt.signingIssuerRef, tt.tlsIssuerRef)
			if result != tt.expected {
				t.Errorf("computeTLSCommonName(%q, %v, %v) = %q, want %q",
					tt.baseName, tt.signingIssuerRef, tt.tlsIssuerRef, result, tt.expected)
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

	register(pb.DeployCertificateAuthorityWorkflowName,
		func(workflow.Context, *pb.DeployCertificateAuthorityInput) (*pb.DeployCertificateAuthorityOutput, error) {
			return nil, nil
		})

	register(pb.DeployCouchDbWorkflowName,
		func(workflow.Context, *pb.DeployCouchDbInput) (*pb.DeployCouchDbOutput, error) {
			return nil, nil
		})

	register(pb.DeployPeerWorkflowName,
		func(workflow.Context, *pb.DeployPeerInput) (*pb.DeployPeerOutput, error) {
			return nil, nil
		})

	register(pb.DeployOrdererWorkflowName,
		func(workflow.Context, *pb.DeployOrdererInput) (*pb.DeployOrdererOutput, error) {
			return nil, nil
		})

	register(pb.DeployChainCodeWorkflowName,
		func(workflow.Context, *pb.DeployChainCodeInput) (*pb.DeployChainCodeOutput, error) {
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

func TestCreateAdminWorkflow(t *testing.T) {
	// Note: Validation tests for CreateAdmin are in TestNewCreateAdminWorkflow above.
	// This test covers execution/orchestration with mocked child workflows.

	tests := []struct {
		name           string
		org            string
		certType       string
		caID           string
		signingCertErr error // First CreateCertificate call (signing)
		tlsCertErr     error // Second CreateCertificate call (TLS)
		expectedErr    error
		validateOutput bool
	}{
		{
			name:           "success - admin created for tls root CA",
			org:            testOrg,
			certType:       "tls",
			caID:           "root",
			signingCertErr: nil,
			tlsCertErr:     nil,
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "success - admin created for org intermediate CA",
			org:            testOrg,
			certType:       "org",
			caID:           "v1",
			signingCertErr: nil,
			tlsCertErr:     nil,
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "error - signing identity creation fails",
			org:            testOrg,
			certType:       "tls",
			caID:           "v1",
			signingCertErr: errors.New("CA enrollment failed"),
			tlsCertErr:     nil,
			expectedErr:    errors.New("failed to create admin signing identity"),
		},
		{
			name:           "error - TLS identity creation fails",
			org:            testOrg,
			certType:       "org",
			caID:           "v1",
			signingCertErr: nil,
			tlsCertErr:     errors.New("TLS CA unavailable"),
			expectedErr:    errors.New("failed to create admin TLS identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Track call count to differentiate between first and second CreateCertificate calls
			createCertCallCount := 0

			// Mock CreateCertificate child workflow with call-order awareness
			env.OnWorkflow(pb.CreateCertificateWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.CreateCertificateInput) (*pb.CreateCertificateOutput, error) {
					createCertCallCount++
					if createCertCallCount == 1 {
						// First call: signing certificate
						if tt.signingCertErr != nil {
							return nil, tt.signingCertErr
						}
						return &pb.CreateCertificateOutput{
							CredentialsRef: input.StorageRef,
						}, nil
					}
					// Second call: TLS certificate
					if tt.tlsCertErr != nil {
						return nil, tt.tlsCertErr
					}
					return &pb.CreateCertificateOutput{
						CredentialsRef: input.StorageRef,
					}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateAdminInput) (*pb.CreateAdminOutput, error) {
				wf, err := workflows.CreateAdmin(ctx, &pb.CreateAdminWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateAdminInput{
				CaRef: &pb.CertificateAuthorityRef{
					CertType: tt.certType,
					Id:       tt.caID,
				},
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
					if result.SigningCredentialsRef == nil {
						t.Error("signing_credentials_ref is nil")
					}
					if result.TlsCredentialsRef == nil {
						t.Error("tls_credentials_ref is nil")
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
			caID:           "v1",
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
			caID:                "v1",
			createTLSKeyPairErr: errors.New("vault quota exceeded"),
			expectedErr:         errors.New("failed to generate TLS key pair"),
		},
		{
			name:         "error - TLS enrollment fails",
			org:          testOrg,
			certType:     "org",
			caID:         "v1",
			enrollTLSErr: errors.New("TLS CA rejected enrollment"),
			expectedErr:  errors.New("failed to bootstrap enroll TLS identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

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
				CaRef: &pb.CertificateAuthorityRef{
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
					if result.SigningCredentialsRef == nil {
						t.Error("signing_credentials_ref is nil")
					}
					if result.TlsCredentialsRef == nil {
						t.Error("tls_credentials_ref is nil")
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
	// This test covers execution/orchestration with mocked child workflows.

	tests := []struct {
		name           string
		org            string
		signingCertErr error // First CreateCertificate call (signing)
		tlsCertErr     error // Second CreateCertificate call (TLS)
		expectedErr    error
		validateOutput bool
	}{
		{
			name:           "success - identity created with signing and TLS certs",
			org:            testOrg,
			signingCertErr: nil,
			tlsCertErr:     nil,
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "error - signing certificate creation fails",
			org:            testOrg,
			signingCertErr: errors.New("CA connection refused"),
			tlsCertErr:     nil, // Never reached
			expectedErr:    errors.New("failed to create signing identity"),
		},
		{
			name:           "error - TLS certificate creation fails",
			org:            testOrg,
			signingCertErr: nil,
			tlsCertErr:     errors.New("TLS enrollment failed"),
			expectedErr:    errors.New("failed to create TLS identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Track call count to differentiate between first and second CreateCertificate calls
			createCertCallCount := 0

			// Mock CreateCertificate child workflow with call-order awareness
			env.OnWorkflow(pb.CreateCertificateWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.CreateCertificateInput) (*pb.CreateCertificateOutput, error) {
					createCertCallCount++
					if createCertCallCount == 1 {
						// First call: signing certificate
						if tt.signingCertErr != nil {
							return nil, tt.signingCertErr
						}
						return &pb.CreateCertificateOutput{
							CredentialsRef: input.StorageRef,
						}, nil
					}
					// Second call: TLS certificate
					if tt.tlsCertErr != nil {
						return nil, tt.tlsCertErr
					}
					return &pb.CreateCertificateOutput{
						CredentialsRef: input.StorageRef,
					}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateIdentityInput) (*pb.CreateIdentityOutput, error) {
				wf, err := workflows.CreateIdentity(ctx, &pb.CreateIdentityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateIdentityInput{
				IdentityName:      "test-peer",
				OrgIssuerRef:      &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
				TlsIssuerRef:      &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				StorageSigningRef: secrets.NewVaultRef("test/signing"),
				StorageTlsRef:     secrets.NewVaultRef("test/tls"),
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
					if result.SigningCredentialsRef == nil {
						t.Error("signing_credentials_ref is nil")
					}
					if result.TlsCredentialsRef == nil {
						t.Error("tls_credentials_ref is nil")
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

func TestDeployCertificateAuthorityWorkflow(t *testing.T) {
	tests := []struct {
		name                         string
		org                          string
		certType                     string
		caID                         string
		signingIssuer                *pb.CertificateAuthorityRef // nil = self-signed
		tlsIssuer                    *pb.CertificateAuthorityRef // nil = self-signed
		createAccessErr              error
		createCertificateDatabaseErr error
		createCertificateErr         error
		createPasswordErr            error
		createDeploymentErr          error
		createBootstrapAdminErr      error
		expectedErr                  error // nil = success
	}{
		{
			name:          "success - intermediate CA deployed with certificate from root",
			org:           testOrg,
			certType:      "tls",
			caID:          "v1",
			signingIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}, // TLS Int signed by TLS Root
			tlsIssuer:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}, // TLS Int TLS from TLS Root
			expectedErr:   nil,
		},
		{
			name:          "success - intermediate org CA deployed",
			org:           testOrg,
			certType:      "org",
			caID:          "v1",
			signingIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}, // Org Int signed by Org Root
			tlsIssuer:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},   // Org Int TLS from TLS Int
			expectedErr:   nil,
		},
		// Note: "invalid cert type" validation is tested in TestNewDeployCertificateAuthorityWorkflow above
		{
			name:            "failure - create access fails",
			org:             testOrg,
			certType:        "tls",
			caID:            "v1",
			signingIssuer:   &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuer:       &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:                         "failure - create certificate database fails",
			org:                          testOrg,
			certType:                     "tls",
			caID:                         "v1",
			signingIssuer:                &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuer:                    &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			createCertificateDatabaseErr: errors.New("postgres connection failed"),
			expectedErr:                  ErrCreateCertificateDatabase,
		},
		{
			name:                 "failure - create certificate fails",
			org:                  testOrg,
			certType:             "tls",
			caID:                 "v1",
			signingIssuer:        &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuer:            &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			createCertificateErr: errors.New("CA connection failed"),
			expectedErr:          ErrCreateCertificate,
		},
		{
			name:              "failure - create password fails",
			org:               testOrg,
			certType:          "tls",
			caID:              "v1",
			signingIssuer:     &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuer:         &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			createPasswordErr: errors.New("random generation failed"),
			expectedErr:       errors.New("failed to generate bootstrap admin credentials"),
		},
		{
			name:                "failure - create deployment fails",
			org:                 testOrg,
			certType:            "tls",
			caID:                "v1",
			signingIssuer:       &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuer:           &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to deploy CA"),
		},
		{
			name:                    "failure - create bootstrap admin fails",
			org:                     testOrg,
			certType:                "tls",
			caID:                    "v1",
			signingIssuer:           &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			tlsIssuer:               &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
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
				Spec: &pb.CertificateAuthoritySpec{
					Self: &pb.CertificateAuthorityRef{
						CertType: tt.certType,
						Id:       tt.caID,
					},
					SigningIssuer:    tt.signingIssuer,
					TlsIssuer:        tt.tlsIssuer,
					ValidityDuration: durationpb.New(15 * 365 * 24 * time.Hour),
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
			caID:           "v1",
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:             "error - keypair generation fails",
			org:              testOrg,
			certType:         "org",
			caID:             "v1",
			createKeyPairErr: errors.New("vault connection failed"),
			expectedErr:      errors.New("failed to generate key pair"),
		},
		{
			name:        "error - identity registration fails",
			org:         testOrg,
			certType:    "tls",
			caID:        "v1",
			registerErr: errors.New("identity already exists"),
			expectedErr: errors.New("failed to register identity"),
		},
		{
			name:        "error - enrollment fails",
			org:         testOrg,
			certType:    "org",
			caID:        "v1",
			enrollErr:   errors.New("CA rejected enrollment"),
			expectedErr: errors.New("failed to enroll identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

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
				IssuerRef:  &pb.CertificateAuthorityRef{CertType: tt.certType, Id: tt.caID},
				StorageRef: secrets.NewVaultRef("test/storage"),
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
					if result.CredentialsRef == nil {
						t.Error("credentials_ref is nil")
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

func TestDeployCouchDbWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		org                 string
		peerID              string
		createAccessErr     error
		createPasswordErr   error
		createDeploymentErr error
		getStatusErr        error
		expectedErr         error
	}{
		{
			name:        "success - CouchDB deployed and ready",
			org:         testOrg,
			peerID:      "peer0",
			expectedErr: nil,
		},
		{
			name:            "error - create vault access fails",
			org:             testOrg,
			peerID:          "peer0",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrDeployCouchDB,
		},
		{
			name:              "error - create credentials fails",
			org:               testOrg,
			peerID:            "peer0",
			createPasswordErr: errors.New("vault unavailable"),
			expectedErr:       errors.New("failed to generate couchdb credentials"),
		},
		{
			name:                "error - create deployment fails",
			org:                 testOrg,
			peerID:              "peer0",
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to create couchdb deployment"),
		},
		{
			name:         "error - CouchDB never ready",
			org:          testOrg,
			peerID:       "peer0",
			getStatusErr: errors.New("pod not ready"),
			expectedErr:  errors.New("CouchDB did not become ready"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Register activities
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

			// Mock Nexus CreateAccess
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

			// Mock CreatePassword activity
			if tt.createPasswordErr != nil {
				env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.createPasswordErr)
			} else {
				env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
					Return(&secretspb.CreatePasswordOutput{}, nil)
			}

			// Mock CreateDeployment child workflow
			mockChildWorkflow[clusterpb.CreateDeploymentOutput](env, clusterpb.CreateDeploymentWorkflowName, tt.createDeploymentErr)

			// Mock GetStatus activity
			if tt.getStatusErr != nil {
				env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.getStatusErr)
			} else {
				env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).
					Return(&kubepb.GetStatusOutput{}, nil)
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployCouchDbInput) (*pb.DeployCouchDbOutput, error) {
				wf, err := workflows.DeployCouchDb(ctx, &pb.DeployCouchDbWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployCouchDbInput{PeerId: tt.peerID})

			err := env.GetWorkflowError()

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
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

func TestNewDeployPeerWorkflow(t *testing.T) {
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.DeployPeerWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - all required fields",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Spec: &pb.PeerSpec{
					Id:        "peer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				},
			}},
			wantErr: false,
		},
		{
			name: "error - nil spec",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Spec: nil,
			}},
			wantErr:   true,
			errSubstr: "spec is required",
		},
		{
			name: "error - missing org_issuer",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Spec: &pb.PeerSpec{
					Id:        "peer0",
					OrgIssuer: nil,
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				},
			}},
			wantErr:   true,
			errSubstr: "spec.org_issuer is required",
		},
		{
			name: "error - missing tls_issuer",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Spec: &pb.PeerSpec{
					Id:        "peer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
					TlsIssuer: nil,
				},
			}},
			wantErr:   true,
			errSubstr: "spec.tls_issuer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployPeerWorkflow(tt.input, fabricCfg)

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

func TestDeployPeerWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		org                 string
		peerID              string
		createAccessErr     error
		createIdentityErr   error
		createDeploymentErr error
		expectedErr         error
	}{
		{
			name:        "success - peer deployed",
			org:         testOrg,
			peerID:      "peer0",
			expectedErr: nil,
		},
		{
			name:            "error - create vault access fails",
			org:             testOrg,
			peerID:          "peer0",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     errors.New("failed to create vault access"),
		},
		{
			name:              "error - create identity fails",
			org:               testOrg,
			peerID:            "peer0",
			createIdentityErr: errors.New("CA enrollment failed"),
			expectedErr:       errors.New("failed to create peer identity"),
		},
		{
			name:                "error - create deployment fails",
			org:                 testOrg,
			peerID:              "peer0",
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to create peer deployment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Mock Nexus CreateAccess
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

			// Mock CreateIdentity child workflow
			mockChildWorkflow[pb.CreateIdentityOutput](env, pb.CreateIdentityWorkflowName, tt.createIdentityErr)

			// Mock CreateDeployment child workflow
			mockChildWorkflow[clusterpb.CreateDeploymentOutput](env, clusterpb.CreateDeploymentWorkflowName, tt.createDeploymentErr)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployPeerInput) (*pb.DeployPeerOutput, error) {
				wf, err := workflows.DeployPeer(ctx, &pb.DeployPeerWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployPeerInput{
				Spec: &pb.PeerSpec{
					Id:        tt.peerID,
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				},
			})

			err := env.GetWorkflowError()

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
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

func TestNewDeployOrdererWorkflow(t *testing.T) {
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.DeployOrdererWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - all required fields",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Spec: &pb.OrdererSpec{
					Id:        "orderer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				},
			}},
			wantErr: false,
		},
		{
			name: "error - nil spec",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Spec: nil,
			}},
			wantErr:   true,
			errSubstr: "spec is required",
		},
		{
			name: "error - missing org_issuer",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Spec: &pb.OrdererSpec{
					Id:        "orderer0",
					OrgIssuer: nil,
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				},
			}},
			wantErr:   true,
			errSubstr: "spec.org_issuer is required",
		},
		{
			name: "error - missing tls_issuer",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Spec: &pb.OrdererSpec{
					Id:        "orderer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
					TlsIssuer: nil,
				},
			}},
			wantErr:   true,
			errSubstr: "spec.tls_issuer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployOrdererWorkflow(tt.input, fabricCfg)

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

func TestDeployOrdererWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		org                 string
		ordererID           string
		createAccessErr     error
		createIdentityErr   error
		createDeploymentErr error
		expectedErr         error
	}{
		{
			name:        "success - orderer deployed",
			org:         testOrg,
			ordererID:   "orderer0",
			expectedErr: nil,
		},
		{
			name:            "error - create vault access fails",
			org:             testOrg,
			ordererID:       "orderer0",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     errors.New("failed to create vault access"),
		},
		{
			name:              "error - create identity fails",
			org:               testOrg,
			ordererID:         "orderer0",
			createIdentityErr: errors.New("CA enrollment failed"),
			expectedErr:       errors.New("failed to create orderer identity"),
		},
		{
			name:                "error - create deployment fails",
			org:                 testOrg,
			ordererID:           "orderer0",
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to create orderer deployment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Mock Nexus CreateAccess
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

			// Mock CreateIdentity child workflow
			mockChildWorkflow[pb.CreateIdentityOutput](env, pb.CreateIdentityWorkflowName, tt.createIdentityErr)

			// Mock CreateDeployment child workflow
			mockChildWorkflow[clusterpb.CreateDeploymentOutput](env, clusterpb.CreateDeploymentWorkflowName, tt.createDeploymentErr)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployOrdererInput) (*pb.DeployOrdererOutput, error) {
				wf, err := workflows.DeployOrderer(ctx, &pb.DeployOrdererWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployOrdererInput{
				Spec: &pb.OrdererSpec{
					Id:        tt.ordererID,
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"},
				},
			})

			err := env.GetWorkflowError()

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
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

// testDeploymentSpec returns a standard deployment spec for testing
func testDeploymentSpec() *pb.OrgFabricDeployment {
	tlsRootRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	tlsIntRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "v1"}
	orgRootRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}
	orgIntRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "v1"}
	caValidity := durationpb.New(15 * 365 * 24 * time.Hour)

	return &pb.OrgFabricDeployment{
		CertificateAuthorities: []*pb.CertificateAuthoritySpec{
			{Self: tlsRootRef, SigningIssuer: nil, TlsIssuer: nil, ValidityDuration: caValidity},
			{Self: tlsIntRef, SigningIssuer: tlsRootRef, TlsIssuer: tlsRootRef, ValidityDuration: caValidity},
			{Self: orgRootRef, SigningIssuer: nil, TlsIssuer: tlsIntRef, ValidityDuration: caValidity},
			{Self: orgIntRef, SigningIssuer: orgRootRef, TlsIssuer: tlsIntRef, ValidityDuration: caValidity},
		},
		Peers: []*pb.PeerSpec{
			{Id: "peer0", OrgIssuer: orgIntRef, TlsIssuer: tlsIntRef},
		},
		Orderers: []*pb.OrdererSpec{
			{Id: "orderer0", OrgIssuer: orgIntRef, TlsIssuer: tlsIntRef},
		},
	}
}

func TestNewBootstrapShardWorkflow(t *testing.T) {
	fabricCfg := fabricconfig.NewFabricConfig(config.OrgConfig{OrgID: testOrg})

	tests := []struct {
		name      string
		input     *pb.BootstrapShardWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - valid deployment spec",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Deployment: testDeploymentSpec(),
			}},
			wantErr: false,
		},
		{
			name: "error - nil deployment",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Deployment: nil,
			}},
			wantErr:   true,
			errSubstr: "deployment is required",
		},
		{
			name: "error - no certificate authorities",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Deployment: &pb.OrgFabricDeployment{
					CertificateAuthorities: []*pb.CertificateAuthoritySpec{},
					Peers:                  []*pb.PeerSpec{{Id: "peer0"}},
					Orderers:               []*pb.OrdererSpec{{Id: "orderer0"}},
				},
			}},
			wantErr:   true,
			errSubstr: "at least one certificate authority is required",
		},
		{
			name: "error - no peers",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Deployment: &pb.OrgFabricDeployment{
					CertificateAuthorities: []*pb.CertificateAuthoritySpec{
						{Self: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}},
					},
					Peers:    []*pb.PeerSpec{},
					Orderers: []*pb.OrdererSpec{{Id: "orderer0"}},
				},
			}},
			wantErr:   true,
			errSubstr: "at least one peer is required",
		},
		{
			name: "error - no orderers",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Deployment: &pb.OrgFabricDeployment{
					CertificateAuthorities: []*pb.CertificateAuthoritySpec{
						{Self: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}},
					},
					Peers:    []*pb.PeerSpec{{Id: "peer0"}},
					Orderers: []*pb.OrdererSpec{},
				},
			}},
			wantErr:   true,
			errSubstr: "at least one orderer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newBootstrapShardWorkflow(tt.input, fabricCfg)

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

func TestBootstrapShardWorkflow(t *testing.T) {
	tests := []struct {
		name                      string
		org                       string
		deployCertStoreErr        error
		deployCertAuthorityErr    error  // Error for CA deployment
		deployCertAuthorityCertID string // Which CA fails (e.g., "tls/root", "org/v1")
		deployCouchDbErr          error  // Error for any CouchDB deployment
		deployCouchDbPeerID       string // Which peer's CouchDB fails (if deployCouchDbErr set)
		deployPeerErr             error  // Error for any peer deployment
		deployPeerID              string // Which peer fails (if deployPeerErr set)
		deployOrdererErr          error
		expectedErr               error
	}{
		{
			name:        "success - full shard bootstrap",
			org:         testOrg,
			expectedErr: nil,
		},
		{
			name:               "error - deploy certificate store fails",
			org:                testOrg,
			deployCertStoreErr: errors.New("postgres deployment failed"),
			expectedErr:        ErrDeployCertificateStore,
		},
		{
			name:                      "error - deploy TLS root CA fails",
			org:                       testOrg,
			deployCertAuthorityErr:    errors.New("CA deployment failed"),
			deployCertAuthorityCertID: "tls/root",
			expectedErr:               errors.New("deploy tls/root"),
		},
		{
			name:                      "error - deploy org intermediate CA fails",
			org:                       testOrg,
			deployCertAuthorityErr:    errors.New("CA deployment failed"),
			deployCertAuthorityCertID: "org/v1",
			expectedErr:               errors.New("deploy org/v1"),
		},
		{
			name:                "error - deploy CouchDB for peer0 fails",
			org:                 testOrg,
			deployCouchDbErr:    errors.New("CouchDB deployment failed"),
			deployCouchDbPeerID: "peer0",
			expectedErr:         errors.New("deploy couchdb for peer0"),
		},
		{
			name:          "error - deploy peer0 fails",
			org:           testOrg,
			deployPeerErr: errors.New("peer deployment failed"),
			deployPeerID:  "peer0",
			expectedErr:   errors.New("deploy peer peer0"),
		},
		{
			name:             "error - deploy orderer fails",
			org:              testOrg,
			deployOrdererErr: errors.New("orderer deployment failed"),
			expectedErr:      errors.New("deploy orderer orderer0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			orgConfig := config.OrgConfig{OrgID: tt.org}
			fabricCfg := fabricconfig.NewFabricConfig(orgConfig)
			workflows := NewWorkflows(fabricCfg)

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Mock DeployCertificateStore child workflow
			mockChildWorkflow[pb.DeployCertificateStoreOutput](env, pb.DeployCertificateStoreWorkflowName, tt.deployCertStoreErr)

			// Mock DeployCertificateAuthority child workflow with specific CA matching
			env.OnWorkflow(pb.DeployCertificateAuthorityWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.DeployCertificateAuthorityInput) (*pb.DeployCertificateAuthorityOutput, error) {
					caRef := input.GetSpec().GetSelf()
					certID := caRef.GetCertType() + "/" + caRef.GetId()
					if tt.deployCertAuthorityErr != nil && certID == tt.deployCertAuthorityCertID {
						return nil, tt.deployCertAuthorityErr
					}
					return &pb.DeployCertificateAuthorityOutput{}, nil
				})

			// Mock DeployCouchDb child workflow
			env.OnWorkflow(pb.DeployCouchDbWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.DeployCouchDbInput) (*pb.DeployCouchDbOutput, error) {
					if tt.deployCouchDbErr != nil && input.GetPeerId() == tt.deployCouchDbPeerID {
						return nil, tt.deployCouchDbErr
					}
					return &pb.DeployCouchDbOutput{}, nil
				})

			// Mock DeployPeer child workflow
			env.OnWorkflow(pb.DeployPeerWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.DeployPeerInput) (*pb.DeployPeerOutput, error) {
					if tt.deployPeerErr != nil && input.GetSpec().GetId() == tt.deployPeerID {
						return nil, tt.deployPeerErr
					}
					return &pb.DeployPeerOutput{}, nil
				})

			// Mock DeployOrderer child workflow
			mockChildWorkflow[pb.DeployOrdererOutput](env, pb.DeployOrdererWorkflowName, tt.deployOrdererErr)

			// Execute workflow with deployment spec
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.BootstrapShardInput) (*pb.BootstrapShardOutput, error) {
				wf, err := workflows.BootstrapShard(ctx, &pb.BootstrapShardWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.BootstrapShardInput{
				Deployment: testDeploymentSpec(),
			})

			err := env.GetWorkflowError()

			if tt.expectedErr == nil {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
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
