package public

import (
	"fmt"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
)

// InterfaceOption provides an option to configure the find request for Interface Revision.
type InterfaceOption func(*InterfaceOptions)

// InterfaceOptions stores Interface filtering parameters.
type InterfaceOptions struct {
	additionalFields string
	filter           gqlpublicapi.InterfaceFilter
}

// Apply is used to configure the InterfaceOption.
func (o *InterfaceOptions) Apply(opts ...InterfaceOption) {
	// Apply overrides
	for _, opt := range opts {
		opt(o)
	}
}

// WithLatestInterfaceRevision adds latestRevision fields for Interface query.
func WithLatestInterfaceRevision(requestedFields InterfaceRevisionQueryFields) InterfaceOption {
	return func(opts *InterfaceOptions) {
		opts.additionalFields = fmt.Sprintf(`
				latestRevision {
					%s
				}`, getIfaceRevisionFieldsFromFlags(requestedFields))
	}
}

// WithInterfaceFilter adds a given filter to Interface query.
func WithInterfaceFilter(filter gqlpublicapi.InterfaceFilter) InterfaceOption {
	return func(opts *InterfaceOptions) {
		opts.filter = filter
	}
}
