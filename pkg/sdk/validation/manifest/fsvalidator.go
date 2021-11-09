package manifest

import (
	"context"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"capact.io/capact/pkg/sdk/manifest"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// FSValidator validates manifests using a OCF specification, which is read from a filesystem.
type FSValidator struct {
	commonValidators []JSONValidator
	kindValidators   map[types.ManifestKind][]JSONValidator
}

// NewDefaultFilesystemValidator returns a new FSValidator.
func NewDefaultFilesystemValidator(fs http.FileSystem, ocfSchemaRootPath string, opts ...ValidatorOption) FileSystemValidator {
	validator := &FSValidator{
		commonValidators: []JSONValidator{
			NewOCFSchemaValidator(fs, ocfSchemaRootPath),
		},
		kindValidators: map[types.ManifestKind][]JSONValidator{
			types.TypeManifestKind: {
				NewTypeValidator(),
			},
			types.InterfaceManifestKind: {
				NewInterfaceValidator(),
			},
		},
	}

	for _, opt := range opts {
		opt(validator)
	}

	return validator
}

// Do validates a manifest.
func (v *FSValidator) Do(ctx context.Context, path string) (ValidationResult, error) {
	yamlBytes, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return ValidationResult{}, err
	}

	metadata, err := manifest.UnmarshalMetadata(yamlBytes)
	if err != nil {
		return newValidationResult(errors.Wrap(err, "failed to read manifest metadata")), nil
	}

	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return newValidationResult(errors.Wrap(err, "cannot convert YAML manifest to JSON")), nil
	}

	validators := append(v.commonValidators, v.kindValidators[metadata.Kind]...)

	var validationErrs []error
	for _, validator := range validators {
		res, err := validator.Do(ctx, metadata, jsonBytes)
		if err != nil {
			validationErrs = append(validationErrs, errors.Wrapf(err, "%s: internal", validator.Name()))
		}

		var prefixedResErrs []error
		for _, resErr := range res.Errors {
			prefixedResErrs = append(prefixedResErrs, errors.Wrap(resErr, validator.Name()))
		}
		validationErrs = append(validationErrs, prefixedResErrs...)
	}

	return newValidationResult(validationErrs...), nil
}
