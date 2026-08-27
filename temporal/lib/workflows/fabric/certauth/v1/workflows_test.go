package certauth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	"github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1/certauthv1nexus"
	fabricpb "github.com/metacensus/infra/temporal/gen/workflows/fabric/v1"
	kubepb "github.com/metacensus/infra/temporal/gen/workflows/kube/v1"
	secretspb "github.com/metacensus/infra/temporal/gen/workflows/secrets/v1"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
	fabricconfig "github.com/metacensus/infra/temporal/lib/workers/fabric/config"
	secrets "github.com/metacensus/infra/temporal/lib/workflows/secrets/v1"
	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const testOrg = "testorg"
const testOperatorID = "metacensus"

// Worker identities under test. Certauth workflows act AS their worker's
// owner: the operator owns "tls" CAs (fabric worker), an org owns "org" CAs
// (shard worker).
var (
	operatorOwner = workerconfig.NewOperatorConfig(testOperatorID)
	orgOwner      = workerconfig.NewOrgConfig(testOrg)
)

// CA refs used across tests
var (
	tlsRootRef = &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "root"}
	tlsIntRef  = &fabricpb.CertificateAuthorityRef{CertType: "tls", Id: "int"}
	orgRootRef = &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "root"}
	orgIntRef  = &fabricpb.CertificateAuthorityRef{CertType: "org", Id: "int"}
)

// =============================================================================
// Constructor / Validation Tests
// =============================================================================

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

func TestNewCreateBootstrapAdminWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		owner     workerconfig.OwnerConfig
		input     *pb.CreateBootstrapAdminWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "success - tls CA on operator worker",
			owner: operatorOwner,
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				CaRef: tlsRootRef,
			}},
			wantErr: false,
		},
		{
			name:  "success - org CA on shard worker",
			owner: orgOwner,
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				CaRef: orgIntRef,
			}},
			wantErr: false,
		},
		{
			name:  "error - nil ca_ref",
			owner: operatorOwner,
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
		{
			name:  "error - shard worker given tls CA (not owned)",
			owner: orgOwner,
			input: &pb.CreateBootstrapAdminWorkflowInput{Req: &pb.CreateBootstrapAdminInput{
				CaRef: tlsRootRef,
			}},
			wantErr:   true,
			errSubstr: "not owned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateBootstrapAdminWorkflow(tt.input, tt.owner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

func TestNewCreateAdminWorkflow(t *testing.T) {
	tests := []struct {
		name      string
		owner     workerconfig.OwnerConfig
		input     *pb.CreateAdminWorkflowInput
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "success - tls CA on operator worker",
			owner: operatorOwner,
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef: tlsRootRef,
			}},
			wantErr: false,
		},
		{
			name:  "success - org CA on shard worker",
			owner: orgOwner,
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef: orgIntRef,
			}},
			wantErr: false,
		},
		{
			name:  "error - nil ca_ref",
			owner: operatorOwner,
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef: nil,
			}},
			wantErr:   true,
			errSubstr: "ca_ref is required",
		},
		{
			name:  "error - operator worker given org CA (not owned)",
			owner: operatorOwner,
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef: orgIntRef,
			}},
			wantErr:   true,
			errSubstr: "not owned",
		},
		{
			name:  "error - foreign signing issuer (crosses ownership boundary)",
			owner: orgOwner,
			input: &pb.CreateAdminWorkflowInput{Req: &pb.CreateAdminInput{
				CaRef:         orgIntRef,
				SigningIssuer: tlsRootRef,
			}},
			wantErr:   true,
			errSubstr: "signing_issuer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf, err := newCreateAdminWorkflow(tt.input, tt.owner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

func TestNewCreateIdentityWorkflow(t *testing.T) {
	validInput := func() *pb.CreateIdentityInput {
		return &pb.CreateIdentityInput{
			CommonName:   "test-identity@testorg",
			OrgIssuerRef: orgIntRef,
			TlsIssuerRef: tlsIntRef,
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
		owner     workerconfig.OwnerConfig
		mutate    func(*pb.CreateIdentityInput)
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "success - all required fields",
			owner:   orgOwner,
			mutate:  func(in *pb.CreateIdentityInput) {},
			wantErr: false,
		},
		{
			name:      "error - missing common name",
			owner:     orgOwner,
			mutate:    func(in *pb.CreateIdentityInput) { in.CommonName = "" },
			wantErr:   true,
			errSubstr: "common_name is required",
		},
		{
			name:      "error - missing org_issuer_ref",
			owner:     orgOwner,
			mutate:    func(in *pb.CreateIdentityInput) { in.OrgIssuerRef = nil },
			wantErr:   true,
			errSubstr: "org_issuer_ref is required",
		},
		{
			name:      "error - missing tls_issuer_ref",
			owner:     orgOwner,
			mutate:    func(in *pb.CreateIdentityInput) { in.TlsIssuerRef = nil },
			wantErr:   true,
			errSubstr: "tls_issuer_ref is required",
		},
		{
			name:      "error - missing signing_storage",
			owner:     orgOwner,
			mutate:    func(in *pb.CreateIdentityInput) { in.SigningStorage = nil },
			wantErr:   true,
			errSubstr: "signing_storage is required",
		},
		{
			name:      "error - missing tls_storage",
			owner:     orgOwner,
			mutate:    func(in *pb.CreateIdentityInput) { in.TlsStorage = nil },
			wantErr:   true,
			errSubstr: "tls_storage is required",
		},
		{
			name:      "error - foreign org issuer (not owned)",
			owner:     orgOwner,
			mutate:    func(in *pb.CreateIdentityInput) { in.OrgIssuerRef = tlsIntRef },
			wantErr:   true,
			errSubstr: "not owned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validInput()
			tt.mutate(input)
			wf, err := newCreateIdentityWorkflow(&pb.CreateIdentityWorkflowInput{Req: input}, tt.owner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestTLSEnrollmentCommonName(t *testing.T) {
	// operator owns tls/* CAs; an org owns org/* CAs. So a tls issuer is foreign
	// (shared) to an org owner, and local to the operator owner.
	tests := []struct {
		name             string
		owner            workerconfig.OwnerConfig
		baseName         string
		signingIssuerRef *fabricpb.CertificateAuthorityRef
		tlsIssuerRef     *fabricpb.CertificateAuthorityRef
		expected         string
	}{
		{
			name:         "self-signed (tls issuer nil) - returns base name",
			owner:        operatorOwner,
			baseName:     "certificate-authority-tls-root",
			tlsIssuerRef: nil,
			expected:     "certificate-authority-tls-root",
		},
		{
			name:             "local same CA issues both - '-tls' suffix",
			owner:            operatorOwner,
			baseName:         "certificate-authority-tls-int",
			signingIssuerRef: tlsRootRef,
			tlsIssuerRef:     tlsRootRef,
			expected:         "certificate-authority-tls-int-tls",
		},
		{
			name:             "local different CAs - no suffix",
			owner:            operatorOwner,
			baseName:         "certificate-authority-tls-int",
			signingIssuerRef: tlsIntRef,
			tlsIssuerRef:     tlsRootRef,
			expected:         "certificate-authority-tls-int",
		},
		{
			name:             "foreign CA - org int getting TLS from operator - owner-qualified",
			owner:            orgOwner,
			baseName:         "certificate-authority-org-int",
			signingIssuerRef: orgRootRef,
			tlsIssuerRef:     tlsIntRef,
			expected:         "certificate-authority-org-int@testorg",
		},
		{
			name:             "foreign CA - org root (self-signed signing) - owner-qualified",
			owner:            orgOwner,
			baseName:         "certificate-authority-org-root",
			signingIssuerRef: nil,
			tlsIssuerRef:     tlsIntRef,
			expected:         "certificate-authority-org-root@testorg",
		},
		{
			name:             "foreign admin TLS cert - double-@ owner-qualified",
			owner:            orgOwner,
			baseName:         "admin@certificate-authority-org-int",
			signingIssuerRef: orgIntRef,
			tlsIssuerRef:     tlsIntRef,
			expected:         "admin@certificate-authority-org-int@testorg",
		},
		{
			name:             "foreign bootstrap-admin TLS cert - double-@ owner-qualified",
			owner:            orgOwner,
			baseName:         "bootstrap-admin@certificate-authority-org-int",
			signingIssuerRef: orgIntRef,
			tlsIssuerRef:     tlsIntRef,
			expected:         "bootstrap-admin@certificate-authority-org-int@testorg",
		},
		{
			name:             "local admin same CA issues both - '-tls' suffix, unqualified",
			owner:            operatorOwner,
			baseName:         "admin@certificate-authority-tls-root",
			signingIssuerRef: tlsRootRef,
			tlsIssuerRef:     tlsRootRef,
			expected:         "admin@certificate-authority-tls-root-tls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TLSEnrollmentCommonName(tt.owner, tt.baseName, tt.signingIssuerRef, tt.tlsIssuerRef)
			if result != tt.expected {
				t.Errorf("TLSEnrollmentCommonName(%s, %q, %v, %v) = %q, want %q",
					tt.owner.OwnerID(), tt.baseName, tt.signingIssuerRef, tt.tlsIssuerRef, result, tt.expected)
			}
		})
	}
}

// TestTLSEnrollmentCommonName_DistinctOwners is the #4 regression: two orgs whose
// CAs share the operator TLS CA must present distinct enrollment CNs, or their
// registry records clobber each other.
func TestTLSEnrollmentCommonName_DistinctOwners(t *testing.T) {
	orgA := workerconfig.NewOrgConfig("org-a")
	orgB := workerconfig.NewOrgConfig("org-b")
	base := "certificate-authority-org-int"

	cnA := TLSEnrollmentCommonName(orgA, base, orgRootRef, tlsIntRef)
	cnB := TLSEnrollmentCommonName(orgB, base, orgRootRef, tlsIntRef)

	if cnA == cnB {
		t.Fatalf("distinct owners produced identical foreign-TLS CN %q (collision not fixed)", cnA)
	}
	if cnA != base+"@org-a" || cnB != base+"@org-b" {
		t.Errorf("unexpected owner-qualified CNs: orgA=%q orgB=%q", cnA, cnB)
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

// TestRegisterSignCertificateNexusService is a regression test for the fx.go
// miswiring where the Fabric Nexus service was registered instead of the
// CertificateAuthority service. It asserts the registered service is
// fabric.certauth.v1.CertificateAuthority and that it handles only SignCertificate.
func TestRegisterSignCertificateNexusService(t *testing.T) {
	registry := &fakeNexusRegistry{}
	if err := registerSignCertificateNexusService(registry); err != nil {
		t.Fatalf("registerSignCertificateNexusService failed: %v", err)
	}
	if len(registry.services) != 1 {
		t.Fatalf("expected 1 registered service, got %d", len(registry.services))
	}
	svc := registry.services[0]
	if svc.Name != certauthv1nexus.CertificateAuthorityServiceName {
		t.Errorf("service name = %q, want %q", svc.Name, certauthv1nexus.CertificateAuthorityServiceName)
	}
	if op := svc.Operation(certauthv1nexus.CertificateAuthoritySignCertificateOperationName); op == nil {
		t.Errorf("service does not handle operation %q", certauthv1nexus.CertificateAuthoritySignCertificateOperationName)
	}
	if op := svc.Operation(certauthv1nexus.CertificateAuthorityCreateIdentityOperationName); op != nil {
		t.Errorf("service unexpectedly handles local-only operation CreateIdentity")
	}
}

// =============================================================================
// Workflow Execution Tests
// =============================================================================
// These tests verify workflow orchestration with mocked activities.
// All inputs are valid - validation is tested separately above.

// registerCertAuthNexusMock registers the CertificateAuthority Nexus service and
// mocks the SignCertificate operation used for cross-owner TLS issuance (a public
// CSR goes out, a public cert + chain come back).
func registerCertAuthNexusMock(t *testing.T, env *testsuite.TestWorkflowEnvironment, opErr error) {
	t.Helper()
	nexusSvc, err := certauthv1nexus.NewCertificateAuthorityNexusService(&certauthv1nexus.UnimplementedCertificateAuthorityNexusHandler{})
	if err != nil {
		t.Fatalf("failed to create certauth nexus service: %v", err)
	}
	env.RegisterNexusService(nexusSvc)
	if opErr != nil {
		env.OnNexusOperation(nexusSvc, certauthv1nexus.CertificateAuthoritySignCertificateOperationName, mock.Anything, mock.Anything).
			Return(nil, opErr)
	} else {
		env.OnNexusOperation(nexusSvc, certauthv1nexus.CertificateAuthoritySignCertificateOperationName, mock.Anything, mock.Anything).
			Return(&nexus.HandlerStartOperationResultSync[*pb.SignCertificateOutput]{
				Value: &pb.SignCertificateOutput{
					CertificatePem: []byte("test-cert"),
					ChainPem:       []byte("test-chain"),
				},
			}, nil)
	}
}

// Named activity stubs for the unified CSR pipeline. These must be top-level
// named functions (not anonymous closures): the Temporal test environment
// aliases anonymous closures to "func1"/"func2"/... which collides with the
// generated workflow ID-expression local activities, causing "reflect: Call
// with too few input arguments" panics.

func stubEnsureKeyPairActivity(_ context.Context, in *secretspb.EnsureKeyPairInput) (*secretspb.EnsureKeyPairOutput, error) {
	return &secretspb.EnsureKeyPairOutput{SecretRef: in.SecretRef}, nil // pragma: allowlist secret
}

func stubGenerateCsrActivity(_ context.Context, _ *pb.GenerateCsrInput) (*pb.GenerateCsrOutput, error) {
	return &pb.GenerateCsrOutput{CsrPem: []byte("test-csr")}, nil
}

func stubSignCsrActivity(_ context.Context, _ *pb.SignCsrInput) (*pb.SignCsrOutput, error) {
	return &pb.SignCsrOutput{CertificatePem: []byte("test-cert"), ChainPem: []byte("test-chain")}, nil
}

func stubBootstrapEnrollCsrActivity(_ context.Context, _ *pb.BootstrapEnrollCsrInput) (*pb.BootstrapEnrollCsrOutput, error) {
	return &pb.BootstrapEnrollCsrOutput{CertificatePem: []byte("test-cert"), ChainPem: []byte("test-chain")}, nil
}

func stubStoreCertificateActivity(_ context.Context, in *secretspb.StoreCertificateInput) (*secretspb.StoreCertificateOutput, error) {
	return &secretspb.StoreCertificateOutput{CertRef: in.CertRef}, nil
}

// registerCSRPipelineStubs registers success-returning stubs for every activity
// in the unified CSR pipeline (EnsureKeyPair -> GenerateCsr -> SignCsr or
// BootstrapEnrollCsr -> StoreCertificate). Tests override individual activities
// with env.OnActivity(...) to inject errors.
func registerCSRPipelineStubs(t *testing.T, env *testsuite.TestWorkflowEnvironment) {
	t.Helper()
	env.RegisterActivityWithOptions(stubEnsureKeyPairActivity,
		activity.RegisterOptions{Name: secretspb.EnsureKeyPairActivityName})
	env.RegisterActivityWithOptions(stubGenerateCsrActivity,
		activity.RegisterOptions{Name: pb.GenerateCsrActivityName})
	env.RegisterActivityWithOptions(stubSignCsrActivity,
		activity.RegisterOptions{Name: pb.SignCsrActivityName})
	env.RegisterActivityWithOptions(stubBootstrapEnrollCsrActivity,
		activity.RegisterOptions{Name: pb.BootstrapEnrollCsrActivityName})
	env.RegisterActivityWithOptions(stubStoreCertificateActivity,
		activity.RegisterOptions{Name: secretspb.StoreCertificateActivityName})
}

func TestCreateBootstrapAdminWorkflow(t *testing.T) {
	// Note: Validation tests for CreateBootstrapAdmin are in TestNewCreateBootstrapAdminWorkflow above.
	// This test covers execution/orchestration with mocked activities.
	//
	// Both the signing identity and a local TLS identity are obtained via the
	// register-less bootstrap CSR pipeline (EnsureKeyPair -> GenerateCsr ->
	// BootstrapEnrollCsr -> StoreCertificate). A foreign TLS issuer (org CA
	// getting TLS from the operator's chain) is reached through the SignCsr-less
	// Nexus SignCertificate path instead.

	tests := []struct {
		name             string
		owner            workerconfig.OwnerConfig
		caRef            *fabricpb.CertificateAuthorityRef
		tlsIssuer        *fabricpb.CertificateAuthorityRef
		getStatusErr     error
		enrollSigningErr error // BootstrapEnrollCsr, "" profile (signing)
		enrollTLSErr     error // BootstrapEnrollCsr, "tls" profile (local TLS)
		nexusTLSErr      error // Nexus SignCertificate (foreign TLS)
		expectedErr      error
		validateOutput   bool
	}{
		{
			name:           "success - bootstrap admin for tls root CA (local TLS enrollment)",
			owner:          operatorOwner,
			caRef:          tlsRootRef,
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "success - bootstrap admin for org int CA (foreign TLS via Nexus)",
			owner:          orgOwner,
			caRef:          orgIntRef,
			tlsIssuer:      tlsIntRef,
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:         "error - CA never ready",
			owner:        operatorOwner,
			caRef:        tlsRootRef,
			getStatusErr: errors.New("pod not ready"),
			expectedErr:  errors.New("CA did not become ready"),
		},
		{
			name:             "error - signing enrollment fails",
			owner:            operatorOwner,
			caRef:            tlsRootRef,
			enrollSigningErr: errors.New("CA enrollment rejected"),
			expectedErr:      errors.New("failed to bootstrap enroll signing identity"),
		},
		{
			name:         "error - TLS enrollment fails (local issuer)",
			owner:        operatorOwner,
			caRef:        tlsRootRef,
			enrollTLSErr: errors.New("TLS CA rejected enrollment"),
			expectedErr:  errors.New("failed to bootstrap enroll TLS identity"),
		},
		{
			name:        "error - foreign TLS issuance via Nexus fails",
			owner:       orgOwner,
			caRef:       orgIntRef,
			tlsIssuer:   tlsIntRef,
			nexusTLSErr: errors.New("TLS Int CA unavailable"),
			expectedErr: errors.New("failed to create TLS identity via CSR flow"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(tt.owner)

			// Foreign TLS issuance goes via the CertificateAuthority Nexus service.
			registerCertAuthNexusMock(t, env, tt.nexusTLSErr)
			// The register-less bootstrap CSR pipeline + shared key/csr/store stubs.
			registerCSRPipelineStubs(t, env)

			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *kubepb.GetStatusInput) (*kubepb.GetStatusOutput, error) {
					return &kubepb.GetStatusOutput{}, nil
				},
				activity.RegisterOptions{Name: kubepb.GetStatusActivityName},
			)

			// Mock GetStatus activity
			if tt.getStatusErr != nil {
				env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).
					Return(nil, tt.getStatusErr)
			} else {
				env.OnActivity(kubepb.GetStatusActivityName, mock.Anything, mock.Anything).
					Return(&kubepb.GetStatusOutput{}, nil)
			}

			// Capture the CSR subject CN per key ref so we can assert the
			// register-less bootstrap paths request a CN equal to the enrollment
			// ID (the bootstrap username). A CN != enrollment ID is Fabric CA
			// error 77 - the regression this guards against.
			var cnMu sync.Mutex
			csrCNByKeyRef := map[string]string{}
			env.OnActivity(pb.GenerateCsrActivityName, mock.Anything, mock.Anything).
				Return(func(ctx context.Context, input *pb.GenerateCsrInput) (*pb.GenerateCsrOutput, error) {
					cnMu.Lock()
					csrCNByKeyRef[input.GetKeyRef().String()] = input.GetCommonName()
					cnMu.Unlock()
					return &pb.GenerateCsrOutput{CsrPem: []byte("test-csr")}, nil
				})

			// Mock BootstrapEnrollCsr - distinguish signing vs TLS by profile.
			// Signing uses "" (empty) profile, TLS uses "tls" profile.
			env.OnActivity(pb.BootstrapEnrollCsrActivityName, mock.Anything, mock.Anything).
				Return(func(ctx context.Context, input *pb.BootstrapEnrollCsrInput) (*pb.BootstrapEnrollCsrOutput, error) {
					if input.GetProfile() == "tls" {
						if tt.enrollTLSErr != nil {
							return nil, tt.enrollTLSErr
						}
					} else {
						if tt.enrollSigningErr != nil {
							return nil, tt.enrollSigningErr
						}
					}
					return &pb.BootstrapEnrollCsrOutput{CertificatePem: []byte("test-cert"), ChainPem: []byte("test-chain")}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateBootstrapAdminInput) (*pb.CreateBootstrapAdminOutput, error) {
				wf, err := workflows.CreateBootstrapAdmin(ctx, &pb.CreateBootstrapAdminWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateBootstrapAdminInput{
				CaRef:     tt.caRef,
				TlsIssuer: tt.tlsIssuer,
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

					// The signing identity is always obtained via register-less
					// bootstrap enrollment, so its CSR CN must equal the CA's
					// bootstrap enrollment ID (Fabric CA error 77 otherwise).
					caConfig := fabricconfig.NewCertificateAuthorityConfig(tt.owner, tt.caRef)
					cnMu.Lock()
					gotSigningCN := csrCNByKeyRef[caConfig.VaultBootstrapAdminSigningKeyRef().String()]
					gotTLSCN := csrCNByKeyRef[caConfig.VaultBootstrapAdminTLSKeyRef().String()]
					cnMu.Unlock()
					if gotSigningCN != caConfig.BootstrapAdminUsername() {
						t.Errorf("signing bootstrap CSR CN = %q, want %q (must equal enrollment ID)",
							gotSigningCN, caConfig.BootstrapAdminUsername())
					}
					// A nil tls_issuer means the local (register-less) TLS branch,
					// whose CN must likewise equal the enrollment ID. A foreign
					// tls_issuer goes through the SignCsr/Nexus path, which
					// registers its own per-identity CN, so it is exempt.
					if tt.tlsIssuer == nil && gotTLSCN != caConfig.BootstrapAdminUsername() {
						t.Errorf("local TLS bootstrap CSR CN = %q, want %q (must equal enrollment ID)",
							gotTLSCN, caConfig.BootstrapAdminUsername())
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
	// This test covers execution/orchestration with mocked activities and Nexus operations.
	//
	// CreateAdmin issues both certs through the unified CSR pipeline. The signing
	// cert is always signed locally via the SignCsr activity (authenticating as
	// the CA's bootstrap admin). The TLS cert is signed locally via SignCsr when
	// the TLS issuer is owned by this worker, or via the Nexus SignCertificate
	// path when it is foreign (org CA admin getting TLS from the operator chain).
	// Signing vs TLS SignCsr calls are distinguished by profile ("" vs "tls").

	tests := []struct {
		name           string
		owner          workerconfig.OwnerConfig
		caRef          *fabricpb.CertificateAuthorityRef
		tlsIssuer      *fabricpb.CertificateAuthorityRef
		signingCertErr error // SignCsr activity, "" profile (signing)
		tlsSignErr     error // SignCsr activity, "tls" profile (local TLS)
		nexusTLSErr    error // Nexus SignCertificate (foreign TLS)
		expectedErr    error
		validateOutput bool
	}{
		{
			name:           "success - admin created for tls root CA (local TLS)",
			owner:          operatorOwner,
			caRef:          tlsRootRef,
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "success - admin created for org intermediate CA (TLS via Nexus)",
			owner:          orgOwner,
			caRef:          orgIntRef,
			tlsIssuer:      tlsIntRef,
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "error - signing identity creation fails",
			owner:          operatorOwner,
			caRef:          tlsIntRef,
			signingCertErr: errors.New("CA enrollment failed"),
			expectedErr:    errors.New("failed to create admin signing identity"),
		},
		{
			name:        "error - TLS identity creation fails (local)",
			owner:       operatorOwner,
			caRef:       tlsIntRef,
			tlsSignErr:  errors.New("TLS CA unavailable"),
			expectedErr: errors.New("failed to create admin TLS identity"),
		},
		{
			name:        "error - TLS identity creation fails (Nexus)",
			owner:       orgOwner,
			caRef:       orgIntRef,
			tlsIssuer:   tlsIntRef,
			nexusTLSErr: errors.New("TLS CA unavailable"),
			expectedErr: errors.New("failed to create admin TLS identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(tt.owner)

			// Foreign TLS issuance goes via the CertificateAuthority Nexus service.
			registerCertAuthNexusMock(t, env, tt.nexusTLSErr)
			registerCSRPipelineStubs(t, env)

			// Capture the privileged signing-path SignCsr request so we can
			// assert the registrar grant is actually wired through (a silent
			// mis-wire here yields a powerless org admin).
			var gotSigningInput *pb.SignCsrInput

			// Mock SignCsr - distinguish signing ("" profile) from local TLS
			// ("tls" profile).
			env.OnActivity(pb.SignCsrActivityName, mock.Anything, mock.Anything).
				Return(func(ctx context.Context, input *pb.SignCsrInput) (*pb.SignCsrOutput, error) {
					if input.GetProfile() == "tls" {
						if tt.tlsSignErr != nil {
							return nil, tt.tlsSignErr
						}
					} else {
						gotSigningInput = input
						if tt.signingCertErr != nil {
							return nil, tt.signingCertErr
						}
					}
					return &pb.SignCsrOutput{CertificatePem: []byte("test-cert"), ChainPem: []byte("test-chain")}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateAdminInput) (*pb.CreateAdminOutput, error) {
				wf, err := workflows.CreateAdmin(ctx, &pb.CreateAdminWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateAdminInput{
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

					// Assert the privileged signing-path CSR is wired correctly:
					// the identity registered must be the CA's admin CN, it must
					// carry the full registrar grant, and it must be registered
					// by the CA's own bootstrap admin (the bootstrap-vs-permanent
					// selection). A stub that ignored these would silently mint a
					// powerless admin.
					caConfig := fabricconfig.NewCertificateAuthorityConfig(tt.owner, tt.caRef)
					si := gotSigningInput
					if si == nil {
						t.Fatal("signing-path SignCsr was never invoked")
					}
					if got, want := si.GetIdentityName(), caConfig.AdminCommonName(); got != want {
						t.Errorf("signing identity name = %q, want %q", got, want)
					}
					if got := si.GetAttributes()["hf.Registrar.Roles"]; got != "*" {
						t.Errorf("hf.Registrar.Roles = %q, want %q", got, "*")
					}
					// Signing issuer defaults to the CA itself, so the registrar
					// is the CA's bootstrap admin.
					wantKeyRef := caConfig.VaultBootstrapAdminSigningKeyRef().String()
					if got := si.GetAdminIdentity().GetKeyRef().String(); got != wantKeyRef {
						t.Errorf("signing registrar key ref = %q, want bootstrap admin %q", got, wantKeyRef)
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
	// This test covers execution/orchestration with mocked activities and Nexus operations.
	//
	// CreateIdentity runs on a shard worker: the signing cert comes from the org's
	// own CA via the local SignCsr activity, and the TLS cert via the
	// CertificateAuthority Nexus SignCertificate path (shared TLS CAs live in the
	// operator namespace, so the TLS issuer is foreign to the org worker).

	tests := []struct {
		name           string
		signingCertErr error // SignCsr activity (local signing)
		tlsCertErr     error // Nexus SignCertificate call (foreign TLS)
		expectedErr    error
		validateOutput bool
	}{
		{
			name:           "success - identity created with signing and TLS certs",
			expectedErr:    nil,
			validateOutput: true,
		},
		{
			name:           "error - signing certificate creation fails",
			signingCertErr: errors.New("CA connection refused"),
			expectedErr:    errors.New("failed to create signing identity"),
		},
		{
			name:        "error - TLS certificate creation fails",
			tlsCertErr:  errors.New("TLS enrollment failed"),
			expectedErr: errors.New("failed to create TLS identity"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(orgOwner)

			registerCertAuthNexusMock(t, env, tt.tlsCertErr)
			registerCSRPipelineStubs(t, env)

			// Mock the local signing SignCsr activity (org issuer is locally owned).
			env.OnActivity(pb.SignCsrActivityName, mock.Anything, mock.Anything).
				Return(func(ctx context.Context, input *pb.SignCsrInput) (*pb.SignCsrOutput, error) {
					if tt.signingCertErr != nil {
						return nil, tt.signingCertErr
					}
					return &pb.SignCsrOutput{CertificatePem: []byte("test-cert"), ChainPem: []byte("test-chain")}, nil
				})

			// Execute workflow
			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.CreateIdentityInput) (*pb.CreateIdentityOutput, error) {
				wf, err := workflows.CreateIdentity(ctx, &pb.CreateIdentityWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.CreateIdentityInput{
				CommonName:   "test-peer@testorg",
				OrgIssuerRef: orgIntRef,
				TlsIssuerRef: tlsIntRef,
				SigningStorage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/signing"),
					CertRef: secrets.NewVaultRef("test/certs/signing"),
				},
				TlsStorage: &fabricpb.IdentityRef{
					KeyRef:  secrets.NewVaultRef("test/keys/tls"),
					CertRef: secrets.NewVaultRef("test/certs/tls"),
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
