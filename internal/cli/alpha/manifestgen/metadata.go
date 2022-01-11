package manifestgen

import (
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"sigs.k8s.io/yaml"
)

// Maintainer holds metadata information about a maintainer.
type Maintainer struct {
	// Email is a email of the maintainer.
	Email string `yaml:"email"`
	// Name is a name of the maintainer.
	Name string `yaml:"name"`
	// URL is a Maintainer URL.
	URL string `yaml:"url"`
}

// MetaDataInfo holds metadata information about manifest file.
type MetaDataInfo struct {
	Name             string        `yaml:"name"`
	Prefix           string        `yaml:"prefix"`
	DocumentationURL string        `yaml:"documentationURL"`
	SupportURL       string        `yaml:"supportURL"`
	IconURL          string        `yaml:"iconURL"`
	License          types.License `yaml:"license"`
	Maintainers      []Maintainer  `yaml:"maintainers"`
}

// Metadata holds generic metadata information for Capact manifests.
type Metadata struct {
	OCFVersion types.OCFVersion   `yaml:"ocfVersion"`
	Kind       types.ManifestKind `yaml:"kind"`
	Metadata   MetaDataInfo       `yaml:"metadata"`
}

// unmarshalMetadata reads the manifest metadata from a bytes slice of a Capact manifest.
func unmarshalMetadata(yamlBytes []byte) (Metadata, error) {
	mm := Metadata{}
	err := yaml.Unmarshal(yamlBytes, &mm)
	if err != nil {
		return mm, err
	}
	return mm, nil
}
