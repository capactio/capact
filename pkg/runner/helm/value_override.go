package helm

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

func readValueOverrides(inlineValues map[string]interface{}, valuesFilePath string) (map[string]interface{}, error) {
	var values map[string]interface{}

	if valuesFilePath == "" {
		return inlineValues, nil
	}

	if len(inlineValues) > 0 && valuesFilePath != "" {
		return nil, errors.New("providing values both inline and from file is currently unsupported")
	}

	bytes, err := ioutil.ReadFile(filepath.Clean(valuesFilePath))
	if err != nil {
		return nil, errors.Wrapf(err, "while reading values from file %q", valuesFilePath)
	}

	if err := yaml.Unmarshal(bytes, &values); err != nil {
		return nil, errors.Wrapf(err, "while parsing %q", valuesFilePath)
	}

	return values, nil
}
