package attributes

import (
	"context"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type AttributeResolver struct{}

func NewResolver() *AttributeResolver {
	return &AttributeResolver{}
}

func (r *AttributeResolver) Attributes(ctx context.Context, filter *gqlpublicapi.AttributeFilter) ([]*gqlpublicapi.Attribute, error) {
	return []*gqlpublicapi.Attribute{dummyAttribute("kubernetes"), dummyAttribute("cloudFoundry")}, nil
}

func (r *AttributeResolver) Attribute(ctx context.Context, path string) (*gqlpublicapi.Attribute, error) {
	return dummyAttribute("kubernetes"), nil
}

func (r *AttributeResolver) Revision(ctx context.Context, obj *gqlpublicapi.Attribute, revision string) (*gqlpublicapi.AttributeRevision, error) {
	return &gqlpublicapi.AttributeRevision{}, nil
}

func dummyAttribute(name string) *gqlpublicapi.Attribute {
	return &gqlpublicapi.Attribute{
		Name:   name,
		Prefix: "cap.attribute.platform",
		Path:   "cap.attribute.platform." + name,
	}
}
