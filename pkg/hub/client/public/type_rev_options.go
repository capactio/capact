package public

import "strings"

// TypeRevisionQueryFields allows configuring which fields should be returned for TypeRevision query.
type TypeRevisionQueryFields uint64

const (
	// TypeRevisionRootFields returns all primitive fields specified on root.
	TypeRevisionRootFields TypeRevisionQueryFields = 1 << iota
	// TypeRevisionMetadataFields returns TypeRevision's metadata fields.
	TypeRevisionMetadataFields
	// TypeRevisionSpecFields for fetching TypeRevision's spec fields only.
	TypeRevisionSpecFields
	// TypeRevisionSpecAdditionalRefsField for fetching TypeRevision's spec.additionalRefs field only.
	TypeRevisionSpecAdditionalRefsField

	typeRevMaxKey
)

// Has returns true if flag is set.
func (f TypeRevisionQueryFields) Has(flag TypeRevisionQueryFields) bool { return f&flag != 0 }

// TypeRevisionOption provides an option to configure the find request for Type Revision.
type TypeRevisionOption func(*TypeRevisionOptions)

// TypeRevisionOptions stores Type Revision filtering parameters.
type TypeRevisionOptions struct {
	fields string
}

// Apply is used to configure the TypeRevisionOption.
func (o *TypeRevisionOptions) Apply(opts ...TypeRevisionOption) {
	o.fields = getTypeRevisionFieldsFromFlags(TypeRevisionRootFields | TypeRevisionMetadataFields | TypeRevisionSpecFields) // defaults to all fields, backward compatible

	// Apply overrides
	for _, opt := range opts {
		opt(o)
	}
}

func getTypeRevisionFieldsFromFlags(queryFields TypeRevisionQueryFields) string {
	var names []string
	for fieldOpt := TypeRevisionRootFields; fieldOpt < typeRevMaxKey; fieldOpt <<= 1 {
		if queryFields.Has(fieldOpt) {
			names = append(names, typeRevisionFieldsRegistry[fieldOpt])
		}
	}
	return strings.Join(names, "\n")
}
