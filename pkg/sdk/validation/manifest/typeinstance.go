package manifest

import (
	"context"
	"encoding/json"
	"fmt"

	graphql "capact.io/capact/pkg/hub/api/graphql/public"
	"k8s.io/utils/strings/slices"

	"capact.io/capact/pkg/hub/client"

	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
)

type TypeJSONSchema struct {
	Properties map[string]struct {
		Id    string `json:"$id"`
		Type  string `json:"type"`
		Title string `json:"title"`
	} `json:"properties"`
}

// ValidateTI is responsible for validating the TypeInstance.
func ValidateTI(ctx context.Context, ti *gqllocalapi.CreateTypeInstanceInput, cl client.Public) (ValidationResult, error) {
	if ti == nil {
		return ValidationResult{}, nil
	}
	var errors []error

	typeRevision, err := cl.FindTypeRevision(ctx, graphql.TypeReference{
		Path:     ti.TypeRef.Path,
		Revision: ti.TypeRef.Revision,
	})
	if err != nil {
		return ValidationResult{}, err
	}

	var typeJSONSchema TypeJSONSchema
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", typeRevision.Spec.JSONSchema)), &typeJSONSchema); err != nil {
		return ValidationResult{}, err
	}

	var validKeys []string
	for key := range typeJSONSchema.Properties {
		validKeys = append(validKeys, key)
	}

	// validate the keys
	mapValues := ti.Value.(map[string]interface{})
	if len(mapValues) > 0 {
		for key := range mapValues {
			if !slices.Contains(validKeys, key) {
				errors = append(errors, fmt.Errorf("key value: %s no defined by the Type", key))
			}
		}
	}

	return ValidationResult{
		Errors: errors,
	}, nil
}
