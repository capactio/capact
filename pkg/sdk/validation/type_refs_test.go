package validation_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestResolveTypeRefsToJSONSchemasFailures(t *testing.T) {
	// given
	tests := map[string]struct {
		givenTypeRefs         validation.TypeRefCollection
		givenHubTypeInstances []*gqlpublicapi.Type
		givenListTypesErr     error
		expectedErrorMsg      string
	}{
		"Not existing TypeRef": {
			givenHubTypeInstances: nil,
			givenTypeRefs: validation.TypeRefCollection{
				"aws-creds": {
					TypeRef: types.TypeRef{
						Path:     "cap.type.aws.auth.creds",
						Revision: "0.1.0",
					},
				},
			},
			expectedErrorMsg: heredoc.Doc(`
		          1 error occurred:
		          	* TypeRef "cap.type.aws.auth.creds:0.1.0" was not found in Hub`),
		},
		"Not existing Revision": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
			},
			givenTypeRefs: validation.TypeRefCollection{
				"aws-creds": {
					TypeRef: types.TypeRef{
						Path:     "cap.type.aws.auth.creds",
						Revision: "1.1.1",
					},
				},
			},
			expectedErrorMsg: heredoc.Doc(`
		          1 error occurred:
		          	* TypeRef "cap.type.aws.auth.creds:1.1.1" was not found in Hub`),
		},
		"Unexpected JSONSchema type": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				func() *gqlpublicapi.Type {
					ti := validation.AWSCredsTypeRevFixture()
					ti.Revisions[0].Spec.JSONSchema = 123 // change type to int, but should be string
					return ti
				}(),
			},
			givenTypeRefs: validation.TypeRefCollection{
				"aws-creds": {
					TypeRef: types.TypeRef{
						Path:     "cap.type.aws.auth.creds",
						Revision: "0.1.0",
					},
				},
			},
			expectedErrorMsg: heredoc.Doc(`
		          1 error occurred:
		          	* unexpected JSONSchema type for "cap.type.aws.auth.creds:0.1.0": expected string, got int`),
		},
		"Hub call error": {
			givenListTypesErr: errors.New("hub error for testing purposes"),
			givenTypeRefs: validation.TypeRefCollection{
				"aws-creds": {},
			},
			expectedErrorMsg: "while fetching JSONSchemas for TypeRefs: hub error for testing purposes",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()
			fakeCli := &validation.FakeHubCli{
				Types:          tc.givenHubTypeInstances,
				ListTypesError: tc.givenListTypesErr,
			}

			// when
			_, err := validation.ResolveTypeRefsToJSONSchemas(ctx, fakeCli, tc.givenTypeRefs)

			// then
			assert.EqualError(t, err, tc.expectedErrorMsg)
		})
	}
}
