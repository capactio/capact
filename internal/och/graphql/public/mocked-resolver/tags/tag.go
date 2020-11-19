package tags

import (
	"context"
	"fmt"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type TagResolver struct{}

func NewResolver() *TagResolver {
	return &TagResolver{}
}

func (r *TagResolver) Tags(ctx context.Context, filter *gqlpublicapi.TagFilter) ([]*gqlpublicapi.Tag, error) {
	tag, err := mockedresolver.MockedTag()
	if err != nil {
		return []*gqlpublicapi.Tag{}, err
	}
	return []*gqlpublicapi.Tag{tag}, nil
}

func (r *TagResolver) Tag(ctx context.Context, path string) (*gqlpublicapi.Tag, error) {
	tag, err := mockedresolver.MockedTag()
	if err != nil {
		return &gqlpublicapi.Tag{}, err
	}
	return tag, nil
}

func (r *TagResolver) Revision(ctx context.Context, obj *gqlpublicapi.Tag, revision string) (*gqlpublicapi.TagRevision, error) {
	tag, err := mockedresolver.MockedTag()
	if err != nil {
		return &gqlpublicapi.TagRevision{}, err
	}
	for _, r := range tag.Revisions {
		if r.Revision == revision {
			return r, nil
		}
	}
	return &gqlpublicapi.TagRevision{}, fmt.Errorf("No tag with revision %s", revision)
}
