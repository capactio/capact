package public

import (
	"fmt"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type ImplementationRevisionNotFoundError struct {
	ref gqlpublicapi.InterfaceReference
}

func NewImplementationRevisionNotFoundError(ref gqlpublicapi.InterfaceReference) *ImplementationRevisionNotFoundError {
	return &ImplementationRevisionNotFoundError{
		ref: ref,
	}
}

func (e *ImplementationRevisionNotFoundError) Error() string {
	return fmt.Sprintf("No ImplementationRevision found for Interface %q in revision %q", e.ref.Path, e.ref.Revision)
}
