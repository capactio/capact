package manifest

import (
	"context"
	"encoding/json"

	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

// RemoteInterfaceValidator is a validator for Interface manifest, which calls Hub in order to do validation checks.
type RemoteInterfaceValidator struct {
	hub Hub
}

// NewRemoteInterfaceValidator creates new RemoteImplementationValidator.
func NewRemoteInterfaceValidator(hub Hub) *RemoteInterfaceValidator {
	return &RemoteInterfaceValidator{
		hub: hub,
	}
}

// Do is a method which triggers the validation.
func (v *RemoteInterfaceValidator) Do(ctx context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	var entity types.Interface
	err := json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Interface type")
	}

	var manifestRefsToCheck []hubpublicgraphql.ManifestReference

	// Input Parameters
	if entity.Spec.Input.Parameters != nil {
		for _, param := range entity.Spec.Input.Parameters.ParametersParameterMap {
			if param.TypeRef == nil {
				continue
			}

			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
				Path:     param.TypeRef.Path,
				Revision: param.TypeRef.Revision,
			})
		}
	}

	// Input TypeInstances
	for _, ti := range entity.Spec.Input.TypeInstances {
		manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
			Path:     ti.TypeRef.Path,
			Revision: ti.TypeRef.Revision,
		})
	}

	// Output TypeInstances
	for _, ti := range entity.Spec.Output.TypeInstances {
		if ti.TypeRef == nil {
			continue
		}

		manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
			Path:     ti.TypeRef.Path,
			Revision: ti.TypeRef.Revision,
		})
	}

	return checkManifestRevisionsExist(ctx, v.hub, manifestRefsToCheck)
}

// Name returns the validator name.
func (v *RemoteInterfaceValidator) Name() string {
	return "RemoteInterfaceValidator"
}
