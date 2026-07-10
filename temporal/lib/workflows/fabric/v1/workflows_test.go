package fabric

import (
	"context"
	"errors"
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
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/durationpb"
)

const testOrg = "testorg"
const testOperatorID = "metacensus"

// Worker identities under test. The fabric worker acts as the operator; a
// shard worker acts as its org. Tests construct workflows with the identity
// matching the worker the workflow would run on.
var (
	operatorOwner = workerconfig.NewOperatorConfig(testOperatorID)
	orgOwner      = workerconfig.NewOrgConfig(testOrg)
)

// CA refs used across tests
var (
	tlsRootRef = &pb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	tlsIntRef  = &pb.CertificateAuthorityRef{CertType: "tls", Id: "int"}
	orgRootRef = &pb.CertificateAuthorityRef{CertType: "org", Id: "root"}
	orgIntRef  = &pb.CertificateAuthorityRef{CertType: "org", Id: "int"}
)

var testCAValidity = durationpb.New(15 * 365 * 24 * time.Hour)

// caSpec builds a CertificateAuthoritySpec for tests
func caSpec(self, signingIssuer, tlsIssuer *pb.CertificateAuthorityRef) *pb.CertificateAuthoritySpec {
	return &pb.CertificateAuthoritySpec{
		Self:             self,
		SigningIssuer:    signingIssuer,
		TlsIssuer:        tlsIssuer,
		ValidityDuration: testCAValidity,
	}
}

// The four CA variants, keyed by the owner whose worker deploys them:
//   - tls/root, tls/int  -> operator (fabric worker)
//   - org/root, org/int  -> org (shard worker)
func tlsRootSpec() *pb.CertificateAuthoritySpec { return caSpec(tlsRootRef, nil, nil) }
func tlsIntSpec() *pb.CertificateAuthoritySpec  { return caSpec(tlsIntRef, tlsRootRef, tlsRootRef) }
func orgRootSpec() *pb.CertificateAuthoritySpec { return caSpec(orgRootRef, nil, tlsIntRef) }
func orgIntSpec() *pb.CertificateAuthoritySpec  { return caSpec(orgIntRef, orgRootRef, tlsIntRef) }

// =============================================================================
// Constructor / Validation Tests
// =============================================================================
// These tests verify input validation without needing a workflow environment.

func TestNewCreateCertificateDatabaseWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		owner     workerconfig.OwnerConfig
		input     *pb.CreateCertificateDatabaseWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "success - tls root CA on operator worker",
			owner: operatorOwner,
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: tlsRootRef,
			}},
			wantErr: false,
		},
		{
			name:  "success - org intermediate CA on shard worker",
			owner: orgOwner,
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: orgIntRef,
			}},
			wantErr: false,
		},
		{
			name:  "error - nil ca_ref",
			owner: operatorOwner,
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
		{
			name:  "error - invalid cert type",
			owner: operatorOwner,
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: &pb.CertificateAuthorityRef{CertType: "invalid", Id: "root"},
			}},
			wantErr:   true,
			errSubstr: "cert_type must be 'tls' or 'org'",
		},
		{
			name:  "error - shard worker given a tls CA (not owned)",
			owner: orgOwner,
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: tlsRootRef,
			}},
			wantErr:   true,
			errSubstr: "not owned",
		},
		{
			name:  "error - operator worker given an org CA (not owned)",
			owner: operatorOwner,
			input: &pb.CreateCertificateDatabaseWorkflowInput{Req: &pb.CreateCertificateDatabaseInput{
				CaRef: orgRootRef,
			}},
			wantErr:   true,
			errSubstr: "not owned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateCertificateDatabaseWorkflow(tt.input, tt.owner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

func TestNewDeployCertificateAuthorityWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		owner     workerconfig.OwnerConfig
		spec      *pb.CertificateAuthoritySpec
		wantErr   bool
		errSubstr string
	}{
		// One accept case per variant on the matching worker
		{name: "success - tls/root on operator worker", owner: operatorOwner, spec: tlsRootSpec()},
		{name: "success - tls/int on operator worker", owner: operatorOwner, spec: tlsIntSpec()},
		{name: "success - org/root on shard worker", owner: orgOwner, spec: orgRootSpec()},
		{name: "success - org/int on shard worker", owner: orgOwner, spec: orgIntSpec()},

		// Ownership rejections
		{
			name:      "error - shard worker given tls/root spec",
			owner:     orgOwner,
			spec:      tlsRootSpec(),
			wantErr:   true,
			errSubstr: "not owned",
		},
		{
			name:      "error - shard worker given tls/int spec",
			owner:     orgOwner,
			spec:      tlsIntSpec(),
			wantErr:   true,
			errSubstr: "not owned",
		},
		{
			name:      "error - operator worker given org/root spec",
			owner:     operatorOwner,
			spec:      orgRootSpec(),
			wantErr:   true,
			errSubstr: "not owned",
		},
		{
			name:      "error - signing issuer crosses ownership boundary",
			owner:     orgOwner,
			spec:      caSpec(orgIntRef, tlsRootRef, tlsIntRef),
			wantErr:   true,
			errSubstr: "signing_issuer",
		},

		// Structural rejections
		{
			name:      "error - nil spec",
			owner:     operatorOwner,
			spec:      nil,
			wantErr:   true,
			errSubstr: "spec is required",
		},
		{
			name:      "error - missing self",
			owner:     operatorOwner,
			spec:      caSpec(nil, nil, nil),
			wantErr:   true,
			errSubstr: "spec.self is required",
		},
		{
			name:      "error - invalid self cert type",
			owner:     operatorOwner,
			spec:      caSpec(&pb.CertificateAuthorityRef{CertType: "bogus", Id: "root"}, nil, nil),
			wantErr:   true,
			errSubstr: "cert_type must be 'tls' or 'org'",
		},
		{
			name:      "error - missing self id",
			owner:     operatorOwner,
			spec:      caSpec(&pb.CertificateAuthorityRef{CertType: "tls"}, nil, nil),
			wantErr:   true,
			errSubstr: "ca id is required",
		},
		{
			name:      "error - invalid tls issuer cert type",
			owner:     operatorOwner,
			spec:      caSpec(tlsIntRef, tlsRootRef, &pb.CertificateAuthorityRef{CertType: "bogus", Id: "root"}),
			wantErr:   true,
			errSubstr: "cert_type must be 'tls' or 'org'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployCertificateAuthorityWorkflow(&pb.DeployCertificateAuthorityWorkflowInput{
				Req: &pb.DeployCertificateAuthorityInput{Spec: tt.spec},
			}, tt.owner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
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
				Spec: &pb.PeerSpec{Id: "peer0", OrgIssuer: orgIntRef, TlsIssuer: tlsIntRef},
			}},
			wantErr: false,
		},
		{
			name:      "error - nil spec",
			input:     &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{Spec: nil}},
			wantErr:   true,
			errSubstr: "spec is required",
		},
		{
			name: "error - missing org_issuer",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Spec: &pb.PeerSpec{Id: "peer0", OrgIssuer: nil, TlsIssuer: tlsIntRef},
			}},
			wantErr:   true,
			errSubstr: "spec.org_issuer is required",
		},
		{
			name: "error - missing tls_issuer",
			input: &pb.DeployPeerWorkflowInput{Req: &pb.DeployPeerInput{
				Spec: &pb.PeerSpec{Id: "peer0", OrgIssuer: orgIntRef, TlsIssuer: nil},
			}},
			wantErr:   true,
			errSubstr: "spec.tls_issuer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployPeerWorkflow(tt.input, orgOwner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
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
				Spec: &pb.OrdererSpec{Id: "orderer0", OrgIssuer: orgIntRef, TlsIssuer: tlsIntRef},
			}},
			wantErr: false,
		},
		{
			name:      "error - nil spec",
			input:     &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{Spec: nil}},
			wantErr:   true,
			errSubstr: "spec is required",
		},
		{
			name: "error - missing org_issuer",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Spec: &pb.OrdererSpec{Id: "orderer0", OrgIssuer: nil, TlsIssuer: tlsIntRef},
			}},
			wantErr:   true,
			errSubstr: "spec.org_issuer is required",
		},
		{
			name: "error - missing tls_issuer",
			input: &pb.DeployOrdererWorkflowInput{Req: &pb.DeployOrdererInput{
				Spec: &pb.OrdererSpec{Id: "orderer0", OrgIssuer: orgIntRef, TlsIssuer: nil},
			}},
			wantErr:   true,
			errSubstr: "spec.tls_issuer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newDeployOrdererWorkflow(tt.input, orgOwner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

func TestNewBootstrapFabricWorkflow(t *testing.T) {
	t.Run("success - valid input", func(t *testing.T) {
		wf, err := newBootstrapFabricWorkflow(&pb.BootstrapFabricWorkflowInput{Req: &pb.BootstrapFabricInput{
			CertificateAuthorities: []*pb.CertificateAuthoritySpec{tlsRootSpec(), tlsIntSpec()},
		}}, operatorOwner)
		assertConstructorResult(t, wf == nil, err, false, "")
	})

	t.Run("error - no certificate authorities", func(t *testing.T) {
		wf, err := newBootstrapFabricWorkflow(&pb.BootstrapFabricWorkflowInput{Req: &pb.BootstrapFabricInput{}}, operatorOwner)
		assertConstructorResult(t, wf == nil, err, true, "at least one certificate authority spec is required")
	})
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
					CertificateAuthorities: []*pb.CertificateAuthoritySpec{orgRootSpec()},
					Peers:                  []*pb.PeerSpec{},
					Orderers:               []*pb.OrdererSpec{{Id: "orderer0"}},
				},
			}},
			wantErr:   true,
			errSubstr: "at least one peer is required",
		},
		{
			name: "error - no orderers",
			input: &pb.BootstrapShardWorkflowInput{Req: &pb.BootstrapShardInput{
				Deployment: &pb.OrgFabricDeployment{
					CertificateAuthorities: []*pb.CertificateAuthoritySpec{orgRootSpec()},
					Peers:                  []*pb.PeerSpec{{Id: "peer0"}},
					Orderers:               []*pb.OrdererSpec{},
				},
			}},
			wantErr:   true,
			errSubstr: "at least one orderer is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newBootstrapShardWorkflow(tt.input, orgOwner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

// assertConstructorResult checks a workflow constructor's (wf, err) pair
func assertConstructorResult(t *testing.T, wfIsNil bool, err error, wantErr bool, errSubstr string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if errSubstr != "" && !strings.Contains(err.Error(), errSubstr) {
			t.Errorf("expected error containing %q, got: %v", errSubstr, err)
		}
		if !wfIsNil {
			t.Error("expected nil workflow on error")
		}
	} else {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wfIsNil {
			t.Fatal("expected non-nil workflow")
		}
	}
}

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

	register(pb.DeployCertificateAuthorityWorkflowName,
		func(workflow.Context, *pb.DeployCertificateAuthorityInput) (*pb.DeployCertificateAuthorityOutput, error) {
			return nil, nil
		})

	register(pb.CreateCertificateDatabaseWorkflowName,
		func(workflow.Context, *pb.CreateCertificateDatabaseInput) (*pb.CreateCertificateDatabaseOutput, error) {
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

func stubGetStatusActivity(ctx context.Context, input *kubepb.GetStatusInput) (*kubepb.GetStatusOutput, error) {
	return &kubepb.GetStatusOutput{}, nil
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
	return &pb.OrgFabricDeployment{
		CertificateAuthorities: []*pb.CertificateAuthoritySpec{
			orgRootSpec(),
			orgIntSpec(),
		},
		Peers: []*pb.PeerSpec{
			{Id: "peer0", OrgIssuer: orgIntRef, TlsIssuer: tlsIntRef},
		},
		Orderers: []*pb.OrdererSpec{
			{Id: "orderer0", OrgIssuer: orgIntRef, TlsIssuer: tlsIntRef},
		},
	}
}

// assertWorkflowResult checks the workflow error against expectations
func assertWorkflowResult(t *testing.T, err error, expectedErr error) {
	t.Helper()
	if expectedErr == nil {
		if err != nil {
			t.Fatalf("unexpected workflow error: %v", err)
		}
	} else {
		if err == nil {
			t.Fatalf("expected error %v, got nil", expectedErr)
		}
		// Temporal serializes workflow errors, breaking Go's error chain, so we check the message.
		if !strings.Contains(err.Error(), expectedErr.Error()) {
			t.Fatalf("expected error containing %q, got: %v", expectedErr, err)
		}
	}
}

// =============================================================================
// Workflow Execution Tests
// =============================================================================

func TestDeployCertificateStoreWorkflow(t *testing.T) {
	tests := []struct {
		name                    string
		owner                   workerconfig.OwnerConfig
		createDeploymentErr     error
		expectedErr             error // nil = success
		validateDeploymentInput bool  // if true, verify the deployment input
	}{
		{
			name:                "failure - create deployment fails",
			owner:               orgOwner,
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to create certificate store deployment"),
		},
		{
			name:                    "success - certificate store deployed on shard worker",
			owner:                   orgOwner,
			expectedErr:             nil,
			validateDeploymentInput: true,
		},
		{
			name:                    "success - certificate store deployed on operator worker",
			owner:                   operatorOwner,
			expectedErr:             nil,
			validateDeploymentInput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)
			env.RegisterActivityWithOptions(stubGetStatusActivity,
				activity.RegisterOptions{Name: kubepb.GetStatusActivityName})

			// Mock Secrets CreateAccess Nexus operation
			registerSecretsNexusMock(t, env, nil)

			// Mock activities
			env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreatePasswordOutput{SecretRef: secrets.NewVaultRef("test/path")}, nil)
			env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).Return(&kubepb.GetStatusOutput{}, nil)

			// Mock CreateDeployment child workflow
			if tt.validateDeploymentInput {
				// The deployment must target the injected owner, not any input value
				env.OnWorkflow(clusterpb.CreateDeploymentWorkflowName, mock.Anything,
					mock.MatchedBy(func(input *clusterpb.CreateDeploymentInput) bool {
						ownerCfg, err := workerconfig.NewOwnerConfig(input.GetOwner())
						return err == nil &&
							ownerCfg.OwnerID() == tt.owner.OwnerID() &&
							ownerCfg.IsOperator() == tt.owner.IsOperator() &&
							input.Component == "certificate-store"
					})).Return(&clusterpb.CreateDeploymentOutput{
					DeploymentName: "certificate-store",
				}, tt.createDeploymentErr)
			} else {
				mockChildWorkflow[clusterpb.CreateDeploymentOutput](env, clusterpb.CreateDeploymentWorkflowName, tt.createDeploymentErr)
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployCertificateStoreInput) (*pb.DeployCertificateStoreOutput, error) {
				workflows := NewWorkflows(tt.owner)
				wf, err := workflows.DeployCertificateStore(ctx, &pb.DeployCertificateStoreWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.DeployCertificateStoreInput{})

			var result *pb.DeployCertificateStoreOutput
			err := env.GetWorkflowResult(&result)

			if tt.expectedErr == nil && err == nil && result == nil {
				t.Fatal("result is nil")
			}
			assertWorkflowResult(t, err, tt.expectedErr)
		})
	}
}

