package main

import (
	"github.com/Project-Voltron/voltron/cmd/argo-runner/runner"
	"github.com/Project-Voltron/voltron/cmd/argo-runner/runner/argo"
	wfclientset "github.com/argoproj/argo/pkg/client/clientset/versioned"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	//"projectvoltron.dev/voltron/cmd/argo-runner/argo"
	//"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// TODO: go with template method pattern or with composable blocks?
func main() {
	// create k8s client
	k8sCfg, err := config.GetConfig()
	exitOnError(err, "while creating k8s config")

	// create Argo workflow client
	wfClient, err := wfclientset.NewForConfig(k8sCfg)
	exitOnError(err, "while creating Argo client")

	stop := signals.SetupSignalHandler()

	argoRunner := argo.NewRunner(wfClient.ArgoprojV1alpha1())

	mgr, err := runner.NewManager(argoRunner)
	exitOnError(err, "while creating runner service")

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")

}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
