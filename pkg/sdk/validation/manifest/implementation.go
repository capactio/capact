package manifest

import (
	"context"
	"encoding/json"
	"fmt"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer/argo"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// ImplementationValidator is a validator for Implementation manifests.
type ImplementationValidator struct{}

// NewImplementationValidator creates new ImplementationValidator.
func NewImplementationValidator() *ImplementationValidator {
	return &ImplementationValidator{}
}

// Do is a method which triggers the validation.
func (v *ImplementationValidator) Do(_ context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	var implEntity types.Implementation
	err := json.Unmarshal(jsonBytes, &implEntity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Implementation type")
	}

	result := ValidationResult{}

	workflowData, ok := implEntity.Spec.Action.Args["workflow"]
	if !ok {
		result.Errors = append(result.Errors, fmt.Errorf("missing workflow key in .spec.action.args"))
		return result, nil
	}

	var workflow argo.Workflow
	if err := mapstructure.Decode(workflowData, &workflow); err != nil {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// Name returns the validator name.
func (v *ImplementationValidator) Name() string {
	return "ImplementationValidator"
}
