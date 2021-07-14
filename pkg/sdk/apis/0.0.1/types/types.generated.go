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
//    vendor, err := UnmarshalVendor(bytes)
//    bytes, err = vendor.Marshal()

package types

import "bytes"
import "errors"
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
	Kind       InterfaceKind       `json:"kind"`               
	Metadata   InterfaceMetadata   `json:"metadata"`           
	OcfVersion string              `json:"ocfVersion"`         
	Revision   string              `json:"revision"`           // Version of the manifest content in the SemVer format.
	Signature  *InterfaceSignature `json:"signature,omitempty"`// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
	Spec       InterfaceSpec       `json:"spec"`               // A container for the Interface specification definition.
}

// A container for the OCF metadata definitions.
type InterfaceMetadata struct {
	Description      string       `json:"description"`               // A short description of the OCF manifest. Must be a non-empty string.
	DisplayName      *string      `json:"displayName,omitempty"`     // The name of the OCF manifest to be displayed in graphical clients.
	DocumentationURL *string      `json:"documentationURL,omitempty"`// Link to documentation page for the OCF manifest.
	IconURL          *string      `json:"iconURL,omitempty"`         // The URL to an icon or a data URL containing an icon.
	Maintainers      []Maintainer `json:"maintainers"`               // The list of maintainers with contact information.
	Name             string       `json:"name"`                      // The name of OCF manifest. Together with the manifest revision property must uniquely; identify this object within the entity sub-tree. Must be a non-empty string. We recommend; using a CLI-friendly name.
	Prefix           *string      `json:"prefix,omitempty"`          // The prefix value is automatically computed and set when storing manifest in Hub.
	SupportURL       *string      `json:"supportURL,omitempty"`      // Link to support page for the OCF manifest.
}

// Holds contact information.
type Maintainer struct {
	Email string  `json:"email"`         // Email address of the person.
	Name  *string `json:"name,omitempty"`// Name of the person.
	URL   *string `json:"url,omitempty"` // URL of the person’s site.
}

// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
type InterfaceSignature struct {
	Hub string `json:"hub"`// The signature signed with the HUB key.
}

// A container for the Interface specification definition.
type InterfaceSpec struct {
	Abstract *bool  `json:"abstract,omitempty"`// If true, the Interface cannot be implemented. CURRENTLY NOT IMPLEMENTED.
	Input    Input  `json:"input"`             // The input schema for Interface action.
	Output   Output `json:"output"`            // The output schema for Interface action.
}

// The input schema for Interface action.
type Input struct {
	Parameters    *Parameters                  `json:"parameters"`             
	TypeInstances map[string]InputTypeInstance `json:"typeInstances,omitempty"`
}

// The input parameters for a given Action.
type Parameter struct {
	JSONSchema *JSONSchema `json:"jsonSchema,omitempty"`
}

// The JSONSchema definition.
type JSONSchema struct {
	Value string `json:"value"`// Inline JSON Schema definition for the parameters.
}

// Object key is an alias of the TypeInstance, used in the Implementation.
type InputTypeInstance struct {
	TypeRef TypeRef `json:"typeRef"`
	Verbs   []Verb  `json:"verbs"`  // The full list of access rights for a given TypeInstance.
}

// The full path to a given Type.
type TypeRef struct {
	Path     string `json:"path"`    // Path of a given Type.
	Revision string `json:"revision"`// Version of the manifest content in the SemVer format.
}

// The output schema for Interface action.
type Output struct {
	TypeInstances map[string]OutputTypeInstance `json:"typeInstances,omitempty"`
}

// Object key is an alias of the TypeInstance, used in the Implementation.
type OutputTypeInstance struct {
	TypeRef *TypeRef `json:"typeRef,omitempty"`
}

// The description of an action and its prerequisites (dependencies). An implementation
// implements at least one interface.
type Implementation struct {
	Kind       ImplementationKind       `json:"kind"`               
	Metadata   ImplementationMetadata   `json:"metadata"`           
	OcfVersion string                   `json:"ocfVersion"`         
	Revision   string                   `json:"revision"`           // Version of the manifest content in the SemVer format.
	Signature  *ImplementationSignature `json:"signature,omitempty"`// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
	Spec       ImplementationSpec       `json:"spec"`               // A container for the Implementation specification definition.
}

