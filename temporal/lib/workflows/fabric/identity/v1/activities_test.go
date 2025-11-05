package identity

import (
	"errors"
	"testing"

	pb "github.com/metacensus/infra/temporal/gen/workflows/fabric/identity/v1"
	fabricmocks "github.com/metacensus/infra/temporal/mocks/workflows/fabric/identity/v1"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

var (
	// Test error sentinels for client errors
	errClientConnectionRefused = errors.New("connection refused")
	errClientUnauthorized      = errors.New("unauthorized")
)

func TestRegisterIdentityActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.RegisterIdentityInput
		clientRes *pb.RegisterIdentityOutput
		clientErr error
	}{
		{
			name: "success - peer identity registration",
			input: &pb.RegisterIdentityInput{
				Name:        "peer0",
				Type:        "peer",
				Affiliation: "org1.department1",
				Attributes: []*pb.Attribute{
					{Name: "hf.Registrar.Roles", Value: "peer", Ecert: true},
				},
				MaxEnrollments: 1,
			},
			clientRes: &pb.RegisterIdentityOutput{
				Name:        "peer0",
				Type:        "peer",
				Secret:      "generated-secret-123",
				Affiliation: "org1.department1",
				Attributes: []*pb.Attribute{
					{Name: "hf.Registrar.Roles", Value: "peer", Ecert: true},
				},
			},
		},
		{
			name: "success - orderer identity registration",
			input: &pb.RegisterIdentityInput{
				Name:        "orderer0",
				Type:        "orderer",
				Affiliation: "org1",
				Attributes: []*pb.Attribute{
					{Name: "hf.Registrar.Roles", Value: "orderer", Ecert: true},
				},
			},
			clientRes: &pb.RegisterIdentityOutput{
				Name:        "orderer0",
				Type:        "orderer",
				Secret:      "generated-secret-456",
				Affiliation: "org1",
				Attributes: []*pb.Attribute{
					{Name: "hf.Registrar.Roles", Value: "orderer", Ecert: true},
				},
			},
		},
		{
			name: "success - client identity with custom secret",
			input: &pb.RegisterIdentityInput{
				Name:        "client1",
				Type:        "client",
				Secret:      "custom-secret",
				Affiliation: "org1",
			},
			clientRes: &pb.RegisterIdentityOutput{
				Name:        "client1",
				Type:        "client",
				Secret:      "custom-secret",
				Affiliation: "org1",
			},
		},
		{
			name: "client connection error propagated",
			input: &pb.RegisterIdentityInput{
				Name:        "peer1",
				Type:        "peer",
				Affiliation: "org1",
			},
			clientErr: errClientConnectionRefused,
		},
		{
			name: "client unauthorized error propagated",
			input: &pb.RegisterIdentityInput{
				Name:        "peer2",
				Type:        "peer",
				Affiliation: "org1",
			},
			clientErr: errClientUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := fabricmocks.NewMockFabricCAClient(t)
			mockClient.EXPECT().
				Register(mock.Anything, tt.input).
				Return(tt.clientRes, tt.clientErr)

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.RegisterIdentity)

			// Execute activity
			val, err := env.ExecuteActivity(activities.RegisterIdentity, tt.input)

			if tt.clientErr != nil {
				// Expect error
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				// Expect success
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}
				var result *pb.RegisterIdentityOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if result.GetName() != tt.clientRes.GetName() ||
					result.GetType() != tt.clientRes.GetType() ||
					result.GetSecret() != tt.clientRes.GetSecret() {
					t.Errorf("result mismatch: got %+v, want %+v", result, tt.clientRes)
				}
			}
		})
	}
}

func TestEnrollIdentityActivity(t *testing.T) {
	tests := []struct {
		name      string
		input     *pb.EnrollIdentityInput
		clientRes *pb.EnrollIdentityOutput
		clientErr error
	}{
		{
			name: "success - enroll with default profile",
			input: &pb.EnrollIdentityInput{
				Name:    "peer0",
				Secret:  "secret-123",
				Profile: "",
				Csr: &pb.CSRInfo{
					Cn: "peer0",
					Hosts: []string{
						"peer0.org1.example.com",
						"peer0",
					},
				},
			},
			clientRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("-----BEGIN CERTIFICATE-----\nMIIC..."),
				PrivateKey:  []byte("-----BEGIN PRIVATE KEY-----\nMIIE..."),
				CaCert:      []byte("-----BEGIN CERTIFICATE-----\nCA..."),
				Profile:     "default",
			},
		},
		{
			name: "success - enroll with TLS profile",
			input: &pb.EnrollIdentityInput{
				Name:    "peer0",
				Secret:  "secret-123",
				Profile: "tls",
				Csr: &pb.CSRInfo{
					Cn: "peer0",
					Hosts: []string{
						"peer0.org1.example.com",
					},
				},
			},
			clientRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("-----BEGIN CERTIFICATE-----\nTLS..."),
				PrivateKey:  []byte("-----BEGIN PRIVATE KEY-----\nTLS..."),
				CaCert:      []byte("-----BEGIN CERTIFICATE-----\nCA..."),
				Profile:     "tls",
			},
		},
		{
			name: "success - enroll orderer",
			input: &pb.EnrollIdentityInput{
				Name:   "orderer0",
				Secret: "secret-456",
				Csr: &pb.CSRInfo{
					Cn: "orderer0",
					Hosts: []string{
						"orderer0.example.com",
					},
				},
			},
			clientRes: &pb.EnrollIdentityOutput{
				Certificate: []byte("-----BEGIN CERTIFICATE-----\nORD..."),
				PrivateKey:  []byte("-----BEGIN PRIVATE KEY-----\nORD..."),
				CaCert:      []byte("-----BEGIN CERTIFICATE-----\nCA..."),
			},
		},
		{
			name: "client connection error propagated",
			input: &pb.EnrollIdentityInput{
				Name:   "peer1",
				Secret: "secret",
			},
			clientErr: errClientConnectionRefused,
		},
		{
			name: "client unauthorized error propagated",
			input: &pb.EnrollIdentityInput{
				Name:   "peer2",
				Secret: "invalid-secret",
			},
			clientErr: errClientUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestActivityEnvironment()

			mockClient := fabricmocks.NewMockFabricCAClient(t)
			mockClient.EXPECT().
				Enroll(mock.Anything, tt.input).
				Return(tt.clientRes, tt.clientErr)

			activities := NewActivities(mockClient)
			env.RegisterActivity(activities.EnrollIdentity)

			// Execute activity
			val, err := env.ExecuteActivity(activities.EnrollIdentity, tt.input)

			if tt.clientErr != nil {
				// Expect error
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				// Expect success
				if err != nil {
					t.Fatalf("unexpected error from ExecuteActivity: %v", err)
				}
				var result *pb.EnrollIdentityOutput
				if err := val.Get(&result); err != nil {
					t.Fatalf("unexpected error from Get: %v", err)
				}
				if len(result.GetCertificate()) == 0 || len(result.GetPrivateKey()) == 0 {
					t.Errorf("result missing certificate or private key: %+v", result)
				}
			}
		})
	}
}
