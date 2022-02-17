package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"capact.io/capact/internal/ptr"
	graphqllocal "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// TypeInstanceEssentialData contains essential TypeInstance Data for validation purpose.
type TypeInstanceEssentialData struct {
	Value   interface{}
	TypeRef types.ManifestRef
	Alias   *string
	ID      *string
}

func (ti *TypeInstanceEssentialData) String() string {
	if ti == nil {
		return ""
	}
	var tiMetadata []string
	if ti.ID != nil {
		tiMetadata = append(tiMetadata, fmt.Sprintf("ID: %s", *ti.ID))
	}
	if ti.Alias != nil {
		tiMetadata = append(tiMetadata, fmt.Sprintf("Alias: %s", *ti.Alias))
	}
	return strings.Join(tiMetadata, ", ")
}

// TypeInstanceValidationHubClient defines Hub methods needed for validation of TypeInstances.
type TypeInstanceValidationHubClient interface {
	HubClient
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]graphqllocal.TypeInstanceTypeReference, error)
}

// ValidateTypeInstancesToCreate  is responsible for validating TypeInstance which do not exist and will be created.
func ValidateTypeInstancesToCreate(ctx context.Context, client TypeInstanceValidationHubClient, typeInstance *graphqllocal.CreateTypeInstancesInput) (Result, error) {
	var typeInstanceCollection []*TypeInstanceEssentialData
	typeRefCollection := TypeRefCollection{}

	for _, ti := range typeInstance.TypeInstances {
		if ti == nil || ti.TypeRef == nil {
			continue
		}
		manifestRef := types.ManifestRef{
			Path:     ti.TypeRef.Path,
			Revision: ti.TypeRef.Revision,
		}
		typeRefCollection[manifestRef.String()] = TypeRef{
			TypeRef: types.TypeRef(manifestRef),
		}
		typeInstanceCollection = append(typeInstanceCollection, &TypeInstanceEssentialData{
			Alias:   ti.Alias,
			Value:   ti.Value,
			TypeRef: manifestRef,
		})
	}

	schemasCollection, err := ResolveTypeRefsToJSONSchemas(ctx, client, typeRefCollection)
	if err != nil {
		return nil, errors.Wrapf(err, "while resolving TypeRefs to JSON Schemas")
	}
	return ValidateTypeInstances(schemasCollection, typeInstanceCollection)
}

// ValidateTypeInstanceToUpdate is responsible for validating TypeInstance which exists and will be updated.
func ValidateTypeInstanceToUpdate(ctx context.Context, client TypeInstanceValidationHubClient, typeInstanceToUpdate []graphqllocal.UpdateTypeInstancesInput) (Result, error) {
	var typeInstanceIds []string
	for _, ti := range typeInstanceToUpdate {
		typeInstanceIds = append(typeInstanceIds, ti.ID)
	}

	typeInstancesTypeRef, err := client.FindTypeInstancesTypeRef(ctx, typeInstanceIds)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding TypeInstance Type reference")
	}

	typeRefCollection := TypeRefCollection{}
	for _, typeReference := range typeInstancesTypeRef {
		manifestRef := types.ManifestRef{
			Path:     typeReference.Path,
			Revision: typeReference.Revision,
		}
		typeRefCollection[manifestRef.String()] = TypeRef{
			TypeRef: types.TypeRef(manifestRef),
		}
	}

	var typeInstanceCollection []*TypeInstanceEssentialData
	for _, ti := range typeInstanceToUpdate {
		if ti.TypeInstance == nil {
			continue
		}
		typeRef, ok := typeInstancesTypeRef[ti.ID]
		if !ok {
			return nil, errors.Wrapf(err, "while finding TypeInstance Type reference for id %q", ti.ID)
		}
		typeInstanceCollection = append(typeInstanceCollection, &TypeInstanceEssentialData{
			ID:      ptr.String(ti.ID),
			Value:   ti.TypeInstance.Value,
			TypeRef: types.ManifestRef(typeRef),
		})
	}

	schemasCollection, err := ResolveTypeRefsToJSONSchemas(ctx, client, typeRefCollection)
	if err != nil {
		return nil, errors.Wrapf(err, "while resolving TypeRefs to JSON Schemas")
	}

	return ValidateTypeInstances(schemasCollection, typeInstanceCollection)
}

//ValidateTypeInstances is responsible for validating TypeInstance.
func ValidateTypeInstances(schemaCollection SchemaCollection, typeInstanceCollection []*TypeInstanceEssentialData) (Result, error) {
	resultBldr := NewResultBuilder("Validation TypeInstances")

	for _, ti := range typeInstanceCollection {
		if _, ok := ti.Value.(map[string]interface{}); !ok {
			return Result{}, errors.New("could not create map from TypeInstance Value")
		}
		valuesJSON, err := json.Marshal(ti.Value)
		if err != nil {
			return Result{}, errors.Wrap(err, "while converting TypeInstance value to JSON bytes")
		}
		if _, ok := schemaCollection[ti.TypeRef.String()]; !ok {
			return Result{}, fmt.Errorf("could not find Schema for type %q", ti.TypeRef.String())
		}

		schemaLoader := gojsonschema.NewStringLoader(schemaCollection[ti.TypeRef.String()].Value)
		dataLoader := gojsonschema.NewBytesLoader(valuesJSON)

		result, err := gojsonschema.Validate(schemaLoader, dataLoader)
		if err != nil {
			return nil, errors.Wrap(err, "while validating JSON schema for TypeInstance")
		}
		if !result.Valid() {
			for _, err := range result.Errors() {
				msg := fmt.Sprintf("TypeInstance(%s)", ti.String())
				resultBldr.ReportIssue(msg, err.String())
			}
		}
	}

	return resultBldr.Result(), nil
}
