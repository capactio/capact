package manifest

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// jsonSchemaCollection defines JSONSchema collection index by the name.
type jsonSchemaCollection map[string]string

// checkJSONSchema07Definition validate a given JSONSchema collection.
// Fast return on internal error, otherwise returns aggregated ValidationResult for all schemas.
func checkJSONSchema07Definition(schemas jsonSchemaCollection) (ValidationResult, error) {
	result := ValidationResult{}

	schemaLoader := gojsonschema.NewReferenceLoader("http://json-schema.org/draft-07/schema")
	schemaDraft07, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return ValidationResult{}, err
	}

	for name, schema := range schemas {
		if err := toJSON(schema); err != nil {
			result.Errors = append(result.Errors, errors.Wrapf(err, "%s: invalid JSON", name))
			continue
		}

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

// toJSON is used to check whether a given string as whole is a valid JSON. It's necessary as the
// gojsonschema.NewStringLoader uses JSON decoder:
//    "Decode reads the next JSON-encoded value from its input (..)"
// As we want to have only one JSON, only the first entry is decoded.
// It allows to have malformed data appended to the first JSON entry and if the input data
// is unmarshalled as a single content, the JSON may be invalid.
func toJSON(str string) error {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js)
}
