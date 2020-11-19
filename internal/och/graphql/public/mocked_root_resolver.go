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

func (r *MockedRootResolver) InterfaceRevision() gqlpublicapi.InterfaceRevisionResolver {
	return interfaces.NewRevisionResolver()
}

func (r *MockedRootResolver) InterfaceGroup() gqlpublicapi.InterfaceGroupResolver {
	return interfacegroups.NewInterfacesResolver()
}

func (r *MockedRootResolver) Implementation() gqlpublicapi.ImplementationResolver {
	return implementations.NewResolver()
}

func (r *MockedRootResolver) ImplementationRevision() gqlpublicapi.ImplementationRevisionResolver {
	return implementations.NewRevisionResolver()
}

func (r *MockedRootResolver) RepoMetadata() gqlpublicapi.RepoMetadataResolver {
	return repometadata.NewResolver()
}

func (r *MockedRootResolver) Tag() gqlpublicapi.TagResolver {
	return tags.NewResolver()
}

func (r *MockedRootResolver) Type() gqlpublicapi.TypeResolver {
	return types.NewResolver()
}
