package validate

import (
	"fmt"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/hashicorp/go-multierror"
)

// ValidationResult holds validation result indexed by name. For example, by TypeInstance name.
type ValidationResult map[string]*multierror.Error

type (
	// Schema holds JSONSchema value and information if instance of this schema is required.
	Schema struct {
		Value    string
		Required bool
	}

	// SchemaCollection defines JSONSchema collection index by name.
	SchemaCollection map[string]Schema
)

type (
	// TypeRef holds TypeRef and information if TypeInstance of this TypeRef is required.
	TypeRef struct {
		types.TypeRef
		Required bool
	}

	// TypeRefCollection defines TypeRef collection index by name.
	TypeRefCollection map[string]TypeRef
)

// MergeSchemaCollection merge input schema collections into one collection.
// Fast error return when name collision is found.
func MergeSchemaCollection(in ...SchemaCollection) (SchemaCollection, error) {
	out := SchemaCollection{}

	for _, collection := range in {
		for name, schema := range collection {
			_, found := out[name]
			if found {
				return nil, fmt.Errorf("cannot merge schema collections, found name collision for %q", name)
			}
			out[name] = schema
		}
	}
	return out, nil
}

// MergeTypeRefCollection merge input typeRef collections into one collection.
// Fast error return when name collision is found.
func MergeTypeRefCollection(in ...TypeRefCollection) (TypeRefCollection, error) {
	out := TypeRefCollection{}

	for _, collection := range in {
		for name, typeRef := range collection {
			_, found := out[name]
			if found {
				return nil, fmt.Errorf("cannot merge input TypeRef collection, found name collision for %q", name)
			}
			out[name] = typeRef
		}
	}
	return out, nil
}
