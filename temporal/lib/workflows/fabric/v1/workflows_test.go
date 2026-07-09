package fabric

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	clusterpb "github.com/metacensus/infra/temporal/gen/workflows/cluster/fabric/v1"
	certauthpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1/certauthv1nexus"
	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	postgrespb "github.com/metacensus/infra/temporal/gen/workflows/postgres/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/secrets/v1/secretsv1nexus"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	postgresv1mocks "github.com/metacensus/infra/temporal/mocks/workflows/postgres/v1"
	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
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
// These tests verify input validation without needing a workflow environment.
// They're fast and focused on the validation logic.

func TestNewCreateCertificateDatabaseWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.CreateCertificateDatabaseWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - tls root CA",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				Org:   testOrg,
				CaRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr: false,
		},
		{
			name: "success - org intermediate CA",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				Org:   testOrg,
				CaRef: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
			}},
			wantErr: false,
		},
		{
			name: "error - missing org",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"},
			}},
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name: "error - nil ca_ref",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				Org:   testOrg,
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
		{
			name: "error - invalid cert type",
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				Org:   testOrg,
				CaRef: &pb.CertificateAuthorityRef{CertType: "invalid", Id: "root"},
			}},
			wantErr:   true,
			errSubstr: "cert_type must be 'tls' or 'org'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateCertificateDatabaseWorkflow(tt.input)

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

func TestNewDeployPeerWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.DeployPeerWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - all required fields",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Org: testOrg,
				Spec: &pb.PeerSpec{
					Id:        "peer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				},
			}},
			wantErr: false,
		},
		{
			name: "error - missing org",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Spec: &pb.PeerSpec{
					Id:        "peer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				},
			}},
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name: "error - nil spec",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Org:  testOrg,
				Spec: nil,
			}},
			wantErr:   true,
			errSubstr: "spec is required",
		},
		{
			name: "error - missing org_issuer",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Org: testOrg,
				Spec: &pb.PeerSpec{
					Id:        "peer0",
					OrgIssuer: nil,
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				},
			}},
			wantErr:   true,
			errSubstr: "spec.org_issuer is required",
		},
		{
			name: "error - missing tls_issuer",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Org: testOrg,
				Spec: &pb.PeerSpec{
					Id:        "peer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
					TlsIssuer: nil,
				},
			}},
			wantErr:   true,
			errSubstr: "spec.tls_issuer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployPeerWorkflow(tt.input)

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

func TestNewDeployOrdererWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.DeployOrdererWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - all required fields",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Org: testOrg,
				Spec: &pb.OrdererSpec{
					Id:        "orderer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				},
			}},
			wantErr: false,
		},
		{
			name: "error - missing org",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Spec: &pb.OrdererSpec{
					Id:        "orderer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				},
			}},
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name: "error - nil spec",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Org:  testOrg,
				Spec: nil,
			}},
			wantErr:   true,
			errSubstr: "spec is required",
		},
		{
			name: "error - missing org_issuer",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Org: testOrg,
				Spec: &pb.OrdererSpec{
					Id:        "orderer0",
					OrgIssuer: nil,
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
				},
			}},
			wantErr:   true,
			errSubstr: "spec.org_issuer is required",
		},
		{
			name: "error - missing tls_issuer",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Org: testOrg,
				Spec: &pb.OrdererSpec{
					Id:        "orderer0",
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
					TlsIssuer: nil,
				},
			}},
			wantErr:   true,
			errSubstr: "spec.tls_issuer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployOrdererWorkflow(tt.input)

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

