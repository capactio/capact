package main

import (
	"context"
	"log"
	"time"

	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/runner"
	"projectvoltron.dev/voltron/pkg/runner/cloudsql"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type noopReporter struct {
}

type Config struct {
	GcpProjectName string `envconfig:"default=projectvoltron"`
	Debug          bool   `envconfig:"default=false"`
}

func (r *noopReporter) Report(ctx context.Context, execCtx runner.ExecutionContext, status interface{}) error {
	return nil
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "failed to load config")

	var logger *zap.Logger
	if cfg.Debug {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	httpClient := httputil.NewClient(30*time.Second, false)

	service, err := sqladmin.NewService(context.Background(), option.WithHTTPClient(httpClient))
	exitOnError(err, "failed to create GCP client")

	cloudsqlRunner := cloudsql.NewRunner(logger, service, cfg.GcpProjectName)

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
