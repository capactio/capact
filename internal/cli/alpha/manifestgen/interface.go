package manifestgen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/jsonschema"
	"github.com/pkg/errors"
)

// GenerateInterfaceManifests generates manifest files for a new Interface.
func GenerateInterfaceManifests(cfg *InterfaceConfig) (map[string]string, error) {
	cfgs := make([]*templatingConfig, 0, 4)

	interfaceGroupCfg, err := getInterfaceGroupTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting InterfaceGroup templating config")
	}
	cfgs = append(cfgs, interfaceGroupCfg)

	interfaceCfg, err := getInterfaceTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Interface templating config")
	}
	cfgs = append(cfgs, interfaceCfg)

	inputTypeCfg, err := getInterfaceInputTypeTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Interface input Type templating config")
	}
	cfgs = append(cfgs, inputTypeCfg)

	outputTypeCfg, err := getInterfaceOutputTypeTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Interface output Type templating config")
	}
	cfgs = append(cfgs, outputTypeCfg)

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

func getInterfaceGroupTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	interfaceGroupPath := getInterfaceGroupPathFromInterfacePath(cfg.ManifestPath)
	groupPrefix, groupName, err := splitPathToPrefixAndName(interfaceGroupPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting InterfaceGroup prefix and path")
	}

	return &templatingConfig{
		Template: interfaceGroupManifestTemplate,
		Input: &interfaceGroupTemplatingInput{
			templatingInput: templatingInput{
				Name:     groupName,
				Prefix:   groupPrefix,
				Revision: cfg.ManifestRevision,
			},
		},
	}, nil
}

func getInterfaceTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	return &templatingConfig{
		Template: interfaceManifestTemplate,
		Input: &interfaceTemplatingInput{
			templatingInput: templatingInput{
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRevision,
			},
		},
	}, nil
}

func getInterfaceInputTypeTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	jsonSchema, err := getInterfaceInputTypeJSONSchema()
	if err != nil {
		return nil, errors.Wrap(err, "while getting input type JSON schema")
	}

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input: &typeTemplatingInput{
			templatingInput: templatingInput{
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRevision,
			},
			JSONSchema: string(jsonSchema),
		},
	}, nil
}

func getInterfaceOutputTypeTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	return &templatingConfig{
		Template: outputTypeManifestTemplate,
		Input: &outputTypeTemplatingInput{
			templatingInput: templatingInput{
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRevision,
			},
		},
	}, nil
}

func getInterfaceInputTypeJSONSchema() ([]byte, error) {
	schema := &jsonschema.Type{
		Type: "object",
	}

	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling JSON schema")
	}

	return schemaBytes, nil
}

func getInterfaceGroupPathFromInterfacePath(ifacePath string) string {
	parts := strings.Split(ifacePath, ".")
	return strings.Join(parts[:len(parts)-1], ".")
}
