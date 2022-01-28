package manifest

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/internal/regexutil"
	"capact.io/capact/pkg/hub/client/public"
	"context"
	"encoding/json"
	"fmt"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
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

	var manifestRefsToCheck []gqlpublicapi.ManifestReference

	// Attributes
	for path, attr := range entity.Metadata.Attributes {
		manifestRefsToCheck = append(manifestRefsToCheck, gqlpublicapi.ManifestReference{
			Path:     path,
			Revision: attr.Revision,
		})
	}

	existRes, err := checkManifestRevisionsExist(ctx, v.hub, manifestRefsToCheck)
	if err != nil {
		return ValidationResult{}, err
	}

	refsCheck, err := v.checkAdditionalRefs(ctx, entity)
	if err != nil {
		return ValidationResult{}, err
	}

	return ValidationResult{
		Errors: append(existRes.Errors, refsCheck.Errors...),
	}, nil
}

func (v *RemoteTypeValidator) checkAdditionalRefs(ctx context.Context, entity types.Type) (ValidationResult, error) {
	res := ValidationResult{}
	if len(entity.Spec.AdditionalRefs) == 0 {
		return res, nil
	}

	// AdditionalRefs should point to a concrete path.
	// It can point only to parent node.
	filter := regexutil.OrStringSlice(entity.Spec.AdditionalRefs)
	gotTypes, err := v.hub.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
		PathPattern: ptr.String(filter),
	}))

	if err != nil {
		return res, err
	}

	for _, item := range gotTypes {
		res.Errors = append(res.Errors, fmt.Errorf("%s cannot be used as parent node as it resolves to concrete Type", item.Path))
	}
	return res, nil
}

// Name returns the validator name.
func (v *RemoteTypeValidator) Name() string {
	return "RemoteTypeValidator"
}
