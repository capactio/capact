package types

import (
	"context"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type TypeResolver struct{}

func NewResolver() *TypeResolver {
	return &TypeResolver{}
}

func (r *TypeResolver) Types(ctx context.Context, filter *gqlpublicapi.TypeFilter) ([]*gqlpublicapi.Type, error) {
	types, err := mockedresolver.MockedTypes()
	if err != nil {
		return []*gqlpublicapi.Type{}, err
	}
	return types, nil
}

func (r *TypeResolver) Type(ctx context.Context, path string) (*gqlpublicapi.Type, error) {
	types, err := mockedresolver.MockedTypes()
	if err != nil {
		return nil, err
	}
	for _, t := range types {
		if t.Path == path {
			return t, nil
		}
	}
	return nil, nil
}

func (r *TypeResolver) Revision(ctx context.Context, obj *gqlpublicapi.Type, revision string) (*gqlpublicapi.TypeRevision, error) {
	for _, rev := range obj.Revisions {
		if rev.Revision == revision {
			return rev, nil
		}
	}
	return nil, nil
}
