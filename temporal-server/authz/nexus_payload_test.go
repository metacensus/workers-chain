package authz

import (
	"strings"
	"testing"

	commonpb "go.temporal.io/api/common/v1"
	nexuspb "go.temporal.io/api/nexus/v1"
	"go.temporal.io/server/api/matchingservice/v1"
	"google.golang.org/protobuf/encoding/protowire"
)

func signCertificateDispatch(t *testing.T, encoding, service, operation, org string) *matchingservice.DispatchNexusTaskRequest {
	t.Helper()

	var data []byte
	switch encoding {
	case encodingProtoJSON:
		if org == "" {
			data = []byte(`{}`)
		} else {
			data = []byte(`{"Requester":{"OrgId":"` + org + `"}}`)
		}
	case encodingProto:
		data = marshalSignCertificateRequester(org)
	default:
		data = []byte(`{}`)
	}

	return &matchingservice.DispatchNexusTaskRequest{
		Request: &nexuspb.Request{
			Variant: &nexuspb.Request_StartOperation{
				StartOperation: &nexuspb.StartOperationRequest{
					Service:   service,
					Operation: operation,
					Payload: &commonpb.Payload{
						Metadata: map[string][]byte{"encoding": []byte(encoding)},
						Data:     data,
					},
				},
			},
		},
	}
}

func marshalSignCertificateRequester(org string) []byte {
	if org == "" {
		return nil
	}
	owner := protowire.AppendTag(nil, ownerOrgIDField, protowire.BytesType)
	owner = protowire.AppendString(owner, org)
	out := protowire.AppendTag(nil, signCertificateRequesterField, protowire.BytesType)
	return protowire.AppendBytes(out, owner)
}

// The frontend Authorizer sees the endpoint name and the payload but not the
// caller (in-workflow dispatch arrives as the internode identity), so its job is
// to bind the per-org endpoint to the requester the payload declares.
func TestCheckCertificateIssuance(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		endpoint  string
		request   any
		wantErr   string
		wantAllow bool
	}{
		{
			name:      "requester matches the endpoint's org",
			endpoint:  "fabric-shard-acme-CreateCertificate",
			request:   signCertificateDispatch(t, encodingProtoJSON, certificateAuthorityService, signCertificateOperation, "acme"),
			wantAllow: true,
		},
		{
			name:      "binary payloads decode too",
			endpoint:  "fabric-shard-acme-CreateCertificate",
			request:   signCertificateDispatch(t, encodingProto, certificateAuthorityService, signCertificateOperation, "acme"),
			wantAllow: true,
		},
		{
			name:     "requester claims another org",
			endpoint: "fabric-shard-acme-CreateCertificate",
			request:  signCertificateDispatch(t, encodingProtoJSON, certificateAuthorityService, signCertificateOperation, "other"),
			wantErr:  "declares requester",
		},
		{
			name:     "requester is absent",
			endpoint: "fabric-shard-acme-CreateCertificate",
			request:  signCertificateDispatch(t, encodingProtoJSON, certificateAuthorityService, signCertificateOperation, ""),
			wantErr:  "declares requester",
		},
		{
			// The fabric worker registers whole proto services, so the endpoint
			// alone does not narrow the operation; this does.
			name:     "another operation on a certificate endpoint",
			endpoint: "fabric-shard-acme-CreateCertificate",
			request:  signCertificateDispatch(t, encodingProtoJSON, certificateAuthorityService, "CreateIdentity", "acme"),
			wantErr:  "serves only",
		},
		{
			name:     "undecodable payload",
			endpoint: "fabric-shard-acme-CreateCertificate",
			request:  signCertificateDispatch(t, "application/json", certificateAuthorityService, signCertificateOperation, ""),
			wantErr:  "cannot decode payload",
		},
		{
			name:     "unexpected request type on a certificate endpoint",
			endpoint: "fabric-shard-acme-CreateCertificate",
			request:  "not a dispatch",
			wantErr:  "not a nexus task dispatch",
		},
		{
			name:      "other endpoints are not inspected",
			endpoint:  EndpointSecretsCreateAccess,
			request:   "not a dispatch",
			wantAllow: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := checkCertificateIssuance(tc.endpoint, tc.request)
			if tc.wantAllow {
				if err != nil {
					t.Fatalf("checkCertificateIssuance() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("checkCertificateIssuance() = nil, want error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("checkCertificateIssuance() error = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// Cancelation and other variants carry no subject to bind.
func TestCheckCertificateIssuanceNonStartVariant(t *testing.T) {
	t.Parallel()
	request := &matchingservice.DispatchNexusTaskRequest{
		Request: &nexuspb.Request{
			Variant: &nexuspb.Request_CancelOperation{
				CancelOperation: &nexuspb.CancelOperationRequest{},
			},
		},
	}
	if err := checkCertificateIssuance("fabric-shard-acme-CreateCertificate", request); err != nil {
		t.Fatalf("checkCertificateIssuance() error = %v, want nil", err)
	}
}
