package main

import (
	"log"

	statusreporter "projectvoltron.dev/voltron/pkg/runner/status-reporter"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"projectvoltron.dev/voltron/pkg/runner"
	"projectvoltron.dev/voltron/pkg/runner/argo"

	wfclientset "github.com/argoproj/argo/pkg/client/clientset/versioned"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	// create k8s client
	k8sCfg, err := config.GetConfig()
	exitOnError(err, "while creating k8s config")

	// create Argo workflow client
	wfCli, err := wfclientset.NewForConfig(k8sCfg)
	exitOnError(err, "while creating Argo client")

	stop := signals.SetupSignalHandler()

	argoRunner := argo.NewRunner(wfCli.ArgoprojV1alpha1())

	// status reporter
	k8sCli, err := client.New(config.GetConfigOrDie(), client.Options{})
	exitOnError(err, "while creating K8s client")

	statusReporter := statusreporter.NewK8sConfigMap(k8sCli)

	// manager
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
