package manifest

import (
	"context"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// FileSystemValidator is an interface, with the Do method.
// Do validates the manifest in filepath and return a ValidationResult.
// If other, not manifest related errors occur, it will return an error.
type FileSystemValidator interface {
	Do(ctx context.Context, filepath string) (ValidationResult, error)
}

// ValidationResult hold the result of the manifest validation.
type ValidationResult struct {
	Errors []error
}

// Valid returns true, if the manifest contains no errors.
func (r *ValidationResult) Valid() bool {
	return len(r.Errors) == 0
}

func newValidationResult(errs ...error) ValidationResult {
	return ValidationResult{
		Errors: errs,
	}
}

// JSONValidator is an interface of validator which takes JSON bytes as input.
type JSONValidator interface {
	Do(ctx context.Context, metadata types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error)
	Name() string
}
