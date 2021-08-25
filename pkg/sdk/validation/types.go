package validation

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	gomultierror "github.com/hashicorp/go-multierror"
)

// Result holds validation result indexed by name. For example, by TypeInstance name.
type Result map[string]*gomultierror.Error

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

// Len returns number of all reported issues.
func (issues Result) Len() int {
	cnt := 0
	for _, issues := range issues {
		if issues == nil {
			continue
		}
		cnt += issues.Len()
	}

	return cnt
}

// ErrorOrNil returns error only if validation issues were reported
// If Result is nil, returns nil.
func (issues *Result) ErrorOrNil() error {
	var msgs []string
	for _, name := range issues.sortedKeys() {
		issue := (*issues)[name]
		if issue == nil {
			continue
		}
		msgs = append(msgs, issue.Error())
	}

	if len(msgs) > 0 {
		return errors.New(strings.Join(msgs, "\n"))
	}
	return nil
}

// sortedKeys returns sorted map keys. Used to have deterministic final error messages.
func (issues Result) sortedKeys() []string {
	keys := make([]string, 0, len(issues))
	for k := range issues {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
