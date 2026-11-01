package workers

import (
	"context"
	"io/fs"
	"log"
	"os"

	"github.com/metacensus/library-temporal-kube/helm"
	"github.com/metacensus/library-temporal-kube/kube"
	"github.com/metacensus/workers-chain/lib/healthcheck"
	"github.com/metacensus/workers-chain/lib/workers/config"
	temporalmtls "github.com/metacensus/workers-chain/lib/workflows/temporalmtls/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
)

// Shared worker infrastructure

// NewWorker creates a Temporal worker from the client and environment configuration
func NewWorker(c client.Client, cfg *config.TemporalConfig) worker.Worker {
	return worker.New(c, cfg.TaskQueue, worker.Options{})
}

// ClusterModule hosts the library's Helm and Kubernetes activities, and the
// Fabric mTLS activity beside them, over this worker's charts and credentials.
var ClusterModule = fx.Options(
	fx.Provide(Charts, RestConfig),
	helm.Module,
	kube.Module(kube.WithFieldManager(config.FieldManager)),
	temporalmtls.Module,
)

// Charts is the chart tree the worker renders: HELM_CHART_DIR, or the
// Dockerfile's COPY at /helm.
func Charts() fs.FS {
	dir := os.Getenv("HELM_CHART_DIR")
	if dir == "" {
		dir = "/helm"
	}
	return os.DirFS(dir)
}

// RestConfig is the worker's cluster credential: the kubeconfig KUBECONFIG
// names, or in-cluster when it names none.
func RestConfig(cfg *config.KubernetesConfig) (*rest.Config, error) {
	return kube.LoadConfig(cfg.KubeconfigPath, "")
}

// RegisterClusterActivities registers what ClusterModule provides.
func RegisterClusterActivities(w worker.Worker, h *helm.Activities, k *kube.Activities, m *temporalmtls.Activities) {
	helm.RegisterActivities(w, h)
	kube.RegisterActivities(w, k)
	temporalmtls.RegisterActivities(w, m)
}

// Shared lifecycle functions

// StartHealthcheck starts the healthcheck HTTP server in a goroutine
func StartHealthcheck(lc fx.Lifecycle, c client.Client, cfg *config.WorkerConfig) {
	s := healthcheck.NewServer(c)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go s.Run(cfg.HealthCheckAddress)
			log.Println("Healthcheck server started at", cfg.HealthCheckAddress)
			return nil
		},
	})
}

// RunWorker runs the Temporal worker until interrupted
func RunWorker(lc fx.Lifecycle, w worker.Worker, temporalCfg *config.TemporalConfig) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Start worker in background
			go func() {
				log.Println("Starting Temporal worker on task queue:", temporalCfg.TaskQueue)
				err := w.Run(worker.InterruptCh())
				if err != nil {
					log.Fatalln("Worker error:", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping worker...")
			w.Stop()
			return nil
		},
	})
}
