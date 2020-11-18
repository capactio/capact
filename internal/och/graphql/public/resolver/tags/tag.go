package tags

import (
	"context"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type TagResolver struct{}

func NewResolver() *TagResolver {
	return &TagResolver{}
}

func (r *TagResolver) Tags(ctx context.Context, filter *gqlpublicapi.TagFilter) ([]*gqlpublicapi.Tag, error) {
	return []*gqlpublicapi.Tag{dummyTag("kubernetes"), dummyTag("cloudFoundry")}, nil
}

func (r *TagResolver) Tag(ctx context.Context, path string) (*gqlpublicapi.Tag, error) {
	return dummyTag("kubernetes"), nil
}

func (r *TagResolver) Revision(ctx context.Context, obj *gqlpublicapi.Tag, revision string) (*gqlpublicapi.TagRevision, error) {
	return &gqlpublicapi.TagRevision{}, nil
}

func dummyTag(name string) *gqlpublicapi.Tag {
	return &gqlpublicapi.Tag{
		Name:   name,
		Prefix: "cap.tag.platform",
		Path:   "cap.tag.platform." + name,
	}
}
