package clusterpolicy

import (
	"sort"

	"github.com/Masterminds/semver/v3"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

const supportedAPIVersionConstraintString = "^0.2"

type SupportedAPIVersionMap map[string]struct{}

func (m SupportedAPIVersionMap) ToStringSlice() []string {
	var strSlice []string
	for key := range m {
		strSlice = append(strSlice, key)
	}

	sort.Strings(strSlice)
	return strSlice
}

func FromYAMLString(in string) (ClusterPolicy, error) {
	bytes := []byte(in)
	err := Validate(bytes)
	if err != nil {
		return ClusterPolicy{}, err
	}

	var policy ClusterPolicy
	if err := yaml.Unmarshal(bytes, &policy); err != nil {
		return ClusterPolicy{}, errors.Wrap(err, "while unmarshalling policy from YAML bytes")
	}

	return policy, nil
}

func Validate(in []byte) error {
	var unmarshalled struct {
		APIVersion semver.Version `json:"apiVersion"`
	}

	if err := yaml.Unmarshal(in, &unmarshalled); err != nil {
		return errors.Wrap(err, "while unmarshalling policy to validate API version")
	}

	constraints, err := semver.NewConstraint(supportedAPIVersionConstraintString)
	if err != nil {
		return errors.Wrap(err, "while parsing SemVer constraints")
	}

	_, errs := constraints.Validate(&unmarshalled.APIVersion)
	if len(errs) > 0 {
		return NewUnsupportedAPIVersionError(errs)
	}

	return nil
}
