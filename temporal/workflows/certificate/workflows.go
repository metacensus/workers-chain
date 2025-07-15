package certificate

import (
	"github.com/metacensus/infra/temporal/workflows/util"
	"go.temporal.io/sdk/worker"
)

func RegisterWorkflows(w worker.Worker) {
	util.RegisterWorkflow(w, GenerateCertificateWorkflow, GenerateCertificateWorkflowName)
	util.RegisterWorkflow(w, FabricCertificatesWorkflow, FabricCertificatesWorkflowName)
	util.RegisterActivity(w, GenerateCertificateActivity, GenerateCertificateActivityName)
}
