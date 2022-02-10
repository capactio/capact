package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTypeInstances(t *testing.T) {
	tests := map[string]struct {
		schemaCollection       SchemaCollection
		typeInstanceCollection []*typeInstanceData
		expError               error
	}{
		"When TypeInstance values do not contain the required property": {
			schemaCollection: SchemaCollection{
				"cap.type.aws.auth.creds:0.1.0": {
					Value:    fmt.Sprintf("%v", AWSCredsTypeRevFixture().Revisions[0].Spec.JSONSchema),
					Required: false,
				},
			},
			typeInstanceCollection: []*typeInstanceData{
				{
					typeRefWithRevision: "cap.type.aws.auth.creds:0.1.0",
					value: map[string]interface{}{
						"test1": "test",
						"test2": "test",
					},
					alias: pointerToAlias("aws-creds"),
				},
			},
			expError: fmt.Errorf("%s", "- Validation TypeInstances \"TypeInstance with alias aws-creds\":\n    * (root): key is required"),
		},
		"When TypeInstance value does not meet Type property constraints": {
			schemaCollection: SchemaCollection{
				"cap.type.aws.elasticsearch.install-input:0.1.0": {
					Value:    fmt.Sprintf("%v", AWSElasticsearchTypeRevFixture().Revisions[0].Spec.JSONSchema),
					Required: false,
				},
			},
			typeInstanceCollection: []*typeInstanceData{
				{
					typeRefWithRevision: "cap.type.aws.elasticsearch.install-input:0.1.0",
					value: map[string]interface{}{
						"replicas": 5,
					},
					id: "5605af48-c34f-4bdc-b2d8-53c679bdfa5a",
				},
			},
			expError: fmt.Errorf("%s", "- Validation TypeInstances \"TypeInstance with id 5605af48-c34f-4bdc-b2d8-53c679bdfa5a\":\n    * replicas: Invalid type. Expected: string, given: integer"),
		},
		"When TypeInstance contain the required property": {
			schemaCollection: SchemaCollection{
				"cap.type.aws.auth.creds:0.1.0": {
					Value:    fmt.Sprintf("%v", AWSCredsTypeRevFixture().Revisions[0].Spec.JSONSchema),
					Required: false,
				},
			},
			typeInstanceCollection: []*typeInstanceData{
				{
					typeRefWithRevision: "cap.type.aws.auth.creds:0.1.0",
					value: map[string]interface{}{
						"key": "aaa",
					},
				},
			},
			expError: nil,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			validationResults, err := validateTypeInstances(tc.schemaCollection, tc.typeInstanceCollection)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expError, validationResults.ErrorOrNil())
		})
	}
}

func pointerToAlias(alias string) *string {
	return &alias
}
