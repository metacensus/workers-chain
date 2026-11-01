package temporalmtls

import (
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// Module provides the activities over the worker's kube.Client.
var Module = fx.Module("temporalmtls", fx.Provide(NewActivities))

// RegisterActivities registers EnsureTemporalMTLSClientCert under its own name.
func RegisterActivities(w worker.Worker, a *Activities) {
	w.RegisterActivity(a.EnsureTemporalMTLSClientCert)
}
