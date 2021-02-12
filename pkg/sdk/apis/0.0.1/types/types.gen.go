// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    interface, err := UnmarshalInterface(bytes)
//    bytes, err = interface.Marshal()
//
//    implementation, err := UnmarshalImplementation(bytes)
//    bytes, err = implementation.Marshal()
//
//    repoMetadata, err := UnmarshalRepoMetadata(bytes)
//    bytes, err = repoMetadata.Marshal()
//
//    attribute, err := UnmarshalAttribute(bytes)
//    bytes, err = attribute.Marshal()
//
//    type, err := UnmarshalType(bytes)
//    bytes, err = type.Marshal()
//
//    typeInstance, err := UnmarshalTypeInstance(bytes)
//    bytes, err = typeInstance.Marshal()
//
//    vendor, err := UnmarshalVendor(bytes)
//    bytes, err = vendor.Marshal()

package types

import "encoding/json"

func UnmarshalInterface(data []byte) (Interface, error) {
	var r Interface
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Interface) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalImplementation(data []byte) (Implementation, error) {
	var r Implementation
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Implementation) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRepoMetadata(data []byte) (RepoMetadata, error) {
	var r RepoMetadata
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RepoMetadata) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAttribute(data []byte) (Attribute, error) {
	var r Attribute
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Attribute) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalType(data []byte) (Type, error) {
	var r Type
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Type) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTypeInstance(data []byte) (TypeInstance, error) {
	var r TypeInstance
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TypeInstance) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalVendor(data []byte) (Vendor, error) {
	var r Vendor
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Vendor) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Interface defines an action signature. It describes the action name, input, and output
// parameters.
type Interface struct {
	Kind       InterfaceKind      `json:"kind"`      
	Metadata   InterfaceMetadata  `json:"metadata"`  
	OcfVersion string             `json:"ocfVersion"`
	Revision   string             `json:"revision"`  // Version of the manifest content in the SemVer format.
	Signature  InterfaceSignature `json:"signature"` // Ensures the authenticity and integrity of a given manifest.
	Spec       InterfaceSpec      `json:"spec"`      // A container for the Interface specification definition.
}

// A container for the OCF metadata definitions.
type InterfaceMetadata struct {
	Description      string       `json:"description"`               // A short description of the OCF manifest. Must be a non-empty string.
	DisplayName      *string      `json:"displayName,omitempty"`     // The name of the OCF manifest to be displayed in graphical clients.
	DocumentationURL *string      `json:"documentationURL,omitempty"`// Link to documentation page for the OCF manifest.
	IconURL          *string      `json:"iconURL,omitempty"`         // The URL to an icon or a data URL containing an icon.
	Maintainers      []Maintainer `json:"maintainers"`               // The list of maintainers with contact information.
	Name             string       `json:"name"`                      // The name of OCF manifest that uniquely identifies this object within the entity sub-tree.; Must be a non-empty string. We recommend using a CLI-friendly name.
	Prefix           *string      `json:"prefix,omitempty"`          // The prefix value is automatically computed and set when storing manifest in OCH.
	SupportURL       *string      `json:"supportURL,omitempty"`      // Link to support page for the OCF manifest.
}

// Holds contact information.
type Maintainer struct {
	Email string  `json:"email"`         // Email address of the person.
	Name  *string `json:"name,omitempty"`// Name of the person.
	URL   *string `json:"url,omitempty"` // URL of the person’s site.
}

// Ensures the authenticity and integrity of a given manifest.
type InterfaceSignature struct {
	Och string `json:"och"`// The signature signed with the HUB key.
}

// A container for the Interface specification definition.
type InterfaceSpec struct {
	Abstract *bool  `json:"abstract,omitempty"`// If true, the Interface cannot be implemented.
	Input    Input  `json:"input"`             // The input schema for Interface action.
	Output   Output `json:"output"`            // The output schema for Interface action.
}

// The input schema for Interface action.
type Input struct {
	Parameters    *Parameters                  `json:"parameters,omitempty"`   // The input parameters passed from User
	TypeInstances map[string]InputTypeInstance `json:"typeInstances,omitempty"`
}

// The input parameters passed from User
type Parameters struct {
	JSONSchema *JSONSchema `json:"jsonSchema,omitempty"`
}

// The JSONSchema definition.
type JSONSchema struct {
	Value string `json:"value"`// Inline JSON Schema definition for the parameters.
}

// Prefix is an alias of the TypeInstance, used in the Implementation
type InputTypeInstance struct {
	TypeRef TypeRef `json:"typeRef"`
	Verbs   []Verb  `json:"verbs"`  // The full list of access rights for a given TypeInstance
}

