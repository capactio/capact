package public

import (
	"strings"
)

// ImplementationRevisionQueryFields allows to configure which fields should be returned for ImplementationRevision query
type ImplementationRevisionQueryFields uint64

const (
	// ImplRevRootFields returns all primitive fields specified on root.
	ImplRevRootFields ImplementationRevisionQueryFields = 1 << iota
	// ImplRevMetadataFields returns ImplementationRevision's metadata fields.
	ImplRevMetadataFields
	// ImplRevAllFields returns all ImplementationRevision fields.
	ImplRevAllFields
	implRevMaxKey
)

// Has returns true if flag is set.
func (f ImplementationRevisionQueryFields) Has(flag ImplementationRevisionQueryFields) bool {
	return f&flag != 0
}

// ListImplementationRevisionsOptions stores Implementation Revision filtering parameters.
type ListImplementationRevisionsOptions struct {
	fields string
}

// Apply is used to configure the ListImplementationRevisionsOption.
func (o *ListImplementationRevisionsOptions) Apply(opts ...ListImplementationRevisionsOption) {
	o.fields = implRevisionFieldsRegistry[ImplRevAllFields] // defaults to all fields, backward compatible
	for _, opt := range opts {
		opt(o)
	}
}

// ListImplementationRevisionsOption provides an option to configure the get request for Implementations.
type ListImplementationRevisionsOption func(*ListImplementationRevisionsOptions)

// WithImplRevCustomFields narrows down the request query fields to the specified ones.
func WithImplRevCustomFields(requestedFlags ImplementationRevisionQueryFields) ListImplementationRevisionsOption {
	return func(ops *ListImplementationRevisionsOptions) {
		if requestedFlags.Has(ImplRevAllFields) {
			ops.fields = implRevisionFieldsRegistry[ImplRevAllFields]
			return
		}

		var names []string
		for fieldOpt := ImplRevRootFields; fieldOpt < implRevMaxKey; fieldOpt <<= 1 {
			if requestedFlags.Has(fieldOpt) {
				names = append(names, implRevisionFieldsRegistry[fieldOpt])
			}
		}
		ops.fields = strings.Join(names, "\n")
	}
}
