package validation

import (
	"context"

	"capact.io/capact/internal/cli/heredoc"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
)

// FakeHubCli provides an easy way to fake real Hub client. It is used for test purposes.
type FakeHubCli struct {
	Types          []*gqlpublicapi.Type
	IDsTypeRefs    map[string]gqllocalapi.TypeInstanceTypeReference
	ListTypesError error
}

// FindTypeInstancesTypeRef returns fake data
func (f *FakeHubCli) FindTypeInstancesTypeRef(_ context.Context, _ []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error) {
	return f.IDsTypeRefs, nil
}

// ListTypes returns fake data
func (f *FakeHubCli) ListTypes(_ context.Context, _ ...public.TypeOption) ([]*gqlpublicapi.Type, error) {
	return f.Types, f.ListTypesError
}

// AWSCredsTypeRevFixture returns test fixture for AWS credentials Type.
func AWSCredsTypeRevFixture() *gqlpublicapi.Type {
	return &gqlpublicapi.Type{
		Path: "cap.type.aws.auth.creds",
		Revisions: []*gqlpublicapi.TypeRevision{
			{
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
			}},
	}
}

// AWSElasticsearchTypeRevFixture returns test fixture for AWS Elasticsearch Type.
func AWSElasticsearchTypeRevFixture() *gqlpublicapi.Type {
	return &gqlpublicapi.Type{
		Path: "cap.type.aws.elasticsearch.install-input",
		Revisions: []*gqlpublicapi.TypeRevision{{
			Metadata: &gqlpublicapi.TypeMetadata{
				Path: "cap.type.aws.elasticsearch.install-input",
			},
			Revision: "0.1.0",
			Spec: &gqlpublicapi.TypeSpec{
				JSONSchema: heredoc.Doc(`
                    {
                      "$schema": "http://json-schema.org/draft-07/schema",
                      "type": "object",
                      "title": "The schema for Elasticsearch input parameters.",
                      "required": ["replicas"],
                      "properties": {
                        "replicas": {
                          "type": "string",
                          "title": "Replica count for the Elasticsearch"
                        }
                      },
                      "additionalProperties": false
                    }`),
			},
		}},
	}
}
