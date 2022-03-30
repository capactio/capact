//go:build localhub
// +build localhub

package localhub

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	typeInstanceTypeRef = "cap.type.testing:0.1.0"
)

type storageSpec struct {
	URL           *string `json:"url,omitempty"`
	AcceptValue   *bool   `json:"acceptValue,omitempty"`
	ContextSchema *string `json:"contextSchema,omitempty"`
}

func TestExternalStorageInputValidation(t *testing.T) {
	ctx := context.Background()
	cli := getLocalClient(t)

	tests := map[string]struct {
		// given
		storageSpec interface{}
		value       map[string]interface{}
		context     interface{}

		// then
		expErrMsg string
	}{
		"Should rejected value": {
			storageSpec: storageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": input value not allowed
              	* Error: rollback externally stored values: External backend "MOCKED_ID": input value not allowed`),
		},
		"Should rejected context": {
			storageSpec: storageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
			},
			context: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": input context not allowed
              	* Error: rollback externally stored values: External backend "MOCKED_ID": input context not allowed`),
		},
		"Should return error that context is not an object": {
			storageSpec: storageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
				ContextSchema: ptr.String(heredoc.Doc(`
					   {
					   	"$id": "#/properties/contextSchema",
					   	"type": "object",
					   	"properties": {
					   		"provider": {
					   			"$id": "#/properties/contextSchema/properties/name",
					   			"type": "string"
					   		}
					   	},
					   	"additionalProperties": false
					   }`)),
			},
			context: "Luke Skywalker",
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": invalid input: context must be object
              	* Error: rollback externally stored values: External backend "MOCKED_ID": invalid input: context must be object`),
		},
		"Should return validation error for context": {
			storageSpec: storageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
				ContextSchema: ptr.String(heredoc.Doc(`
					   {
					   	"$id": "#/properties/contextSchema",
					   	"type": "object",
					   	"properties": {
					   		"provider": {
					   			"$id": "#/properties/contextSchema/properties/name",
					   			"type": "string"
					   		}
					   	},
					   	"additionalProperties": false
					   }`)),
			},
			context: map[string]interface{}{
				"provider": true,
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": invalid input: context/provider must be string
              	* Error: rollback externally stored values: External backend "MOCKED_ID": invalid input: context/provider must be string`),
		},
		"Should reject value and context": {
			storageSpec: storageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			context: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": input value not allowed
              	* Error: rollback externally stored values: External backend "MOCKED_ID": input value not allowed`),
		},

		// Invalid Storage TypeInstance
		"Should reject usage of backend without URL field": {
			storageSpec: storageSpec{
				AcceptValue: ptr.Bool(false),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'url'
              	* Error: rollback externally stored values: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'url'`),
		},
		"Should reject usage of backend without AcceptValue field": {
			storageSpec: storageSpec{
				URL: ptr.String("http://localhost:5000/fake"),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'acceptValue'
              	* Error: rollback externally stored values: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'acceptValue'`),
		},
		"Should reject usage of backend without URL and AcceptValue fields": {
			storageSpec: map[string]interface{}{
				"other-data": true,
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'url', spec.value must have required property 'acceptValue'
              	* Error: rollback externally stored values: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'url', spec.value must have required property 'acceptValue'`),
		},
		"Should reject usage of backend with wrong context schema": {
			storageSpec: storageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
				ContextSchema: ptr.String(heredoc.Doc(`
					   yaml: true`)),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: failed to process the TypeInstance's backend "MOCKED_ID": invalid spec.context: Unexpected token y in JSON at position 0
              	* Error: rollback externally stored values: failed to process the TypeInstance's backend "MOCKED_ID": invalid spec.context: Unexpected token y in JSON at position 0`),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			externalStorageID, cleanup := registerExternalStorage(ctx, t, cli, tc.storageSpec)
			defer cleanup()

			// when
			_, err := cli.CreateTypeInstance(ctx, &gqllocalapi.CreateTypeInstanceInput{
				TypeRef: typeRef(typeInstanceTypeRef),
				Value:   tc.value,
				Backend: &gqllocalapi.TypeInstanceBackendInput{
					ID:      externalStorageID,
					Context: tc.context,
				},
			})

			require.Error(t, err)

			regex := regexp.MustCompile(`\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`)
			gotErr := regex.ReplaceAllString(err.Error(), "MOCKED_ID")

			// then
			assert.Equal(t, tc.expErrMsg, gotErr)
		})
	}

	// sanity check
	familyDetails, err := cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		TypeRef: &gqllocalapi.TypeRefFilterInput{
			Path:     typeRef(typeInstanceTypeRef).Path,
			Revision: ptr.String(typeRef(typeInstanceTypeRef).Revision),
		},
	}, local.WithFields(local.TypeInstanceRootFields))
	require.NoError(t, err)
	assert.Len(t, familyDetails, 0)
}

func registerExternalStorage(ctx context.Context, t *testing.T, cli *local.Client, value interface{}) (string, func()) {
	t.Helper()

	externalStorageID, err := cli.CreateTypeInstance(ctx, fixExternalStorage(value))
	require.NoError(t, err)
	require.NotEmpty(t, externalStorageID)

	return externalStorageID, func() {
		_ = cli.DeleteTypeInstance(ctx, externalStorageID)
	}
}

func fixExternalStorage(value interface{}) *gqllocalapi.CreateTypeInstanceInput {
	return &gqllocalapi.CreateTypeInstanceInput{
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.example.filesystem.storage",
			Revision: "0.1.0",
		},
		Value: value,
	}
}

func typeRef(in string) *gqllocalapi.TypeInstanceTypeReferenceInput {
	out := strings.Split(in, ":")
	return &gqllocalapi.TypeInstanceTypeReferenceInput{Path: out[0], Revision: out[1]}
}
