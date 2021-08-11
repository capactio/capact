package manifest

import (
	"context"
	"encoding/json"

	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

// RemoteTypeValidator is a validator for Type manifest, which calls Hub in order to do validation checks.
type RemoteTypeValidator struct {
	hub Hub
}

// NewRemoteTypeValidator creates new RemoteTypeValidator.
func NewRemoteTypeValidator(hub Hub) *RemoteTypeValidator {
	return &RemoteTypeValidator{
		hub: hub,
	}
}

// Do is a method which triggers the validation.
func (v *RemoteTypeValidator) Do(ctx context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	var entity types.Type
	err := json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Type type")
	}

	var manifestRefsToCheck []hubpublicgraphql.ManifestReference

	// Attributes
	for path, attr := range entity.Metadata.Attributes {
		manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
			Path:     path,
			Revision: attr.Revision,
		})
	}

	return checkManifestRevisionsExist(ctx, v.hub, manifestRefsToCheck)
}

// Name returns the validator name.
func (v *RemoteTypeValidator) Name() string {
	return "RemoteTypeValidator"
}
