package manifestgen

import (
	"fmt"
	"sort"
	"strings"

	"capact.io/capact/pkg/sdk/manifest"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
)

// TerraformConfig stores input parameters for Terraform based content generation
type TerraformConfig struct {
	Config

	ModulePath                string
	ModuleSourceURL           string
	InterfacePathWithRevision string
	Provider                  Provider
}

type terraformTemplatingInput struct {
	templatingInput

	InterfacePath     string
	InterfaceRevision string
	ModuleSourceURL   string
	Outputs           []outputVariable
	Provider          Provider
}

// GenerateTerraformManifests generates manifest files for a Terraform module based Implementation
func GenerateTerraformManifests(cfg *TerraformConfig) (map[string]string, error) {
	input, err := getTerraformTemplatingInput(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while getting templating input")
	}

	cfgs := []*templatingConfig{
		{
			Template: typeManifestTemplate,
			Input:    input,
		},
		{
			Template: terraformImplementationManifestTemplate,
			Input:    input,
		},
	}

	generated, err := generateManifests(cfgs)
	if err != nil {
		return nil, errors.Wrap(err, "while generating manifests")
	}

	result := make(map[string]string, len(generated))

	for _, m := range generated {
		metadata, err := manifest.GetMetadata([]byte(m))
		if err != nil {
			return nil, errors.Wrap(err, "while getting metadata for manifest")
		}

		manifestPath := fmt.Sprintf("%s.%s", metadata.Metadata.Prefix, metadata.Metadata.Name)

		result[manifestPath] = m
	}

	return result, nil
}

func getTerraformTemplatingInput(cfg *TerraformConfig) (*terraformTemplatingInput, error) {
	module, diags := tfconfig.LoadModule(cfg.ModulePath)
	if diags.Err() != nil {
		return nil, errors.Wrap(diags.Err(), "while loading Terraform module")
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

	prefix, name := splitPathToPrefixAndName(cfg.ManifestPath)

	input := &terraformTemplatingInput{
		templatingInput: templatingInput{
			Name:      name,
			Prefix:    prefix,
			Revision:  cfg.ManifestRevision,
			Variables: make([]inputVariable, 0, len(module.Variables)),
		},
		InterfacePath:     interfacePath,
		InterfaceRevision: interfaceRevision,
		ModuleSourceURL:   cfg.ModuleSourceURL,
		Outputs:           make([]outputVariable, 0, len(module.Outputs)),
		Provider:          cfg.Provider,
	}

	for _, tfVar := range module.Variables {
		// Skip default for now, as there are problems, when it is a multiline string or with doublequotes in it.
		input.Variables = append(input.Variables, inputVariable{
			Name:        tfVar.Name,
			Type:        getTypeFromTerraformType(tfVar.Type),
			Description: tfVar.Description,
		})
	}

	sort.Slice(input.Variables, func(i, j int) bool {
		return input.Variables[i].Name < input.Variables[j].Name
	})

	for _, tfOut := range module.Outputs {
		input.Outputs = append(input.Outputs, outputVariable{
			Name: tfOut.Name,
		})
	}

	sort.Slice(input.Outputs, func(i, j int) bool {
		return input.Outputs[i].Name < input.Outputs[j].Name
	})

	return input, nil
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
