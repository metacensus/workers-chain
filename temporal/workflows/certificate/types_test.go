package certificate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnrollment_ToSecretData(t *testing.T) {
	enrollment := &Enrollment{
		Certificate: []byte("test-cert"),
		PrivateKey:  []byte("test-key"),
		CACert:      []byte("test-ca"),
		Profile:     "sign",
	}

	data := enrollment.ToSecretData()

	assert.Equal(t, []byte("test-cert"), data["cert.pem"])
	assert.Equal(t, []byte("test-key"), data["key.pem"])
	assert.Equal(t, []byte("test-ca"), data["ca.pem"])
}

func TestEnrollment_ToTLSSecretData(t *testing.T) {
	enrollment := &Enrollment{
		Certificate: []byte("test-cert"),
		PrivateKey:  []byte("test-key"),
		CACert:      []byte("test-ca"),
		Profile:     "tls",
	}

	data := enrollment.ToTLSSecretData()

	assert.Equal(t, []byte("test-cert"), data["tls.crt"])
	assert.Equal(t, []byte("test-key"), data["tls.key"])
	assert.Equal(t, []byte("test-ca"), data["ca.crt"])
}

func TestCSRConfig_ToCSRInfo(t *testing.T) {
	csrConfig := &CSRConfig{
		Hosts: []string{
			"localhost",
			"127.0.0.1",
		},
		CommonName:   "test.example.com",
		Organization: "Test Org",
		Country:      "US",
		Province:     "CA",
		Locality:     "San Francisco",
	}

	csrInfo := csrConfig.ToCSRInfo()

	assert.Equal(t, "test.example.com", csrInfo.CN)
	assert.Equal(t, []string{"Test Org"}, csrInfo.Names[0].O)
	assert.Equal(t, []string{"US"}, csrInfo.Names[0].C)
	assert.Equal(t, []string{"CA"}, csrInfo.Names[0].ST)
	assert.Equal(t, []string{"San Francisco"}, csrInfo.Names[0].L)
	assert.Equal(t, []string{"localhost", "127.0.0.1"}, csrInfo.Hosts)
}

func TestBootstrapConfig_Validation(t *testing.T) {
	// Test valid config
	validConfig := BootstrapConfig{
		CAEndpoint: "http://localhost:7054",
		OrgNodeOUs: NodeOUConfig{
			Enable: true,
			Client: map[string]string{
				"certificateAuthorities": "true",
				"clientOU":               "true",
			},
			Peer: map[string]string{
				"certificateAuthorities": "true",
				"peerOU":                 "true",
			},
			Admin: map[string]string{
				"certificateAuthorities": "true",
				"adminOU":                "true",
			},
			Orderer: map[string]string{
				"certificateAuthorities": "true",
				"ordererOU":              "true",
			},
		},
		Namespace: "fabric",
	}

	// This should not panic
	assert.NotEmpty(t, validConfig.CAEndpoint)
	assert.True(t, validConfig.OrgNodeOUs.Enable)
	assert.NotEmpty(t, validConfig.Namespace)

	// Test empty CAEndpoint
	emptyConfig := BootstrapConfig{
		CAEndpoint: "",
		OrgNodeOUs: NodeOUConfig{
			Enable: true,
		},
		Namespace: "fabric",
	}

	assert.Empty(t, emptyConfig.CAEndpoint)
}

