package main

import (
	"log"

	"capact.io/capact/pkg/runner"
	glabapi "capact.io/capact/pkg/runner/gitlab-api-runner"
	statusreporter "capact.io/capact/pkg/runner/status-reporter"

	"github.com/vrischmann/envconfig"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	var cfg glabapi.Config
	err := envconfig.InitWithPrefix(&cfg, "RUNNER")
	exitOnError(err, "while loading configuration")

	stop := signals.SetupSignalHandler()

	glabAPIRunner := glabapi.NewRESTRunner(cfg)

	statusReporter := statusreporter.NewNoop()

	// create and run manager
	mgr, err := runner.NewManager(glabAPIRunner, statusReporter)
	exitOnError(err, "while creating runner manager")

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
