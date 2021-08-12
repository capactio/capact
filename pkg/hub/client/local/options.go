package local

import (
	"strings"
)

// TypeInstancesQueryFields allows to configure which fields should be returned for TypeInstance query.
type TypeInstancesQueryFields uint64

const (
	// TypeInstanceRootFields returns all primitive fields specified on root.
	TypeInstanceRootFields TypeInstancesQueryFields = 1 << iota
	// TypeInstanceTypeRefFields returns TypeInstance's TypeRef fields.
	TypeInstanceTypeRefFields
	// TypeInstanceUsedByIDField returns IDs for UsedBy field.
	TypeInstanceUsedByIDField
	// TypeInstanceUsesIDField returns IDs for Uses field.
	TypeInstanceUsesIDField
	// TypeInstanceLatestResourceVersionField returns resourceVersion for LatestResourceVersion field.
	TypeInstanceLatestResourceVersionField
	// TypeInstanceAllFields returns all TypeInstance fields.
	TypeInstanceAllFields
	// TypeInstanceAllFieldsWithUses returns all TypeInstance fields with UsedBy and Uses.
	// It may generate a huge payload as the UsedBy and Uses fields has relations to next UsedBy and Uses fields.
	TypeInstanceAllFieldsWithUses
	maxKey
)

// Has returns true if flag is set
func (f TypeInstancesQueryFields) Has(flag TypeInstancesQueryFields) bool { return f&flag != 0 }

// Clear clears the given flag
func (f *TypeInstancesQueryFields) Clear(flag TypeInstancesQueryFields) { *f &= ^flag }

// TypeInstancesOptions stores configurations for TypeInstances request.
type TypeInstancesOptions struct {
	fields string
}

func newTypeInstancesOptions(defaultFieldsKey TypeInstancesQueryFields) *TypeInstancesOptions {
	return &TypeInstancesOptions{fields: typeInstancesFieldsRegistry[defaultFieldsKey]}
}

// Apply is used to configure the ListImplementationRevisionsOptions.
func (o *TypeInstancesOptions) Apply(opts ...TypeInstancesOption) {
	// Apply overrides
	for _, opt := range opts {
		opt(o)
	}
}

// TypeInstancesOption provides an option to configure the list request for TypeInstances.
type TypeInstancesOption func(*TypeInstancesOptions)

// WithCustomFields narrows down the request query fields to the specified ones.
func WithCustomFields(i TypeInstancesQueryFields) TypeInstancesOption {
	return func(ops *TypeInstancesOptions) {
		ops.fields = getTypeInstanceFieldsFromFlags(i)
	}
}

func getTypeInstanceFieldsFromFlags(i TypeInstancesQueryFields) string {
	var names []string
	for key := TypeInstanceRootFields; key < maxKey; key <<= 1 {
		if i.Has(key) {
			names = append(names, typeInstancesFieldsRegistry[key])
		}
	}
	return strings.Join(names, "\n")
}
