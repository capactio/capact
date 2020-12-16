package dto

type InterfaceRevision struct {
	Metadata *GenericMetadata `json:"metadata"`
	Revision string           `json:"revision"`
	Spec     *InterfaceSpec   `json:"spec"`
	// List Implementations for a given Interface
	//Implementations []*Implementation `json:"implementations"`
	Signature       *Signature        `json:"signature"`
}
type GenericMetadata struct {
	Name             string        `json:"name"`
	Prefix           *string       `json:"prefix"`
	Path             *string       `json:"path"`
	DisplayName      *string       `json:"displayName"`
	Description      string        `json:"description"`
	Maintainers      []*Maintainer `json:"maintainers"`
	DocumentationURL *string       `json:"documentationURL"`
	SupportURL       *string       `json:"supportURL"`
	IconURL          *string       `json:"iconURL"`
}

type Maintainer struct {
	Name  *string `json:"name"`
	Email string  `json:"email"`
	URL   *string `json:"url"`
}

type Signature struct {
	Och string `json:"och"`
}
