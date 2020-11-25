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

	interfaceResolver              gqlpublicapi.InterfaceResolver
	interfaceRevisionResolver      gqlpublicapi.InterfaceRevisionResolver
	interfaceGroupResolver         gqlpublicapi.InterfaceGroupResolver
	implementationResolver         gqlpublicapi.ImplementationResolver
	implementationRevisionResolver gqlpublicapi.ImplementationRevisionResolver
	repoMetadataResolver           gqlpublicapi.RepoMetadataResolver
	tagResolver                    gqlpublicapi.TagResolver
	typeResolver                   gqlpublicapi.TypeResolver
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
		interfaceResolver:              interfaces.NewResolver(),
		interfaceRevisionResolver:      interfaces.NewRevisionResolver(),
		interfaceGroupResolver:         interfacegroups.NewInterfacesResolver(),
		implementationResolver:         implementations.NewResolver(),
		implementationRevisionResolver: implementations.NewRevisionResolver(),
		repoMetadataResolver:           repometadata.NewResolver(),
		tagResolver:                    tags.NewResolver(),
		typeResolver:                   types.NewResolver(),
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
	return r.interfaceResolver
}

func (r *MockedRootResolver) InterfaceRevision() gqlpublicapi.InterfaceRevisionResolver {
	return r.interfaceRevisionResolver
}

func (r *MockedRootResolver) InterfaceGroup() gqlpublicapi.InterfaceGroupResolver {
	return r.interfaceGroupResolver
}

func (r *MockedRootResolver) Implementation() gqlpublicapi.ImplementationResolver {
	return r.implementationResolver
}

func (r *MockedRootResolver) ImplementationRevision() gqlpublicapi.ImplementationRevisionResolver {
	return r.implementationRevisionResolver
}

func (r *MockedRootResolver) RepoMetadata() gqlpublicapi.RepoMetadataResolver {
	return r.repoMetadataResolver
}

func (r *MockedRootResolver) Tag() gqlpublicapi.TagResolver {
	return r.tagResolver
}

func (r *MockedRootResolver) Type() gqlpublicapi.TypeResolver {
	return r.typeResolver
}
