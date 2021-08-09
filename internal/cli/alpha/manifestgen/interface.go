package manifestgen

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// InterfaceConfig stores the input parameters for Interface content generation
type InterfaceConfig struct {
	Config
}

// GenerateInterfaceManifests generates manifest files for a new Interface.
func GenerateInterfaceManifests(cfg *InterfaceConfig) (map[string]string, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	interfaceGroupPath := getInterfaceGroupPathFromInterfacePath(cfg.ManifestPath)
	groupPrefix, groupName, err := splitPathToPrefixAndName(interfaceGroupPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting InterfaceGroup prefix and path")
	}

	interfaceInput := &templatingInput{
		Name:     name,
		Prefix:   prefix,
		Revision: cfg.ManifestRevision,
	}

	interfaceGroupInput := &templatingInput{
		Name:     groupName,
		Prefix:   groupPrefix,
		Revision: cfg.ManifestRevision,
	}

	cfgs := []*templatingConfig{
		{
			Template: interfaceGroupManifestTemplate,
			Input:    interfaceGroupInput,
		},
		{
			Template: typeManifestTemplate,
			Input:    interfaceInput,
		},
		{
			Template: outputTypeManifestTemplate,
			Input:    interfaceInput,
		},
		{
			Template: interfaceManifestTemplate,
			Input:    interfaceInput,
		},
	}

	generated, err := generateManifests(cfgs)
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

func getInterfaceGroupPathFromInterfacePath(ifacePath string) string {
	parts := strings.Split(ifacePath, ".")
	return strings.Join(parts[:len(parts)-1], ".")
}
