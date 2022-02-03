package validation_test

import (
	"context"
	"fmt"
	"testing"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTI(t *testing.T) {
	tests := map[string]struct {
		types        []*gqlpublicapi.Type
		typeInstance validation.TypeInstanceValidation
		expError     error
	}{
		"When TypeInstance values do not contain the required property": {
			types: []*gqlpublicapi.Type{validation.AWSCredsTypeRevFixture()},
			typeInstance: validation.TypeInstanceValidation{
				TypeRef: types.TypeRef{
					Path:     "cap.type.aws.auth.creds",
					Revision: "0.1.0",
				},
				Value: map[string]interface{}{
					"test1": "test",
					"test2": "test",
				},
			},
			expError: fmt.Errorf("%s", "- TypeInstance value \"\":\n    * (root): key is required"),
		},
		"When TypeInstance value does not meet Type property constraints": {
			types: []*gqlpublicapi.Type{validation.AWSElasticsearchTypeRevFixture()},
			typeInstance: validation.TypeInstanceValidation{
				TypeRef: types.TypeRef{
					Path:     "cap.type.aws.elasticsearch.install-input",
					Revision: "0.1.0",
				},
				Value: map[string]interface{}{
					"replicas": 5,
				},
			},
			expError: fmt.Errorf("%s", "- TypeInstance value \"\":\n    * replicas: Invalid type. Expected: string, given: integer"),
		},
		"When TypeInstance contain the required property": {
			types: []*gqlpublicapi.Type{validation.AWSCredsTypeRevFixture()},
			typeInstance: validation.TypeInstanceValidation{
				TypeRef: types.TypeRef{
					Path:     "cap.type.aws.auth.creds",
					Revision: "0.1.0",
				},
				Value: map[string]interface{}{
					"key": "aaa",
				},
			},
			expError: nil,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			hubCli := validation.FakeHubCli{
				Types: tc.types,
			}

			// when
			validationResults, err := validation.ValidateTI(context.Background(), &tc.typeInstance, &hubCli)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expError, validationResults.ErrorOrNil())
		})
	}
}
