package runner

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// DefaultFilePermissions are the default file permissions
// of the output artifact files created by the runners.
const DefaultFilePermissions = 0644

// SaveToFile saves the bytes to a file under the path.
func SaveToFile(path string, bytes []byte) error {
	err := ioutil.WriteFile(path, bytes, DefaultFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "while writing file to %q", path)
	}
	return nil
}

// NestingOutputUnderValue write output to "value" key in YAML.
func NestingOutputUnderValue(output []byte) ([]byte, error) {
	unmarshalled := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(output), &unmarshalled)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling output to map[string]interface{}")
	}

	nestingOutputValue := map[string]interface{}{
		"value": unmarshalled,
	}
	result, err := yaml.Marshal(&nestingOutputValue)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling output under a value key")
	}
	return result, nil
}
