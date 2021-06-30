package terraform

import (
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform/configs"
	"github.com/zclconf/go-cty/cty"
)

// LoadVariablesFromFiles loads and merges multiple files with Terraform variables.
// Variables from subsequent files are overriding the variables in files before.
func LoadVariablesFromFiles(paths ...string) (map[string]cty.Value, error) {
	p := configs.NewParser(nil)

	values := map[string]cty.Value{}

	for _, path := range paths {
		loadedValues, diag := p.LoadValuesFile(path)
		if diag.HasErrors() {
			return nil, diag.Errs()[0]
		}

		for k, v := range loadedValues {
			values[k] = v
		}
	}

	return values, nil
}

// MarshalVariables outputs the provided variables as a bytestream in HCL format.
func MarshalVariables(variables map[string]cty.Value) []byte {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for k, v := range variables {
		rootBody.SetAttributeValue(k, v)
	}

	return f.Bytes()
}