func TestCreateCertificateDatabaseWorkflow(t *testing.T) {
	tests := []struct {
		name                   string
		owner                  workerconfig.OwnerConfig
		caRef                  *pb.CertificateAuthorityRef
		generatePasswordErr    error
		createDatabaseErr      error
		createUserErr          error
		grantPrivilegesErr     error
		expectedErr            error
		validateCredentialsRef bool
	}{
		{
			name:                   "success - database created for tls root CA on operator worker",
			owner:                  operatorOwner,
			caRef:                  tlsRootRef,
			expectedErr:            nil,
			validateCredentialsRef: true,
		},
		{
			name:                   "success - database created for org intermediate CA on shard worker",
			owner:                  orgOwner,
			caRef:                  orgIntRef,
			expectedErr:            nil,
			validateCredentialsRef: true,
		},
		{
			name:                "error - generate password fails",
			owner:               operatorOwner,
			caRef:               tlsRootRef,
			generatePasswordErr: errors.New("vault unavailable"),
			expectedErr:         errors.New("failed to generate credentials"),
		},
		{
			name:              "error - create database fails",
			owner:             operatorOwner,
			caRef:             tlsRootRef,
			createDatabaseErr: errors.New("postgres connection refused"),
			expectedErr:       errors.New("failed to create database"),
		},
		{
			name:          "error - create user fails",
			owner:         orgOwner,
			caRef:         orgIntRef,
			createUserErr: errors.New("user already exists with different password"),
			expectedErr:   errors.New("failed to create user"),
		},
		{
			name:               "error - grant privileges fails",
			owner:              orgOwner,
			caRef:              orgIntRef,
			grantPrivilegesErr: errors.New("insufficient permissions"),
			expectedErr:        errors.New("failed to grant privileges"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)
			env.RegisterActivityWithOptions(stubGetStatusActivity,
				activity.RegisterOptions{Name: kubepb.GetStatusActivityName})

			// Mock PostgreSQL activities - must register BEFORE OnActivity calls
			mockPostgresActivities := postgresv1mocks.NewMockPostgresActivities(t)

			if tt.createDatabaseErr != nil {
				mockPostgresActivities.EXPECT().
					CreateDatabase(mock.Anything, mock.Anything).
					Return(nil, tt.createDatabaseErr).
					Maybe()
			} else {
				mockPostgresActivities.EXPECT().
					CreateDatabase(mock.Anything, mock.Anything).
					Return(&postgrespb.CreateDatabaseOutput{}, nil).
					Maybe()
			}

			if tt.createUserErr != nil {
				mockPostgresActivities.EXPECT().
					CreateUser(mock.Anything, mock.Anything).
					Return(nil, tt.createUserErr).
					Maybe()
			} else {
				mockPostgresActivities.EXPECT().
					CreateUser(mock.Anything, mock.Anything).
					Return(&postgrespb.CreateUserOutput{}, nil).
					Maybe()
			}

			if tt.grantPrivilegesErr != nil {
				mockPostgresActivities.EXPECT().
					GrantPrivileges(mock.Anything, mock.Anything).
					Return(nil, tt.grantPrivilegesErr).
					Maybe()
			} else {
				mockPostgresActivities.EXPECT().
					GrantPrivileges(mock.Anything, mock.Anything).
					Return(&postgrespb.GrantPrivilegesOutput{}, nil).
					Maybe()
			}

			// Register activities - must be before OnActivity calls
			postgrespb.RegisterPostgresActivities(env, mockPostgresActivities)

			// Mock activities using OnActivity (must come after all RegisterActivity calls)
			env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).
				Return(&kubepb.GetStatusOutput{}, nil)

			if tt.generatePasswordErr != nil {
				env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.generatePasswordErr).Maybe()
			} else {
				env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
					Return(&secretspb.CreatePasswordOutput{
						SecretRef: secrets.NewVaultRef("fabric/" + tt.owner.OwnerID() + "/ca/" + tt.caRef.GetCertType() + "/" + tt.caRef.GetId() + "/db/credentials"),
					}, nil).Maybe()
			}

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateCertificateDatabaseInput) (*pb.CreateCertificateDatabaseOutput, error) {
				workflows := NewWorkflows(tt.owner)
				wf, err := workflows.CreateCertificateDatabase(ctx, &pb.CreateCertificateDatabaseWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateCertificateDatabaseInput{
				CaRef: tt.caRef,
			})

			var result *pb.CreateCertificateDatabaseOutput
			err := env.GetWorkflowResult(&result)

			assertWorkflowResult(t, err, tt.expectedErr)

			if tt.expectedErr == nil && tt.validateCredentialsRef {
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
		})
	}
}

func TestDeployCouchDbWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		peerID              string
		createAccessErr     error
		createPasswordErr   error
		createDeploymentErr error
		getStatusErr        error
		expectedErr         error
	}{
		{
			name:        "success - CouchDB deployed and ready",
			peerID:      "peer0",
			expectedErr: nil,
		},
		{
			name:            "error - create vault access fails",
			peerID:          "peer0",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:              "error - create credentials fails",
			peerID:            "peer0",
			createPasswordErr: errors.New("vault unavailable"),
			expectedErr:       errors.New("failed to generate couchdb credentials"),
		},
		{
			name:                "error - create deployment fails",
			peerID:              "peer0",
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to create couchdb deployment"),
		},
		{
			name:         "error - CouchDB never ready",
			peerID:       "peer0",
			getStatusErr: errors.New("pod not ready"),
			expectedErr:  errors.New("CouchDB did not become ready"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(orgOwner)

			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)
			env.RegisterActivityWithOptions(stubGetStatusActivity,
				activity.RegisterOptions{Name: kubepb.GetStatusActivityName})

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
			}, &pb.DeployCouchDbInput{PeerId: tt.peerID})

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

func TestDeployPeerWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		peerID              string
		createAccessErr     error
		createIdentityErr   error
		createDeploymentErr error
		expectedErr         error
	}{
		{
			name:        "success - peer deployed",
			peerID:      "peer0",
			expectedErr: nil,
		},
		{
			name:            "error - create vault access fails",
			peerID:          "peer0",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:              "error - create identity fails",
			peerID:            "peer0",
			createIdentityErr: errors.New("CA enrollment failed"),
			expectedErr:       errors.New("failed to create peer identity"),
		},
		{
			name:                "error - create deployment fails",
			peerID:              "peer0",
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to create peer deployment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(orgOwner)

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
				Spec: &pb.PeerSpec{Id: tt.peerID, OrgIssuer: orgIntRef, TlsIssuer: tlsIntRef},
			})

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

func TestDeployOrdererWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		ordererID           string
		createAccessErr     error
		createIdentityErr   error
		createDeploymentErr error
		expectedErr         error
	}{
		{
			name:        "success - orderer deployed",
			ordererID:   "orderer0",
			expectedErr: nil,
		},
		{
			name:            "error - create vault access fails",
			ordererID:       "orderer0",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:              "error - create identity fails",
			ordererID:         "orderer0",
			createIdentityErr: errors.New("CA enrollment failed"),
			expectedErr:       errors.New("failed to create orderer identity"),
		},
		{
			name:                "error - create deployment fails",
			ordererID:           "orderer0",
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to create orderer deployment"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(orgOwner)

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
				Spec: &pb.OrdererSpec{Id: tt.ordererID, OrgIssuer: orgIntRef, TlsIssuer: tlsIntRef},
			})

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

func TestBootstrapFabricWorkflow(t *testing.T) {
	tests := []struct {
		name               string
		deployCertStoreErr error
		deployCAErr        error  // Error for any CA deployment
		deployCAFailID     string // Which CA fails (matched on spec.self.id)
		expectedErr        error
	}{
		{
			name:        "success - certificate store and TLS CA chain deployed in order",
			expectedErr: nil,
		},
		{
			name:               "error - deploy certificate store fails",
			deployCertStoreErr: errors.New("postgres deployment failed"),
			expectedErr:        ErrDeployCertificateStore,
		},
		{
			name:           "error - deploy TLS root CA fails",
			deployCAErr:    errors.New("CA deployment failed"),
			deployCAFailID: "root",
			expectedErr:    errors.New("deploy certificate authority (tls-root)"),
		},
		{
			name:           "error - deploy TLS intermediate CA fails",
			deployCAErr:    errors.New("CA deployment failed"),
			deployCAFailID: "int",
			expectedErr:    errors.New("deploy certificate authority (tls-int)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(operatorOwner)

			registerChildWorkflowStubs(env)

			mockChildWorkflow[pb.DeployCertificateStoreOutput](
				env, pb.DeployCertificateStoreWorkflowName, tt.deployCertStoreErr)

			env.OnWorkflow(pb.DeployCertificateAuthorityWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.DeployCertificateAuthorityInput) (*pb.DeployCertificateAuthorityOutput, error) {
					if tt.deployCAErr != nil && input.GetSpec().GetSelf().GetId() == tt.deployCAFailID {
						return nil, tt.deployCAErr
					}
					return &pb.DeployCertificateAuthorityOutput{}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.BootstrapFabricInput) (*pb.BootstrapFabricOutput, error) {
				wf, err := workflows.BootstrapFabric(ctx, &pb.BootstrapFabricWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.BootstrapFabricInput{
				CertificateAuthorities: []*pb.CertificateAuthoritySpec{tlsRootSpec(), tlsIntSpec()},
			})

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

func TestBootstrapShardWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		deployCertStoreErr  error
		deployCAErr         error  // Error for any CA deployment
		deployCAFailID      string // Which CA fails (matched on spec.self.id)
		deployCouchDbErr    error  // Error for any CouchDB deployment
		deployCouchDbPeerID string // Which peer's CouchDB fails (if deployCouchDbErr set)
		deployPeerErr       error  // Error for any peer deployment
		deployPeerID        string // Which peer fails (if deployPeerErr set)
		deployOrdererErr    error
		expectedErr         error
	}{
		{
			name:        "success - full shard bootstrap",
			expectedErr: nil,
		},
		{
			name:               "error - deploy certificate store fails",
			deployCertStoreErr: errors.New("postgres deployment failed"),
			expectedErr:        ErrDeployCertificateStore,
		},
		{
			name:           "error - deploy org root CA fails",
			deployCAErr:    errors.New("org CA deployment failed"),
			deployCAFailID: "root",
			expectedErr:    errors.New("deploy certificate authority (org-root)"),
		},
		{
			name:           "error - deploy org intermediate CA fails",
			deployCAErr:    errors.New("org CA deployment failed"),
			deployCAFailID: "int",
			expectedErr:    errors.New("deploy certificate authority (org-int)"),
		},
		{
			name:                "error - deploy CouchDB for peer0 fails",
			deployCouchDbErr:    errors.New("CouchDB deployment failed"),
			deployCouchDbPeerID: "peer0",
			expectedErr:         errors.New("deploy couchdb for peer0"),
		},
		{
			name:          "error - deploy peer0 fails",
			deployPeerErr: errors.New("peer deployment failed"),
			deployPeerID:  "peer0",
			expectedErr:   errors.New("deploy peer peer0"),
		},
		{
			name:             "error - deploy orderer fails",
			deployOrdererErr: errors.New("orderer deployment failed"),
			expectedErr:      errors.New("deploy orderer orderer0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(orgOwner)

			registerChildWorkflowStubs(env)

			mockChildWorkflow[pb.DeployCertificateStoreOutput](
				env, pb.DeployCertificateStoreWorkflowName, tt.deployCertStoreErr)

			env.OnWorkflow(pb.DeployCertificateAuthorityWorkflowName, mock.Anything, mock.Anything).
				Return(func(ctx workflow.Context, input *pb.DeployCertificateAuthorityInput) (*pb.DeployCertificateAuthorityOutput, error) {
					if tt.deployCAErr != nil && input.GetSpec().GetSelf().GetId() == tt.deployCAFailID {
						return nil, tt.deployCAErr
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

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

// =============================================================================
// DeployCertificateAuthority Execution Tests
// =============================================================================
// One fixture per variant (tls/root, tls/int, org/root, org/int) exercising
// the single DeployCertificateAuthority workflow.

// executeDeployCA runs DeployCertificateAuthority with the given owner and spec
func executeDeployCA(env *testsuite.TestWorkflowEnvironment, owner workerconfig.OwnerConfig, spec *pb.CertificateAuthoritySpec) {
	env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.DeployCertificateAuthorityInput) (*pb.DeployCertificateAuthorityOutput, error) {
		workflows := NewWorkflows(owner)
		wf, err := workflows.DeployCertificateAuthority(ctx, &pb.DeployCertificateAuthorityWorkflowInput{Req: input})
		if err != nil {
			return nil, err
		}
		return wf.Execute(ctx)
	}, &pb.DeployCertificateAuthorityInput{Spec: spec})
}

// TestDeployCertificateAuthorityTlsRoot: operator worker, both identities self-signed
func TestDeployCertificateAuthorityTlsRoot(t *testing.T) {
	tests := []struct {
		name                         string
		createAccessErr              error
		createCertificateDatabaseErr error
		createKeyPairErr             error
		createDeploymentErr          error
		createBootstrapAdminErr      error
		createAdminErr               error
		expectedErr                  error
	}{
		{
			name:        "success - TLS Root CA deployed with self-signed certs",
			expectedErr: nil,
		},
		{
			name:            "failure - create access fails",
			createAccessErr: errors.New("vault policy creation failed"),
			expectedErr:     ErrCreateAccess,
		},
		{
			name:                         "failure - create certificate database fails",
			createCertificateDatabaseErr: errors.New("postgres unavailable"),
			expectedErr:                  ErrCreateCertificateDatabase,
		},
		{
			name:             "failure - create key pair fails",
			createKeyPairErr: errors.New("vault unavailable"),
			expectedErr:      errors.New("failed to generate signing key pair"),
		},
		{
			name:                "failure - create deployment fails",
			createDeploymentErr: errors.New("kubernetes error"),
			expectedErr:         errors.New("failed to deploy CA"),
		},
		{
			name:                    "failure - create bootstrap admin fails",
			createBootstrapAdminErr: errors.New("enrollment failed"),
			expectedErr:             ErrCreateBootstrapAdmin,
		},
		{
			name:           "failure - create admin fails",
			createAdminErr: errors.New("registration failed"),
			expectedErr:    ErrCreateAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)

			registerSecretsNexusMock(t, env, tt.createAccessErr)

			mockChildWorkflow[pb.CreateCertificateDatabaseOutput](
				env, pb.CreateCertificateDatabaseWorkflowName, tt.createCertificateDatabaseErr)

			if tt.createKeyPairErr != nil {
				env.OnActivity(secretspb.CreateKeyPairActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.createKeyPairErr)
			} else {
				env.OnActivity(secretspb.CreateKeyPairActivityName, mock.Anything, mock.Anything).
					Return(&secretspb.CreateKeyPairOutput{}, nil)
			}
			env.OnActivity(secretspb.CreateSelfSignedCertificateActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreateSelfSignedCertificateOutput{}, nil)
			env.OnActivity(secretspb.CreatePasswordActivityName, mock.Anything, mock.Anything).
				Return(&secretspb.CreatePasswordOutput{}, nil)

			mockChildWorkflow[clusterpb.CreateDeploymentOutput](
				env, clusterpb.CreateDeploymentWorkflowName, tt.createDeploymentErr)
			mockChildWorkflow[certauthpb.CreateBootstrapAdminOutput](
				env, certauthpb.CreateBootstrapAdminWorkflowName, tt.createBootstrapAdminErr)
			mockChildWorkflow[certauthpb.CreateAdminOutput](
				env, certauthpb.CreateAdminWorkflowName, tt.createAdminErr)

			executeDeployCA(env, operatorOwner, tlsRootSpec())

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

// TestDeployCertificateAuthorityTlsInt: operator worker, both signing and TLS
// certificates issued by the (local) TLS Root via direct child workflow calls.
func TestDeployCertificateAuthorityTlsInt(t *testing.T) {
	tests := []struct {
		name                 string
		createCertificateErr error // CreateCertificate child (signing + tls, both local)
		expectedErr          error
	}{
		{
			name:        "success - TLS Intermediate CA deployed",
			expectedErr: nil,
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

			registerChildWorkflowStubs(env)
			registerSecretsActivityStubs(env)

			registerSecretsNexusMock(t, env, nil)

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

			executeDeployCA(env, operatorOwner, tlsIntSpec())

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

// TestDeployCertificateAuthorityOrgRoot: shard worker, self-signed signing
// certificate, TLS certificate from the (foreign) shared TLS Intermediate via Nexus.
func TestDeployCertificateAuthorityOrgRoot(t *testing.T) {
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

			executeDeployCA(env, orgOwner, orgRootSpec())

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

// TestDeployCertificateAuthorityOrgInt: shard worker, signing certificate from
// the (local) Org Root via child workflow, TLS certificate via Nexus.
func TestDeployCertificateAuthorityOrgInt(t *testing.T) {
	tests := []struct {
		name           string
		signingCertErr error // CreateCertificate child (signing cert from Org Root, local)
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

			executeDeployCA(env, orgOwner, orgIntSpec())

			assertWorkflowResult(t, env.GetWorkflowError(), tt.expectedErr)
		})
	}
}

// TestDeployCertificateAuthorityOwnershipRejection verifies the workflow fails
// fast (before any side effects) when the spec violates ownership rules.
func TestDeployCertificateAuthorityOwnershipRejection(t *testing.T) {
	tests := []struct {
		name  string
		owner workerconfig.OwnerConfig
		spec  *pb.CertificateAuthoritySpec
	}{
		{name: "shard worker given tls/root spec", owner: orgOwner, spec: tlsRootSpec()},
		{name: "shard worker given tls/int spec", owner: orgOwner, spec: tlsIntSpec()},
		{name: "operator worker given org/root spec", owner: operatorOwner, spec: orgRootSpec()},
		{name: "operator worker given org/int spec", owner: operatorOwner, spec: orgIntSpec()},
		{name: "signing issuer crosses ownership boundary", owner: orgOwner, spec: caSpec(orgIntRef, tlsRootRef, tlsIntRef)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			// No mocks: the workflow must fail during validation, before
			// touching any activity, child workflow, or Nexus service.
			executeDeployCA(env, tt.owner, tt.spec)

			err := env.GetWorkflowError()
			if err == nil {
				t.Fatal("expected ownership validation error, got nil")
			}
			if !strings.Contains(err.Error(), ErrNotOwned.Error()) {
				t.Fatalf("expected error containing %q, got: %v", ErrNotOwned.Error(), err)
			}
		})
	}
}
