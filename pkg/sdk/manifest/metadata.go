package manifest

import (
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// UnmarshalMetadata loads essential manifest metadata (kind and OCF version) from YAML bytes.
func UnmarshalMetadata(yamlBytes []byte) (types.ManifestMetadata, error) {
	mm := types.ManifestMetadata{}
	err := yaml.Unmarshal(yamlBytes, &mm)
	if err != nil {
		return mm, err
	}

	if mm.OCFVersion == "" || mm.Kind == "" {
		return mm, errors.New("OCFVersion and Kind must not be empty")
	}

	return mm, nil
}
