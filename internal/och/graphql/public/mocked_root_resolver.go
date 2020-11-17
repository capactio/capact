package public

import (
	"projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver/implementations"
	interfacegroups "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver/interface-groups"
	"projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver/interfaces"
	repometadata "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver/repo-metadata"
	"projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver/tags"
	"projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver/types"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type MockedRootResolver struct {
	queryResolver mockedQueryResolver
}

func NewMockedRootResolver() *MockedRootResolver {
	return &MockedRootResolver{
		queryResolver: mockedQueryResolver{
			ImplementationResolver: implementations.NewResolver(),
			InterfaceResolver:      interfaces.NewResolver(),
			InterfaceGroupResolver: interfacegroups.NewResolver(),
			RepoMetadataResolver:   repometadata.NewResolver(),
			TagResolver:            tags.NewResolver(),
			TypeResolver:           types.NewResolver(),
		},
	}
}

func (r *MockedRootResolver) Query() gqlpublicapi.QueryResolver {
	return r.queryResolver
}

type mockedQueryResolver struct {
	*implementations.ImplementationResolver
	*interfaces.InterfaceResolver
	*interfacegroups.InterfaceGroupResolver
	*repometadata.RepoMetadataResolver
	*tags.TagResolver
	*types.TypeResolver
}

func (r *MockedRootResolver) Interface() gqlpublicapi.InterfaceResolver {
	return interfaces.NewResolver()
}

func (r *MockedRootResolver) Implementation() gqlpublicapi.ImplementationResolver {
	return implementations.NewResolver()
}
