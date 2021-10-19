package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"sigs.k8s.io/yaml"

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

// KubeconfigTypeInstanceFieldKey
const KubeconfigTypeInstanceFieldKey = "config"

// Config holds the input parameters for the Helm runner binary.
type Config struct {
	OptionalKubeconfigTI string `envconfig:"optional"`
	Command              helm.CommandType
	HelmReleasePath      string `envconfig:"optional"`
	HelmDriver           string `envconfig:"default=secrets"`
	RepositoryCachePath  string `envconfig:"default=/tmp/helm"`
	Output               struct {
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

	kcPath, cleanup, err := getOptionalKubeconfigPathIfExists(cfg.OptionalKubeconfigTI)
	exitOnError(err, "while getting optional kubeconfig path")
	defer cleanup()

	if kcPath != "" {
		k8sCfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kcPath},
			&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{}}).ClientConfig()
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

func noopFunc() error {
	return nil
}
func getOptionalKubeconfigPathIfExists(path string) (string, func() error, error) {
	f, err := os.Stat(path)
	switch {
	case err == nil:
		if f.IsDir() {
			return "", nil, errors.New("RUNNER_OPTIONAL_KUBECONFIG_TI cannot be dir, must be a file")
		}
	case os.IsNotExist(err):
		return "", noopFunc, nil
	default:
		return "", nil, err
	}

	kcfg, err := extractKubeconfigFromTI(path)
	if err != nil {
		return "", nil, err
	}

	file, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	cleanup := func() error {
		return os.Remove(file.Name())
	}

	if _, err := file.Write(kcfg); err != nil {
		return "", nil, err
	}

	return file.Name(), cleanup, err
}

func extractKubeconfigFromTI(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	raw := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	rawKubeconfig, found := raw[KubeconfigTypeInstanceFieldKey]
	if !found {
		return nil, fmt.Errorf("TypeInstance doesn't have %q field", KubeconfigTypeInstanceFieldKey)
	}

	kcfg, err := yaml.Marshal(rawKubeconfig)
	if err != nil {
		return nil, err
	}
	return kcfg, nil
}
