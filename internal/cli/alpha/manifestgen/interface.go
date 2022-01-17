package manifestgen

import (
	"encoding/json"
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	"github.com/alecthomas/jsonschema"
	"github.com/pkg/errors"
)

type genManifestFn func(cfg *InterfaceConfig) (*templatingConfig, error)

// GenerateInterfaceManifests generates collection of manifests for a new Interface.
func GenerateInterfaceManifests(cfg *InterfaceConfig) (ManifestCollection, error) {
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

	return createManifestCollection(generated)
}

// GenerateInterfaceTemplatingConfig generates Interface templating config.
func GenerateInterfaceTemplatingConfig(cfg *InterfaceConfig) (ManifestCollection, error) {
	return generateManifestCollection(cfg, []genManifestFn{getInterfaceTemplatingConfig})
}

// GenerateInterfaceGroupTemplatingConfigFromInterfaceCfg generates InterfaceGroup templating config from interface config.
func GenerateInterfaceGroupTemplatingConfigFromInterfaceCfg(cfg *InterfaceConfig) (ManifestCollection, error) {
	cfg.ManifestPath = getInterfaceGroupPathFromInterfacePath(cfg.ManifestPath)
	return generateManifestCollection(cfg, []genManifestFn{getInterfaceGroupTemplatingConfig})
}

// GenerateInterfaceGroupTemplatingConfig generates InterfaceGroup templating config.
func GenerateInterfaceGroupTemplatingConfig(cfg *InterfaceConfig) (ManifestCollection, error) {
	// TODO: created dedicated InterfaceGroupConfig
	return generateManifestCollection(cfg, []genManifestFn{getInterfaceGroupTemplatingConfig})
}

// GenerateTypeTemplatingConfig generates Type templating config.
func GenerateTypeTemplatingConfig(cfg *InterfaceConfig) (ManifestCollection, error) {
	return generateManifestCollection(cfg, []genManifestFn{getInterfaceTypeTemplatingConfig})
}

// GenerateInputTypeTemplatingConfig generates Input Type templating config.
func GenerateInputTypeTemplatingConfig(cfg *InterfaceConfig) (ManifestCollection, error) {
	return generateManifestCollection(cfg, []genManifestFn{getInterfaceInputTypeTemplatingConfig})
}

// GenerateOutputTypeTemplatingConfig generates Output Type templating config.
func GenerateOutputTypeTemplatingConfig(cfg *InterfaceConfig) (ManifestCollection, error) {
	return generateManifestCollection(cfg, []genManifestFn{getInterfaceOutputTypeTemplatingConfig})
}

func generateManifestCollection(cfg *InterfaceConfig, fnList []genManifestFn) (ManifestCollection, error) {
	cfgs := make([]*templatingConfig, 0, 4)
	for _, fn := range fnList {
		interfaceCfg, err := fn(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "while getting Interface templating config")
		}
		cfgs = append(cfgs, interfaceCfg)
	}
	generated, err := generateManifests(cfgs)
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	return createManifestCollection(generated)
}

func getInterfaceGroupTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	groupPrefix, groupName, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting InterfaceGroup prefix and path")
	}

	return &templatingConfig{
		Template: interfaceGroupManifestTemplate,
		Input: &interfaceGroupTemplatingInput{
			templatingInput: templatingInput{
				Metadata: cfg.ManifestMetadata,
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

	var inputPath, inputRevision, outputPath, outputRevision string

	inputPathSlice := strings.SplitN(cfg.InputPathWithRevision, ":", 2)
	if len(inputPathSlice) == 2 {
		inputPath = inputPathSlice[0]
		inputRevision = inputPathSlice[1]
	}

	outputPathSlice := strings.SplitN(cfg.OutputPathWithRevision, ":", 2)
	if len(outputPathSlice) == 2 {
		outputPath = outputPathSlice[0]
		outputRevision = outputPathSlice[1]
	}

	return &templatingConfig{
		Template: interfaceManifestTemplate,
		Input: &interfaceTemplatingInput{
			templatingInput: templatingInput{
				Metadata: cfg.ManifestMetadata,
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRevision,
			},
			InputTypeName:      inputPath,
			InputTypeRevision:  inputRevision,
			OutputTypeName:     outputPath,
			OutputTypeRevision: outputRevision,
		},
	}, nil
}

func getInterfaceTypeTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	jsonSchema, err := getInterfaceInputTypeJSONSchema()
	if err != nil {
		return nil, errors.Wrap(err, "while getting input type JSON schema")
	}

	cfg.ManifestMetadata.DisplayName = ptr.String(fmt.Sprintf("Type %s.%s", prefix, name))
	cfg.ManifestMetadata.Description = "Description of the Type"

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input: &typeTemplatingInput{
			templatingInput: templatingInput{
				Metadata: cfg.ManifestMetadata,
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRevision,
			},
			JSONSchema: string(jsonSchema),
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

	cfg.ManifestMetadata.DisplayName = ptr.String(fmt.Sprintf("Input for %s.%s", prefix, name))
	cfg.ManifestMetadata.Description = fmt.Sprintf("Input for the \"%s.%s Action\"", prefix, name)

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input: &typeTemplatingInput{
			templatingInput: templatingInput{
				Metadata: cfg.ManifestMetadata,
				Name:     getDefaultInputTypeName(name),
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

	cfg.ManifestMetadata.DisplayName = ptr.String(fmt.Sprintf("%s config", prefix))
	cfg.ManifestMetadata.Description = fmt.Sprintf("Type representing a %s config", prefix)

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input: &typeTemplatingInput{
			templatingInput: templatingInput{
				Metadata: cfg.ManifestMetadata,
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
