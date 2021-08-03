package manifest

import "sigs.k8s.io/yaml"

// Metadata holds generic metadata information for Capact manifests
type Metadata struct {
	OCFVersion ocfVersion `yaml:"ocfVersion"`
	Kind       kind       `yaml:"kind"`
	Metadata   struct {
		Name   string `yaml:"name"`
		Prefix string `yaml:"prefix"`
	} `yaml:"metadata"`
}

// GetMetadata reads the manifest metadata from a byteslice of a Capact manifest
func GetMetadata(yamlBytes []byte) (Metadata, error) {
	mm := Metadata{}
	err := yaml.Unmarshal(yamlBytes, &mm)
	if err != nil {
		return mm, err
	}
	return mm, nil
}
