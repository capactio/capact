package policy

import (
	"fmt"

	"capact.io/capact/internal/ptr"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// maxBackendLookupForTypeRef defines maximum number of iteration to find a matching backend based on TypeRef path pattern.
const maxBackendLookupForTypeRef = 30

// TypeInstancePolicy holds the Policy for TypeInstance.
type TypeInstancePolicy struct {
	Rules []RulesForTypeInstance `json:"rules"`
}

// RulesForTypeInstance holds a single policy rule for a TypeInstance.
// +kubebuilder:object:generate=true
type RulesForTypeInstance struct {
	TypeRef types.ManifestRefWithOptRevision `json:"typeRef"`
	Backend TypeInstanceBackend              `json:"backend"`
}

// TypeInstanceBackend holds a Backend description to be used for storing a given TypeInstance.
// +kubebuilder:object:generate=true
type TypeInstanceBackend struct {
	TypeInstanceReference `json:",inline"`
}

// TypeInstanceBackendCollection knows which Backend should be used for a given TypeInstance based on the TypeRef.
type TypeInstanceBackendCollection struct {
	byTypeRef map[string]TypeInstanceBackend
	byAlias   map[string]TypeInstanceBackend
}

// SetByTypeRef associates a given TypeRef with a given storage backend instance.
func (t *TypeInstanceBackendCollection) SetByTypeRef(ref types.ManifestRefWithOptRevision, backend TypeInstanceBackend) {
	if t.byTypeRef == nil {
		t.byTypeRef = map[string]TypeInstanceBackend{}
	}
	t.byTypeRef[t.key(ref)] = backend
}

// GetByTypeRef returns storage backend for a given TypeRef.
// If backend for an explicit TypeRef is not found, the pattern matching is used.
//
// For example, if TypeRef is `cap.type.capactio.examples.message:0.1.0`:
//    - cap.type.capactio.examples.*:0.1.0
//    - cap.type.capactio.examples.*
//    - cap.type.capactio.*:0.1.0
//    - cap.type.capactio.*
//    - cap.type.*:0.1.0
//    - cap.type.*
//    - cap.*:0.1.0
//    - cap.*
//
// If both methods fail, default backend is returned.
func (t TypeInstanceBackendCollection) GetByTypeRef(typeRef types.TypeRef) (TypeInstanceBackend, bool) {
	// 1. Try the explicit TypeRef
	backend, found := t.byTypeRef[t.key(types.ManifestRefWithOptRevision{
		Path:     typeRef.Path,
		Revision: ptr.String(typeRef.Revision),
	})]
	if found {
		return backend, true
	}

	// 2. Try to find matching pattern for a given TypeRef.
	var (
		subPath    = typeRef.Path
		iterations = 0
	)

	for {
		if fmt.Sprintf("%s.", subPath) == types.OCFPathPrefix || iterations > maxBackendLookupForTypeRef {
			break
		}
		subPath = types.TrimLastNodeFromOCFPath(subPath)

		keyPatterns := []string{
			fmt.Sprintf("%s.*:%s", subPath, typeRef.Revision), // first try to match with revision
			fmt.Sprintf("%s.*", subPath),                      // later check for path pattern only
		}
		for _, pattern := range keyPatterns {
			backend, found := t.byTypeRef[pattern]
			if found {
				return backend, true
			}
		}
		iterations++
	}

	return TypeInstanceBackend{}, false
}

// SetByAlias associates a given alias with a given storage backend instance.
func (t *TypeInstanceBackendCollection) SetByAlias(name string, backend TypeInstanceBackend) {
	if t.byAlias == nil {
		t.byAlias = map[string]TypeInstanceBackend{}
	}
	t.byAlias[name] = backend
}

// GetByAlias returns backend associated with a given alias.
func (t *TypeInstanceBackendCollection) GetByAlias(name string) (TypeInstanceBackend, bool) {
	backend, found := t.byAlias[name]
	return backend, found
}

// GetAll returns all registered storage backends both for aliases and TypeRefs.
func (t *TypeInstanceBackendCollection) GetAll() map[string]TypeInstanceBackend {
	out := map[string]TypeInstanceBackend{}
	for k, v := range t.byAlias {
		out[k] = v
	}
	for k, v := range t.byTypeRef {
		out[k] = v
	}
	return out
}

func (t TypeInstanceBackendCollection) key(typeRef types.ManifestRefWithOptRevision) string {
	if typeRef.Revision != nil && *typeRef.Revision != "" {
		return fmt.Sprintf("%s:%s", typeRef.Path, *typeRef.Revision)
	}
	return typeRef.Path
}
