package certauth

import (
	"context"
	"errors"
	"strings"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/certauth/v1"
	workerconfig "github.com/metacensus/infra/temporal/lib/workers/config"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

// TestNewSignCertificateWorkflow validates the ownership boundary: the issuer
// CA MUST be owned by the executing worker's owner. A caller cannot coerce a
// worker into signing with a CA it does not own.
func TestNewSignCertificateWorkflow(t *testing.T) {
	validReq := func() *pb.SignCertificateInput {
		return &pb.SignCertificateInput{
			IssuerRef:  tlsIntRef,
			CommonName: "peer0@" + testOrg,
			CsrPem:     []byte("-----BEGIN CERTIFICATE REQUEST-----\nx\n-----END CERTIFICATE REQUEST-----"),
			Requester:  orgOwner.OwnerProto(),
		}
	}

	tests := []struct {
		name      string
		owner     workerconfig.OwnerConfig
		mutate    func(*pb.SignCertificateInput)
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "valid - operator owns tls issuer",
			owner: operatorOwner,
		},
		{
			name:      "invalid - missing issuer_ref",
			owner:     operatorOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.IssuerRef = nil },
			wantErr:   true,
			errSubstr: "issuer_ref is required",
		},
		{
			name:      "invalid - missing common_name",
			owner:     operatorOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.CommonName = "" },
			wantErr:   true,
			errSubstr: "common_name is required",
		},
		{
			name:      "invalid - missing csr_pem",
			owner:     operatorOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.CsrPem = nil },
			wantErr:   true,
			errSubstr: "csr_pem is required",
		},
		{
			name:      "invalid - operator does not own org issuer",
			owner:     operatorOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.IssuerRef = orgIntRef },
			wantErr:   true,
			errSubstr: "does not own",
		},
		{
			name:      "invalid - shard does not own tls issuer",
			owner:     orgOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.IssuerRef = tlsIntRef },
			wantErr:   true,
			errSubstr: "does not own",
		},
		// Cross-owner issuance is for leaf certificates only. Intermediate CAs and
		// registrars are always minted against a locally-owned issuer, so these
		// requests can never be legitimate here.
		{
			name:      "privileged - ca profile",
			owner:     operatorOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.Profile = "ca" },
			wantErr:   true,
			errSubstr: ErrPrivilegedRequest.Error(),
		},
		{
			name:  "privileged - intermediate CA attribute",
			owner: operatorOwner,
			mutate: func(r *pb.SignCertificateInput) {
				r.Attributes = map[string]string{"hf.IntermediateCA": "true"}
			},
			wantErr:   true,
			errSubstr: ErrPrivilegedRequest.Error(),
		},
		{
			name:  "privileged - registrar attribute",
			owner: operatorOwner,
			mutate: func(r *pb.SignCertificateInput) {
				r.Attributes = map[string]string{"hf.Registrar.Roles": "*"}
			},
			wantErr:   true,
			errSubstr: ErrPrivilegedRequest.Error(),
		},
		{
			name:  "privileged - revoker attribute",
			owner: operatorOwner,
			mutate: func(r *pb.SignCertificateInput) {
				r.Attributes = map[string]string{"hf.Revoker": "true"}
			},
			wantErr:   true,
			errSubstr: ErrPrivilegedRequest.Error(),
		},
		{
			name:   "valid - tls profile with ordinary attributes",
			owner:  operatorOwner,
			mutate: func(r *pb.SignCertificateInput) { r.Profile = "tls"; r.Attributes = map[string]string{"role": "peer"} },
		},
		// The requester is bound to the per-org Nexus endpoint by the Temporal
		// server, so the issuer can scope the subject to it: a shard cannot
		// request an identity or a service name belonging to another org.
		{
			name:      "invalid - missing requester",
			owner:     operatorOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.Requester = nil },
			wantErr:   true,
			errSubstr: "requester.org_id is required",
		},
		{
			name:      "out of scope - common name qualified with another org",
			owner:     operatorOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.CommonName = "peer0@otherorg" },
			wantErr:   true,
			errSubstr: ErrSubjectOutOfScope.Error(),
		},
		{
			name:      "out of scope - common name not org-qualified",
			owner:     operatorOwner,
			mutate:    func(r *pb.SignCertificateInput) { r.CommonName = "peer0" },
			wantErr:   true,
			errSubstr: ErrSubjectOutOfScope.Error(),
		},
		{
			name:  "out of scope - host in another org's namespace",
			owner: operatorOwner,
			mutate: func(r *pb.SignCertificateInput) {
				r.Hosts = []string{"peer0.fabric-shard-otherorg.svc.cluster.local"}
			},
			wantErr:   true,
			errSubstr: ErrSubjectOutOfScope.Error(),
		},
		{
			name:  "out of scope - host in the operator namespace",
			owner: operatorOwner,
			mutate: func(r *pb.SignCertificateInput) {
				r.Hosts = []string{"certificate-authority-tls-int.fabric.svc.cluster.local"}
			},
			wantErr:   true,
			errSubstr: ErrSubjectOutOfScope.Error(),
		},
		{
			name:  "valid - hosts within the requester's namespace",
			owner: operatorOwner,
			mutate: func(r *pb.SignCertificateInput) {
				r.Hosts = []string{
					"peer0",
					"peer0.fabric-shard-" + testOrg,
					"peer0.fabric-shard-" + testOrg + ".svc",
					"peer0.fabric-shard-" + testOrg + ".svc.cluster.local",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validReq()
			if tt.mutate != nil {
				tt.mutate(req)
			}
			wf, err := newSignCertificateWorkflow(&pb.SignCertificateWorkflowInput{Req: req}, tt.owner)
			assertConstructorResult(t, wf == nil, err, tt.wantErr, tt.errSubstr)
		})
	}
}

