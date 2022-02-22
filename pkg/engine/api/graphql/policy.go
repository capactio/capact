package graphql

// The types had to be moved out from generated models to add `omitempty` tags.

//InterfacePolicy represents Interface Policy.
type InterfacePolicy struct {
	Default *DefaultInterfaceData `json:"default,omitempty"`
	Rules   []*RulesForInterface  `json:"rules"`
}

// PolicyRule represents a single policy rule.
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

// ManifestReferenceWithOptionalRevision is used to represent a manifest reference with an optional revision property.
type ManifestReferenceWithOptionalRevision struct {
	Path     string  `json:"path"`
	Revision *string `json:"revision,omitempty"`
}

// PolicyRuleInjectData describes injection data for a given Policy rule.
type PolicyRuleInjectData struct {
	RequiredTypeInstances   []*RequiredTypeInstanceReference   `json:"requiredTypeInstances,omitempty"`
	AdditionalParameters    []*AdditionalParameter             `json:"additionalParameters,omitempty"`
	AdditionalTypeInstances []*AdditionalTypeInstanceReference `json:"additionalTypeInstances,omitempty"`
}

// RequiredTypeInstanceReference is used to represent required TypeInstance injection for a given Implementation.
type RequiredTypeInstanceReference struct {
	ID          string  `json:"id"`
	Description *string `json:"description,omitempty"`
}

// AdditionalTypeInstanceReference is used to represent additional TypeInstance injection for a given Implementation.
type AdditionalTypeInstanceReference struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
