package validation

import (
	"context"
	"encoding/json"
	"fmt"

	graphqllocal "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type typeInstanceData struct {
	alias               *string
	id                  string
	value               interface{}
	typeRefWithRevision string
}

// TypeInstanceValidationHubClient defines Hub methods needed for validation of TypeInstances.
type TypeInstanceValidationHubClient interface {
	HubClient
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]graphqllocal.TypeInstanceTypeReference, error)
}

// ValidateTypeInstancesToCreate  is responsible for validating TypeInstance which do not exist and will be created.
func ValidateTypeInstancesToCreate(ctx context.Context, client TypeInstanceValidationHubClient, typeInstance *graphqllocal.CreateTypeInstancesInput) (Result, error) {
	var typeInstanceCollection []*typeInstanceData
	typeRefCollection := TypeRefCollection{}

	for _, ti := range typeInstance.TypeInstances {
		if ti == nil {
			continue
		}
		typeRef := types.TypeRef{
			Path:     ti.TypeRef.Path,
			Revision: ti.TypeRef.Revision,
		}
		name := getManifestPathWithRevision(ti.TypeRef.Path, ti.TypeRef.Revision)
		typeRefCollection[name] = TypeRef{
			TypeRef: typeRef,
		}
		typeInstanceCollection = append(typeInstanceCollection, &typeInstanceData{
			alias:               ti.Alias,
			value:               ti.Value,
			typeRefWithRevision: name,
		})
	}

	schemasCollection, err := ResolveTypeRefsToJSONSchemas(ctx, client, typeRefCollection)
	if err != nil {
		return nil, errors.Wrapf(err, "while resolving TypeRefs to JSON Schemas")
	}
	return validateTypeInstances(schemasCollection, typeInstanceCollection)
}

// ValidateTypeInstanceToUpdate  is responsible for validating TypeInstance which exists and will be updated.
func ValidateTypeInstanceToUpdate(ctx context.Context, client TypeInstanceValidationHubClient, typeInstanceToUpdate []graphqllocal.UpdateTypeInstancesInput) (Result, error) {
	var typeInstanceIds []string
	idToTypeNameMap := map[string]string{}
	for _, ti := range typeInstanceToUpdate {
		typeInstanceIds = append(typeInstanceIds, ti.ID)
	}

	typeInstancesTypeRef, err := client.FindTypeInstancesTypeRef(ctx, typeInstanceIds)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding TypeInstance Type reference")
	}

	typeRefCollection := TypeRefCollection{}
	for id, typeReference := range typeInstancesTypeRef {
		name := getManifestPathWithRevision(typeReference.Path, typeReference.Revision)
		typeRefCollection[name] = TypeRef{
			TypeRef: types.TypeRef{
				Path:     typeReference.Path,
				Revision: typeReference.Revision,
			},
		}
		idToTypeNameMap[id] = name
	}

	var typeInstanceCollection []*typeInstanceData
	for _, ti := range typeInstanceToUpdate {
		if ti.TypeInstance == nil {
			continue
		}
		typeInstanceCollection = append(typeInstanceCollection, &typeInstanceData{
			id:                  ti.ID,
			value:               ti.TypeInstance.Value,
			typeRefWithRevision: idToTypeNameMap[ti.ID],
		})
	}

	schemasCollection, err := ResolveTypeRefsToJSONSchemas(ctx, client, typeRefCollection)
	if err != nil {
		return nil, errors.Wrapf(err, "while resolving TypeRefs to JSON Schemas")
	}

	return validateTypeInstances(schemasCollection, typeInstanceCollection)
}

func validateTypeInstances(schemaCollection SchemaCollection, typeInstanceCollection []*typeInstanceData) (Result, error) {
	resultBldr := NewResultBuilder("Validation TypeInstances")

	for _, ti := range typeInstanceCollection {
		if _, ok := ti.value.(map[string]interface{}); !ok {
			return Result{}, errors.New("could not create map from TypeInstance Value")
		}
		valuesJSON, err := json.Marshal(ti.value)
		if err != nil {
			return Result{}, errors.Wrap(err, "while converting TypeInstance value to JSON bytes")
		}
		if _, ok := schemaCollection[ti.typeRefWithRevision]; !ok {
			return Result{}, fmt.Errorf("could not find Schema for type %s", ti.typeRefWithRevision)
		}

		schemaLoader := gojsonschema.NewStringLoader(schemaCollection[ti.typeRefWithRevision].Value)
		dataLoader := gojsonschema.NewBytesLoader(valuesJSON)

		result, err := gojsonschema.Validate(schemaLoader, dataLoader)
		if err != nil {
			return nil, errors.Wrap(err, "while validating JSON schema for TypeInstance")
		}
		if !result.Valid() {
			for _, err := range result.Errors() {
				msg := ""
				if ti.alias != nil {
					msg = fmt.Sprintf("TypeInstance with alias %s", *ti.alias)
				} else if ti.id != "" {
					msg = fmt.Sprintf("TypeInstance with id %s", ti.id)
				}
				resultBldr.ReportIssue(msg, err.String())
			}
		}
	}

	return resultBldr.Result(), nil
}

func getManifestPathWithRevision(path string, revision string) string {
	return path + ":" + revision
}
