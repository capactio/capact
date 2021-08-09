package action

import (
	"context"
	"testing"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"capact.io/capact/internal/cli/heredoc"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

var InterfaceRaw = []byte(`
spec:
  input:
    parameters:
      - name: input-parameters
        jsonSchema: |-
          {
            "$schema": "http://json-schema.org/draft-07/schema",
            "type": "object",
            "required": [ "key" ],
            "properties": {
              "key": {
                "type": "boolean",
                "title": "Key"
              }
            }
          }
      - name: db-settings
        jsonSchema: |-
          {
            "$schema": "http://json-schema.org/draft-07/schema",
            "type": "object",
            "required": [ "key" ],
            "properties": {
              "key": {
                "type": "boolean",
                "title": "Key"
              }
            }
          }
      - name: aws-creds
        typeRef:
          path: cap.type.aws.auth.creds
          revision: 0.1.0`)

func Test_ValidateParameters(t *testing.T) {
	// given
	iface := &gqlpublicapi.InterfaceRevision{}
	require.NoError(t, yaml.Unmarshal(InterfaceRaw, iface))

	tests := map[string]struct {
		givenHubTypeInstances []*gqlpublicapi.Type
		givenParameters       map[string]string
		expectedIssues        string
	}{
		"Happy path JSON": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				fixAWSCredsType(),
			},
			givenParameters: map[string]string{
				"input-parameters": `{"key": true}`,
				"db-settings":      `{"key": true}`,
				"aws-creds":        `{"key": "true"}`,
			},
		},
		"Happy path YAML": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				fixAWSCredsType(),
			},
			givenParameters: map[string]string{
				"input-parameters": `key: true`,
				"db-settings":      `key: true`,
				"aws-creds":        `key: "true"`,
			},
		},
		"Not found `aws-creds`": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				fixAWSCredsType(),
			},
			givenParameters: map[string]string{
				"input-parameters": `{"key": true}`,
				"db-settings":      `{"key": true}`,
			},
			expectedIssues: heredoc.Doc(`
        	            	- Parameters "aws-creds":
        	            	    * not found but it's required`),
		},
		"Invalid parameters": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				fixAWSCredsType(),
			},
			givenParameters: map[string]string{
				"input-parameters": `{"key": "true"}`,
				"db-settings":      `{"key": "true"}`,
				"aws-creds":        `{"key": "true"}`,
			},
			expectedIssues: heredoc.Doc(`
        	            	- Parameters "db-settings":
        	            	    * key: Invalid type. Expected: boolean, given: string
        	            	- Parameters "input-parameters":
        	            	    * key: Invalid type. Expected: boolean, given: string`),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeCli := &fakeHubCli{
				Types: tc.givenHubTypeInstances,
			}

			validator := NewValidator(fakeCli)

			// when
			ifaceSchemas, err := validator.LoadIfaceInputParametersSchemas(context.Background(), iface)
			// then
			require.NoError(t, err)
			require.Len(t, ifaceSchemas, 3)

			// when
			result, err := validator.ValidateParameters(ifaceSchemas, tc.givenParameters)
			// then
			require.NoError(t, err)

			if tc.expectedIssues == "" {
				assert.NoError(t, result.ErrorOrNil())
			} else {
				assert.EqualError(t, result.ErrorOrNil(), tc.expectedIssues)
			}
		})
	}
}

var InterfaceInputTypesRaw = []byte(`
spec:
  input:
    typeInstances:
      - name: database
        typeRef:
          path: cap.type.db.connection
          revision: 0.1.0
      - name: config
        typeRef:
          path: cap.type.mattermost.config
          revision: 0.1.0`)

func Test_ValidateTypeInstances(t *testing.T) {
	// given
	iface := &gqlpublicapi.InterfaceRevision{}
	require.NoError(t, yaml.Unmarshal(InterfaceInputTypesRaw, iface))

	tests := map[string]struct {
		givenHubTypeInstances map[string]gqllocalapi.TypeInstanceTypeReference
		givenTypeInstances    []types.InputTypeInstanceRef
		expectedIssues        string
	}{
		"Happy path": {
			givenHubTypeInstances: map[string]gqllocalapi.TypeInstanceTypeReference{
				"id-database": {
					Path:     "cap.type.db.connection",
					Revision: "0.1.0",
				},
				"id-config": {
					Path:     "cap.type.mattermost.config",
					Revision: "0.1.0",
				},
			},
			givenTypeInstances: []types.InputTypeInstanceRef{
				{Name: "config", ID: "id-config"},
				{Name: "database", ID: "id-database"},
			},
		},
		"Revision mismatch": {
			givenHubTypeInstances: map[string]gqllocalapi.TypeInstanceTypeReference{
				"id-database": {
					Path:     "cap.type.db.connection",
					Revision: "0.1.0",
				},
				"id-config": {
					Path:     "cap.type.mattermost.config",
					Revision: "0.1.1",
				},
			},
			givenTypeInstances: []types.InputTypeInstanceRef{
				{Name: "config", ID: "id-config"},
				{Name: "database", ID: "id-database"},
			},
			expectedIssues: heredoc.Doc(`
                    - TypeInstances "config":
                        * must be in Revision "0.1.0" but it's "0.1.1"`),
		},
		"Type mismatch": {
			givenHubTypeInstances: map[string]gqllocalapi.TypeInstanceTypeReference{
				"id-database": {
					Path:     "cap.type.db.connection",
					Revision: "0.1.0",
				},
				"id-config": {
					Path:     "cap.type.slack.config",
					Revision: "0.1.0",
				},
			},
			givenTypeInstances: []types.InputTypeInstanceRef{
				{Name: "config", ID: "id-config"},
				{Name: "database", ID: "id-database"},
			},
			expectedIssues: heredoc.Doc(`
                    - TypeInstances "config":
                        * must be of Type "cap.type.mattermost.config" but it's "cap.type.slack.config"`),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeCli := &fakeHubCli{
				IDsTypeRefs: tc.givenHubTypeInstances,
			}

			validator := NewValidator(fakeCli)

			// when
			ifaceTypes, err := validator.LoadIfaceInputTypeInstanceRefs(context.Background(), iface)
			// then
			require.NoError(t, err)
			require.Len(t, ifaceTypes, 2)

			// when
			result, err := validator.ValidateTypeInstances(ifaceTypes, tc.givenTypeInstances)
			// then
			require.NoError(t, err)

			if tc.expectedIssues == "" {
				assert.NoError(t, result.ErrorOrNil())
			} else {
				assert.EqualError(t, result.ErrorOrNil(), tc.expectedIssues)
			}
		})
	}
}

type fakeHubCli struct {
	Types       []*gqlpublicapi.Type
	IDsTypeRefs map[string]gqllocalapi.TypeInstanceTypeReference
}

func (f *fakeHubCli) FindTypeInstancesTypeRef(_ context.Context, ids []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error) {
	return f.IDsTypeRefs, nil
}

func (f *fakeHubCli) ListTypeRefRevisionsJSONSchemas(_ context.Context, filter gqlpublicapi.TypeFilter) ([]*gqlpublicapi.Type, error) {
	return f.Types, nil
}

func fixAWSCredsType() *gqlpublicapi.Type {
	return &gqlpublicapi.Type{
		Path: "cap.type.aws.auth.creds",
		Revisions: []*gqlpublicapi.TypeRevision{
			{
				Revision: "0.1.0",
				Spec: &gqlpublicapi.TypeSpec{
					JSONSchema: "{}",
				},
			},
		},
	}
}