func TestBootstrapAdminParams_Validation(t *testing.T) {
	// Test valid params
	validParams := BootstrapAdminParams{
		CAEndpoint:     "http://localhost:7054",
		AdminID:        "test-admin",
		AdminSecret:    "test-secret",
		DefaultAdminID: "admin",
		DefaultSecret:  "adminpw",
		Attributes: map[string]string{
			"hf.Registrar.Roles":         "admin",
			"hf.Registrar.Attributes":    "*",
			"hf.Registrar.DelegateRoles": "admin",
		},
		VaultPath: "fabric/test/admin",
	}

	assert.NotEmpty(t, validParams.CAEndpoint)
	assert.NotEmpty(t, validParams.AdminID)
	assert.NotEmpty(t, validParams.AdminSecret)
	assert.NotEmpty(t, validParams.DefaultAdminID)
	assert.NotEmpty(t, validParams.DefaultSecret)
	assert.NotEmpty(t, validParams.VaultPath)
	assert.Len(t, validParams.Attributes, 3)

	// Test empty CAEndpoint
	emptyParams := BootstrapAdminParams{
		CAEndpoint:     "",
		AdminID:        "test-admin",
		AdminSecret:    "test-secret",
		DefaultAdminID: "admin",
		DefaultSecret:  "adminpw",
		Attributes: map[string]string{
			"hf.Registrar.Roles":         "admin",
			"hf.Registrar.Attributes":    "*",
			"hf.Registrar.DelegateRoles": "admin",
		},
		VaultPath: "fabric/test/admin",
	}

	assert.Empty(t, emptyParams.CAEndpoint)
}

func TestAdminIdentity_Validation(t *testing.T) {
	// Test valid admin identity
	validIdentity := &AdminIdentity{
		ID:          "test-admin",
		Secret:      "test-secret",
		Certificate: []byte("test-cert"),
		PrivateKey:  []byte("test-key"),
		CACert:      []byte("test-ca"),
		Attributes: map[string]string{
			"hf.Registrar.Roles":         "admin",
			"hf.Registrar.Attributes":    "*",
			"hf.Registrar.DelegateRoles": "admin",
		},
		CAEndpoint: "http://localhost:7054",
	}

	assert.NotEmpty(t, validIdentity.ID)
	assert.NotEmpty(t, validIdentity.Secret)
	assert.NotEmpty(t, validIdentity.Certificate)
	assert.NotEmpty(t, validIdentity.PrivateKey)
	assert.NotEmpty(t, validIdentity.CACert)
	assert.NotEmpty(t, validIdentity.CAEndpoint)
	assert.Len(t, validIdentity.Attributes, 3)

	// Test empty CAEndpoint
	emptyIdentity := &AdminIdentity{
		ID:          "test-admin",
		Secret:      "test-secret",
		Certificate: []byte("test-cert"),
		PrivateKey:  []byte("test-key"),
		CACert:      []byte("test-ca"),
		Attributes: map[string]string{
			"hf.Registrar.Roles":         "admin",
			"hf.Registrar.Attributes":    "*",
			"hf.Registrar.DelegateRoles": "admin",
		},
		CAEndpoint: "",
	}

	assert.Empty(t, emptyIdentity.CAEndpoint)
}

func TestBootstrapAdminResult_Validation(t *testing.T) {
	// Test successful result
	successResult := &BootstrapAdminResult{
		AdminIdentity: &AdminIdentity{
			ID:          "test-admin",
			Secret:      "test-secret",
			Certificate: []byte("test-cert"),
			PrivateKey:  []byte("test-key"),
			CACert:      []byte("test-ca"),
			Attributes: map[string]string{
				"hf.Registrar.Roles":         "admin",
				"hf.Registrar.Attributes":    "*",
				"hf.Registrar.DelegateRoles": "admin",
			},
			CAEndpoint: "http://localhost:7054",
		},
		Success: true,
		Message: "Admin bootstrap completed successfully",
	}

	assert.True(t, successResult.Success)
	assert.NotEmpty(t, successResult.Message)
	assert.NotNil(t, successResult.AdminIdentity)

	// Test failed result
	failedResult := &BootstrapAdminResult{
		AdminIdentity: nil,
		Success:       false,
		Message:       "Admin bootstrap failed",
	}

	assert.False(t, failedResult.Success)
	assert.NotEmpty(t, failedResult.Message)
	assert.Nil(t, failedResult.AdminIdentity)
}
