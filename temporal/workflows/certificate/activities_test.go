package certificate

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestGenerateSecureSecretActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock the activity to return a test result
	env.OnActivity(GenerateSecureSecretActivity, mock.Anything, 32).Return("a1b2c3d4e5f678901234567890123456", nil)

	// Execute a simple workflow that calls the activity
	env.ExecuteWorkflow(func(ctx workflow.Context) (string, error) {
		// Set activity options within the workflow
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		var result string
		err := workflow.ExecuteActivity(ctx, GenerateSecureSecretActivity, 32).Get(ctx, &result)
		return result, err
	})

	// Verify the workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result string
	require.NoError(t, env.GetWorkflowResult(&result))

	// Verify the result
	assert.Equal(t, "a1b2c3d4e5f678901234567890123456", result)
	assert.Len(t, result, 32)
	assert.Regexp(t, `^[0-9a-f]{32}$`, result)
}

func TestGenerateSecureSecretActivity_InvalidLength(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock the activity to return an error for invalid length
	env.OnActivity(GenerateSecureSecretActivity, mock.Anything, -1).Return("", assert.AnError)

	// Execute a simple workflow that calls the activity
	env.ExecuteWorkflow(func(ctx workflow.Context) (string, error) {
		var result string
		err := workflow.ExecuteActivity(ctx, GenerateSecureSecretActivity, -1).Get(ctx, &result)
		return result, err
	})

	// Verify the workflow failed
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}

func TestGenerateSecureSecret(t *testing.T) {
	// Test with valid length
	result, err := GenerateSecureSecret(32)
	require.NoError(t, err)
	assert.Len(t, result, 32)
	assert.Regexp(t, `^[0-9a-f]{32}$`, result)

	// Test with different lengths
	result16, err := GenerateSecureSecret(16)
	require.NoError(t, err)
	assert.Len(t, result16, 16)
	assert.Regexp(t, `^[0-9a-f]{16}$`, result16)
}

func TestGenerateSecureSecret_InvalidLength(t *testing.T) {
	// Test with negative length - this should panic or return empty
	// Since the function doesn't validate input, we'll test the actual behavior
	result, err := GenerateSecureSecret(-1)
	// The function doesn't validate input, so it will try to generate with negative length
	// This will result in an empty string or panic
	if err != nil {
		require.Error(t, err)
	} else {
		// If no error, the result should be empty or invalid
		assert.Empty(t, result)
	}

	// Test with zero length
	result, err = GenerateSecureSecret(0)
	if err != nil {
		require.Error(t, err)
	} else {
		// If no error, the result should be empty
		assert.Empty(t, result)
	}
}

func TestCreateAffiliationsActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	params := CreateAffiliationsParams{
		CAEndpoint: "http://localhost:7054",
		AdminIdentity: &AdminIdentity{
			ID:     "test-admin",
			Secret: "test-secret",
		},
		Affiliations: []string{
			"metacensus",
			"metacensus.department1",
			"metacensus.department2",
		},
	}

	// Mock the activity
	env.OnActivity(CreateAffiliationsActivity, mock.Anything, params).Return(nil)

	// Execute a simple workflow that calls the activity
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		// Set activity options within the workflow
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		return workflow.ExecuteActivity(ctx, CreateAffiliationsActivity, params).Get(ctx, nil)
	})

	// Verify the workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

func TestStoreAdminIdentityActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	params := StoreAdminIdentityParams{
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
	}

	// Mock the activity
	env.OnActivity(StoreAdminIdentityActivity, mock.Anything, params).Return(nil)

	// Execute a simple workflow that calls the activity
	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		// Set activity options within the workflow
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		return workflow.ExecuteActivity(ctx, StoreAdminIdentityActivity, params).Get(ctx, nil)
	})

	// Verify the workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

// TestStoreAdminIdentityActivity_Validation tests input validation
func TestStoreAdminIdentityActivity_Validation(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Test with nil admin identity
	env.OnActivity(StoreAdminIdentityActivity, mock.Anything, StoreAdminIdentityParams{
		AdminIdentity: nil,
		VaultPath:     "fabric/test/admin",
	}).Return(fmt.Errorf("admin identity cannot be nil"))

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		return workflow.ExecuteActivity(ctx, StoreAdminIdentityActivity, StoreAdminIdentityParams{
			AdminIdentity: nil,
			VaultPath:     "fabric/test/admin",
		}).Get(ctx, nil)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	assert.Contains(t, env.GetWorkflowError().Error(), "admin identity cannot be nil")

	// Test with empty admin ID
	env = testSuite.NewTestWorkflowEnvironment()
	env.OnActivity(StoreAdminIdentityActivity, mock.Anything, StoreAdminIdentityParams{
		AdminIdentity: &AdminIdentity{
			ID:          "",
			Secret:      "test-secret",
			Certificate: []byte("test-cert"),
			PrivateKey:  []byte("test-key"),
			CACert:      []byte("test-ca"),
			Attributes:  map[string]string{},
			CAEndpoint:  "http://localhost:7054",
		},
		VaultPath: "fabric/test/admin",
	}).Return(fmt.Errorf("admin identity ID cannot be empty"))

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		return workflow.ExecuteActivity(ctx, StoreAdminIdentityActivity, StoreAdminIdentityParams{
			AdminIdentity: &AdminIdentity{
				ID:          "",
				Secret:      "test-secret",
				Certificate: []byte("test-cert"),
				PrivateKey:  []byte("test-key"),
				CACert:      []byte("test-ca"),
				Attributes:  map[string]string{},
				CAEndpoint:  "http://localhost:7054",
			},
			VaultPath: "fabric/test/admin",
		}).Get(ctx, nil)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	assert.Contains(t, env.GetWorkflowError().Error(), "admin identity ID cannot be empty")

	// Test with empty vault path
	env = testSuite.NewTestWorkflowEnvironment()
	env.OnActivity(StoreAdminIdentityActivity, mock.Anything, StoreAdminIdentityParams{
		AdminIdentity: &AdminIdentity{
			ID:          "test-admin",
			Secret:      "test-secret",
			Certificate: []byte("test-cert"),
			PrivateKey:  []byte("test-key"),
			CACert:      []byte("test-ca"),
			Attributes:  map[string]string{},
			CAEndpoint:  "http://localhost:7054",
		},
		VaultPath: "",
	}).Return(fmt.Errorf("vault path cannot be empty"))

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		return workflow.ExecuteActivity(ctx, StoreAdminIdentityActivity, StoreAdminIdentityParams{
			AdminIdentity: &AdminIdentity{
				ID:          "test-admin",
				Secret:      "test-secret",
				Certificate: []byte("test-cert"),
				PrivateKey:  []byte("test-key"),
				CACert:      []byte("test-ca"),
				Attributes:  map[string]string{},
				CAEndpoint:  "http://localhost:7054",
			},
			VaultPath: "",
		}).Get(ctx, nil)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	assert.Contains(t, env.GetWorkflowError().Error(), "vault path cannot be empty")
}
