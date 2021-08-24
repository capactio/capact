package policy

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// FromYAMLString reads the Policy from the input string.
func FromYAMLString(in string) (Policy, error) {
	bytes := []byte(in)
	var policy Policy
	if err := yaml.Unmarshal(bytes, &policy); err != nil {
		return Policy{}, errors.Wrap(err, "while unmarshalling Policy from YAML bytes")
	}

	return policy, nil
}
