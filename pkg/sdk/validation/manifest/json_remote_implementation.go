package manifest

import (
	"context"
	"encoding/json"
	"strings"

	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
)

// RemoteImplementationValidator is a validator for Implementation manifest, which calls Hub in order to do validation checks.
type RemoteImplementationValidator struct {
	hub Hub
}

// NewRemoteImplementationValidator creates new RemoteImplementationValidator.
func NewRemoteImplementationValidator(hub Hub) *RemoteImplementationValidator {
	return &RemoteImplementationValidator{
		hub: hub,
	}
}

// Do is a method which triggers the validation.
func (v *RemoteImplementationValidator) Do(ctx context.Context, _ types.ManifestMetadata, jsonBytes []byte) (ValidationResult, error) {
	var entity types.Implementation
	err := json.Unmarshal(jsonBytes, &entity)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while unmarshalling JSON into Implementation type")
	}

	var manifestRefsToCheck []hubpublicgraphql.ManifestReference

	// Attributes
	for path, attr := range entity.Metadata.Attributes {
		manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
			Path:     path,
			Revision: attr.Revision,
		})
	}

	// AdditionalParameters
	if entity.Spec.AdditionalInput != nil {
		// Parameters
		for _, param := range entity.Spec.AdditionalInput.Parameters {
			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference(param.TypeRef))
		}

		// TypeInstances
		for _, ti := range entity.Spec.AdditionalInput.TypeInstances {
			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference(ti.TypeRef))
		}
	}

	// AdditionalOutput
	if entity.Spec.AdditionalOutput != nil {
		for _, ti := range entity.Spec.AdditionalOutput.TypeInstances {
			if ti.TypeRef == nil {
				continue
			}

			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference(*ti.TypeRef))
		}
	}

	// Implements
	for _, implementsItem := range entity.Spec.Implements {
		manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference(implementsItem))
	}

	// Requires
	for requiresKey, requiresValue := range entity.Spec.Requires {
		var itemsToCheck []types.RequireEntity
		itemsToCheck = append(itemsToCheck, requiresValue.OneOf...)
		itemsToCheck = append(itemsToCheck, requiresValue.AllOf...)
		itemsToCheck = append(itemsToCheck, requiresValue.AnyOf...)

		for _, requiresSubItem := range itemsToCheck {
			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
				Path:     strings.Join([]string{requiresKey, requiresSubItem.Name}, "."),
				Revision: requiresSubItem.Revision,
			})
		}
	}

	// Imports
	for _, importsItem := range entity.Spec.Imports {
		for _, method := range importsItem.Methods {
			manifestRefsToCheck = append(manifestRefsToCheck, hubpublicgraphql.ManifestReference{
				Path:     strings.Join([]string{importsItem.InterfaceGroupPath, method.Name}, "."),
				Revision: method.Revision,
			})
		}
	}

	return checkManifestRevisionsExist(ctx, v.hub, manifestRefsToCheck)
}

// Name returns the validator name.
func (v *RemoteImplementationValidator) Name() string {
	return "RemoteImplementationValidator"
}