// The full path to the Type from which the TypeInstance is created.
type TypeRef struct {
	Path     string  `json:"path"`              // Path of a given Type
	Revision *string `json:"revision,omitempty"`// Version of the manifest content in the SemVer format.
}

// The output schema for Interface action.
type Output struct {
	TypeInstances map[string]OutputTypeInstance `json:"typeInstances,omitempty"`
}

// Prefix is an alias of the TypeInstance, used in the Implementation
type OutputTypeInstance struct {
	TypeRef *TypeRef `json:"typeRef,omitempty"`
}

// The description of an action and its prerequisites (dependencies). An implementation
// implements at least one interface.
type Implementation struct {
	Kind       ImplementationKind      `json:"kind"`      
	Metadata   ImplementationMetadata  `json:"metadata"`  
	OcfVersion string                  `json:"ocfVersion"`
	Revision   string                  `json:"revision"`  // Version of the manifest content in the SemVer format.
	Signature  ImplementationSignature `json:"signature"` // Ensures the authenticity and integrity of a given manifest.
	Spec       ImplementationSpec      `json:"spec"`      // A container for the Implementation specification definition.
}

// A container for the OCF metadata definitions.
type ImplementationMetadata struct {
	Description      string                       `json:"description"`               // A short description of the OCF manifest. Must be a non-empty string.
	DisplayName      *string                      `json:"displayName,omitempty"`     // The name of the OCF manifest to be displayed in graphical clients.
	DocumentationURL *string                      `json:"documentationURL,omitempty"`// Link to documentation page for the OCF manifest.
	IconURL          *string                      `json:"iconURL,omitempty"`         // The URL to an icon or a data URL containing an icon.
	Maintainers      []Maintainer                 `json:"maintainers"`               // The list of maintainers with contact information.
	Name             string                       `json:"name"`                      // The name of OCF manifest that uniquely identifies this object within the entity sub-tree.; Must be a non-empty string. We recommend using a CLI-friendly name.
	Prefix           *string                      `json:"prefix,omitempty"`          // The prefix value is automatically computed and set when storing manifest in OCH.
	SupportURL       *string                      `json:"supportURL,omitempty"`      // Link to support page for the OCF manifest.
	Attributes       map[string]MetadataAttribute `json:"attributes,omitempty"`      
	License          License                      `json:"license"`                   // This entry allows you to specify a license, so people know how they are permitted to use; it, and what kind of restrictions you are placing on it.
}

// The attribute object contains OCF Attributes references. It provides generic
// categorization for Implementations, Types and TypeInstances. Attributes are used to
// filter out a specific Implementation.
type MetadataAttribute struct {
	Revision string `json:"revision"`// The exact Attribute revision.
}

// This entry allows you to specify a license, so people know how they are permitted to use
// it, and what kind of restrictions you are placing on it.
type License struct {
	Name *string `json:"name,omitempty"`// If you are using a common license such as BSD-2-Clause or MIT, add a current SPDX license; identifier for the license you’re using e.g. BSD-3-Clause. If your package is licensed; under multiple common licenses, use an SPDX license expression syntax version 2.0 string,; e.g. (ISC OR GPL-3.0)
	Ref  *string `json:"ref,omitempty"` // If you are using a license that hasn’t been assigned an SPDX identifier, or if you are; using a custom license, use the direct link to the license file e.g.; https://raw.githubusercontent.com/project/v1/license.md. The resource under given link; MUST be immutable and publicly accessible.
}

// Ensures the authenticity and integrity of a given manifest.
type ImplementationSignature struct {
	Och string `json:"och"`// The signature signed with the HUB key.
}

// A container for the Implementation specification definition.
type ImplementationSpec struct {
	Action           Action             `json:"action"`                    // An explanation about the purpose of this instance.
	AdditionalInput  *AdditionalInput   `json:"additionalInput,omitempty"` // Specifies additional input for a given Implementation
	AdditionalOutput *AdditionalOutput  `json:"additionalOutput,omitempty"`// Specifies additional output for a given Implementation
	AppVersion       string             `json:"appVersion"`                // The supported application versions in SemVer2 format.
	Implements       []Implement        `json:"implements"`                // Defines what kind of interfaces this implementation fulfills.
	Imports          []Import           `json:"imports,omitempty"`         // List of external Interfaces that this Implementation requires to be able to execute the; action.
	Requires         map[string]Require `json:"requires,omitempty"`        // List of the system prerequisites that need to be present on the cluster. There has to be; an Instance for every concrete type.
}