// TestSignCertificateWorkflow covers execution/orchestration with a mocked
// SignCsr activity. The issuer is always owned by the executing worker; the
// workflow returns exactly the public cert + chain produced by SignCsr.
func TestSignCertificateWorkflow(t *testing.T) {
	wantCert := []byte("-----BEGIN CERTIFICATE-----\nleaf\n-----END CERTIFICATE-----")
	wantChain := []byte("-----BEGIN CERTIFICATE-----\nleaf\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nissuer\n-----END CERTIFICATE-----")

	tests := []struct {
		name        string
		signErr     error
		expectedErr string
	}{
		{name: "success - operator tls issuer signs CSR"},
		{name: "error - SignCsr activity fails", signErr: errors.New("CA rejected CSR"), expectedErr: "failed to sign CSR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()

			workflows := NewWorkflows(operatorOwner)

			env.RegisterActivityWithOptions(
				func(ctx context.Context, input *pb.SignCsrInput) (*pb.SignCsrOutput, error) {
					return &pb.SignCsrOutput{}, nil
				},
				activity.RegisterOptions{Name: pb.SignCsrActivityName},
			)
			if tt.signErr != nil {
				env.OnActivity(pb.SignCsrActivityName, mock.Anything, mock.Anything).Return(nil, tt.signErr)
			} else {
				env.OnActivity(pb.SignCsrActivityName, mock.Anything, mock.Anything).
					Return(&pb.SignCsrOutput{CertificatePem: wantCert, ChainPem: wantChain}, nil)
			}

			env.ExecuteWorkflow(func(ctx workflow.Context, input *pb.SignCertificateInput) (*pb.SignCertificateOutput, error) {
				wf, err := workflows.SignCertificate(ctx, &pb.SignCertificateWorkflowInput{Req: input})
				if err != nil {
					return nil, err
				}
				return wf.Execute(ctx)
			}, &pb.SignCertificateInput{
				IssuerRef:    tlsIntRef,
				CommonName:   "peer0@" + testOrg,
				IdentityType: "peer",
				CsrPem:       []byte("-----BEGIN CERTIFICATE REQUEST-----\nx\n-----END CERTIFICATE REQUEST-----"),
				Requester:    orgOwner.OwnerProto(),
			})

			var result *pb.SignCertificateOutput
			err := env.GetWorkflowResult(&result)

			if tt.expectedErr == "" {
				if err != nil {
					t.Fatalf("unexpected workflow error: %v", err)
				}
				if string(result.GetCertificatePem()) != string(wantCert) {
					t.Errorf("certificate = %q, want %q", result.GetCertificatePem(), wantCert)
				}
				if string(result.GetChainPem()) != string(wantChain) {
					t.Errorf("chain = %q, want %q", result.GetChainPem(), wantChain)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
				}
			}
		})
	}
}
