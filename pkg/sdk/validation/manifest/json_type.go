package manifest

import (
	"context"
	"encoding/json"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/pkg/errors"
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

	return validateJSONSchema07Definition(jsonSchemaCollection{
		"spec.jsonSchema.value": typeEntity.Spec.JSONSchema.Value,
	})
}

// Name returns the validator name.
func (v *TypeValidator) Name() string {
	return "TypeValidator"
}
