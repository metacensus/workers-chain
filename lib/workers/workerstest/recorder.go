// Package workerstest records the workflows a worker registers.
package workerstest

import (
	"slices"
	"testing"

	clusterpb "github.com/metacensus/workers-chain/gen/cluster/fabric/v1"
	certauthpb "github.com/metacensus/workers-chain/gen/fabric/certauth/v1"
	fabricpb "github.com/metacensus/workers-chain/gen/fabric/v1"
	clusterfabric "github.com/metacensus/workers-chain/lib/cluster/fabric/v1"
	certauth "github.com/metacensus/workers-chain/lib/fabric/certauth/v1"
	fabric "github.com/metacensus/workers-chain/lib/fabric/v1"
	"github.com/metacensus/workers-chain/lib/workers/config"
	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
)

// recorder is a worker.Worker that records the workflow names registered on
// it; any other method panics.
type recorder struct {
	worker.Worker
	workflows []string
}

func (r *recorder) RegisterWorkflowWithOptions(_ any, options workflow.RegisterOptions) {
	r.workflows = append(r.workflows, options.Name)
}

func (r *recorder) RegisterNexusService(*nexus.Service) {}

// Registered runs invoke as owner's worker and returns the workflow names it
// registers, sorted.
func Registered(t *testing.T, owner config.OwnerConfig, invoke fx.Option) []string {
	t.Helper()
	operator := config.NewOperatorConfig("operator")
	w := &recorder{}
	app := fx.New(
		fx.NopLogger,
		clusterfabric.Module, certauth.Module, fabric.Module,
		fx.Supply(&operator, fx.Annotate(owner, fx.As(new(config.OwnerConfig)))),
		fx.Supply(fx.Annotate(w, fx.As(new(worker.Worker)))),
		invoke,
	)
	if err := app.Err(); err != nil {
		t.Fatal(err)
	}
	slices.Sort(w.workflows)
	return w.workflows
}

// AllExcept is every workflow the cluster, certauth and fabric services
// declare, less the names given, sorted.
func AllExcept(names ...string) []string {
	operator := config.NewOperatorConfig("operator")
	all := &recorder{}
	clusterpb.RegisterClusterWorkflows(all, clusterfabric.NewWorkflows(&operator))
	certauthpb.RegisterCertificateAuthorityWorkflows(all, certauth.NewWorkflows(operator))
	fabricpb.RegisterFabricWorkflows(all, fabric.NewWorkflows(operator, &operator))
	kept := slices.DeleteFunc(all.workflows, func(n string) bool { return slices.Contains(names, n) })
	slices.Sort(kept)
	return kept
}
