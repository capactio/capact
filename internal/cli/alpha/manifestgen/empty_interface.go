package manifestgen

import (
	"strings"

	"github.com/pkg/errors"
)

// GenerateEmptyManifests generates collection of manifest to be filled by the content developer.
func GenerateEmptyManifests(cfg *EmptyImplementationConfig) (ManifestCollection, error) {
	cfgs := make([]*templatingConfig, 0, 2)

	inputTypeCfg, err := getEmptyInputTypeTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Implementation templating config")
	}
	cfgs = append(cfgs, inputTypeCfg)

	implCfg, err := getEmptyImplementationTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Implementation templating config")
	}
	cfgs = append(cfgs, implCfg)

	generated, err := generateManifests(cfgs)
	if err != nil {
		return nil, errors.Wrap(err, "while generating empty Implementation manifests")
	}

	return createManifestCollection(generated)
}

func getEmptyImplementationTemplatingConfig(cfg *EmptyImplementationConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting prefix and path for manifests")
	}

	var (
		interfacePath     = cfg.InterfacePathWithRevision
		interfaceRevision = "0.1.0"
	)

	pathSlice := strings.Split(cfg.InterfacePathWithRevision, ":")
	if len(pathSlice) == 2 {
		interfacePath = pathSlice[0]
		interfaceRevision = pathSlice[1]
	}

	input := &emptyImplementationTemplatingInput{
		templatingInput: templatingInput{
			Metadata: cfg.ManifestMetadata,
			Name:     name,
			Prefix:   prefix,
			Revision: cfg.ManifestRevision,
		},
		InterfacePath:     interfacePath,
		InterfaceRevision: interfaceRevision,
	}

	return &templatingConfig{
		Template: emptyImplementationManifestTemplate,
		Input:    input,
	}, nil
}

func getEmptyInputTypeTemplatingConfig(cfg *EmptyImplementationConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting prefix and path for manifests")
	}

	input := &typeTemplatingInput{
		templatingInput: templatingInput{
			Metadata: cfg.ManifestMetadata,
			Name:     name,
			Prefix:   prefix,
			Revision: cfg.ManifestRevision,
		},
	}

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input:    input,
	}, nil
}
