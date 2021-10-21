package main

import (
	"log"

	"capact.io/capact/pkg/runner"
	"capact.io/capact/pkg/runner/helm"
	statusreporter "capact.io/capact/pkg/runner/status-reporter"
	"k8s.io/client-go/rest"

	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Config holds the input parameters for the Helm runner binary.
type Config struct {
	// Kubeconfig to be used by the runner
	KubeConfig          string `envconfig:"optional"`
	Command             helm.CommandType
	HelmReleasePath     string `envconfig:"optional"`
	HelmDriver          string `envconfig:"default=secrets"`
	RepositoryCachePath string `envconfig:"default=/tmp/helm"`
	Output              struct {
		HelmReleaseFilePath string `envconfig:"default=/tmp/helm-release.yaml"`
		// Extracting resource metadata from Kubernetes as outputs
		AdditionalFilePath string `envconfig:"default=/tmp/additional.yaml"`
	}
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "RUNNER")
	exitOnError(err, "while loading configuration")

	stop := signals.SetupSignalHandler()
	var k8sCfg *rest.Config
	if cfg.KubeConfig != "" {
		k8sCfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: cfg.KubeConfig},
			&clientcmd.ConfigOverrides{
				ClusterInfo: clientcmdapi.Cluster{
					Server: "",
				},
				CurrentContext: "",
			}).ClientConfig()
	} else {
		k8sCfg, err = config.GetConfig()
	}

	exitOnError(err, "while creating k8s config")

	helmConfig := helm.Config{
		Command:             cfg.Command,
		HelmReleasePath:     cfg.HelmReleasePath,
		HelmDriver:          cfg.HelmDriver,
		RepositoryCachePath: cfg.RepositoryCachePath,
		Output:              cfg.Output,
	}
	helmRunner := helm.NewRunner(k8sCfg, helmConfig)

	statusReporter := statusreporter.NewNoop()

	// create and run manager
	mgr, err := runner.NewManager(helmRunner, statusReporter)
	exitOnError(err, "while creating runner manager")

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
