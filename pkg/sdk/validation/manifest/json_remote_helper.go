package manifest

import (
	"context"
	"fmt"

	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"github.com/pkg/errors"
)

// Hub is an interface for Hub GraphQL client methods needed for the remote validation.
type Hub interface {
	CheckManifestRevisionsExist(ctx context.Context, manifestRefs []hubpublicgraphql.ManifestReference) (map[hubpublicgraphql.ManifestReference]bool, error)
	FindInterfaceRevision(ctx context.Context, ref hubpublicgraphql.InterfaceReference, opts ...public.InterfaceRevisionOption) (*hubpublicgraphql.InterfaceRevision, error)
}

func checkManifestRevisionsExist(ctx context.Context, hub Hub, manifestRefsToCheck []hubpublicgraphql.ManifestReference) (ValidationResult, error) {
	if len(manifestRefsToCheck) == 0 {
		return ValidationResult{}, nil
	}

	res, err := hub.CheckManifestRevisionsExist(ctx, manifestRefsToCheck)
	if err != nil {
		return ValidationResult{}, errors.Wrap(err, "while checking if manifest revisions exist")
	}

	var validationErrs []error
	for typeRef, exists := range res {
		if exists {
			continue
		}

		validationErrs = append(validationErrs, fmt.Errorf("manifest revision '%s:%s' doesn't exist in Hub", typeRef.Path, typeRef.Revision))
	}

	return ValidationResult{Errors: validationErrs}, nil
}

func getTypesForImplementation(ctx context.Context, interfacePath string, hub Hub) (hubpublicgraphql.InterfaceInput, error) {
	iface, err := hub.FindInterfaceRevision(ctx, hubpublicgraphql.InterfaceReference{
		Path: interfacePath,
	}, public.WithInterfaceRevisionFields(public.InterfaceRevisionInputFields))
	if err != nil {
		return hubpublicgraphql.InterfaceInput{}, errors.Wrap(err, "while looking for Interface definition")
	}
	if iface == nil {
		return hubpublicgraphql.InterfaceInput{}, fmt.Errorf("interface %s was not found in Hub", interfacePath)
	}

	return *iface.Spec.Input, nil
}
