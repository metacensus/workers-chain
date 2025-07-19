package certificate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestBootstrapMetacensusOrgWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register the child workflow
	env.RegisterWorkflow(BootstrapAdminWorkflow)

	// Mock activities for the child workflow
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

	// Mock GenerateSecureSecretActivity
	env.OnActivity(GenerateSecureSecretActivity, mock.Anything, 32).Return("test-admin-secret", nil)

	// Mock CreateAffiliationsActivity
	env.OnActivity(CreateAffiliationsActivity, mock.Anything, mock.Anything).Return(nil)

	// Execute the workflow
	config := BootstrapConfig{
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

	env.ExecuteWorkflow(BootstrapMetacensusOrgWorkflow, config)

	// Verify the workflow completed successfully
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// Verify all activities were called
	env.AssertExpectations(t)
}

func TestBootstrapMetacensusOrgWorkflow_GenerateSecretFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock GenerateSecureSecretActivity to fail
	env.OnActivity(GenerateSecureSecretActivity, mock.Anything, 32).Return("", assert.AnError)

	// Execute the workflow
	config := BootstrapConfig{
		CAEndpoint: "http://localhost:7054",
		OrgNodeOUs: NodeOUConfig{
			Enable: true,
		},
		Namespace: "fabric",
	}

	env.ExecuteWorkflow(BootstrapMetacensusOrgWorkflow, config)

	// Verify the workflow failed
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	assert.Contains(t, env.GetWorkflowError().Error(), "failed to generate admin secret")
}

func TestBootstrapMetacensusOrgWorkflow_CreateAffiliationsFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register the child workflow
	env.RegisterWorkflow(BootstrapAdminWorkflow)

	// Mock activities for the child workflow
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

	// Mock GenerateSecureSecretActivity
	env.OnActivity(GenerateSecureSecretActivity, mock.Anything, 32).Return("test-admin-secret", nil)

	// Mock CreateAffiliationsActivity to fail
	env.OnActivity(CreateAffiliationsActivity, mock.Anything, mock.Anything).Return(assert.AnError)

	// Execute the workflow
	config := BootstrapConfig{
		CAEndpoint: "http://localhost:7054",
		OrgNodeOUs: NodeOUConfig{
			Enable: true,
		},
		Namespace: "fabric",
	}

	env.ExecuteWorkflow(BootstrapMetacensusOrgWorkflow, config)

	// Verify the workflow failed
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	assert.Contains(t, env.GetWorkflowError().Error(), "failed to create affiliations")
}
