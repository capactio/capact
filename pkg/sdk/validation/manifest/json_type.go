package manifest

import (
	"context"
	"encoding/json"
	"fmt"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// TypeValidator is a validator for Type manifest.
type TypeValidator struct{}

// NewTypeValidator creates new TypeValidator.
func NewTypeValidator() *TypeValidator {
	return &TypeValidator{}
}

// Do is a method which triggers the validation.
func (v *TypeValidator) Do(_ context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	var typeEntity types.Type
	err := json.Unmarshal(jsonBytes, &typeEntity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Type type")
	}

	jsonSchemaStr := typeEntity.Spec.JSONSchema.Value
	schemaLoader := gojsonschema.NewReferenceLoader("http://json-schema.org/draft-07/schema")
	manifestLoader := gojsonschema.NewStringLoader(jsonSchemaStr)

	jsonSchemaValidationResult, err := gojsonschema.Validate(schemaLoader, manifestLoader)
	if err != nil {
		return newValidationResult(errors.Wrap(err, "spec.jsonSchema.value")), nil
	}

	result := ValidationResult{}
	for _, err := range jsonSchemaValidationResult.Errors() {
		result.Errors = append(result.Errors, fmt.Errorf("%v", err.String()))
	}

	return result, nil
}

// Name returns the validator name.
func (v *TypeValidator) Name() string {
	return "TypeValidator"
}
