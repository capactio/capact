package manifest

import (
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// jsonSchemaCollection defines JSONSchema collection index by the name.
type jsonSchemaCollection map[string]string

// validateJSONSchema07Definition validate a given JSONSchema collection.
// Fast return on internal error, otherwise returns aggregated ValidationResult for all schemas.
func validateJSONSchema07Definition(schemas jsonSchemaCollection) (ValidationResult, error) {
	result := ValidationResult{}

	schemaLoader := gojsonschema.NewReferenceLoader("http://json-schema.org/draft-07/schema")
	schemaDraft07, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return ValidationResult{}, err
	}

	for name, schema := range schemas {
		manifestLoader := gojsonschema.NewStringLoader(schema)

		jsonSchemaValidationResult, err := schemaDraft07.Validate(manifestLoader)
		if err != nil {
			return newValidationResult(errors.Wrap(err, name)), nil
		}

		for _, err := range jsonSchemaValidationResult.Errors() {
			result.Errors = append(result.Errors, errors.New(err.String()))
		}
	}

	return result, nil
}
