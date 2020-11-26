package main

import (
	"log"

	statusreporter "projectvoltron.dev/voltron/internal/k8s-engine/status-reporter"
	"projectvoltron.dev/voltron/pkg/runner"
	"projectvoltron.dev/voltron/pkg/runner/argo"

	wfclientset "github.com/argoproj/argo/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	stop := signals.SetupSignalHandler()

	// create k8s client
	k8sCfg, err := config.GetConfig()
	exitOnError(err, "while creating k8s config")

	// create Argo workflow client
	wfCli, err := wfclientset.NewForConfig(k8sCfg)
	exitOnError(err, "while creating Argo client")

	argoRunner := argo.NewRunner(wfCli)

	// create status reporter
	k8sCli, err := client.New(k8sCfg, client.Options{})
	exitOnError(err, "while creating K8s client")

	statusReporter := statusreporter.NewK8sSecret(k8sCli)

	// create and run manager
	mgr, err := runner.NewManager(argoRunner, statusReporter)
	exitOnError(err, "while creating runner manager")

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
