package manifest_test

import (
	"context"
	"fmt"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation/manifest"
	"github.com/stretchr/testify/assert"
)

var (
	genericStorageSchema = heredoc.Doc(`
	{
		"$schema": "http://json-schema.org/draft-07/schema",
		"type": "object",
		"title": "The generic storage backend schema",
		"required": [
		  "required"
		],
		"properties": {
		  "required": {
			"$id": "#/properties/required",
			"type": "array",
			"minItems": 3,
			"allOf": [
			  {
				"contains": {
				  "const": "url"
				}
			  },
			  {
				"contains": {
				  "const": "acceptValue"
				}
			  },
			  {
				"contains": {
				  "const": "contextSchema"
				}
			  }
			]
		  }
		}
	  }						
	`)
)

func TestValidateBackendStorageSchema(t *testing.T) {
	tests := map[string]struct {
		typeInput        *types.Type
		errorMsg         string
		validationErrors []error
	}{
		"when there is a Type that does not extend a Hub storage": {
			typeInput:        fixType("", "{}"),
			errorMsg:         "",
			validationErrors: nil,
		},
		"when there is a correct storage Type schema": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`
						{
							"required": [
							  "acceptValue",
							  "url",
							  "contextSchema"
							]						  
						}
						`)),
			errorMsg:         "",
			validationErrors: nil,
		},
		"when there is schema that is invalid against the generic storage backend schema": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`
						{
							"required": [
							  "acceptValue",
							  "url"
							]						  
						}
						`)),
			errorMsg: "",
			validationErrors: []error{
				fmt.Errorf("while validating JSON Schema against the generic storage backend schema (`%s`):: required: At least one of the items must match", types.GenericBackendStorageSchemaTypePath),
				fmt.Errorf("while validating JSON Schema against the generic storage backend schema (`%s`):: required.0: required.0 does not match: \"contextSchema\"", types.GenericBackendStorageSchemaTypePath),
				fmt.Errorf("while validating JSON Schema against the generic storage backend schema (`%s`):: required: Must validate all the schemas (allOf)", types.GenericBackendStorageSchemaTypePath),
				fmt.Errorf("while validating JSON Schema against the generic storage backend schema (`%s`):: required: Array must have at least 3 items", types.GenericBackendStorageSchemaTypePath),
			},
		},
		"when there is invalid JSON schema": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`{test}`)),
			errorMsg:         "while validating JSON schema against the generic storage backend schema: invalid character 't' looking for beginning of object key string",
			validationErrors: nil,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()
			hubCli := fakeHub{
				knownTypes: []*gqlpublicapi.Type{
					fixGQLTypeSchema(
						types.GenericBackendStorageSchemaTypePath,
						genericStorageSchema,
					),
				},
			}
			validator := manifest.NewRemoteTypeValidator(&hubCli)

			//when
			results, err := validator.ValidateBackendStorageSchema(ctx, *tc.typeInput)

			//then
			if err != nil {
				assert.Equal(t, tc.errorMsg, err.Error())
			}
			assert.Equal(t, tc.validationErrors, results.Errors)
		})
	}
}

func fixType(refs string, schema string) *types.Type {
	return &types.Type{
		Spec: types.TypeSpec{
			AdditionalRefs: []string{refs},
			JSONSchema: types.JSONSchema{
				Value: schema,
			},
		},
	}
}

func fixGQLTypeSchema(path string, schema string) *gqlpublicapi.Type {
	return &gqlpublicapi.Type{
		Path: path,
		LatestRevision: &gqlpublicapi.TypeRevision{
			Spec: &gqlpublicapi.TypeSpec{
				JSONSchema: schema,
			},
		},
	}
}
