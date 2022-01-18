package manifestgen

import (
	"encoding/json"
	"fmt"
	"strings"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/alecthomas/jsonschema"
	"github.com/pkg/errors"
)

type genManifestFn func(cfg *InterfaceConfig) (*templatingConfig, error)

// GenerateInterfaceManifests generates collection of manifests for a new Interface.
func GenerateInterfaceManifests(cfg *InterfaceConfig) (ManifestCollection, error) {
	cfgs := make([]*templatingConfig, 0, 4)

	interfaceGroupCfg, err := getInterfaceGroupTemplatingConfig(&InterfaceGroupConfig{
		Config: cfg.Config,
	})
	if err != nil {
		return nil, errors.Wrap(err, "while getting InterfaceGroup templating config")
	}
	cfgs = append(cfgs, interfaceGroupCfg)

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

	trimmedInterfacePath := strings.TrimPrefix(cfg.ManifestRef.Path, "cap.interface.")
	outputsuffix := strings.Split(trimmedInterfacePath, ".")
	pathWithoutLastName := strings.Join(outputsuffix[:len(outputsuffix)-1], ".")

	cfg.InputTypeRef = types.ManifestRef{
		Path:     common.CreateManifestPath(types.TypeManifestKind, trimmedInterfacePath+"-input"),
		Revision: cfg.ManifestRef.Revision,
	}

	cfg.OutputTypeRef = types.ManifestRef{
		Path:     common.CreateManifestPath(types.TypeManifestKind, pathWithoutLastName) + ".config",
		Revision: cfg.ManifestRef.Revision,
	}

	interfaceCfg, err := getInterfaceTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Interface templating config")
	}
	cfgs = append(cfgs, interfaceCfg)

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

// GenerateInterfaceGroupTemplatingConfigFromInterfacePath generates InterfaceGroup templating config from interface config.
func GenerateInterfaceGroupTemplatingConfigFromInterfacePath(cfg *InterfaceGroupConfig) (ManifestCollection, error) {
	cfg.ManifestRef.Path = getInterfaceGroupPathFromInterfacePath(cfg.ManifestRef.Path)

	interfaceCfg, err := getInterfaceGroupTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Interface templating config")
	}
	generated, err := generateManifests([]*templatingConfig{interfaceCfg})
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	return createManifestCollection(generated)
}

// GenerateInterfaceGroupTemplatingConfig generates InterfaceGroup templating config.
func GenerateInterfaceGroupTemplatingConfig(cfg *InterfaceGroupConfig) (ManifestCollection, error) {
	interfaceCfg, err := getInterfaceGroupTemplatingConfig(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Interface templating config")
	}
	generated, err := generateManifests([]*templatingConfig{interfaceCfg})
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	return createManifestCollection(generated)
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

func getInterfaceGroupTemplatingConfig(cfg *InterfaceGroupConfig) (*templatingConfig, error) {
	groupPrefix, groupName, err := splitPathToPrefixAndName(cfg.ManifestRef.Path)
	if err != nil {
		return nil, errors.Wrap(err, "while getting InterfaceGroup prefix and path")
	}

	return &templatingConfig{
		Template: interfaceGroupManifestTemplate,
		Input: &interfaceGroupTemplatingInput{
			templatingInput: templatingInput{
				Name:     groupName,
				Prefix:   groupPrefix,
				Revision: cfg.ManifestRef.Revision,
			},
			Metadata: cfg.Metadata,
		},
	}, nil
}

func getInterfaceTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestRef.Path)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	return &templatingConfig{
		Template: interfaceManifestTemplate,
		Input: &interfaceTemplatingInput{
			templatingInput: templatingInput{
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRef.Revision,
			},
			Metadata: cfg.Metadata,
			InputRef: types.ManifestRef{
				Path:     cfg.InputTypeRef.Path,
				Revision: cfg.InputTypeRef.Revision,
			},
			OutputRef: types.ManifestRef{
				Path:     cfg.OutputTypeRef.Path,
				Revision: cfg.OutputTypeRef.Revision,
			},
		},
	}, nil
}

func getInterfaceTypeTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestRef.Path)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	jsonSchema, err := getInterfaceInputTypeJSONSchema()
	if err != nil {
		return nil, errors.Wrap(err, "while getting input type JSON schema")
	}

	typeMetadata := types.TypeMetadata{
		DocumentationURL: cfg.Metadata.DocumentationURL,
		IconURL:          cfg.Metadata.IconURL,
		SupportURL:       cfg.Metadata.SupportURL,
		Maintainers:      cfg.Metadata.Maintainers,
		DisplayName:      ptr.String(fmt.Sprintf("Type %s.%s", prefix, name)),
		Description:      "Description of the Type",
	}

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input: &typeTemplatingInput{
			templatingInput: templatingInput{
				Name:     name,
				Prefix:   prefix,
				Revision: cfg.ManifestRef.Revision,
			},
			Metadata:   typeMetadata,
			JSONSchema: string(jsonSchema),
		},
	}, nil
}

func getInterfaceInputTypeTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestRef.Path)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	jsonSchema, err := getInterfaceInputTypeJSONSchema()
	if err != nil {
		return nil, errors.Wrap(err, "while getting input type JSON schema")
	}

	typeMetadata := types.TypeMetadata{
		DocumentationURL: cfg.Metadata.DocumentationURL,
		IconURL:          cfg.Metadata.IconURL,
		SupportURL:       cfg.Metadata.SupportURL,
		Maintainers:      cfg.Metadata.Maintainers,
		DisplayName:      ptr.String(fmt.Sprintf("Input for %s.%s", prefix, name)),
		Description:      fmt.Sprintf("Input for the \"%s.%s Action\"", prefix, name),
	}

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input: &typeTemplatingInput{
			templatingInput: templatingInput{
				Name:     getDefaultInputTypeName(name),
				Prefix:   prefix,
				Revision: cfg.ManifestRef.Revision,
			},
			Metadata:   typeMetadata,
			JSONSchema: string(jsonSchema),
		},
	}, nil
}

func getInterfaceOutputTypeTemplatingConfig(cfg *InterfaceConfig) (*templatingConfig, error) {
	prefix, _, err := splitPathToPrefixAndName(cfg.ManifestRef.Path)
	if err != nil {
		return nil, errors.Wrap(err, "while getting path and prefix for manifests")
	}

	typeMetadata := types.TypeMetadata{
		DocumentationURL: cfg.Metadata.DocumentationURL,
		IconURL:          cfg.Metadata.IconURL,
		SupportURL:       cfg.Metadata.SupportURL,
		Maintainers:      cfg.Metadata.Maintainers,
		DisplayName:      ptr.String(fmt.Sprintf("%s config", prefix)),
		Description:      fmt.Sprintf("Type representing a %s config", prefix),
	}

	return &templatingConfig{
		Template: typeManifestTemplate,
		Input: &typeTemplatingInput{
			templatingInput: templatingInput{
				Name:     "config",
				Prefix:   prefix,
				Revision: cfg.ManifestRef.Revision,
			},
			Metadata: typeMetadata,
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
