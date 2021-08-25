package manifestgen

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/alecthomas/jsonschema"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"
)

// GenerateTerraformManifests generates manifest files for a Terraform module based Implementation
func GenerateTerraformManifests(cfg *TerraformConfig) (map[string]string, error) {
	module, diags := tfconfig.LoadModule(cfg.ModulePath)
	if diags.Err() != nil {
		return nil, errors.Wrap(diags.Err(), "while loading Terraform module")
	}

	cfgs := make([]*templatingConfig, 0, 2)

	inputTypeCfg, err := getTerraformInputTypeTemplatingConfig(cfg, module)
	if err != nil {
		return nil, errors.Wrap(err, "while getting input Type templating config")
	}
	cfgs = append(cfgs, inputTypeCfg)

	implCfg, err := getTerraformImplementationTemplatingConfig(cfg, module)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Implementation templating config")
	}
	cfgs = append(cfgs, implCfg)

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

func getTerraformInputTypeTemplatingConfig(cfg *TerraformConfig, module *tfconfig.Module) (*templatingConfig, error) {
	prefix, name, err := splitPathToPrefixAndName(cfg.ManifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "while getting prefix and path for manifests")
	}

	jsonSchema, err := getTerraformInputTypeJSONSchema(module.Variables)
	if err != nil {
		return nil, errors.Wrap(err, "while getting input type JSON Schema")
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

func getTerraformImplementationTemplatingConfig(cfg *TerraformConfig, module *tfconfig.Module) (*templatingConfig, error) {
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

	input := &terraformImplementationTemplatingInput{
		templatingInput: templatingInput{
			Name:     name,
			Prefix:   prefix,
			Revision: cfg.ManifestRevision,
		},
		InterfacePath:     interfacePath,
		InterfaceRevision: interfaceRevision,
		ModuleSourceURL:   cfg.ModuleSourceURL,
		Provider:          cfg.Provider,
		Outputs:           make([]*tfconfig.Output, 0, len(module.Outputs)),
		Variables:         make([]*tfconfig.Variable, 0, len(module.Variables)),
	}

	for i := range module.Variables {
		input.Variables = append(input.Variables, module.Variables[i])
	}

	sort.Slice(input.Variables, func(i, j int) bool {
		return input.Variables[i].Name < input.Variables[j].Name
	})

	for i := range module.Outputs {
		input.Outputs = append(input.Outputs, module.Outputs[i])
	}

	sort.Slice(input.Outputs, func(i, j int) bool {
		return input.Outputs[i].Name < input.Outputs[j].Name
	})

	return &templatingConfig{
		Template: terraformImplementationManifestTemplate,
		Input:    input,
	}, nil
}

func getTerraformInputTypeJSONSchema(variables map[string]*tfconfig.Variable) ([]byte, error) {
	schema := &jsonschema.Type{
		Title:      "",
		Properties: orderedmap.New(),
	}

	for _, value := range variables {
		propSchema := &jsonschema.Type{
			Title:       value.Name,
			Type:        getTypeFromTerraformType(value.Type),
			Description: value.Description,
		}
		schema.Properties.Set(value.Name, propSchema)
	}

	schema.Properties.Sort(func(a, b *orderedmap.Pair) bool {
		return a.Key() < b.Key()
	})

	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling JSON schema")
	}

	return schemaBytes, nil
}

// Terraform types: https://www.terraform.io/docs/language/expressions/types.html
func getTypeFromTerraformType(t string) string {
	if strings.HasPrefix(t, "list") || strings.HasPrefix(t, "tuple") {
		return "array"
	}

	switch t {
	case "string":
		return "string"
	case "number":
		return "number"
	case "bool":
		return "boolean"
	case "null":
		return "null"
	}

	return "object"
}
