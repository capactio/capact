package public

import (
	"projectvoltron.dev/voltron/internal/och/graphql/public/resolver/attributes"
	"projectvoltron.dev/voltron/internal/och/graphql/public/resolver/implementations"
	interfacegroups "projectvoltron.dev/voltron/internal/och/graphql/public/resolver/interface-groups"
	"projectvoltron.dev/voltron/internal/och/graphql/public/resolver/interfaces"
	repometadata "projectvoltron.dev/voltron/internal/och/graphql/public/resolver/repo-metadata"
	"projectvoltron.dev/voltron/internal/och/graphql/public/resolver/types"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type RootResolver struct {
	queryResolver queryResolver

	interfaceResolver              gqlpublicapi.InterfaceResolver
	interfaceGroupResolver         gqlpublicapi.InterfaceGroupResolver
	implementationResolver         gqlpublicapi.ImplementationResolver
	implementationRevisionResolver gqlpublicapi.ImplementationRevisionResolver
	repoMetadataResolver           gqlpublicapi.RepoMetadataResolver
	attributeResolver              gqlpublicapi.AttributeResolver
	typeResolver                   gqlpublicapi.TypeResolver
}

func NewRootResolver() *RootResolver {
	return &RootResolver{
		queryResolver: queryResolver{
			ImplementationResolver: implementations.NewResolver(),
			InterfaceResolver:      interfaces.NewResolver(),
			InterfaceGroupResolver: interfacegroups.NewResolver(),
			RepoMetadataResolver:   repometadata.NewResolver(),
			AttributeResolver:      attributes.NewResolver(),
			TypeResolver:           types.NewResolver(),
		},
		interfaceResolver:              interfaces.NewResolver(),
		interfaceGroupResolver:         interfacegroups.NewInterfacesResolver(),
		implementationResolver:         implementations.NewResolver(),
		implementationRevisionResolver: implementations.NewRevisionResolver(),
		repoMetadataResolver:           repometadata.NewResolver(),
		attributeResolver:              attributes.NewResolver(),
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
	*attributes.AttributeResolver
	*types.TypeResolver
}

func (r *RootResolver) Interface() gqlpublicapi.InterfaceResolver {
	return r.interfaceResolver
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

func (r *RootResolver) Attribute() gqlpublicapi.AttributeResolver {
	return r.attributeResolver
}

func (r *RootResolver) Type() gqlpublicapi.TypeResolver {
	return r.typeResolver
}
