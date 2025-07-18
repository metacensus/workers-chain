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

	// Composite workflows
	util.RegisterWorkflow(w, ProvisionPeerWorkflow, ProvisionPeerWorkflowName)
	util.RegisterWorkflow(w, ProvisionOrdererWorkflow, ProvisionOrdererWorkflowName)
	util.RegisterWorkflow(w, ProvisionClientWorkflow, ProvisionClientWorkflowName)

	// Bootstrap workflows
	util.RegisterWorkflow(w, BootstrapMetacensusOrgWorkflow, BootstrapMetacensusOrgWorkflowName)

	// Test workflows
	util.RegisterWorkflow(w, TestFabricCAWorkflow, TestFabricCAWorkflowName)

	// Legacy workflows (to be replaced)
	util.RegisterWorkflow(w, GenerateCertificateWorkflow, GenerateCertificateWorkflowName)
	util.RegisterWorkflow(w, FabricCertificatesWorkflow, FabricCertificatesWorkflowName)

	// Activities
	util.RegisterActivity(w, RegisterIdentityActivity, "certificate.RegisterIdentityActivity")
	util.RegisterActivity(w, EnrollIdentityActivity, "certificate.EnrollIdentityActivity")
	util.RegisterActivity(w, AssembleMSPActivity, "certificate.AssembleMSPActivity")
	util.RegisterActivity(w, StoreSecretActivity, "certificate.StoreSecretActivity")
	util.RegisterActivity(w, GenerateCertificateActivity, GenerateCertificateActivityName)
}