// An explanation about the purpose of this instance.
type Action struct {
	Args            map[string]interface{} `json:"args"`           // Holds all parameters that should be passed to the selected runner, for example repoUrl,; or chartName for the Helm3 runner.
	RunnerInterface string                 `json:"runnerInterface"`// The Interface of a Runner, which handles the execution, for example,; cap.interface.runner.helm3.run
}

// Specifies additional input for a given Implementation
type AdditionalInput struct {
	TypeInstances map[string]InputTypeInstance `json:"typeInstances,omitempty"`
}

// Specifies additional output for a given Implementation
type AdditionalOutput struct {
	TypeInstanceRelations map[string]TypeInstanceRelation `json:"typeInstanceRelations"`  // Defines the relations between all output TypeInstances
	TypeInstances         map[string]OutputTypeInstance   `json:"typeInstances,omitempty"`
}

// Prefix is an alias of the TypeInstance, used in the Implementation
type TypeInstanceRelation struct {
	Uses []string `json:"uses,omitempty"`// Uses contains all dependant TypeInstances
}

type Implement struct {
	Path     string  `json:"path"`              // The Interface path, for example cap.interfaces.db.mysql.install
	Revision *string `json:"revision,omitempty"`// The exact Interface revision.
}

type Import struct {
	Alias              *string         `json:"alias,omitempty"`     // The alias for the full name of the imported group name. It can be used later in the; workflow definition instead of using full name.
	AppVersion         *string         `json:"appVersion,omitempty"`// The supported application versions in SemVer2 format.
	InterfaceGroupPath string          `json:"interfaceGroupPath"`  // The name of the Interface Group that contains specific actions that you want to import,; for example cap.interfaces.db.mysql
	Methods            []MethodElement `json:"methods"`             // The list of all required actions’ names that must be imported.
}

type MethodElement struct {
	Name     string  `json:"name"`              // The name of the action for a given Interface group, e.g. install
	Revision *string `json:"revision,omitempty"`// Revision of the Interface for a given action. If not specified, the latest revision is; used.
}

// Prefix MUST be an abstract node and represents a core abstract Type e.g.
// cap.core.type.platform. Custom Types are not allowed.
type Require struct {
	AllOf []RequireEntity `json:"allOf,omitempty"`// All of the given types MUST have an Instance on the cluster. Element on the list MUST; resolves to concrete Type.
	AnyOf []RequireEntity `json:"anyOf,omitempty"`// Any (one or more) of the given types MUST have an Instance on the cluster. Element on the; list MUST resolves to concrete Type.
	OneOf []RequireEntity `json:"oneOf,omitempty"`// Exactly one of the given types MUST have an Instance on the cluster. Element on the list; MUST resolves to concrete Type.
}

type RequireEntity struct {
	Alias            *string                `json:"alias,omitempty"`           // If provided, the TypeInstance of the Type, configured in policy, is injected to the; workflow under the alias.
	Name             string                 `json:"name"`                      // The name of the Type. Root prefix can be skipped if it’s a core Type. If it is a custom; Type then it MUST be defined as full path to that Type. Custom Type MUST extend the; abstract node which is defined as a root prefix for that entry.
	Revision         string                 `json:"revision"`                  // The exact revision of the given Type.
	ValueConstraints map[string]interface{} `json:"valueConstraints,omitempty"`// Holds the configuration constraints for the given entry. It needs to be valid against the; Type JSONSchema.
}

// RepoMetadata stores metadata about the Open Capability Hub.
type RepoMetadata struct {
	Kind       RepoMetadataKind      `json:"kind"`      
	Metadata   InterfaceMetadata     `json:"metadata"`  
	OcfVersion string                `json:"ocfVersion"`
	Revision   string                `json:"revision"`  // Version of the manifest content in the SemVer format.
	Signature  RepoMetadataSignature `json:"signature"` // Ensures the authenticity and integrity of a given manifest.
	Spec       RepoMetadataSpec      `json:"spec"`      // A container for the RepoMetadata definition.
}

// Ensures the authenticity and integrity of a given manifest.
type RepoMetadataSignature struct {
	Och string `json:"och"`// The signature signed with the HUB key.
}

// A container for the RepoMetadata definition.
type RepoMetadataSpec struct {
	Implementation *ImplementationClass `json:"implementation,omitempty"`// Holds configuration for the OCF Implementation entities.
	OcfVersion     OcfVersion           `json:"ocfVersion"`              // Holds information about supported OCF versions in OCH server.
	OchVersion     string               `json:"ochVersion"`              // Defines the OCH version in SemVer2 format.
}

