package validation

import (
	"fmt"
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTypeInstances(t *testing.T) {
	tests := map[string]struct {
		schemaCollection       SchemaCollection
		typeInstanceCollection []*TypeInstanceEssentialData
		expError               error
	}{
		"When TypeInstance values do not contain the required property": {
			schemaCollection: SchemaCollection{
				"cap.type.aws.auth.creds:0.1.0": {
					Value:    fmt.Sprintf("%v", AWSCredsTypeRevFixture().Revisions[0].Spec.JSONSchema),
					Required: false,
				},
			},
			typeInstanceCollection: []*TypeInstanceEssentialData{
				{
					TypeRef: types.ManifestRef{
						Path:     "cap.type.aws.auth.creds",
						Revision: "0.1.0",
					},
					Value: map[string]interface{}{
						"test1": "test",
						"test2": "test",
					},
					Alias: ptr.String("aws-creds"),
				},
			},
			expError: fmt.Errorf("%s", "- Validation TypeInstances \"TypeInstance(ID: ,Alias: aws-creds)\":\n    * (root): key is required"),
		},
		"When TypeInstance value does not meet Type property constraints": {
			schemaCollection: SchemaCollection{
				"cap.type.aws.elasticsearch.install-input:0.1.0": {
					Value:    fmt.Sprintf("%v", AWSElasticsearchTypeRevFixture().Revisions[0].Spec.JSONSchema),
					Required: false,
				},
			},
			typeInstanceCollection: []*TypeInstanceEssentialData{
				{
					TypeRef: types.ManifestRef{
						Path:     "cap.type.aws.elasticsearch.install-input",
						Revision: "0.1.0",
					},
					Value: map[string]interface{}{
						"replicas": 5,
					},
					ID: ptr.String("5605af48-c34f-4bdc-b2d8-53c679bdfa5a"),
				},
			},
			expError: fmt.Errorf("%s", "- Validation TypeInstances \"TypeInstance(ID: 5605af48-c34f-4bdc-b2d8-53c679bdfa5a,Alias: )\":\n    * replicas: Invalid type. Expected: string, given: integer"),
		},
		"When TypeInstance contain the required property": {
			schemaCollection: SchemaCollection{
				"cap.type.aws.auth.creds:0.1.0": {
					Value:    fmt.Sprintf("%v", AWSCredsTypeRevFixture().Revisions[0].Spec.JSONSchema),
					Required: false,
				},
			},
			typeInstanceCollection: []*TypeInstanceEssentialData{
				{
					TypeRef: types.ManifestRef{
						Path:     "cap.type.aws.auth.creds",
						Revision: "0.1.0",
					},
					Value: map[string]interface{}{
						"key": "aaa",
					},
				},
			},
			expError: nil,
		},
		"When there is a collection of TypeInstance with an incorrect value": {
			schemaCollection: SchemaCollection{
				"cap.type.aws.auth.creds:0.1.0": {
					Value:    fmt.Sprintf("%v", AWSCredsTypeRevFixture().Revisions[0].Spec.JSONSchema),
					Required: false,
				},
			},
			typeInstanceCollection: []*TypeInstanceEssentialData{
				{
					TypeRef: types.ManifestRef{
						Path:     "cap.type.aws.auth.creds",
						Revision: "0.1.0",
					},
					Value: map[string]interface{}{
						"test1": "test",
					},
					Alias: ptr.String("aws-creds"),
				},
				{
					TypeRef: types.ManifestRef{
						Path:     "cap.type.aws.auth.creds",
						Revision: "0.1.0",
					},
					Value: map[string]interface{}{
						"test2": "test",
					},
					Alias: ptr.String("aws-creds-2"),
				},
			},
			expError: fmt.Errorf("%s", "- Validation TypeInstances \"TypeInstance(ID: ,Alias: aws-creds)\":\n    * (root): key is required\n- Validation TypeInstances \"TypeInstance(ID: ,Alias: aws-creds-2)\":\n    * (root): key is required"),
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			validationResults, err := ValidateTypeInstances(tc.schemaCollection, tc.typeInstanceCollection)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expError, validationResults.ErrorOrNil())
		})
	}
}
