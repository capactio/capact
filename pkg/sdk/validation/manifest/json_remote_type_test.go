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
        "$schema":"http://json-schema.org/draft-07/schema",
        "type":"object",
        "title":"The generic storage backend schema",
        "required":[
          "required",
          "properties"
        ],
        "properties":{
          "required":{
            "$id":"#/properties/required",
            "type":"array",
            "minItems":3,
            "allOf":[
              {
                "contains":{
                  "const":"url"
                }
              },
              {
                "contains":{
                  "const":"acceptValue"
                }
              },
              {
                "contains":{
                  "const":"contextSchema"
                }
              }
            ]
          },
          "properties":{
            "$id":"#/properties/properties",
            "type":"object",
            "minProperties":3,
            "maxProperties":3,
            "properties":{
              "url":{
                "type":"object",
                "properties":{
                  "type":{
                    "const":"string"
                  }
                },
				"required":[
				  "type"
				]
              },
              "contextSchema":{
                "type":"object",
                "oneOf":[
                  {
                    "properties":{
                      "const":{
                        "type":"object"
                      }
                    },
                    "required":[
                      "const"
                    ]
                  },
                  {
                    "properties":{
                      "type":{
                        "const":"null"
                      }
                    },
                    "required":[
                      "type",
                      "const"
                    ]
                  }
                ]
              },
              "acceptValue":{
                "type":"object",
                "properties":{
                  "type":{
                    "const":"boolean"
                  }
                },
                "required":[
                  "type",
                  "const"
                ]
              }
            }
          },
          "additionalProperties":{
            "const":false
          }
        }
      }`)
	errorMsgPrefix = "while validating JSON Schema against the generic storage backend schema"
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
					  "url",
					  "acceptValue",
					  "contextSchema"
					],
					"properties": {
					  "url": {
						  "type": "string",
						  "format": "uri"
					  },
					  "contextSchema": {
						"type": "null",
						"const": null
					  },
					  "acceptValue": {
						"type": "boolean",
						"const": true
					  }
					},
					"additionalProperties": false
				  }`)),
			errorMsg:         "",
			validationErrors: nil,
		},
		"when there is schema that does not have all fields in required section": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`
				{
					"required": [
					  "acceptValue",
					  "url"
					],
					"properties": {
					  "url": {
						  "type": "string",
						  "format": "uri"
					  },
					  "contextSchema": {
						"type": "null",
						"const": null
					  },
					  "acceptValue": {
						"type": "boolean",
						"const": true
					  }
					},
					"additionalProperties": false
				}`)),
			errorMsg: "",
			validationErrors: []error{
				fmt.Errorf("%s (`%s`):: required: At least one of the items must match", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
				fmt.Errorf("%s (`%s`):: required.0: required.0 does not match: \"contextSchema\"", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
				fmt.Errorf("%s (`%s`):: required: Must validate all the schemas (allOf)", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
				fmt.Errorf("%s (`%s`):: required: Array must have at least 3 items", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
			},
		},
		"when there is schema that has invalid url property": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`
				{
					"required": [
					  "acceptValue",
					  "contextSchema",
					  "url"
					],
					"properties": {
					  "url": {
						  "type": "test",
						  "format": "uri"
					  },
					  "contextSchema": {
						"type": "null",
						"const": null
					  },
					  "acceptValue": {
						"type": "boolean",
						"const": true
					  }
					},
					"additionalProperties": false
				}`)),
			errorMsg: "",
			validationErrors: []error{
				fmt.Errorf("%s (`%s`):: properties.url.type: properties.url.type does not match: \"string\"", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
			},
		},
		"when there is schema that has invalid contextSchema property": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`
				{
					"required": [
					  "acceptValue",
					  "contextSchema",
					  "url"
					],
					"properties": {
					  "url": {
						  "type": "string",
						  "format": "uri"
					  },
					  "contextSchema": {
						"type": "test",
						"const": "null"
					  },
					  "acceptValue": {
						"type": "boolean",
						"const": true
					  }
					},
					"additionalProperties": false
				}`)),
			errorMsg: "",
			validationErrors: []error{
				fmt.Errorf("%s (`%s`):: properties.contextSchema: Must validate one and only one schema (oneOf)", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
				fmt.Errorf("%s (`%s`):: properties.contextSchema.type: properties.contextSchema.type does not match: \"null\"", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
			},
		},
		"when there is schema that has invalid acceptValue property": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`
				{
					"required": [
					  "acceptValue",
					  "contextSchema",
					  "url"
					],
					"properties": {
					  "url": {
						  "type": "string",
						  "format": "uri"
					  },
					  "contextSchema": {
						"type": "null",
						"const": null
					  },
					  "acceptValue": {
						"type": "boolean"
					  }
					},
					"additionalProperties": false
				}`)),
			errorMsg: "",
			validationErrors: []error{
				fmt.Errorf("%s (`%s`):: properties.acceptValue: const is required", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
			},
		},
		"when there is schema that has invalid additionalProperties property": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`
				{
					"required": [
					  "acceptValue",
					  "contextSchema",
					  "url"
					],
					"properties": {
					  "url": {
						  "type": "string",
						  "format": "uri"
					  },
					  "contextSchema": {
						"type": "null",
						"const": null
					  },
					  "acceptValue": {
						"type": "boolean",
						"const": false
					  }
					},
					"additionalProperties": true
				}`)),
			errorMsg: "",
			validationErrors: []error{
				fmt.Errorf("%s (`%s`):: additionalProperties: additionalProperties does not match: false", errorMsgPrefix, types.GenericBackendStorageSchemaTypePath),
			},
		},
		"when there is invalid JSON schema": {
			typeInput: fixType(
				"cap.core.type.hub.storage",
				heredoc.Doc(`{test}`)),
			errorMsg:         fmt.Sprintf("%s: invalid character 't' looking for beginning of object key string", errorMsgPrefix),
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