// A container for the OCF metadata definitions.
type ImplementationMetadata struct {
	Description      string                       `json:"description"`               // A short description of the OCF manifest. Must be a non-empty string.
	DisplayName      *string                      `json:"displayName,omitempty"`     // The name of the OCF manifest to be displayed in graphical clients.
	DocumentationURL *string                      `json:"documentationURL,omitempty"`// Link to documentation page for the OCF manifest.
	IconURL          *string                      `json:"iconURL,omitempty"`         // The URL to an icon or a data URL containing an icon.
	Maintainers      []Maintainer                 `json:"maintainers"`               // The list of maintainers with contact information.
	Name             string                       `json:"name"`                      // The name of OCF manifest. Together with the manifest revision property must uniquely; identify this object within the entity sub-tree. Must be a non-empty string. We recommend; using a CLI-friendly name.
	Prefix           *string                      `json:"prefix,omitempty"`          // The prefix value is automatically computed and set when storing manifest in Hub.
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

// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
type ImplementationSignature struct {
	Hub string `json:"hub"`// The signature signed with the HUB key.
}

// A container for the Implementation specification definition.
type ImplementationSpec struct {
	Action                      Action                                `json:"action"`                     // Definition of an action that should be executed.
	AdditionalInput             *AdditionalInput                      `json:"additionalInput,omitempty"`  // Specifies additional input for the Implementation.
	AdditionalOutput            *AdditionalOutput                     `json:"additionalOutput,omitempty"` // Specifies additional output for a given Implementation.
	AppVersion                  string                                `json:"appVersion"`                 // The supported application versions in SemVer2 format. Currently not used for filtering of; Implementations.
	Implements                  []Implement                           `json:"implements"`                 // Defines what kind of Interfaces this Implementation fulfills.
	Imports                     []Import                              `json:"imports,omitempty"`          // List of external Interfaces that this Implementation requires to be able to execute the; action.
	OutputTypeInstanceRelations map[string]OutputTypeInstanceRelation `json:"outputTypeInstanceRelations"`// Defines all output TypeInstances to upload with relations between them. It relates to; both optional and required TypeInstances. No TypeInstance name specified here means it; won't be uploaded to Hub after workflow run.
	Requires                    map[string]Require                    `json:"requires,omitempty"`         // List of the system prerequisites that need to be present on the cluster.
}

// Definition of an action that should be executed.
type Action struct {
	Args            map[string]interface{} `json:"args"`           // Holds all parameters that should be passed to the selected runner, for example repoUrl,; or chartName for the Helm3 runner.
	RunnerInterface string                 `json:"runnerInterface"`// The Interface of a Runner, which handles the execution, for example,; cap.interface.runner.helm3.run
}

// Specifies additional input for the Implementation.
type AdditionalInput struct {
	Parameters    map[string]interface{}       `json:"parameters,omitempty"`   // Specifies additional input parameters for the Implementation
	TypeInstances map[string]InputTypeInstance `json:"typeInstances,omitempty"`
}

// Specifies additional output for a given Implementation.
type AdditionalOutput struct {
	TypeInstances map[string]OutputTypeInstance `json:"typeInstances,omitempty"`
}

type Implement struct {
	Path     string  `json:"path"`              // The Interface path, for example cap.interfaces.db.mysql.install
	Revision *string `json:"revision,omitempty"`// The exact Interface revision.
}

type Import struct {
	Alias              *string  `json:"alias,omitempty"`     // The alias for the full name of the imported group name. It can be used later in the; workflow definition instead of using full name.
	AppVersion         *string  `json:"appVersion,omitempty"`// The supported application versions in SemVer2 format. CURRENTLY NOT IMPLEMENTED.
	InterfaceGroupPath string   `json:"interfaceGroupPath"`  // The name of the InterfaceGroup that contains specific actions that you want to import,; for example cap.interfaces.db.mysql
	Methods            []Method `json:"methods"`             // The list of all required actions’ names that must be imported.
}

type Method struct {
	Name     string  `json:"name"`              // The name of the action for a given InterfaceGroup, e.g. install.
	Revision *string `json:"revision,omitempty"`// Revision of the Interface for a given action. If not specified, the latest revision is; used.
}

// Object key is an alias of the TypeInstance, used in the Implementation
type OutputTypeInstanceRelation struct {
	Uses []string `json:"uses,omitempty"`// Contains all dependant TypeInstances
}

// Prefix MUST be an abstract node and represents a core abstract Type e.g.
// cap.core.type.platform. Custom Types are not allowed.
type Require struct {
	AllOf []RequireEntity `json:"allOf,omitempty"`// All of the given types MUST have an TypeInstance on the cluster. Element on the list MUST; resolves to concrete Type.
	AnyOf []RequireEntity `json:"anyOf,omitempty"`// Any (one or more) of the given types MUST have an TypeInstance on the cluster. Element on; the list MUST resolves to concrete Type.
	OneOf []RequireEntity `json:"oneOf,omitempty"`// Exactly one of the given types MUST have an TypeInstance on the cluster. Element on the; list MUST resolves to concrete Type.
}

type RequireEntity struct {
	Alias            *string                `json:"alias,omitempty"`           // If provided, the TypeInstance of the Type, configured in policy, is injected to the; workflow under the alias.
	Name             string                 `json:"name"`                      // The name of the Type. Root prefix can be skipped if it’s a core Type. If it is a custom; Type then it MUST be defined as full path to that Type. Custom Type MUST extend the; abstract node which is defined as a root prefix for that entry. Support for custom Types; is CURRENTLY NOT IMPLEMENTED.
	Revision         string                 `json:"revision"`                  // The exact revision of the given Type.
	ValueConstraints map[string]interface{} `json:"valueConstraints,omitempty"`// Holds the configuration constraints for the given entry. It needs to be valid against the; Type JSONSchema. CURRENTLY NOT IMPLEMENTED.
}

// RepoMetadata stores metadata about the Capact Hub.
type RepoMetadata struct {
	Kind       RepoMetadataKind       `json:"kind"`               
	Metadata   InterfaceMetadata      `json:"metadata"`           
	OcfVersion string                 `json:"ocfVersion"`         
	Revision   string                 `json:"revision"`           // Version of the manifest content in the SemVer format.
	Signature  *RepoMetadataSignature `json:"signature,omitempty"`// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
	Spec       RepoMetadataSpec       `json:"spec"`               // A container for the RepoMetadata definition.
}

// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
type RepoMetadataSignature struct {
	Hub string `json:"hub"`// The signature signed with the HUB key.
}

// A container for the RepoMetadata definition.
type RepoMetadataSpec struct {
	HubVersion     string               `json:"hubVersion"`              // Defines the Hub version in SemVer2 format.
	Implementation *ImplementationClass `json:"implementation,omitempty"`// Holds configuration for the OCF Implementation entities. CURRENTLY NOT IMPLEMENTED.
	OcfVersion     OcfVersion           `json:"ocfVersion"`              // Holds information about supported OCF versions in Hub server.
}

// Holds configuration for the OCF Implementation entities. CURRENTLY NOT IMPLEMENTED.
type ImplementationClass struct {
	AppVersion *AppVersion `json:"appVersion,omitempty"`// Defines the configuration for the appVersion field.
}

// Defines the configuration for the appVersion field.
type AppVersion struct {
	SemVerTaggingStrategy *SemVerTaggingStrategy `json:"semVerTaggingStrategy,omitempty"`// Defines the tagging strategy.
}

// Defines the tagging strategy.
type SemVerTaggingStrategy struct {
	Latest Latest `json:"latest"`// Defines the strategy for which version the tag Latest should be applied. You configure; this while running Hub.
}

// Defines the strategy for which version the tag Latest should be applied. You configure
// this while running Hub.
type Latest struct {
	PointsTo *PointsTo `json:"pointsTo,omitempty"`// An explanation about the purpose of this instance.
}

// Holds information about supported OCF versions in Hub server.
type OcfVersion struct {
	Default   string   `json:"default"`  // The default OCF version that is supported by the Hub. It should be the stored version.
	Supported []string `json:"supported"`// The supported OCF version that Hub is able to serve. In general, the Hub takes the stored; version and converts it to the supported one. CURRENTLY NOT IMPLEMENTED.
}

// Attribute is used to categorize Implementations, Types and TypeInstances. For example,
// you can use `cap.core.attribute.workload.stateful` Attribute to find and filter Stateful
// Implementations.
type Attribute struct {
	Kind       AttributeKind       `json:"kind"`               
	Metadata   InterfaceMetadata   `json:"metadata"`           
	OcfVersion string              `json:"ocfVersion"`         
	Revision   string              `json:"revision"`           // Version of the manifest content in the SemVer format.
	Signature  *AttributeSignature `json:"signature,omitempty"`// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
	Spec       *AttributeSpec      `json:"spec,omitempty"`     // A container for the Attribute specification definition.
}

// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
type AttributeSignature struct {
	Hub string `json:"hub"`// The signature signed with the HUB key.
}

// A container for the Attribute specification definition.
type AttributeSpec struct {
	AdditionalRefs []string `json:"additionalRefs,omitempty"`// List of the full path of additional parent nodes the Attribute is attached to. The parent; nodes MUST reside under “cap.core.attribute” or “cap.attribute” subtree. The connection; means that the Attribute becomes a child of the referenced parent nodes. In a result, the; Attribute has multiple parents.
}

// Primitive, that holds the JSONSchema which describes that Type. It’s also used for
// validation. There are core and custom Types. Type can be also a composition of other
// Types.
type Type struct {
	Kind       TypeKind       `json:"kind"`               
	Metadata   TypeMetadata   `json:"metadata"`           
	OcfVersion string         `json:"ocfVersion"`         
	Revision   string         `json:"revision"`           // Version of the manifest content in the SemVer format.
	Signature  *TypeSignature `json:"signature,omitempty"`// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
	Spec       TypeSpec       `json:"spec"`               // A container for the Type specification definition.
}

// A container for the OCF metadata definitions.
type TypeMetadata struct {
	Description      string                       `json:"description"`               // A short description of the OCF manifest. Must be a non-empty string.
	DisplayName      *string                      `json:"displayName,omitempty"`     // The name of the OCF manifest to be displayed in graphical clients.
	DocumentationURL *string                      `json:"documentationURL,omitempty"`// Link to documentation page for the OCF manifest.
	IconURL          *string                      `json:"iconURL,omitempty"`         // The URL to an icon or a data URL containing an icon.
	Maintainers      []Maintainer                 `json:"maintainers"`               // The list of maintainers with contact information.
	Name             string                       `json:"name"`                      // The name of OCF manifest. Together with the manifest revision property must uniquely; identify this object within the entity sub-tree. Must be a non-empty string. We recommend; using a CLI-friendly name.
	Prefix           *string                      `json:"prefix,omitempty"`          // The prefix value is automatically computed and set when storing manifest in Hub.
	SupportURL       *string                      `json:"supportURL,omitempty"`      // Link to support page for the OCF manifest.
	Attributes       map[string]MetadataAttribute `json:"attributes,omitempty"`      
}

// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
type TypeSignature struct {
	Hub string `json:"hub"`// The signature signed with the HUB key.
}

// A container for the Type specification definition.
type TypeSpec struct {
	AdditionalRefs []string   `json:"additionalRefs,omitempty"`// List of the full path of additional parent nodes the Type is attached to. The parent; nodes MUST reside under “cap.core.type” or “cap.type” subtree. The connection means that; the Type becomes a child of the referenced parent nodes. In a result, the Type has; multiple parents.
	JSONSchema     JSONSchema `json:"jsonSchema"`              
}

// Vendor manifests are currently not used. They will be part of the Hub federation feature.
type Vendor struct {
	Kind       VendorKind        `json:"kind"`               
	Metadata   InterfaceMetadata `json:"metadata"`           
	OcfVersion string            `json:"ocfVersion"`         
	Revision   string            `json:"revision"`           // Version of the manifest content in the SemVer format.
	Signature  *VendorSignature  `json:"signature,omitempty"`// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
	Spec       VendorSpec        `json:"spec"`               // A container for the Vendor specification definition.
}

// Ensures the authenticity and integrity of a given manifest. CURRENTLY NOT IMPLEMENTED.
type VendorSignature struct {
	Hub string `json:"hub"`// The signature signed with the HUB key.
}

// A container for the Vendor specification definition.
type VendorSpec struct {
	Federation Federation `json:"federation"`// Holds configuration for vendor federation.
}

// Holds configuration for vendor federation.
type Federation struct {
	URI string `json:"uri"`// The URI of the external Hub.
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

type VendorKind string
const (
	KindVendor VendorKind = "Vendor"
)

type Parameters struct {
	AnythingArray []interface{}
	Bool          *bool
	Double        *float64
	Integer       *int64
	ParameterMap  map[string]Parameter
	String        *string
}

func (x *Parameters) UnmarshalJSON(data []byte) error {
	x.AnythingArray = nil
	x.ParameterMap = nil
	object, err := unmarshalUnion(data, &x.Integer, &x.Double, &x.Bool, &x.String, true, &x.AnythingArray, false, nil, true, &x.ParameterMap, false, nil, true)
	if err != nil {
		return err
	}
	if object {
	}
	return nil
}

func (x *Parameters) MarshalJSON() ([]byte, error) {
	return marshalUnion(x.Integer, x.Double, x.Bool, x.String, x.AnythingArray != nil, x.AnythingArray, false, nil, x.ParameterMap != nil, x.ParameterMap, false, nil, true)
}

func unmarshalUnion(data []byte, pi **int64, pf **float64, pb **bool, ps **string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) (bool, error) {
	if pi != nil {
		*pi = nil
	}
	if pf != nil {
		*pf = nil
	}
	if pb != nil {
		*pb = nil
	}
	if ps != nil {
		*ps = nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return false, err
	}

	switch v := tok.(type) {
	case json.Number:
		if pi != nil {
			i, err := v.Int64()
			if err == nil {
				*pi = &i
				return false, nil
			}
		}
		if pf != nil {
			f, err := v.Float64()
			if err == nil {
				*pf = &f
				return false, nil
			}
			return false, errors.New("Unparsable number")
		}
		return false, errors.New("Union does not contain number")
	case float64:
		return false, errors.New("Decoder should not return float64")
	case bool:
		if pb != nil {
			*pb = &v
			return false, nil
		}
		return false, errors.New("Union does not contain bool")
	case string:
		if haveEnum {
			return false, json.Unmarshal(data, pe)
		}
		if ps != nil {
			*ps = &v
			return false, nil
		}
		return false, errors.New("Union does not contain string")
	case nil:
		if nullable {
			return false, nil
		}
		return false, errors.New("Union does not contain null")
	case json.Delim:
		if v == '{' {
			if haveObject {
				return true, json.Unmarshal(data, pc)
			}
			if haveMap {
				return false, json.Unmarshal(data, pm)
			}
			return false, errors.New("Union does not contain object")
		}
		if v == '[' {
			if haveArray {
				return false, json.Unmarshal(data, pa)
			}
			return false, errors.New("Union does not contain array")
		}
		return false, errors.New("Cannot handle delimiter")
	}
	return false, errors.New("Cannot unmarshal union")

}

func marshalUnion(pi *int64, pf *float64, pb *bool, ps *string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) ([]byte, error) {
	if pi != nil {
		return json.Marshal(*pi)
	}
	if pf != nil {
		return json.Marshal(*pf)
	}
	if pb != nil {
		return json.Marshal(*pb)
	}
	if ps != nil {
		return json.Marshal(*ps)
	}
	if haveArray {
		return json.Marshal(pa)
	}
	if haveObject {
		return json.Marshal(pc)
	}
	if haveMap {
		return json.Marshal(pm)
	}
	if haveEnum {
		return json.Marshal(pe)
	}
	if nullable {
		return json.Marshal(nil)
	}
	return nil, errors.New("Union must not be null")
}
