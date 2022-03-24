package local

import (
	"strings"
)

// TypeInstancesQueryFields allows configuring which fields should be returned for TypeInstance query.
type TypeInstancesQueryFields uint64

const (
	// TypeInstanceRootFields returns all primitive fields specified on root.
	TypeInstanceRootFields TypeInstancesQueryFields = 1 << iota
	// TypeInstanceTypeRefFields returns TypeInstance's TypeRef fields.
	TypeInstanceTypeRefFields
	// TypeInstanceBackendFields returns TypeInstance's Backend fields.
	TypeInstanceBackendFields
	// TypeInstanceUsedByIDField returns IDs for UsedBy field.
	TypeInstanceUsedByIDField
	// TypeInstanceUsesIDField returns IDs for Uses field.
	TypeInstanceUsesIDField
	// TypeInstanceLatestResourceVersionVersionField returns resourceVersion for LatestResourceVersion field.
	TypeInstanceLatestResourceVersionVersionField
	// TypeInstanceLatestResourceVersionFields returns TypeInstance's LatestResourceVersion fields.
	TypeInstanceLatestResourceVersionFields
	// TypeInstanceAllFields returns all TypeInstance fields.
	TypeInstanceAllFields
	// TypeInstanceUsesAllFields returns TypeInstance's Uses field.
	TypeInstanceUsesAllFields
	// TypeInstanceUsedByAllFields returns TypeInstance's UsedBy field.
	TypeInstanceUsedByAllFields

	typeInstanceMaxKey
)

// Holds combined options for easy usage.
const (
	// TypeInstanceAllFieldsWithRelations returns all TypeInstance fields with UsedBy and Uses.
	// It may generate a huge payload as the UsedBy and Uses fields has relations to next UsedBy and Uses fields.
	TypeInstanceAllFieldsWithRelations = TypeInstanceAllFields | TypeInstanceUsesAllFields | TypeInstanceUsedByAllFields
)

// Has returns true if flag is set.
func (f TypeInstancesQueryFields) Has(flag TypeInstancesQueryFields) bool { return f&flag != 0 }

// Clear clears the given flag.
func (f *TypeInstancesQueryFields) Clear(flag TypeInstancesQueryFields) { *f &= ^flag }

// TypeInstancesOptions stores configurations for TypeInstances request.
type TypeInstancesOptions struct {
	fields string
}

func newTypeInstancesOptions(defaultFields TypeInstancesQueryFields) *TypeInstancesOptions {
	return &TypeInstancesOptions{fields: getTypeInstanceFieldsFromFlags(defaultFields)}
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

// WithFields narrows down the request query fields to the specified ones.
func WithFields(queryFields TypeInstancesQueryFields) TypeInstancesOption {
	return func(opts *TypeInstancesOptions) {
		opts.fields = getTypeInstanceFieldsFromFlags(queryFields)
	}
}

func getTypeInstanceFieldsFromFlags(queryFields TypeInstancesQueryFields) string {
	var names []string
	for key := TypeInstanceRootFields; key < typeInstanceMaxKey; key <<= 1 {
		if queryFields.Has(key) {
			names = append(names, typeInstancesFieldsRegistry[key])
		}
	}
	return strings.Join(names, "\n")
}
