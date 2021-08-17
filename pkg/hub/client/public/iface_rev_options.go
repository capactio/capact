package public

import "strings"

// InterfaceRevisionQueryFields allows to configure which fields should be returned for InterfaceRevision query.
type InterfaceRevisionQueryFields uint64

const (
	// InterfaceRevisionRootFields returns all primitive fields specified on root.
	InterfaceRevisionRootFields InterfaceRevisionQueryFields = 1 << iota
	// InterfaceRevisionMetadataFields returns InterfaceRevision's metadata fields.
	InterfaceRevisionMetadataFields
	// InterfaceRevisionImplementationRevisionsMetadata returns ImplementationRevisions' metadata fields for a given InterfaceRevision.
	InterfaceRevisionImplementationRevisionsMetadata
	// InterfaceRevisionInputFields returns InterfaceRevision's input data fields.
	InterfaceRevisionInputFields
	// InterfaceRevisionAllFields returns all InterfaceRevision fields.
	InterfaceRevisionAllFields

	ifaceRevMaxKey
)

// Has returns true if flag is set.
func (f InterfaceRevisionQueryFields) Has(flag InterfaceRevisionQueryFields) bool { return f&flag != 0 }

// InterfaceRevisionOption provides an option to configure the find request for Interface Revision.
type InterfaceRevisionOption func(*InterfaceRevisionOptions)

// InterfaceRevisionOptions stores Interface Revision filtering parameters.
type InterfaceRevisionOptions struct {
	fields string
}

// Apply is used to configure the InterfaceRevisionOption.
func (o *InterfaceRevisionOptions) Apply(opts ...InterfaceRevisionOption) {
	o.fields = ifaceRevisionFieldsRegistry[InterfaceRevisionAllFields] // defaults to all fields, backward compatible

	// Apply overrides
	for _, opt := range opts {
		opt(o)
	}
}

// WithInterfaceRevisionFields narrows down the request query fields to the specified ones.
func WithInterfaceRevisionFields(queryFields InterfaceRevisionQueryFields) InterfaceRevisionOption {
	return func(opts *InterfaceRevisionOptions) {
		opts.fields = getIfaceRevisionFieldsFromFlags(queryFields)
	}
}

func getIfaceRevisionFieldsFromFlags(queryFields InterfaceRevisionQueryFields) string {
	if queryFields.Has(InterfaceRevisionAllFields) {
		return ifaceRevisionFieldsRegistry[InterfaceRevisionAllFields]
	}

	var names []string
	for fieldOpt := InterfaceRevisionRootFields; fieldOpt < ifaceRevMaxKey; fieldOpt <<= 1 {
		if queryFields.Has(fieldOpt) {
			names = append(names, ifaceRevisionFieldsRegistry[fieldOpt])
		}
	}
	return strings.Join(names, "\n")
}