// Holds configuration for the OCF Implementation entities.
type ImplementationClass struct {
	AppVersion *AppVersion `json:"appVersion,omitempty"`// Defines the configuration for the appVersion field.
}

// Defines the configuration for the appVersion field.
type AppVersion struct {
	SemVerTaggingStrategy *SemVerTaggingStrategy `json:"semVerTaggingStrategy,omitempty"`// Defines the tagging strategy.
}

// Defines the tagging strategy.
type SemVerTaggingStrategy struct {
	Latest Latest `json:"latest"`// Defines the strategy for which version the tag Latest should be applied. You configure; this while running OCH.
}

// Defines the strategy for which version the tag Latest should be applied. You configure
// this while running OCH.
type Latest struct {
	PointsTo *PointsTo `json:"pointsTo,omitempty"`// An explanation about the purpose of this instance.
}

// Holds information about supported OCF versions in OCH server.
type OcfVersion struct {
	Default   string   `json:"default"`  // The default OCF version that is supported by the OCH. It should be the stored version.
	Supported []string `json:"supported"`// The supported OCF version that OCH is able to serve. In general, the OCH takes the stored; version and converts it to the supported one.
}

// Attribute is a primitive, which is used to categorize Implementations. You can use
// Attributes to find and filter Implementations.
type Attribute struct {
	Kind       AttributeKind      `json:"kind"`          
	Metadata   InterfaceMetadata  `json:"metadata"`      
	OcfVersion string             `json:"ocfVersion"`    
	Revision   string             `json:"revision"`      // Version of the manifest content in the SemVer format.
	Signature  AttributeSignature `json:"signature"`     // Ensures the authenticity and integrity of a given manifest.
	Spec       *AttributeSpec     `json:"spec,omitempty"`// A container for the Attribute specification definition.
}

// Ensures the authenticity and integrity of a given manifest.
type AttributeSignature struct {
	Och string `json:"och"`// The signature signed with the HUB key.
}

// A container for the Attribute specification definition.
type AttributeSpec struct {
	AdditionalRefs []string `json:"additionalRefs,omitempty"`// List of the full path of additional parent nodes the Attribute is attached to. The parent; nodes MUST reside under “cap.core.attribute” or “cap.attribute” subtree. The connection; means that the Attribute becomes a child of the referenced parent nodes. In a result, the; Attribute has multiple parents.
}

// Primitive, that holds the JSONSchema which describes that Type. It’s also used for
// validation. There are core and custom Types. Type can be also a composition of other
// Types.
type Type struct {
	Kind       TypeKind      `json:"kind"`      
	Metadata   TypeMetadata  `json:"metadata"`  
	OcfVersion string        `json:"ocfVersion"`
	Revision   string        `json:"revision"`  // Version of the manifest content in the SemVer format.
	Signature  TypeSignature `json:"signature"` // Ensures the authenticity and integrity of a given manifest.
	Spec       TypeSpec      `json:"spec"`      // A container for the Type specification definition.
}

// A container for the OCF metadata definitions.
type TypeMetadata struct {
	Description      string                       `json:"description"`               // A short description of the OCF manifest. Must be a non-empty string.
	DisplayName      *string                      `json:"displayName,omitempty"`     // The name of the OCF manifest to be displayed in graphical clients.
	DocumentationURL *string                      `json:"documentationURL,omitempty"`// Link to documentation page for the OCF manifest.
	IconURL          *string                      `json:"iconURL,omitempty"`         // The URL to an icon or a data URL containing an icon.
	Maintainers      []Maintainer                 `json:"maintainers"`               // The list of maintainers with contact information.
	Name             string                       `json:"name"`                      // The name of OCF manifest that uniquely identifies this object within the entity sub-tree.; Must be a non-empty string. We recommend using a CLI-friendly name.
	Prefix           *string                      `json:"prefix,omitempty"`          // The prefix value is automatically computed and set when storing manifest in OCH.
	SupportURL       *string                      `json:"supportURL,omitempty"`      // Link to support page for the OCF manifest.
	Attributes       map[string]MetadataAttribute `json:"attributes,omitempty"`      
}

// Ensures the authenticity and integrity of a given manifest.
type TypeSignature struct {
	Och string `json:"och"`// The signature signed with the HUB key.
}

// A container for the Type specification definition.
type TypeSpec struct {
	AdditionalRefs []string   `json:"additionalRefs,omitempty"`// List of the full path of additional parent nodes the Type is attached to. The parent; nodes MUST reside under “cap.core.type” or “cap.type” subtree. The connection means that; the Type becomes a child of the referenced parent nodes. In a result, the Type has; multiple parents.
	JSONSchema     JSONSchema `json:"jsonSchema"`              
}

