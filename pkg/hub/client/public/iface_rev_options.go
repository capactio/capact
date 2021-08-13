package public

import "strings"

// InterfaceRevisionQueryFields allows to configure which fields should be returned for InterfaceRevision query.
type InterfaceRevisionQueryFields uint64

const (
	// IfaceRevRootFields returns all primitive fields specified on root.
	IfaceRevRootFields InterfaceRevisionQueryFields = 1 << iota
	// IfaceRevMetadataFields returns InterfaceRevision's metadata fields.
	IfaceRevMetadataFields
	// IfaceRevImplRevisionsMetadata returns ImplementationRevisions's metadata fields for a given InterfaceRevision.
	IfaceRevImplRevisionsMetadata
	// IfaceRevInputFields returns InterfaceRevision's input data fields.
	IfaceRevInputFields
	// IfaceRevAllFields returns all InterfaceRevision fields.
	IfaceRevAllFields

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
	o.fields = ifaceRevisionFieldsRegistry[IfaceRevAllFields] // defaults to all fields, backward compatible

	// Apply overrides
	for _, opt := range opts {
		opt(o)
	}
}

// WithIfaceRevCustomFields narrows down the request query fields to the specified ones.
func WithIfaceRevCustomFields(queryFields InterfaceRevisionQueryFields) InterfaceRevisionOption {
	return func(opts *InterfaceRevisionOptions) {
		opts.fields = getIfaceRevisionFieldsFromFlags(queryFields)
	}
}

func getIfaceRevisionFieldsFromFlags(queryFields InterfaceRevisionQueryFields) string {
	if queryFields.Has(IfaceRevAllFields) {
		return ifaceRevisionFieldsRegistry[IfaceRevAllFields]
	}

	var names []string
	for fieldOpt := IfaceRevRootFields; fieldOpt < ifaceRevMaxKey; fieldOpt <<= 1 {
		if queryFields.Has(fieldOpt) {
			names = append(names, ifaceRevisionFieldsRegistry[fieldOpt])
		}
	}
	return strings.Join(names, "\n")
}
