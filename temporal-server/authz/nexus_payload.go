package authz

import (
	"encoding/json"
	"fmt"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/server/api/matchingservice/v1"
	"google.golang.org/protobuf/encoding/protowire"
)

// Wire names generated for the certauth Nexus service. Keep in sync with
// gen/fabric/certauth/v1/certauthv1nexus in this repository.
const (
	certificateAuthorityService = "fabric.certauth.v1.CertificateAuthority"
	signCertificateOperation    = "SignCertificate"
)

// Payload encodings produced by the Temporal SDK's default data converter.
const (
	encodingProtoJSON = "json/protobuf"
	encodingProto     = "binary/protobuf"
)

// Protobuf field numbers for the subset of SignCertificateInput / Owner we
// inspect. Keep in sync with proto/fabric/certauth/v1 and proto/fabric/v1 in
// this repository. Defined locally so the server binary does not import the
// worker gen package (protoc-gen-go-temporal conflicts with Temporal server's
// own MethodOptions extension number).
const (
	signCertificateRequesterField = 8 // SignCertificateInput.requester
	ownerOrgIDField               = 2 // Owner.org_id
)

// checkCertificateIssuance binds a per-org certificate endpoint to the request
// it carries: only SignCertificate may be invoked, and the requester it
// declares must be the org that owns the endpoint.
//
// Together with the caller→endpoint allowlist in the history service, this is
// what makes a foreign-org identity unrequestable: the caller cannot reach
// another org's endpoint, and cannot use its own endpoint to ask for another
// org's certificate. The handler then scopes CN/SAN to the declared requester.
//
// Endpoints other than per-org certificate issuance are not inspected.
func checkCertificateIssuance(endpoint string, request any) error {
	org, ok := OrgForCreateCertificateEndpoint(endpoint)
	if !ok {
		return nil
	}

	dispatch, ok := request.(*matchingservice.DispatchNexusTaskRequest)
	if !ok {
		return fmt.Errorf("endpoint %q: request is not a nexus task dispatch", endpoint)
	}

	// Cancelation and other variants carry no subject material to bind.
	start := dispatch.GetRequest().GetStartOperation()
	if start == nil {
		return nil
	}

	if start.GetService() != certificateAuthorityService || start.GetOperation() != signCertificateOperation {
		return fmt.Errorf("endpoint %q serves only %s/%s, got %s/%s",
			endpoint, certificateAuthorityService, signCertificateOperation,
			start.GetService(), start.GetOperation())
	}

	requester, err := decodeRequesterOrg(start.GetPayload())
	if err != nil {
		return fmt.Errorf("endpoint %q: %w", endpoint, err)
	}
	if requester != org {
		return fmt.Errorf("endpoint %q is owned by org %q, request declares requester %q",
			endpoint, org, requester)
	}
	return nil
}

// signCertificateJSON is the json/protobuf shape of SignCertificateInput for
// the fields we care about (json_name tags from the proto).
type signCertificateJSON struct {
	Requester *ownerJSON `json:"Requester"`
}

type ownerJSON struct {
	OrgID string `json:"OrgId"`
}

func decodeRequesterOrg(payload *commonpb.Payload) (string, error) {
	encoding := string(payload.GetMetadata()["encoding"])
	switch encoding {
	case encodingProtoJSON:
		var input signCertificateJSON
		if err := json.Unmarshal(payload.GetData(), &input); err != nil {
			return "", fmt.Errorf("cannot decode payload: %w", err)
		}
		if input.Requester == nil {
			return "", nil
		}
		return input.Requester.OrgID, nil
	case encodingProto:
		return requesterOrgFromProto(payload.GetData())
	default:
		return "", fmt.Errorf("cannot decode payload with encoding %q", encoding)
	}
}

func requesterOrgFromProto(data []byte) (string, error) {
	requesterBytes, ok := protoBytesField(data, signCertificateRequesterField)
	if !ok {
		return "", nil
	}
	org, ok := protoStringField(requesterBytes, ownerOrgIDField)
	if !ok {
		return "", nil
	}
	return org, nil
}

func protoBytesField(data []byte, fieldNumber protowire.Number) ([]byte, bool) {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return nil, false
		}
		data = data[n:]
		if num == fieldNumber && typ == protowire.BytesType {
			v, m := protowire.ConsumeBytes(data)
			if m < 0 {
				return nil, false
			}
			return v, true
		}
		m := protowire.ConsumeFieldValue(num, typ, data)
		if m < 0 {
			return nil, false
		}
		data = data[m:]
	}
	return nil, false
}

func protoStringField(data []byte, fieldNumber protowire.Number) (string, bool) {
	v, ok := protoBytesField(data, fieldNumber)
	if !ok {
		return "", false
	}
	return string(v), true
}
