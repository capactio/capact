package main

import (
	"log"

	"projectvoltron.dev/voltron/pkg/runner"
	statusreporter "projectvoltron.dev/voltron/pkg/runner/status-reporter"
	"projectvoltron.dev/voltron/pkg/runner/terraform"

	"github.com/vrischmann/envconfig"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	var cfg terraform.Config
	err := envconfig.InitWithPrefix(&cfg, "RUNNER")
	exitOnError(err, "while loading configuration")

	stop := signals.SetupSignalHandler()

	terraformRunner := terraform.NewTerraformRunner(cfg)

	statusReporter := statusreporter.NewNoop()

	// create and run manager
	mgr, err := runner.NewManager(terraformRunner, statusReporter)
	exitOnError(err, "while creating runner manager")

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
