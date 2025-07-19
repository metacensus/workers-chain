package certificate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestBootstrapAdminWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock the activities
	env.OnActivity(CreateAdminIdentityActivity, mock.Anything, CreateAdminIdentityParams{
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
	}).Return(&AdminIdentity{
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
	}, nil)

	env.OnActivity(StoreAdminIdentityActivity, mock.Anything, StoreAdminIdentityParams{
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
		VaultPath: "fabric/test/admin",
	}).Return(nil)

	env.OnActivity(RevokeDefaultAdminActivity, mock.Anything, RevokeDefaultAdminParams{
		CAEndpoint:     "http://localhost:7054",
		DefaultAdminID: "admin",
		DefaultSecret:  "adminpw",
	}).Return(nil)

	// Execute the workflow
	params := BootstrapAdminParams{
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

	env.ExecuteWorkflow(BootstrapAdminWorkflow, params)

	// Verify the workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result BootstrapAdminResult
	require.NoError(t, env.GetWorkflowResult(&result))

	// Verify the result
	assert.True(t, result.Success)
	assert.Equal(t, "Admin bootstrap completed successfully", result.Message)
	assert.NotNil(t, result.AdminIdentity)
	assert.Equal(t, "test-admin", result.AdminIdentity.ID)
	assert.Equal(t, "test-secret", result.AdminIdentity.Secret)
	assert.Equal(t, "http://localhost:7054", result.AdminIdentity.CAEndpoint)

	// Verify all activities were called
	env.AssertExpectations(t)
}

func TestBootstrapAdminWorkflow_CreateAdminIdentityFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock the CreateAdminIdentityActivity to fail
	env.OnActivity(CreateAdminIdentityActivity, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	// Execute the workflow
	params := BootstrapAdminParams{
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

	env.ExecuteWorkflow(BootstrapAdminWorkflow, params)

	// Verify the workflow failed
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())

	// The workflow should return an error, not a result
	var result BootstrapAdminResult
	err := env.GetWorkflowResult(&result)
	require.Error(t, err) // Should not be able to get a result when workflow fails
}

func TestBootstrapAdminWorkflow_StoreAdminIdentityFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock successful CreateAdminIdentityActivity
	env.OnActivity(CreateAdminIdentityActivity, mock.Anything, mock.Anything).Return(&AdminIdentity{
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
	}, nil)

	// Mock StoreAdminIdentityActivity to fail
	env.OnActivity(StoreAdminIdentityActivity, mock.Anything, mock.Anything).Return(assert.AnError)

	// Execute the workflow
	params := BootstrapAdminParams{
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

	env.ExecuteWorkflow(BootstrapAdminWorkflow, params)

	// Verify the workflow failed
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())

	// The workflow should return an error, not a result
	var result BootstrapAdminResult
	err := env.GetWorkflowResult(&result)
	require.Error(t, err) // Should not be able to get a result when workflow fails
}

func TestBootstrapAdminWorkflow_RevokeDefaultAdminFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock successful CreateAdminIdentityActivity
	env.OnActivity(CreateAdminIdentityActivity, mock.Anything, mock.Anything).Return(&AdminIdentity{
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
	}, nil)

	// Mock successful StoreAdminIdentityActivity
	env.OnActivity(StoreAdminIdentityActivity, mock.Anything, mock.Anything).Return(nil)

	// Mock RevokeDefaultAdminActivity to fail
	env.OnActivity(RevokeDefaultAdminActivity, mock.Anything, mock.Anything).Return(assert.AnError)

	// Execute the workflow
	params := BootstrapAdminParams{
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

	env.ExecuteWorkflow(BootstrapAdminWorkflow, params)

	// Verify the workflow completed successfully (revoke failure is not fatal)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result BootstrapAdminResult
	require.NoError(t, env.GetWorkflowResult(&result))

	// Verify the result indicates success (revoke failure is logged but doesn't fail the workflow)
	assert.True(t, result.Success)
	assert.Equal(t, "Admin bootstrap completed successfully", result.Message)
}
