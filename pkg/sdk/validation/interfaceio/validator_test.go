package interfaceio

import (
	"context"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

var interfaceRevisionRaw = []byte(`
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

func TestValidateInterfaceInputParameters(t *testing.T) {
	// given
	iface := &gqlpublicapi.InterfaceRevision{}
	require.NoError(t, yaml.Unmarshal(interfaceRevisionRaw, iface))

	tests := map[string]struct {
		givenHubTypeInstances []*gqlpublicapi.Type
		givenParameters       types.ParametersCollection
		expectedIssues        string
	}{
		"Happy path JSON": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
			},
			givenParameters: types.ParametersCollection{
				"input-parameters": `{"key": true}`,
				"db-settings":      `{"key": true}`,
				"aws-creds":        `{"key": "true"}`,
			},
		},
		"Happy path YAML": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
			},
			givenParameters: types.ParametersCollection{
				"input-parameters": `key: true`,
				"db-settings":      `key: true`,
				"aws-creds":        `key: "true"`,
			},
		},
		"Not found `aws-creds`": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
			},
			givenParameters: types.ParametersCollection{
				"input-parameters": `{"key": true}`,
				"db-settings":      `{"key": true}`,
			},
			expectedIssues: heredoc.Doc(`
        	            	- Parameters "aws-creds":
        	            	    * required but missing input parameters`),
		},
		"Invalid parameters": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
			},
			givenParameters: types.ParametersCollection{
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
			ctx := context.Background()
			fakeCli := &validation.FakeHubCli{
				Types: tc.givenHubTypeInstances,
			}

			validator := NewValidator(fakeCli)

			// when
			ifaceSchemas, err := validator.LoadInputParametersSchemas(ctx, iface)
			// then
			require.NoError(t, err)
			require.Len(t, ifaceSchemas, 3)

			// when
			result, err := validator.ValidateParameters(ctx, ifaceSchemas, tc.givenParameters)
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

func TestValidateParametersNoop(t *testing.T) {
	tests := map[string]struct {
		givenIface      *gqlpublicapi.InterfaceRevision
		givenParameters types.ParametersCollection
	}{
		"Should do nothing on nil": {
			givenIface:      nil,
			givenParameters: nil,
		},
		"Should do nothing on zero values": {
			givenIface:      &gqlpublicapi.InterfaceRevision{},
			givenParameters: types.ParametersCollection{},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()

			validator := NewValidator(&validation.FakeHubCli{})

			// when
			ifaceSchemas, err := validator.LoadInputParametersSchemas(ctx, tc.givenIface)
			// then
			require.NoError(t, err)
			require.Len(t, ifaceSchemas, 0)

			// when
			result, err := validator.ValidateParameters(ctx, ifaceSchemas, tc.givenParameters)
			// then
			require.NoError(t, err)
			assert.NoError(t, result.ErrorOrNil())
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

func TestValidateTypeInstances(t *testing.T) {
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
		"not found required TypeInstance": {
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
			},
			expectedIssues: heredoc.Doc(`
		           - TypeInstances "database":
		               * required but missing TypeInstance of type cap.type.db.connection:0.1.0`),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()
			fakeCli := &validation.FakeHubCli{
				IDsTypeRefs: tc.givenHubTypeInstances,
			}

			validator := NewValidator(fakeCli)

			// when
			ifaceTypes, err := validator.LoadInputTypeInstanceRefs(ctx, iface)
			// then
			require.NoError(t, err)
			require.Len(t, ifaceTypes, 2)

			// when
			result, err := validator.ValidateTypeInstances(ctx, ifaceTypes, tc.givenTypeInstances)
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

func TestValidateTypeInstancesNoop(t *testing.T) {
	// given
	tests := map[string]struct {
		givenIface         *gqlpublicapi.InterfaceRevision
		givenTypeInstances []types.InputTypeInstanceRef
	}{
		"Should do nothing on nil": {
			givenIface:         nil,
			givenTypeInstances: nil,
		},
		"Should do nothing on zero values": {
			givenIface:         &gqlpublicapi.InterfaceRevision{},
			givenTypeInstances: []types.InputTypeInstanceRef{},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()

			validator := NewValidator(&validation.FakeHubCli{})

			// when
			ifaceTypes, err := validator.LoadInputTypeInstanceRefs(ctx, tc.givenIface)
			// then
			require.NoError(t, err)
			require.Len(t, ifaceTypes, 0)

			// when
			result, err := validator.ValidateTypeInstances(ctx, ifaceTypes, tc.givenTypeInstances)
			// then
			require.NoError(t, err)
			assert.NoError(t, result.ErrorOrNil())
		})
	}
}