// TypeInstance is an instance a given Type. It stores JSON object that is valid against
// JSON Schema from Type.
type TypeInstance struct {
	Kind            TypeInstanceKind     `json:"kind"`           
	Metadata        TypeInstanceMetadata `json:"metadata"`       
	OcfVersion      string               `json:"ocfVersion"`     
	ResourceVersion int64                `json:"resourceVersion"`// Resource version of the manifest content. Every time the manifest content changes, the; resource version is updated automatically on server side.
	Spec            TypeInstanceSpec     `json:"spec"`           // A container for the TypeInstance specification definition.
}

type TypeInstanceMetadata struct {
	Attributes map[string]MetadataAttribute `json:"attributes,omitempty"`
	ID         string                       `json:"id"`                  // The unique identifier of a given TypeInstance.
}

// A container for the TypeInstance specification definition.
type TypeInstanceSpec struct {
	Instrumentation *Instrumentation       `json:"instrumentation,omitempty"`// Holds the details about instrumentation for given Type Instance
	TypeRef         TypeRef                `json:"typeRef"`                  
	Value           map[string]interface{} `json:"value"`                    // Holds the JSON object for a given TypeInstance with arbitrary data. It needs to be valid; against the Type JSONSchema.
}

// Holds the details about instrumentation for given Type Instance
type Instrumentation struct {
	Health  *Health  `json:"health,omitempty"` // Details about health of a given TypeInstance
	Metrics *Metrics `json:"metrics,omitempty"`// Details about metrics exposed for a given TypeInstance
}

// Details about health of a given TypeInstance
type Health struct {
	Method *MethodEnum `json:"method,omitempty"`// Method of HTTP request
	URL    *string     `json:"url,omitempty"`   // URL of health endpoint
}

// Details about metrics exposed for a given TypeInstance
type Metrics struct {
	Dashboards []Dashboard `json:"dashboards,omitempty"`// Metrics dashboards
	Regex      *string     `json:"regex,omitempty"`     // Regex for exposed metrics
	URL        *string     `json:"url,omitempty"`       // URL of metrics endpoint
}

// Stores details of a metrics dashboard
type Dashboard struct {
	URL *string `json:"url,omitempty"`// URL of metrics dashboard
}

// Remote OCH repositories can be mounted under the vendor sub-tree in the local repository.
// OCF Vendor manifest stores connection details of the external OCH, such as URI of the
// repository (base path) or federation strategy.
type Vendor struct {
	Kind       VendorKind        `json:"kind"`      
	Metadata   InterfaceMetadata `json:"metadata"`  
	OcfVersion string            `json:"ocfVersion"`
	Revision   string            `json:"revision"`  // Version of the manifest content in the SemVer format.
	Signature  VendorSignature   `json:"signature"` // Ensures the authenticity and integrity of a given manifest.
	Spec       VendorSpec        `json:"spec"`      // A container for the Vendor specification definition.
}

// Ensures the authenticity and integrity of a given manifest.
type VendorSignature struct {
	Och string `json:"och"`// The signature signed with the HUB key.
}

// A container for the Vendor specification definition.
type VendorSpec struct {
	Federation Federation `json:"federation"`// Holds configuration for vendor federation.
}

// Holds configuration for vendor federation.
type Federation struct {
	URI string `json:"uri"`// The URI of the external OCH.
}

type InterfaceKind string
const (
	KindInterface InterfaceKind = "Interface"
)

type Verb string
const (
	VerbCreate Verb = "create"
	VerbDelete Verb = "delete"
	VerbGet Verb = "get"
	VerbList Verb = "list"
	VerbUpdate Verb = "update"
)

type ImplementationKind string
const (
	KindImplementation ImplementationKind = "Implementation"
)

type RepoMetadataKind string
const (
	KindRepoMetadata RepoMetadataKind = "RepoMetadata"
)

// An explanation about the purpose of this instance.
type PointsTo string
const (
	PointsToEdge PointsTo = "Edge"
	PointsToStable PointsTo = "Stable"
)

type AttributeKind string
const (
	KindAttribute AttributeKind = "Attribute"
)

type TypeKind string
const (
	KindType TypeKind = "Type"
)

type TypeInstanceKind string
const (
	KindTypeInstance TypeInstanceKind = "TypeInstance"
)

// Method of HTTP request
type MethodEnum string
const (
	MethodGET MethodEnum = "GET"
	MethodPOST MethodEnum = "POST"
)

type VendorKind string
const (
	KindVendor VendorKind = "Vendor"
)
