package helm

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"
)

// KubeconfigTypeInstanceFieldKey defines property name under which the kubeconfig is stored in TypeInstance.
const KubeconfigTypeInstanceFieldKey = "config"

func (r *helmRunner) loadKubeconfig() (*rest.Config, error) {
	_, err := setKubeconfigEnvIfTypeInstanceExists(r.cfg.OptionalKubeconfigTI, r.log)
	if err != nil {
		return nil, errors.Wrap(err, "while setting optional kubeconfig")
	}
	//defer func() {
	//	merr := multierror.Append(err, cleanup())
	//	err = merr.ErrorOrNil()
	//}()

	return config.GetConfig()
}

func noopFunc() error {
	return nil
}

func setKubeconfigEnvIfTypeInstanceExists(path string, log *zap.Logger) (func() error, error) {
	if path == "" {
		log.Debug("optional Kubeconfig TI not specified")
		return noopFunc, nil
	}

	f, err := os.Stat(path)
	switch {
	case err == nil:
		if f.IsDir() {
			return nil, errors.New("RUNNER_OPTIONAL_KUBECONFIG_TI cannot be dir, must be a file")
		}
	case os.IsNotExist(err):
		log.Debug("optional Kubeconfig TI specified but file does not exist")
		return noopFunc, nil
	default:
		return nil, err
	}

	kcfg, err := extractKubeconfigFromTI(path)
	if err != nil {
		return nil, err
	}

	file, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cleanup := func() error {
		return os.Remove(file.Name())
	}

	if _, err := file.Write(kcfg); err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("set optional kubeconfig TI by changing %s env", clientcmd.RecommendedConfigPathEnvVar),
		zap.String("from", os.Getenv(clientcmd.RecommendedConfigPathEnvVar)),
		zap.String("to", file.Name()),
	)

	// override the env for current process
	if err = os.Setenv(clientcmd.RecommendedConfigPathEnvVar, file.Name()); err != nil {
		return nil, errors.Wrapf(err, "while setting up %s env", clientcmd.RecommendedConfigPathEnvVar)
	}

	return cleanup, err
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
