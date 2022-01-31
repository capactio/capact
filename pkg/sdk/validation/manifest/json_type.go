package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/pkg/errors"
)

const (
	coreTypePrefix   = "cap.core.type."
	customTypePrefix = "cap.type."
)

// TypeValidator is a validator for Type manifest.
type TypeValidator struct{}

// NewTypeValidator creates new TypeValidator.
func NewTypeValidator() *TypeValidator {
	return &TypeValidator{}
}

// Do is a method which triggers the validation.
func (v *TypeValidator) Do(_ context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	var entity types.Type
	err := json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Type type")
	}

	var resNodes []error
	for _, ref := range entity.Spec.AdditionalRefs {
		if strings.HasPrefix(ref, coreTypePrefix) || strings.HasPrefix(ref, customTypePrefix) {
			continue
		}
		resNodes = append(resNodes, fmt.Errorf("spec.additionalRefs: %q is not allowed. It can refer only to a parent node under %q or %q", ref, coreTypePrefix, customTypePrefix))
	}

	resSchema, err := checkJSONSchema07Definition(jsonSchemaCollection{
		"spec.jsonSchema.value": entity.Spec.JSONSchema.Value,
	})
	if err != nil {
		return ValidationResult{}, err
	}
	return ValidationResult{
		Errors: append(resNodes, resSchema.Errors...),
	}, nil
}

// Name returns the validator name.
func (v *TypeValidator) Name() string {
	return "TypeValidator"
}
