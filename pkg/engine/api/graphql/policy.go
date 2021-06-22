package graphql

// The types had to be moved out from generated models to add `omitempty` tags.

// PolicyRule represents a single policy rule
type PolicyRule struct {
	ImplementationConstraints *PolicyRuleImplementationConstraints `json:"implementationConstraints,omitempty"`
	Inject                    *PolicyRuleInjectData                `json:"inject,omitempty"`
}

// PolicyRuleImplementationConstraints represent the constraints, which must be meet by an Implementation,
// so the rule will match.
type PolicyRuleImplementationConstraints struct {
	// Refers a specific required TypeInstance by path and optional revision.
	Requires []*ManifestReferenceWithOptionalRevision `json:"requires,omitempty"`
	// Refers a specific Attribute by path and optional revision.
	Attributes []*ManifestReferenceWithOptionalRevision `json:"attributes,omitempty"`
	// Refers a specific Implementation with exact path.
	Path *string `json:"path,omitempty"`
}

// ManifestReferenceWithOptionalRevision is used to represent a Manifest Reference with an optional revision property
type ManifestReferenceWithOptionalRevision struct {
	Path     string  `json:"path"`
	Revision *string `json:"revision,omitempty"`
}
