package validate

import (
	"fmt"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/hashicorp/go-multierror"
)

type ValidationResult map[string]*multierror.Error

type (
	Schema struct {
		Value    string
		Required bool
	}

	SchemaCollection map[string]Schema
)

type (
	TypeRef struct {
		types.TypeRef
		Required bool
	}

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
