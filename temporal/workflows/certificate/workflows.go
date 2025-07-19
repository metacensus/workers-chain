package certificate

import (
	"github.com/metacensus/infra/temporal/workflows/util"
	"go.temporal.io/sdk/worker"
)

func RegisterWorkflows(w worker.Worker) {
	// Base Fabric CA workflows
	util.RegisterWorkflow(w, RegisterIdentityWorkflow, RegisterIdentityWorkflowName)
	util.RegisterWorkflow(w, EnrollIdentityWorkflow, EnrollIdentityWorkflowName)
	util.RegisterWorkflow(w, AssembleMSPWorkflow, AssembleMSPWorkflowName)
	util.RegisterWorkflow(w, StoreSecretWorkflow, StoreSecretWorkflowName)

	// Admin management workflows
	util.RegisterWorkflow(w, BootstrapAdminWorkflow, BootstrapAdminWorkflowName)

	// Composite workflows
	util.RegisterWorkflow(w, ProvisionPeerWorkflow, ProvisionPeerWorkflowName)
	util.RegisterWorkflow(w, ProvisionOrdererWorkflow, ProvisionOrdererWorkflowName)
	util.RegisterWorkflow(w, ProvisionClientWorkflow, ProvisionClientWorkflowName)

	// Bootstrap workflows
	util.RegisterWorkflow(w, BootstrapMetacensusOrgWorkflow, BootstrapMetacensusOrgWorkflowName)

	// Test workflows
	util.RegisterWorkflow(w, TestFabricCAWorkflow, TestFabricCAWorkflowName)

	// Activities
	util.RegisterActivity(w, RegisterIdentityActivity, "certificate.RegisterIdentityActivity")
	util.RegisterActivity(w, EnrollIdentityActivity, "certificate.EnrollIdentityActivity")
	util.RegisterActivity(w, AssembleMSPActivity, "certificate.AssembleMSPActivity")
	util.RegisterActivity(w, StoreSecretActivity, "certificate.StoreSecretActivity")
	util.RegisterActivity(w, CreateAdminIdentityActivity, "certificate.CreateAdminIdentityActivity")
	util.RegisterActivity(w, StoreAdminIdentityActivity, "certificate.StoreAdminIdentityActivity")
	util.RegisterActivity(w, RevokeDefaultAdminActivity, "certificate.RevokeDefaultAdminActivity")
	util.RegisterActivity(w, CreateAffiliationsActivity, "certificate.CreateAffiliationsActivity")
	util.RegisterActivity(w, GenerateSecureSecretActivity, "certificate.GenerateSecureSecretActivity")
}
