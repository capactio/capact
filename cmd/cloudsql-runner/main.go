package main

import (
	"context"
	"log"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"projectvoltron.dev/voltron/pkg/runner"
	"projectvoltron.dev/voltron/pkg/runner/cloudsql"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type noopReporter struct {
}

func (r *noopReporter) Report(ctx context.Context, execCtx runner.ExecutionContext, status interface{}) error {
	return nil
}

func main() {
	service, err := sqladmin.NewService(context.Background())
	exitOnError(err, "failed to create GCP client")

	cloudsqlRunner := cloudsql.NewRunner(service)

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
