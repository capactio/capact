// Holds manually added types
package types

// InterfaceRef holds the full path and revision to the Interface
type InterfaceRef struct {
	Path     string  `json:"path"`               // Path of a given Interface
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
