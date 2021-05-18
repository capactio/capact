package graphql

// The types had to be moved out from generated models to add `omitempty` tags.

type PolicyRule struct {
	ImplementationConstraints *PolicyRuleImplementationConstraints `json:"implementationConstraints,omitempty"`
	InjectTypeInstances       []*TypeInstanceReference             `json:"injectTypeInstances,omitempty"`
}

type PolicyRuleImplementationConstraints struct {
	// Refers a specific required TypeInstance by path and optional revision.
	Requires []*ManifestReferenceWithOptionalRevision `json:"requires,omitempty"`
	// Refers a specific Attribute by path and optional revision.
	Attributes []*ManifestReferenceWithOptionalRevision `json:"attributes,omitempty"`
	// Refers a specific Implementation with exact path.
	Path *string `json:"path,omitempty"`
}

type ManifestReferenceWithOptionalRevision struct {
	Path     string  `json:"path"`
	Revision *string `json:"revision,omitempty"`
}
