package manifest

import (
	"context"
	"encoding/json"
	"fmt"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/pkg/errors"
)

// InterfaceValidator is a validator for Interface manifest, which executes static validation.
type InterfaceValidator struct{}

// NewInterfaceValidator creates new ImplementationValidator.
func NewInterfaceValidator() *InterfaceValidator {
	return &InterfaceValidator{}
}

// Do is a method which triggers the validation.
func (v *InterfaceValidator) Do(_ context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	var entity types.Interface
	err := json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Interface type")
	}

	if entity.Spec.Input.Parameters == nil {
		return ValidationResult{}, nil
	}

	toValidate := jsonSchemaCollection{}
	for name, param := range entity.Spec.Input.Parameters.ParametersParameterMap {
		if param.JSONSchema == nil {
			continue
		}

		key := fmt.Sprintf("spec.input.parameters.%s.jsonSchema.value", name)
		toValidate[key] = param.JSONSchema.Value
	}

	return validateJSONSchema07Definition(toValidate)
}

// Name returns the validator name.
func (v *InterfaceValidator) Name() string {
	return "InterfaceValidator"
}
