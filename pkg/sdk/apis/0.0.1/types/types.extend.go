// Package types holds manually added types.
package types

// InterfaceRef holds the full path and revision to the Interface
type InterfaceRef ManifestRef

// ImplementationRef holds the full path and revision to the Implementation
type ImplementationRef ManifestRef

// AttributeRef holds the full path and revision to the Attribute
type AttributeRef ManifestRef

// ManifestRef holds the full path and the revision to a given manifest.
// +kubebuilder:object:generate=true
type ManifestRef struct {
	Path     string  `json:"path"`               // Path of a given manifest
	Revision *string `json:"revision,omitempty"` // Version of the manifest content in the SemVer format.
}

// InputTypeInstanceRef holds input TypeInstance reference.
type InputTypeInstanceRef struct {
	// Name refers to input TypeInstance name used in rendered Action.
	// Name is not unique as there may be multiple TypeInstances with the same name on different levels of Action workflow.
	Name string `json:"name"`

	// ID is a unique identifier for the input TypeInstance.
	ID string `json:"id"`
}

// TypeRefWithOptRevision specify type by path and optional revision.
type TypeRefWithOptRevision struct {
	// Path of a given Type.
	Path string `json:"path"`
	// Version of the manifest content in the SemVer format.
	Revision *string `json:"revision"`
}
