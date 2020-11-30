package main

import (
	"context"
	"log"

	"github.com/vrischmann/envconfig"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"projectvoltron.dev/voltron/pkg/runner"
	"projectvoltron.dev/voltron/pkg/runner/cloudsql"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type noopReporter struct {
}

type Config struct {
	GcpProjectName string `envconfig:"default=projectvoltron"`
}

func (r *noopReporter) Report(ctx context.Context, execCtx runner.ExecutionContext, status interface{}) error {
	return nil
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "failed to load config")

	service, err := sqladmin.NewService(context.Background())
	exitOnError(err, "failed to create GCP client")

	cloudsqlRunner := cloudsql.NewRunner(service, cfg.GcpProjectName)

	mgr, err := runner.NewManager(cloudsqlRunner, &noopReporter{})
	exitOnError(err, "failed to create manager")

	stop := signals.SetupSignalHandler()

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
