package tags

import (
	"context"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type TagResolver struct{}

func NewResolver() *TagResolver {
	return &TagResolver{}
}

func (r *TagResolver) Tags(ctx context.Context, filter *gqlpublicapi.TagFilter) ([]*gqlpublicapi.Tag, error) {
	tags, err := mockedresolver.MockedTags()
	if err != nil {
		return []*gqlpublicapi.Tag{}, err
	}
	return tags, nil
}

func (r *TagResolver) Tag(ctx context.Context, path string) (*gqlpublicapi.Tag, error) {
	tags, err := mockedresolver.MockedTags()
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		if tag.Path == path {
			return tag, nil
		}
	}
	return nil, nil
}

func (r *TagResolver) Revision(ctx context.Context, obj *gqlpublicapi.Tag, revision string) (*gqlpublicapi.TagRevision, error) {
	for _, rev := range obj.Revisions {
		if rev.Revision == revision {
			return rev, nil
		}
	}
	return nil, nil
}
