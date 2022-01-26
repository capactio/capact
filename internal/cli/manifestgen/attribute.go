package manifestgen

import (
	"github.com/pkg/errors"
)

// GenerateAttributeTemplatingConfig generates an attribute templating config.
func GenerateAttributeTemplatingConfig(cfg *AttributeConfig) (ManifestCollection, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestRef.Path)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	tc := &templatingConfig{
		Template: attributeManifestTemplate,
		Input: &attributeTemplatingInput{
			templatingInput: templatingInput{
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRef.Revision,
			},
			Metadata: cfg.Metadata,
		},
	}

	generated, err := generateManifests([]*templatingConfig{tc})
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	return createManifestCollection(generated)
}
