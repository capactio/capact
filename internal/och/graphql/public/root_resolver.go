package public

import (
	"projectvoltron.dev/voltron/internal/och/graphql/public/resolver/implementations"
	interfacegroups "projectvoltron.dev/voltron/internal/och/graphql/public/resolver/interface-groups"
	"projectvoltron.dev/voltron/internal/och/graphql/public/resolver/interfaces"
	repometadata "projectvoltron.dev/voltron/internal/och/graphql/public/resolver/repo-metadata"
	"projectvoltron.dev/voltron/internal/och/graphql/public/resolver/tags"
	"projectvoltron.dev/voltron/internal/och/graphql/public/resolver/types"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type RootResolver struct {
	queryResolver queryResolver

	interfaceResolver              gqlpublicapi.InterfaceResolver
	interfaceRevisionResolver      gqlpublicapi.InterfaceRevisionResolver
	interfaceGroupResolver         gqlpublicapi.InterfaceGroupResolver
	implementationResolver         gqlpublicapi.ImplementationResolver
	implementationRevisionResolver gqlpublicapi.ImplementationRevisionResolver
	repoMetadataResolver           gqlpublicapi.RepoMetadataResolver
	tagResolver                    gqlpublicapi.TagResolver
	typeResolver                   gqlpublicapi.TypeResolver
}

func NewRootResolver() *RootResolver {
	return &RootResolver{
		queryResolver: queryResolver{
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

func (r *RootResolver) Query() gqlpublicapi.QueryResolver {
	return r.queryResolver
}

type queryResolver struct {
	*implementations.ImplementationResolver
	*interfaces.InterfaceResolver
	*interfacegroups.InterfaceGroupResolver
	*repometadata.RepoMetadataResolver
	*tags.TagResolver
	*types.TypeResolver
}

func (r *RootResolver) Interface() gqlpublicapi.InterfaceResolver {
	return r.interfaceResolver
}

func (r *RootResolver) InterfaceRevision() gqlpublicapi.InterfaceRevisionResolver {
	return r.interfaceRevisionResolver
}

func (r *RootResolver) InterfaceGroup() gqlpublicapi.InterfaceGroupResolver {
	return r.interfaceGroupResolver
}

func (r *RootResolver) Implementation() gqlpublicapi.ImplementationResolver {
	return r.implementationResolver
}

func (r *RootResolver) ImplementationRevision() gqlpublicapi.ImplementationRevisionResolver {
	return r.implementationRevisionResolver
}

func (r *RootResolver) RepoMetadata() gqlpublicapi.RepoMetadataResolver {
	return r.repoMetadataResolver
}

func (r *RootResolver) Tag() gqlpublicapi.TagResolver {
	return r.tagResolver
}

func (r *RootResolver) Type() gqlpublicapi.TypeResolver {
	return r.typeResolver
}
