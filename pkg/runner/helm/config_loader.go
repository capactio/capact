package helm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"
)

// KubeconfigInput defines type under which the kubeconfig is stored in TypeInstance.
type KubeconfigInput struct {
	Value KubeconfigContent `yaml:"value"`
}

// KubeconfigContent defines type for kubeconfig TypeInstance value content.
type KubeconfigContent struct {
	Config map[string]interface{} `yaml:"config"`
}

func (r *helmRunner) loadKubeconfig() (*rest.Config, error) {
	err := setKubeconfigEnvIfTypeInstanceExists(r.cfg.OptionalKubeconfigTI, r.log)
	if err != nil {
		return nil, errors.Wrap(err, "while setting optional kubeconfig")
	}

	return config.GetConfig()
}

func setKubeconfigEnvIfTypeInstanceExists(path string, log *zap.Logger) error {
	if path == "" {
		log.Debug("optional Kubeconfig TI not specified")
		return nil
	}

	f, err := os.Stat(path)
	switch {
	case err == nil:
		if f.IsDir() {
			return errors.New("RUNNER_OPTIONAL_KUBECONFIG_TI cannot be dir, must be a file")
		}
	case os.IsNotExist(err):
		log.Debug("optional Kubeconfig TI specified but file does not exist")
		return nil
	default:
		return err
	}

	kcfg, err := extractKubeconfigFromTI(path)
	if err != nil {
		return err
	}

	return SetNewKubeconfig(kcfg, log)
}

// SetNewKubeconfig creates a temporary kubeconfig file from bytes and sets environment variable to that file.
func SetNewKubeconfig(kubeconfig []byte, log *zap.Logger) error {
	file, err := ioutil.TempFile("", "kubeconfig")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(kubeconfig); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("set optional kubeconfig TI by changing %s env", clientcmd.RecommendedConfigPathEnvVar),
		zap.String("from", os.Getenv(clientcmd.RecommendedConfigPathEnvVar)),
		zap.String("to", file.Name()),
	)

	// override the env for current process
	if err = os.Setenv(clientcmd.RecommendedConfigPathEnvVar, file.Name()); err != nil {
		return errors.Wrapf(err, "while setting up %s env", clientcmd.RecommendedConfigPathEnvVar)
	}
	return nil
}

func extractKubeconfigFromTI(path string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, errors.Wrapf(err, "while reading Kubeconfig TypeInstance path")
	}
	kubeconfigInput := KubeconfigInput{}
	if err := yaml.Unmarshal(data, &kubeconfigInput); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling Kubeconfig TypeInstance")
	}

	kcfg, err := yaml.Marshal(kubeconfigInput.Value.Config)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshaling Kubeconfig TypeInstance")
	}

	return kcfg, nil
}
