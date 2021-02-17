package clusterpolicy

import (
	"sort"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

var SupportedAPIVersions = SupportedAPIVersionMap{
	"0.1.0": {},
}

type SupportedAPIVersionMap map[string]struct{}

func (m SupportedAPIVersionMap) ToStringSlice() []string {
	var strSlice []string
	for key := range m {
		strSlice = append(strSlice, key)
	}

	sort.Strings(strSlice)
	return strSlice
}

func FromYAMLBytes(in []byte) (ClusterPolicy, error) {
	err := Validate(in)
	if err != nil {
		return ClusterPolicy{}, err
	}

	var policy ClusterPolicy
	if err := yaml.Unmarshal(in, &policy); err != nil {
		return ClusterPolicy{}, errors.Wrap(err, "while unmarshalling policy from YAML bytes")
	}

	return policy, nil
}

// TODO: Use https://github.com/Masterminds/semver and validate only major and minor versions
func Validate(in []byte) error {
	var unmarshalled struct {
		APIVersion string `json:"apiVersion"`
	}

	if err := yaml.Unmarshal(in, &unmarshalled); err != nil {
		return errors.Wrap(err, "while unmarshalling policy to validate API version")
	}

	if _, ok := SupportedAPIVersions[unmarshalled.APIVersion]; !ok {
		return NewUnsupportedAPIVersionError(SupportedAPIVersions.ToStringSlice())
	}

	return nil
}
