// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    implementation, err := UnmarshalImplementation(bytes)
//    bytes, err = implementation.Marshal()

package quicktype

import "encoding/json"

func UnmarshalImplementation(data []byte) (Implementation, error) {
	var r Implementation
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Implementation) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// The description of an action and its prerequisites (dependencies). An implementation
// implements at least one interface.
type Implementation struct {
	Kind       Kind      `json:"kind"`      
	Metadata   Metadata  `json:"metadata"`  
	OcfVersion string    `json:"ocfVersion"`
	Revision   string    `json:"revision"`  // Version of the manifest content in the SemVer format.
	Signature  Signature `json:"signature"` // Ensures the authenticity and integrity of a given manifest.
	Spec       Spec      `json:"spec"`      // A container for the Implementation specification definition.
}

// A container for the OCF metadata definitions.
type Metadata struct {
	Description      string         `json:"description"`               // A short description of the OCF manifest. Must be a non-empty string.
	DisplayName      *string        `json:"displayName,omitempty"`     // The name of the OCF manifest to be displayed in graphical clients.
	DocumentationURL *string        `json:"documentationURL,omitempty"`// Link to documentation page for the OCF manifest.
	IconURL          *string        `json:"iconURL,omitempty"`         // The URL to an icon or a data URL containing an icon.
	Maintainers      []Maintainer   `json:"maintainers"`               // The list of maintainers with contact information.
	Name             string         `json:"name"`                      // The name of OCF manifest that uniquely identifies this object within the entity sub-tree.; Must be a non-empty string. We recommend using a CLI-friendly name.
	Prefix           *string        `json:"prefix,omitempty"`          // The prefix value is automatically computed and set when storing manifest in OCH.
	SupportURL       *string        `json:"supportURL,omitempty"`      // Link to support page for the OCF manifest.
	License          License        `json:"license"`                   // This entry allows you to specify a license, so people know how they are permitted to use; it, and what kind of restrictions you are placing on it.
	Tags             map[string]Tag `json:"tags,omitempty"`            // The tags is a list of key value, OCF Tags. Describes the OCF Implementation (provides; generic categorization) and are used to filter out a specific Implementation.
}

// This entry allows you to specify a license, so people know how they are permitted to use
// it, and what kind of restrictions you are placing on it.
type License struct {
	Name *string `json:"name,omitempty"`// If you are using a common license such as BSD-2-Clause or MIT, add a current SPDX license; identifier for the license you’re using e.g. BSD-3-Clause. If your package is licensed; under multiple common licenses, use an SPDX license expression syntax version 2.0 string,; e.g. (ISC OR GPL-3.0)
	Ref  *string `json:"ref,omitempty"` // If you are using a license that hasn’t been assigned an SPDX identifier, or if you are; using a custom license, use the direct link to the license file e.g.; https://raw.githubusercontent.com/project/v1/license.md. The resource under given link; MUST be immutable and publicly accessible.
}

// Holds contact information.
type Maintainer struct {
	Email string  `json:"email"`         // Email address of the person.
	Name  *string `json:"name,omitempty"`// Name of the person.
	URL   *string `json:"url,omitempty"` // URL of the person’s site.
}

type Tag struct {
	Revision string `json:"revision"`
}

// Ensures the authenticity and integrity of a given manifest.
type Signature struct {
	Och string `json:"och"`
}

// A container for the Implementation specification definition.
type Spec struct {
	Action     Action             `json:"action"`            // An explanation about the purpose of this instance.
	AppVersion string             `json:"appVersion"`        // The supported application versions in SemVer2 format.
	Implements []Implement        `json:"implements"`        // Defines what kind of interfaces this implementation fulfills.
	Imports    []Import           `json:"imports,omitempty"` // List of external Interfaces that this Implementation requires to be able to execute the; action.
	Requires   map[string]Require `json:"requires,omitempty"`// List of the system prerequisites that need to be present on the cluster. There has to be; an Instance for every concrete type.
}

// An explanation about the purpose of this instance.
type Action struct {
	Args map[string]interface{} `json:"args"`// Holds all parameters that should be passed to the selected runner, for example repoUrl,; or chartName for the Helm3 runner.
	Type string                 `json:"type"`// The Interface or Implementation of a runner, which handles the execution, for example,; cap.interface.runner.helm3.run
}

type Implement struct {
	Name     string  `json:"name"`              // The Interface name, for example cap.interfaces.db.mysql.install
	Revision *string `json:"revision,omitempty"`// The Interface revision.
}

type Import struct {
	Alias      *string  `json:"alias,omitempty"`     // The alias for the full name of the imported group name. It can be used later in the; workflow definition instead of using full name.
	AppVersion *string  `json:"appVersion,omitempty"`// The supported application versions in SemVer2 format.
	Methods    []string `json:"methods"`             // The list of all required actions’ names that must be imported.
	Name       string   `json:"name"`                // The name of the group that holds specific actions that you want to import, for example; cap.interfaces.db.mysql
}

// Prefix MUST be an abstract node and represents a core abstract Type e.g.
// cap.core.type.platform. Custom Types are not allowed.
type Require struct {
	AllOf []RequireEntity `json:"allOf,omitempty"`// All of the given types MUST have an Instance on the cluster. Element on the list MUST; resolves to concrete Type.
	AnyOf []RequireEntity `json:"anyOf,omitempty"`// Any (one or more) of the given types MUST have an Instance on the cluster. Element on the; list MUST resolves to concrete Type.
	OneOf []RequireEntity `json:"oneOf,omitempty"`// Exactly one of the given types MUST have an Instance on the cluster. Element on the list; MUST resolves to concrete Type.
}

type RequireEntity struct {
	Name     string                 `json:"name"`           // The name of the Type. Root prefix can be skipped if it’s a core Type. If it is a custom; Type then it MUST be defined as full path to that Type. Custom Type MUST extend the; abstract node which is defined as a root prefix for that entry.
	Revision string                 `json:"revision"`       // The revision version of the given Type.
	Value    map[string]interface{} `json:"value,omitempty"`// Holds the configuration constraints for the given entry. It needs to be valid against the; Type JSONSchema.
}

type Kind string
const (
	KindImplementation Kind = "Implementation"
)
