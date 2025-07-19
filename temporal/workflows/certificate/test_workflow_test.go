package certificate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestTestFabricCAWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register the child workflow
	env.RegisterWorkflow(BootstrapMetacensusOrgWorkflow)

	// Register the grandchild workflow
	env.RegisterWorkflow(BootstrapAdminWorkflow)

	// Mock activities for the grandchild workflow
	env.OnActivity(CreateAdminIdentityActivity, mock.Anything, mock.Anything).Return(&AdminIdentity{
		ID:          "metacensus-admin",
		Secret:      "test-admin-secret",
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

	env.OnActivity(StoreAdminIdentityActivity, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(RevokeDefaultAdminActivity, mock.Anything, mock.Anything).Return(nil)

	// Mock activities for the child workflow
	env.OnActivity(GenerateSecureSecretActivity, mock.Anything, 32).Return("test-admin-secret", nil)
	env.OnActivity(CreateAffiliationsActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute the workflow
	env.ExecuteWorkflow(TestFabricCAWorkflow)

	// Verify the workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

func TestTestFabricCAWorkflow_BootstrapFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock the child workflow to fail
	env.OnActivity(BootstrapMetacensusOrgWorkflow, mock.Anything, mock.Anything).Return(assert.AnError)

	// Execute the workflow
	env.ExecuteWorkflow(TestFabricCAWorkflow)

	// Verify the workflow failed
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	assert.Contains(t, env.GetWorkflowError().Error(), "bootstrap workflow failed")
}
