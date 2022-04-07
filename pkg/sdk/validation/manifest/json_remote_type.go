package manifest

import (
	"context"
	"encoding/json"
	"fmt"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/internal/regexutil"
	"capact.io/capact/pkg/hub/client/public"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
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

	resExist, err := checkManifestRevisionsExist(ctx, v.hub, manifestRefsToCheck)
	if err != nil {
		return ValidationResult{}, err
	}

	resNodes, err := v.checkAdditionalRefs(ctx, entity)
	if err != nil {
		return ValidationResult{}, err
	}

	resBackendStorage, err := v.validateBackendStorageSchema(ctx, entity)
	if err != nil {
		return ValidationResult{}, err
	}

	errorLists := [][]error{
		resExist.Errors,
		resNodes.Errors,
		resBackendStorage.Errors,
	}

	var validationResult ValidationResult
	for _, errorList := range errorLists {
		validationResult.Errors = append(validationResult.Errors, errorList...)
	}

	return validationResult, nil
}

func (v *RemoteTypeValidator) checkAdditionalRefs(ctx context.Context, entity types.Type) (ValidationResult, error) {
	res := ValidationResult{}
	if len(entity.Spec.AdditionalRefs) == 0 {
		return res, nil
	}

	// AdditionalRefs cannot point to a concrete path.
	// It must point to a parent (abstract) node.
	filter := regexutil.OrStringSlice(entity.Spec.AdditionalRefs)
	gotTypes, err := v.hub.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
		PathPattern: ptr.String(filter),
	}))

	if err != nil {
		return res, err
	}

	for _, item := range gotTypes {
		res.Errors = append(res.Errors, fmt.Errorf("%q cannot be used as parent node as it resolves to concrete Type", item.Path))
	}
	return res, nil
}

func (v *RemoteTypeValidator) validateBackendStorageSchema(ctx context.Context, entity types.Type) (ValidationResult, error) {
	res := ValidationResult{}

	if !entity.IsExtendingHubStorage() {
		return res, nil
	}

	opts := []public.TypeOption{
		public.WithTypeFilter(gqlpublicapi.TypeFilter{
			PathPattern: ptr.String(types.GenericBackendStorageSchemaTypePath),
		}),
		public.WithTypeLatestRevision(public.TypeRevisionSpecFields),
	}

	gotTypes, err := v.hub.ListTypes(ctx, opts...)
	if err != nil {
		return res, errors.Wrap(err, "while fetching Types")
	}

	if len(gotTypes) == 0 || gotTypes[0].LatestRevision == nil || gotTypes[0].LatestRevision.Spec == nil {
		return res, fmt.Errorf("cannot find generic backend storage schema")
	}

	storageBytes, err := json.Marshal(gotTypes[0].LatestRevision.Spec.JSONSchema)
	if err != nil {
		return res, errors.Wrap(err, "while marshaling generic backend storage schema to bytes")
	}

	var storageSchema string
	err = json.Unmarshal(storageBytes, &storageSchema)
	if err != nil {
		return res, errors.Wrap(err, "while converting generic backend storage schema to string")
	}

	schemaLoader := gojsonschema.NewStringLoader(storageSchema)
	dataLoader := gojsonschema.NewStringLoader(entity.Spec.JSONSchema.Value)

	result, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		return res, errors.Wrap(err, "while validating JSON schema against the generic storage backend schema")
	}

	for _, err := range result.Errors() {
		res.Errors = append(res.Errors, fmt.Errorf("while validating JSON Schema against the generic storage backend schema (`%s`):: %s", types.GenericBackendStorageSchemaTypePath, err.String()))
	}

	return res, nil
}

// Name returns the validator name.
func (v *RemoteTypeValidator) Name() string {
	return "RemoteTypeValidator"
}
