package public

import (
	"fmt"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
)

// TypeOption provides an option to configure the find request for Type.
type TypeOption func(*TypeOptions)

// TypeOptions stores Type filtering parameters.
type TypeOptions struct {
	additionalFields string
	Filter           gqlpublicapi.TypeFilter
}

// Apply is used to configure the TypeOption.
func (o *TypeOptions) Apply(opts ...TypeOption) {
	// Apply overrides
	for _, opt := range opts {
		opt(o)
	}
}

// WithTypeRevisions adds revisions field for Type query.
func WithTypeRevisions(requestedFields TypeRevisionQueryFields) TypeOption {
	return func(opts *TypeOptions) {
		opts.additionalFields = fmt.Sprintf(`
				revisions {
					%s
				}`, getTypeRevisionFieldsFromFlags(requestedFields))
	}
}

// WithTypeLatestRevision adds latestRevision field for Type query.
func WithTypeLatestRevision(requestedFields TypeRevisionQueryFields) TypeOption {
	return func(opts *TypeOptions) {
		opts.additionalFields = fmt.Sprintf(`
				latestRevision {
					%s
				}`, getTypeRevisionFieldsFromFlags(requestedFields))
	}
}

// WithTypeFilter adds a given filter to Type query.
func WithTypeFilter(filter gqlpublicapi.TypeFilter) TypeOption {
	return func(opts *TypeOptions) {
		opts.Filter = filter
	}
}
