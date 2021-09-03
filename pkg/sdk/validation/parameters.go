package validation

import (
	"context"
	"strings"

	"capact.io/capact/internal/ctxutil"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

// ValidateParameters validates that a given input parameters are valid against JSONSchema defined in SchemaCollection.
func ValidateParameters(ctx context.Context, header string, paramsSchemas SchemaCollection, parameters types.ParametersCollection) (Result, error) {
	resultBldr := NewResultBuilder(header)

	// 1. Check that all required parameters are specified
	for name, schema := range paramsSchemas {
		val, found := parameters[name]
		if schema.Required && (!found || strings.TrimSpace(val) == "") {
			delete(parameters, name) // delete to don't validate against JSONSchema
			resultBldr.ReportIssue(name, "required but missing input parameters")
		}
	}

	// 2. Validate input parameters against JSONSchema
	for paramName, paramData := range parameters {
		if ctxutil.ShouldExit(ctx) { // validation may cause additional resource usage, so stop if not needed
			return nil, ctx.Err()
		}

		// Ensure that it's in JSON format.
		// It's not a problem if it's already a JSON.
		paramDataJSON, err := yaml.YAMLToJSON([]byte(paramData))
		if err != nil {
			return nil, err
		}

		schema, found := paramsSchemas[paramName]
		if !found {
			resultBldr.ReportIssue(paramName, "Unknown parameter. Cannot validate it against JSONSchema.")
			continue
		}

		schemaLoader := gojsonschema.NewStringLoader(schema.Value)
		dataLoader := gojsonschema.NewBytesLoader(paramDataJSON)

		result, err := gojsonschema.Validate(schemaLoader, dataLoader)
		if err != nil {
			return nil, err
		}

		if !result.Valid() {
			for _, err := range result.Errors() {
				resultBldr.ReportIssue(paramName, err.String())
			}
		}
	}

	return resultBldr.Result(), nil
}
