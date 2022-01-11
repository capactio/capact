package manifestgen

import (
	"github.com/pkg/errors"
)

// GenerateAttributeTemplatingConfig generates an attribute templating config.
func GenerateAttributeTemplatingConfig(cfg *AttributeConfig) (ManifestCollection, error) {
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

	return createManifestCollection(generated)
}
