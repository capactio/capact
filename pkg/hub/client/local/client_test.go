package local

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO(review): THIS FILE WILL BE REMOVED BEFORE MERGING. IT WAS ADDED ONLY FOR DEMO/PR TESTING PURPOSES.

const (
	RunValidation       = "RUN_VALIDATION"
	typeInstanceTypeRef = "cap.type.testing:0.1.0"
)

type StorageSpec struct {
	URL           *string `json:"url,omitempty"`
	AcceptValue   *bool   `json:"acceptValue,omitempty"`
	ContextSchema *string `json:"contextSchema,omitempty"`
}

// Prerequisite:
//   Before running this test, make sure that Local Hub is running:
//     cd hub-js; APP_NEO4J_ENDPOINT=bolt://localhost:7687 APP_NEO4J_PASSWORD=okon APP_HUB_MODE=local npm run dev
//
// To run this test, execute:
// RUN_VALIDATION=true go test ./pkg/hub/client/local/ -v -count 1
func TestExternalStorageInputValidation(t *testing.T) {
	run := os.Getenv(RunValidation)
	if run == "" {
		t.Skipf("skipping running example test as the env %s is not provided", RunValidation)
	}

	ctx := context.Background()
	cli := NewDefaultClient("http://localhost:8080/graphql")

	tests := map[string]struct {
		// given
		storageSpec interface{}
		value       map[string]interface{}
		context     interface{}

		// then
		expErrMsg string
	}{
		"Should rejected value": {
			storageSpec: StorageSpec{
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
			storageSpec: StorageSpec{
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
			storageSpec: StorageSpec{
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
			storageSpec: StorageSpec{
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
			storageSpec: StorageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			context: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			// TODO(review): currently it's a an early return, is it sufficient?
			// if not, we will need to an support for throwing multierr to print aggregated data in higher layer.
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": input value not allowed
              	* Error: rollback externally stored values: External backend "MOCKED_ID": input value not allowed`),
		},

		// Invalid Storage TypeInstance
		"Should reject usage of backend without URL field": {
			storageSpec: StorageSpec{
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
			storageSpec: StorageSpec{
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
			storageSpec: StorageSpec{
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
	}, WithFields(TypeInstanceRootFields))
	require.NoError(t, err)
	assert.Len(t, familyDetails, 0)
}

// ======= HELPERS =======

func registerExternalStorage(ctx context.Context, t *testing.T, cli *Client, value interface{}) (string, func()) {
	t.Helper()

	externalStorageID, err := cli.CreateTypeInstance(ctx, fixExternalDotenvStorage(value))
	require.NoError(t, err)
	require.NotEmpty(t, externalStorageID)

	return externalStorageID, func() {
		_ = cli.DeleteTypeInstance(ctx, externalStorageID)
	}
}

func fixExternalDotenvStorage(value interface{}) *gqllocalapi.CreateTypeInstanceInput {
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
