package validation_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestResolveTypeRefsToJSONSchemasFailures(t *testing.T) {
	// given
	tests := map[string]struct {
		givenTypeRefs                           validation.TypeRefCollection
		givenHubTypeInstances                   []*gqlpublicapi.TypeRevision
		givenListTypeRefRevisionsJSONSchemasErr error
		expectedErrorMsg                        string
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
			givenHubTypeInstances: []*gqlpublicapi.TypeRevision{
				fixAWSCredsTypeRev(),
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
			givenHubTypeInstances: []*gqlpublicapi.TypeRevision{
				func() *gqlpublicapi.TypeRevision {
					ti := fixAWSCredsTypeRev()
					ti.Spec.JSONSchema = 123 // change type to int, but should be string
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
			givenListTypeRefRevisionsJSONSchemasErr: errors.New("hub error for testing purposes"),
			givenTypeRefs: validation.TypeRefCollection{
				"aws-creds": {},
			},
			expectedErrorMsg: "while fetching JSONSchemas for input TypeRefs: hub error for testing purposes",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()
			fakeCli := &fakeHubCli{
				Types:                                tc.givenHubTypeInstances,
				ListTypeRefRevisionsJSONSchemasError: tc.givenListTypeRefRevisionsJSONSchemasErr,
			}

			// when
			_, err := validation.ResolveTypeRefsToJSONSchemas(ctx, fakeCli, tc.givenTypeRefs)

			// then
			assert.EqualError(t, err, tc.expectedErrorMsg)
		})
	}
}

func fixAWSCredsTypeRev() *gqlpublicapi.TypeRevision {
	return &gqlpublicapi.TypeRevision{
		Metadata: &gqlpublicapi.TypeMetadata{
			Path: "cap.type.aws.auth.creds",
		},
		Revision: "0.1.0",
		Spec: &gqlpublicapi.TypeSpec{
			JSONSchema: heredoc.Doc(`
                    {
                      "$schema": "http://json-schema.org/draft-07/schema",
                      "type": "object",
                      "required": [ "key" ],
                      "properties": {
                        "key": {
                          "type": "string"
                        }
                      }
                    }`),
		},
	}
}

type fakeHubCli struct {
	Types                                []*gqlpublicapi.TypeRevision
	IDsTypeRefs                          map[string]gqllocalapi.TypeInstanceTypeReference
	ListTypeRefRevisionsJSONSchemasError error
}

func (f *fakeHubCli) FindTypeInstancesTypeRef(_ context.Context, _ []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error) {
	return f.IDsTypeRefs, nil
}

func (f *fakeHubCli) ListTypeRefRevisionsJSONSchemas(_ context.Context, _ gqlpublicapi.TypeFilter) ([]*gqlpublicapi.TypeRevision, error) {
	return f.Types, f.ListTypeRefRevisionsJSONSchemasError
}
