package attributes

import (
	"context"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type AttributeResolver struct{}

func NewResolver() *AttributeResolver {
	return &AttributeResolver{}
}

func (r *AttributeResolver) Attributes(ctx context.Context, filter *gqlpublicapi.AttributeFilter) ([]*gqlpublicapi.Attribute, error) {
	attributes, err := mockedresolver.MockedAttributes()
	if err != nil {
		return []*gqlpublicapi.Attribute{}, err
	}
	return attributes, nil
}

func (r *AttributeResolver) Attribute(ctx context.Context, path string) (*gqlpublicapi.Attribute, error) {
	attributes, err := mockedresolver.MockedAttributes()
	if err != nil {
		return nil, err
	}
	for _, attribute := range attributes {
		if attribute.Path == path {
			return attribute, nil
		}
	}
	return nil, nil
}

func (r *AttributeResolver) Revision(ctx context.Context, obj *gqlpublicapi.Attribute, revision string) (*gqlpublicapi.AttributeRevision, error) {
	for _, rev := range obj.Revisions {
		if rev.Revision == revision {
			return rev, nil
		}
	}
	return nil, nil
}
