package manifestgen

import (
	"fmt"

	"github.com/pkg/errors"
)

// GenerateAttributeTemplatingConfig generates an attribute templating config.
func GenerateAttributeTemplatingConfig(cfg *AttributeConfig) (map[string]string, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	tc := &templatingConfig{
		Template: attributeManifestTemplate,
		Input: &attributeTemplatingInput{
			templatingInput: templatingInput{
				Metadata: cfg.ManifestMetadata,
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRevision,
			},
		},
	}

	generated, err := generateManifests([]*templatingConfig{tc})
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	result := make(map[string]string, len(generated))

	for _, m := range generated {
		metadata, err := unmarshalMetadata([]byte(m))
		if err != nil {
			return nil, errors.Wrap(err, "while getting metadata for manifest")
		}
		manifestPath := fmt.Sprintf("%s.%s", metadata.Metadata.Prefix, metadata.Metadata.Name)
		result[manifestPath] = m
	}

	return result, nil
}
