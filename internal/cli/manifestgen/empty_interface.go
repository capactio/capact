package manifestgen

import (
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

// GenerateEmptyManifests generates collection of manifest to be filled by the content developer.
func GenerateEmptyManifests(cfg *EmptyImplementationConfig) (ManifestCollection, error) {
	cfgs := make([]*templatingConfig, 0, 2)

	additionalInputTypeCfg, err := getEmptyAdditionalInputTypeTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Implementation templating config")
	}
	cfgs = append(cfgs, additionalInputTypeCfg)

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
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestRef.Path)
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
			Name:     name,
			Prefix:   prefix,
			Revision: cfg.ManifestRef.Revision,
		},
		Metadata:            cfg.Metadata,
		AdditionalInputName: cfg.AdditionalInputTypeName,
		InterfaceRef: types.ManifestRef{
			Path:     interfacePath,
			Revision: interfaceRevision,
		},
	}

	return &templatingConfig{
		Template: emptyImplementationManifestTemplate,
		Input:    input,
	}, nil
}

func getEmptyAdditionalInputTypeTemplatingConfig(cfg *EmptyImplementationConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestRef.Path)
	if err != nil {
		return nil, errors.Wrap(err, "while getting prefix and path for manifests")
	}

	typeMetadata := types.TypeMetadata{
		DocumentationURL: cfg.Metadata.DocumentationURL,
		IconURL:          cfg.Metadata.IconURL,
		SupportURL:       cfg.Metadata.SupportURL,
		Maintainers:      cfg.Metadata.Maintainers,
		DisplayName:      ptr.String(fmt.Sprintf("Additional input for %s", name)),
		Description:      fmt.Sprintf("Additional input for the \"%s Action\"", name),
	}

	input := &typeTemplatingInput{
		templatingInput: templatingInput{
			Name:     cfg.AdditionalInputTypeName,
			Prefix:   prefix,
			Revision: cfg.ManifestRef.Revision,
		},
		Metadata: typeMetadata,
	}

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input:    input,
	}, nil
}