func TestNewBootstrapFabricWorkflow(t *testing.T) {
	t.Run("success - valid input", func(t *testing.T) {
		wf, err := newBootstrapFabricWorkflow(&pb.BootstrapFabricWorkflowInput{Req: &pb.BootstrapFabricInput{
			Org: testOperatorID,
		}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wf == nil {
			t.Fatal("expected non-nil workflow")
		}
	})

	t.Run("error - missing org", func(t *testing.T) {
		wf, err := newBootstrapFabricWorkflow(&pb.BootstrapFabricWorkflowInput{Req: &pb.BootstrapFabricInput{}})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "org is required") {
			t.Errorf("expected error containing %q, got: %v", "org is required", err)
		}
		if wf != nil {
			t.Error("expected nil workflow on error")
		}
	})
}

func TestNewBootstrapTlsCertificateAuthoritiesWorkflow(t *testing.T) {
	t.Run("success - valid input", func(t *testing.T) {
		wf, err := newBootstrapTlsCertificateAuthoritiesWorkflow(&pb.BootstrapTlsCertificateAuthoritiesWorkflowInput{
			Req: &pb.BootstrapTlsCertificateAuthoritiesInput{Org: testOperatorID},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wf == nil {
			t.Fatal("expected non-nil workflow")
		}
	})

	t.Run("error - missing org", func(t *testing.T) {
		wf, err := newBootstrapTlsCertificateAuthoritiesWorkflow(&pb.BootstrapTlsCertificateAuthoritiesWorkflowInput{
			Req: &pb.BootstrapTlsCertificateAuthoritiesInput{},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "org is required") {
			t.Errorf("expected error containing %q, got: %v", "org is required", err)
		}
		if wf != nil {
			t.Error("expected nil workflow on error")
		}
	})
}

func TestNewBootstrapOrgCertificateAuthoritiesWorkflow(t *testing.T) {
	t.Run("success - valid input", func(t *testing.T) {
		wf, err := newBootstrapOrgCertificateAuthoritiesWorkflow(&pb.BootstrapOrgCertificateAuthoritiesWorkflowInput{
			Req: &pb.BootstrapOrgCertificateAuthoritiesInput{Org: testOrg},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wf == nil {
			t.Fatal("expected non-nil workflow")
		}
	})

	t.Run("error - missing org", func(t *testing.T) {
		wf, err := newBootstrapOrgCertificateAuthoritiesWorkflow(&pb.BootstrapOrgCertificateAuthoritiesWorkflowInput{
			Req: &pb.BootstrapOrgCertificateAuthoritiesInput{},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "org is required") {
			t.Errorf("expected error containing %q, got: %v", "org is required", err)
		}
		if wf != nil {
			t.Error("expected nil workflow on error")
		}
	})
}

func TestNewDeployTlsRootCertificateAuthorityWorkflow(t *testing.T) {
	t.Run("success - valid input", func(t *testing.T) {
		wf, err := newDeployTlsRootCertificateAuthorityWorkflow(&pb.DeployTlsRootCertificateAuthorityWorkflowInput{
			Req: &pb.DeployTlsRootCertificateAuthorityInput{Org: testOperatorID},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wf == nil {
			t.Fatal("expected non-nil workflow")
		}
	})

	t.Run("error - missing org", func(t *testing.T) {
		wf, err := newDeployTlsRootCertificateAuthorityWorkflow(&pb.DeployTlsRootCertificateAuthorityWorkflowInput{
			Req: &pb.DeployTlsRootCertificateAuthorityInput{},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "org must not be empty") {
			t.Errorf("expected error containing %q, got: %v", "org must not be empty", err)
		}
		if wf != nil {
			t.Error("expected nil workflow on error")
		}
	})
}

func TestNewDeployOrgRootCertificateAuthorityWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.DeployOrgRootCertificateAuthorityWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - valid input",
			input: &pb.DeployOrgRootCertificateAuthorityWorkflowInput{Req: &pb.DeployOrgRootCertificateAuthorityInput{
				Org:       testOrg,
				TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			}},
			wantErr: false,
		},
		{
			name: "error - missing org",
			input: &pb.DeployOrgRootCertificateAuthorityWorkflowInput{Req: &pb.DeployOrgRootCertificateAuthorityInput{
				TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
			}},
			wantErr:   true,
			errSubstr: "org must not be empty",
		},
		{
			name: "error - missing tls_issuer",
			input: &pb.DeployOrgRootCertificateAuthorityWorkflowInput{Req: &pb.DeployOrgRootCertificateAuthorityInput{
				Org: testOrg,
			}},
			wantErr:   true,
			errSubstr: "tls_issuer must not be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployOrgRootCertificateAuthorityWorkflow(tt.input)

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

func TestNewBootstrapShardWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.BootstrapShardWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - valid deployment spec",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Org:        testOrg,
				Deployment: testDeploymentSpec(),
			}},
			wantErr: false,
		},
		{
			name: "error - missing org",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Deployment: testDeploymentSpec(),
			}},
			wantErr:   true,
			errSubstr: "org is required",
		},
		{
			name: "error - nil deployment",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Org:        testOrg,
				Deployment: nil,
			}},
			wantErr:   true,
			errSubstr: "deployment is required",
		},
		{
			name: "error - no certificate authorities",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Org: testOrg,
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
				Org: testOrg,
				Deployment: &pb.OrgFabricDeployment{
					CertificateAuthorities: []*pb.CertificateAuthoritySpec{
						{Self: &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}},
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
				Org: testOrg,
				Deployment: &pb.OrgFabricDeployment{
					CertificateAuthorities: []*pb.CertificateAuthoritySpec{
						{Self: &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}},
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
			wf, err := newBootstrapShardWorkflow(tt.input)

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
// Test Helpers
// =============================================================================

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

	register(pb.BootstrapTlsCertificateAuthoritiesWorkflowName,
		func(workflow.Context, *pb.BootstrapTlsCertificateAuthoritiesInput) (*pb.BootstrapTlsCertificateAuthoritiesOutput, error) {
			return nil, nil
		})

	register(pb.BootstrapOrgCertificateAuthoritiesWorkflowName,
		func(workflow.Context, *pb.BootstrapOrgCertificateAuthoritiesInput) (*pb.BootstrapOrgCertificateAuthoritiesOutput, error) {
			return nil, nil
		})

	register(pb.DeployTlsRootCertificateAuthorityWorkflowName,
		func(workflow.Context, *pb.DeployTlsRootCertificateAuthorityInput) (*pb.DeployTlsRootCertificateAuthorityOutput, error) {
			return nil, nil
		})

	register(pb.DeployTlsIntermediateCertificateAuthorityWorkflowName,
		func(workflow.Context, *pb.DeployTlsIntermediateCertificateAuthorityInput) (*pb.DeployTlsIntermediateCertificateAuthorityOutput, error) {
			return nil, nil
		})

	register(pb.DeployOrgRootCertificateAuthorityWorkflowName,
		func(workflow.Context, *pb.DeployOrgRootCertificateAuthorityInput) (*pb.DeployOrgRootCertificateAuthorityOutput, error) {
			return nil, nil
		})

	register(pb.DeployOrgIntermediateCertificateAuthorityWorkflowName,
		func(workflow.Context, *pb.DeployOrgIntermediateCertificateAuthorityInput) (*pb.DeployOrgIntermediateCertificateAuthorityOutput, error) {
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

	// certauth (certificate authority) child workflows
	register(certauthpb.CreateCertificateWorkflowName,
		func(workflow.Context, *certauthpb.CreateCertificateInput) (*certauthpb.CreateCertificateOutput, error) {
			return nil, nil
		})

	register(certauthpb.CreateBootstrapAdminWorkflowName,
		func(workflow.Context, *certauthpb.CreateBootstrapAdminInput) (*certauthpb.CreateBootstrapAdminOutput, error) {
			return nil, nil
		})

	register(certauthpb.CreateAdminWorkflowName,
		func(workflow.Context, *certauthpb.CreateAdminInput) (*certauthpb.CreateAdminOutput, error) {
			return nil, nil
		})

	register(certauthpb.CreateIdentityWorkflowName,
		func(workflow.Context, *certauthpb.CreateIdentityInput) (*certauthpb.CreateIdentityOutput, error) {
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

// Named activity stub functions. These must be top-level named functions (not
// anonymous closures): the Temporal test environment registers activity aliases
// by the function's short name, and anonymous closures all shorten to "func1",
// "func2", etc. which collides with the generated child-workflow ID-expression
// local activities (also anonymous), causing mocks to intercept the wrong call.

func stubCreateKeyPairActivity(ctx context.Context, input *secretspb.CreateKeyPairInput) (*secretspb.CreateKeyPairOutput, error) {
	return &secretspb.CreateKeyPairOutput{SecretRef: input.SecretRef}, nil
}

func stubCreateSelfSignedCertificateActivity(ctx context.Context, input *secretspb.CreateSelfSignedCertificateInput) (*secretspb.CreateSelfSignedCertificateOutput, error) {
	return &secretspb.CreateSelfSignedCertificateOutput{}, nil
}

func stubCreatePasswordActivity(ctx context.Context, input *secretspb.CreatePasswordInput) (*secretspb.CreatePasswordOutput, error) {
	return &secretspb.CreatePasswordOutput{SecretRef: input.SecretRef}, nil
}

// registerSecretsActivityStubs registers the secrets activities used by the CA
// deployment workflows so they can be mocked with OnActivity.
func registerSecretsActivityStubs(env *testsuite.TestWorkflowEnvironment) {
	env.RegisterActivityWithOptions(stubCreateKeyPairActivity,
		activity.RegisterOptions{Name: secretspb.CreateKeyPairActivityName})
	env.RegisterActivityWithOptions(stubCreateSelfSignedCertificateActivity,
		activity.RegisterOptions{Name: secretspb.CreateSelfSignedCertificateActivityName})
	env.RegisterActivityWithOptions(stubCreatePasswordActivity,
		activity.RegisterOptions{Name: secretspb.CreatePasswordActivityName})
}

// registerSecretsNexusMock registers the Secrets Nexus service and mocks the
// CreateAccess operation (used by workflows to provision Vault access).
func registerSecretsNexusMock(t *testing.T, env *testsuite.TestWorkflowEnvironment, opErr error) {
	t.Helper()
	secretsNexusSvc, nexusErr := secretsv1nexus.NewSecretsNexusService(&secretsv1nexus.UnimplementedSecretsNexusHandler{})
	if nexusErr != nil {
		t.Fatalf("failed to create secrets nexus service: %v", nexusErr)
	}
	env.RegisterNexusService(secretsNexusSvc)
	if opErr != nil {
		env.OnNexusOperation(secretsNexusSvc, secretsv1nexus.SecretsCreateAccessOperationName, mock.Anything, mock.Anything).
			Return(nil, opErr)
	} else {
		env.OnNexusOperation(secretsNexusSvc, secretsv1nexus.SecretsCreateAccessOperationName, mock.Anything, mock.Anything).
			Return(&nexus.HandlerStartOperationResultSync[*secretspb.CreateAccessOutput]{Value: &secretspb.CreateAccessOutput{}}, nil)
	}
}

// registerCertAuthNexusMock registers the CertificateAuthority Nexus service and
// mocks the CreateCertificate operation (cross-namespace TLS issuance by the
// shared TLS CAs on the fabric worker).
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
			Return(&nexus.HandlerStartOperationResultSync[*certauthpb.CreateCertificateOutput]{
				Value: &certauthpb.CreateCertificateOutput{
					Identity: &pb.IdentityRef{
						KeyRef:  secrets.NewVaultRef("test/keys"),
						CertRef: secrets.NewVaultRef("test/certs"),
					},
				},
			}, nil)
	}
}

// testDeploymentSpec returns a standard deployment spec for testing
func testDeploymentSpec() *pb.OrgFabricDeployment {
	tlsIntRef := &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"}
	orgRootRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}
	orgIntRef := &pb.CertificateAuthorityRef{CertType: "org", Id: "int"}
	caValidity := durationpb.New(15 * 365 * 24 * time.Hour)

	return &pb.OrgFabricDeployment{
		CertificateAuthorities: []*pb.CertificateAuthoritySpec{
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

// =============================================================================
// Workflow Execution Tests
// =============================================================================

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
			registerSecretsNexusMock(t, env, nil)

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
				workflows := NewWorkflows()
				wf, err := workflows.DeployCertificateStore(ctx, &pb.DeployCertificateStoreWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployCertificateStoreInput{Org: tt.org})

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
		generatePasswordErr    error
		createDatabaseErr      error
		createUserErr          error
		grantPrivilegesErr     error
		expectedErr            error
		validateCredentialsRef bool
	}{
		{
			name:                   "success - database created for tls root CA",
			org:                    testOrg,
			certType:               "tls",
			caID:                   "root",
			expectedErr:            nil,
			validateCredentialsRef: true,
		},
		{
			name:                   "success - database created for org intermediate CA",
			org:                    testOrg,
			certType:               "org",
			caID:                   "int",
			expectedErr:            nil,
			validateCredentialsRef: true,
		},
		{
			name:                "error - generate password fails",
			org:                 testOrg,
			certType:            "tls",
			caID:                "root",
			generatePasswordErr: errors.New("vault unavailable"),
			expectedErr:         errors.New("failed to generate credentials"),
		},
		{
			name:              "error - create database fails",
			org:               testOrg,
			certType:          "tls",
			caID:              "root",
			createDatabaseErr: errors.New("postgres connection refused"),
			expectedErr:       errors.New("failed to create database"),
		},
		{
			name:          "error - create user fails",
			org:           testOrg,
			certType:      "org",
			caID:          "int",
			createUserErr: errors.New("user already exists with different password"),
			expectedErr:   errors.New("failed to create user"),
		},
		{
			name:               "error - grant privileges fails",
			org:                testOrg,
			certType:           "org",
			caID:               "int",
			grantPrivilegesErr: errors.New("insufficient permissions"),
			expectedErr:        errors.New("failed to grant privileges"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

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
				workflows := NewWorkflows()
				wf, err := workflows.CreateCertificateDatabase(ctx, &pb.CreateCertificateDatabaseWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateCertificateDatabaseInput{
				Org: tt.org,
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
			expectedErr:     ErrCreateAccess,
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

			workflows := NewWorkflows()

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
			registerSecretsNexusMock(t, env, tt.createAccessErr)

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
			}, &pb.DeployCouchDbInput{Org: tt.org, PeerId: tt.peerID})

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
			expectedErr:     ErrCreateAccess,
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

			workflows := NewWorkflows()

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Mock Nexus CreateAccess
			registerSecretsNexusMock(t, env, tt.createAccessErr)

			// Mock CreateIdentity child workflow (certauth)
			mockChildWorkflow[certauthpb.CreateIdentityOutput](env, certauthpb.CreateIdentityWorkflowName, tt.createIdentityErr)

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
				Org: tt.org,
				Spec: &pb.PeerSpec{
					Id:        tt.peerID,
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
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
			expectedErr:     ErrCreateAccess,
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

			workflows := NewWorkflows()

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Mock Nexus CreateAccess
			registerSecretsNexusMock(t, env, tt.createAccessErr)

			// Mock CreateIdentity child workflow (certauth)
			mockChildWorkflow[certauthpb.CreateIdentityOutput](env, certauthpb.CreateIdentityWorkflowName, tt.createIdentityErr)

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
				Org: tt.org,
				Spec: &pb.OrdererSpec{
					Id:        tt.ordererID,
					OrgIssuer: &pb.CertificateAuthorityRef{CertType: "org", Id: "int"},
					TlsIssuer: &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
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

func TestBootstrapFabricWorkflow(t *testing.T) {
	tests := []struct {
		name            string
		bootstrapTlsErr error
		expectedErr     error
	}{
		{
			name:        "success - shared fabric infrastructure bootstrapped",
			expectedErr: nil,
		},
		{
			name:            "error - bootstrap TLS CAs fails",
			bootstrapTlsErr: errors.New("TLS root deployment failed"),
			expectedErr:     errors.New("bootstrap TLS certificate authorities"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()

			registerChildWorkflowStubs(env)

			// Mock BootstrapTlsCertificateAuthorities child workflow
			mockChildWorkflow[pb.BootstrapTlsCertificateAuthoritiesOutput](
				env, pb.BootstrapTlsCertificateAuthoritiesWorkflowName, tt.bootstrapTlsErr)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.BootstrapFabricInput) (*pb.BootstrapFabricOutput, error) {
				wf, err := workflows.BootstrapFabric(ctx, &pb.BootstrapFabricWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.BootstrapFabricInput{
				Org:                     testOperatorID,
				TlsRootValidity:         durationpb.New(15 * 365 * 24 * time.Hour),
				TlsIntermediateValidity: durationpb.New(5 * 365 * 24 * time.Hour),
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

func TestBootstrapTlsCertificateAuthoritiesWorkflow(t *testing.T) {
	tests := []struct {
		name               string
		deployCertStoreErr error
		deployTlsRootErr   error
		deployTlsIntErr    error
		expectedErr        error
	}{
		{
			name:        "success - TLS CA chain bootstrapped",
			expectedErr: nil,
		},
		{
			name:               "error - deploy certificate store fails",
			deployCertStoreErr: errors.New("postgres deployment failed"),
			expectedErr:        ErrDeployCertificateStore,
		},
		{
			name:             "error - deploy TLS root CA fails",
			deployTlsRootErr: errors.New("CA deployment failed"),
			expectedErr:      errors.New("deploy TLS Root CA"),
		},
		{
			name:            "error - deploy TLS intermediate CA fails",
			deployTlsIntErr: errors.New("CA deployment failed"),
			expectedErr:     errors.New("deploy TLS Intermediate CA"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()

			registerChildWorkflowStubs(env)

			mockChildWorkflow[pb.DeployCertificateStoreOutput](
				env, pb.DeployCertificateStoreWorkflowName, tt.deployCertStoreErr)
			mockChildWorkflow[pb.DeployTlsRootCertificateAuthorityOutput](
				env, pb.DeployTlsRootCertificateAuthorityWorkflowName, tt.deployTlsRootErr)
			mockChildWorkflow[pb.DeployTlsIntermediateCertificateAuthorityOutput](
				env, pb.DeployTlsIntermediateCertificateAuthorityWorkflowName, tt.deployTlsIntErr)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.BootstrapTlsCertificateAuthoritiesInput) (*pb.BootstrapTlsCertificateAuthoritiesOutput, error) {
				wf, err := workflows.BootstrapTlsCertificateAuthorities(ctx, &pb.BootstrapTlsCertificateAuthoritiesWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.BootstrapTlsCertificateAuthoritiesInput{
				Org:                     testOperatorID,
				TlsRootValidity:         durationpb.New(15 * 365 * 24 * time.Hour),
				TlsIntermediateValidity: durationpb.New(5 * 365 * 24 * time.Hour),
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

func TestBootstrapOrgCertificateAuthoritiesWorkflow(t *testing.T) {
	tests := []struct {
		name               string
		deployCertStoreErr error
		deployOrgRootErr   error
		deployOrgIntErr    error
		expectedErr        error
	}{
		{
			name:        "success - org CA chain bootstrapped",
			expectedErr: nil,
		},
		{
			name:               "error - deploy certificate store fails",
			deployCertStoreErr: errors.New("postgres deployment failed"),
			expectedErr:        ErrDeployCertificateStore,
		},
		{
			name:             "error - deploy org root CA fails",
			deployOrgRootErr: errors.New("CA deployment failed"),
			expectedErr:      errors.New("deploy Org Root CA"),
		},
		{
			name:            "error - deploy org intermediate CA fails",
			deployOrgIntErr: errors.New("CA deployment failed"),
			expectedErr:     errors.New("deploy Org Intermediate CA"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()

			registerChildWorkflowStubs(env)

			mockChildWorkflow[pb.DeployCertificateStoreOutput](
				env, pb.DeployCertificateStoreWorkflowName, tt.deployCertStoreErr)
			mockChildWorkflow[pb.DeployOrgRootCertificateAuthorityOutput](
				env, pb.DeployOrgRootCertificateAuthorityWorkflowName, tt.deployOrgRootErr)
			mockChildWorkflow[pb.DeployOrgIntermediateCertificateAuthorityOutput](
				env, pb.DeployOrgIntermediateCertificateAuthorityWorkflowName, tt.deployOrgIntErr)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.BootstrapOrgCertificateAuthoritiesInput) (*pb.BootstrapOrgCertificateAuthoritiesOutput, error) {
				wf, err := workflows.BootstrapOrgCertificateAuthorities(ctx, &pb.BootstrapOrgCertificateAuthoritiesWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.BootstrapOrgCertificateAuthoritiesInput{
				Org:                     testOrg,
				OrgRootValidity:         durationpb.New(15 * 365 * 24 * time.Hour),
				OrgIntermediateValidity: durationpb.New(5 * 365 * 24 * time.Hour),
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

func TestBootstrapShardWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		org                 string
		bootstrapOrgCAsErr  error
		deployCouchDbErr    error  // Error for any CouchDB deployment
		deployCouchDbPeerID string // Which peer's CouchDB fails (if deployCouchDbErr set)
		deployPeerErr       error  // Error for any peer deployment
		deployPeerID        string // Which peer fails (if deployPeerErr set)
		deployOrdererErr    error
		expectedErr         error
	}{
		{
			name:        "success - full shard bootstrap",
			org:         testOrg,
			expectedErr: nil,
		},
		{
			name:               "error - bootstrap org CAs fails",
			org:                testOrg,
			bootstrapOrgCAsErr: errors.New("org CA deployment failed"),
			expectedErr:        errors.New("bootstrap Org certificate authorities"),
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

			workflows := NewWorkflows()

			// Register child workflow stubs
			registerChildWorkflowStubs(env)

			// Mock BootstrapOrgCertificateAuthorities child workflow
			mockChildWorkflow[pb.BootstrapOrgCertificateAuthoritiesOutput](
				env, pb.BootstrapOrgCertificateAuthoritiesWorkflowName, tt.bootstrapOrgCAsErr)

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
				Org:        tt.org,
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

// =============================================================================
// Tests for Split CA Deployment Workflows
// =============================================================================

func TestDeployTlsRootCertificateAuthorityWorkflow(t *testing.T) {
	tests := []struct {
		name                         string
		org                          string
		createAccessErr              error
		createCertificateDatabaseErr error
		createKeyPairErr             error
		createSelfSignedCertErr      error
		createPasswordErr            error
		createDeploymentErr          error
		createBootstrapAdminErr      error
		createAdminErr               error
		expectedErr                  error
	}{
		{
			name:        "success - TLS Root CA deployed with self-signed certs",
			org:         testOperatorID,
			expectedErr: nil,
		},
		{
			name:            "failure - create access fails",
			org:             testOperatorID,
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:                         "failure - create certificate database fails",
			org:                          testOperatorID,
			createCertificateDatabaseErr: errors.New("postgres unavailable"),
			expectedErr:                  ErrCreateCertificateDatabase,
		},
		{
			name:             "failure - create key pair fails",
			org:              testOperatorID,
			createKeyPairErr: errors.New("vault unavailable"),
			expectedErr:      errors.New("failed to generate signing key pair"),
		},
		{
			name:                "failure - create deployment fails",
			org:                 testOperatorID,
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to deploy CA"),
		},
		{
			name:                    "failure - create bootstrap admin fails",
			org:                     testOperatorID,
			createBootstrapAdminErr: errors.New("enrollment failed"),
			expectedErr:             ErrCreateBootstrapAdmin,
		},
		{
			name:           "failure - create admin fails",
			org:            testOperatorID,
			createAdminErr: errors.New("registration failed"),
			expectedErr:    ErrCreateAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()
			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)

			// Mock Nexus call to secrets.v1.CreateAccess
			registerSecretsNexusMock(t, env, tt.createAccessErr)

			// Mock CreateCertificateDatabase child workflow
			mockChildWorkflow[pb.CreateCertificateDatabaseOutput](
				env,
				pb.CreateCertificateDatabaseWorkflowName,
				tt.createCertificateDatabaseErr,
			)

			// Mock CreateKeyPair activity
			if tt.createKeyPairErr != nil {
				env.OnActivity(secretspb.CreateKeyPairActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.createKeyPairErr)
			} else {
				env.OnActivity(secretspb.CreateKeyPairActivityName, mock.Anything, mock.Anything).
					Return(&secretspb.CreateKeyPairOutput{}, nil)
			}

			// Mock CreateSelfSignedCertificate activity
			env.OnActivity(secretspb.CreateSelfSignedCertificateActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreateSelfSignedCertificateOutput{}, tt.createSelfSignedCertErr)

			// Mock CreatePassword activity
			env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreatePasswordOutput{}, tt.createPasswordErr)

			// Mock CreateDeployment child workflow
			mockChildWorkflow[clusterpb.CreateDeploymentOutput](
				env,
				clusterpb.CreateDeploymentWorkflowName,
				tt.createDeploymentErr,
			)

			// Mock CreateBootstrapAdmin child workflow (certauth)
			mockChildWorkflow[certauthpb.CreateBootstrapAdminOutput](
				env,
				certauthpb.CreateBootstrapAdminWorkflowName,
				tt.createBootstrapAdminErr,
			)

			// Mock CreateAdmin child workflow (certauth)
			mockChildWorkflow[certauthpb.CreateAdminOutput](
				env,
				certauthpb.CreateAdminWorkflowName,
				tt.createAdminErr,
			)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployTlsRootCertificateAuthorityInput) (*pb.DeployTlsRootCertificateAuthorityOutput, error) {
				wf, err := workflows.DeployTlsRootCertificateAuthority(ctx, &pb.DeployTlsRootCertificateAuthorityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployTlsRootCertificateAuthorityInput{
				Org:              tt.org,
				ValidityDuration: durationpb.New(15 * 365 * 24 * time.Hour),
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

func TestDeployTlsIntermediateCertificateAuthorityWorkflow(t *testing.T) {
	tests := []struct {
		name                 string
		createAccessErr      error
		createCertificateErr error // CreateCertificate child (signing + tls, both direct)
		expectedErr          error
	}{
		{
			name:        "success - TLS Intermediate CA deployed",
			expectedErr: nil,
		},
		{
			name:            "failure - create access fails",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:                 "failure - create certificate fails",
			createCertificateErr: errors.New("CA enrollment failed"),
			expectedErr:          ErrCreateCertificate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()
			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)

			registerSecretsNexusMock(t, env, tt.createAccessErr)

			mockChildWorkflow[pb.CreateCertificateDatabaseOutput](
				env, pb.CreateCertificateDatabaseWorkflowName, nil)

			// Both the signing (profile "ca") and TLS (profile "tls") certificates
			// come from the operator-owned TLS Root via direct child workflow calls.
			mockChildWorkflow[certauthpb.CreateCertificateOutput](
				env, certauthpb.CreateCertificateWorkflowName, tt.createCertificateErr)

			env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreatePasswordOutput{}, nil)

			mockChildWorkflow[clusterpb.CreateDeploymentOutput](
				env, clusterpb.CreateDeploymentWorkflowName, nil)
			mockChildWorkflow[certauthpb.CreateBootstrapAdminOutput](
				env, certauthpb.CreateBootstrapAdminWorkflowName, nil)
			mockChildWorkflow[certauthpb.CreateAdminOutput](
				env, certauthpb.CreateAdminWorkflowName, nil)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployTlsIntermediateCertificateAuthorityInput) (*pb.DeployTlsIntermediateCertificateAuthorityOutput, error) {
				wf, err := workflows.DeployTlsIntermediateCertificateAuthority(ctx, &pb.DeployTlsIntermediateCertificateAuthorityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployTlsIntermediateCertificateAuthorityInput{
				Org:              testOperatorID,
				ValidityDuration: durationpb.New(5 * 365 * 24 * time.Hour),
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

func TestDeployOrgRootCertificateAuthorityWorkflow(t *testing.T) {
	tests := []struct {
		name            string
		createAccessErr error
		nexusTlsCertErr error // Nexus CreateCertificate call (TLS cert from shared TLS Int)
		expectedErr     error
	}{
		{
			name:        "success - Org Root CA deployed with TLS cert via Nexus",
			expectedErr: nil,
		},
		{
			name:            "failure - create access fails",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:            "failure - Nexus TLS certificate fails",
			nexusTlsCertErr: errors.New("TLS Int CA unavailable"),
			expectedErr:     ErrCreateCertificate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()
			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)

			registerSecretsNexusMock(t, env, tt.createAccessErr)

			// The TLS cert comes from the shared TLS Intermediate CA in the
			// operator namespace, via the CertificateAuthority Nexus service.
			registerCertAuthNexusMock(t, env, tt.nexusTlsCertErr)

			mockChildWorkflow[pb.CreateCertificateDatabaseOutput](
				env, pb.CreateCertificateDatabaseWorkflowName, nil)

			env.OnActivity(secretspb.CreateKeyPairActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreateKeyPairOutput{}, nil)
			env.OnActivity(secretspb.CreateSelfSignedCertificateActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreateSelfSignedCertificateOutput{}, nil)
			env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreatePasswordOutput{}, nil)

			mockChildWorkflow[clusterpb.CreateDeploymentOutput](
				env, clusterpb.CreateDeploymentWorkflowName, nil)
			mockChildWorkflow[certauthpb.CreateBootstrapAdminOutput](
				env, certauthpb.CreateBootstrapAdminWorkflowName, nil)
			mockChildWorkflow[certauthpb.CreateAdminOutput](
				env, certauthpb.CreateAdminWorkflowName, nil)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployOrgRootCertificateAuthorityInput) (*pb.DeployOrgRootCertificateAuthorityOutput, error) {
				wf, err := workflows.DeployOrgRootCertificateAuthority(ctx, &pb.DeployOrgRootCertificateAuthorityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployOrgRootCertificateAuthorityInput{
				Org:              testOrg,
				ValidityDuration: durationpb.New(15 * 365 * 24 * time.Hour),
				TlsIssuer:        &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
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

func TestDeployOrgIntermediateCertificateAuthorityWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		signingCertErr error // CreateCertificate child (signing cert from Org Root, direct)
		nexusTlsErr    error // Nexus CreateCertificate call (TLS cert from shared TLS Int)
		expectedErr    error
	}{
		{
			name:        "success - Org Intermediate CA deployed",
			expectedErr: nil,
		},
		{
			name:           "failure - signing certificate fails",
			signingCertErr: errors.New("Org Root CA unavailable"),
			expectedErr:    ErrCreateCertificate,
		},
		{
			name:        "failure - Nexus TLS certificate fails",
			nexusTlsErr: errors.New("TLS Int CA unavailable"),
			expectedErr: ErrCreateCertificate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows()
			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)

			registerSecretsNexusMock(t, env, nil)

			// The TLS cert comes from the shared TLS Intermediate CA via Nexus;
			// the signing cert comes from the Org Root via a direct child workflow.
			registerCertAuthNexusMock(t, env, tt.nexusTlsErr)

			mockChildWorkflow[pb.CreateCertificateDatabaseOutput](
				env, pb.CreateCertificateDatabaseWorkflowName, nil)
			mockChildWorkflow[certauthpb.CreateCertificateOutput](
				env, certauthpb.CreateCertificateWorkflowName, tt.signingCertErr)

			env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreatePasswordOutput{}, nil)

			mockChildWorkflow[clusterpb.CreateDeploymentOutput](
				env, clusterpb.CreateDeploymentWorkflowName, nil)
			mockChildWorkflow[certauthpb.CreateBootstrapAdminOutput](
				env, certauthpb.CreateBootstrapAdminWorkflowName, nil)
			mockChildWorkflow[certauthpb.CreateAdminOutput](
				env, certauthpb.CreateAdminWorkflowName, nil)

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployOrgIntermediateCertificateAuthorityInput) (*pb.DeployOrgIntermediateCertificateAuthorityOutput, error) {
				wf, err := workflows.DeployOrgIntermediateCertificateAuthority(ctx, &pb.DeployOrgIntermediateCertificateAuthorityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployOrgIntermediateCertificateAuthorityInput{
				Org:              testOrg,
				ValidityDuration: durationpb.New(5 * 365 * 24 * time.Hour),
				TlsIssuer:        &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"},
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
