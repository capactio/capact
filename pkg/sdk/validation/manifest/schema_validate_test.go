package manifest

import (
	"testing"

	"capact.io/capact/internal/cli/heredoc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateJSONSchema07DefinitionSuccess(t *testing.T) {
	// given
	validJSONSchema := heredoc.Doc(`
			{
			  "$schema": "http://json-schema.org/draft-07/schema",
			  "type": "object",
			  "required": [ "key" ],
			  "properties": {
				"key": {
				  "type": "string"
				}
			  }
			}`)

	// when
	res, err := validateJSONSchema07Definition(jsonSchemaCollection{
		"valid-schema": validJSONSchema,
	})

	// then
	require.NoError(t, err)
	assert.Empty(t, res.Errors)
}

func TestValidateJSONSchema07DefinitionFailures(t *testing.T) {
	tests := map[string]struct {
		JSONSchema string
		errMsg     string
	}{
		"Invalid JSONSchema": {
			JSONSchema: `{ "invalid" - schema]`,
			errMsg:     `schema-name: invalid JSON: invalid character '-' after object key`,
		},
		"Valid JSONSchema with appended random characters": {
			JSONSchema: heredoc.Doc(`
				{
				  "$schema": "http://json-schema.org/draft-07/schema",
				  "type": "object",
				  "properties": {
					"key": {
					  "type": "string"
					}
				  }
				}^&*()`),
			errMsg: `schema-name: invalid JSON: invalid character '^' after top-level value`,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			res, err := validateJSONSchema07Definition(jsonSchemaCollection{
				"schema-name": tc.JSONSchema,
			})

			// then
			require.NoError(t, err)

			assert.False(t, res.Valid())

			require.Len(t, res.Errors, 1)
			assert.EqualError(t, res.Errors[0], tc.errMsg)
		})
	}
}
