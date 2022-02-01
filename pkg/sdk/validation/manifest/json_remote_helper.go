package manifest

import (
	"context"
	"fmt"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"

	"github.com/pkg/errors"
)

// Hub is an interface for Hub GraphQL client methods needed for the remote validation.
type Hub interface {
	CheckManifestRevisionsExist(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error)
	FindInterfaceRevision(ctx context.Context, ref gqlpublicapi.InterfaceReference, opts ...public.InterfaceRevisionOption) (*gqlpublicapi.InterfaceRevision, error)
	ListTypes(ctx context.Context, opts ...public.TypeOption) ([]*gqlpublicapi.Type, error)
}

func checkManifestRevisionsExist(ctx context.Context, hub Hub, manifestRefsToCheck []gqlpublicapi.ManifestReference) (ValidationResult, error) {
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
