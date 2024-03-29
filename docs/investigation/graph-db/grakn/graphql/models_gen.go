// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package graphql

import (
	"fmt"
	"io"
	"strconv"
)

type MetadataBaseFields interface {
	IsMetadataBaseFields()
}

type TypeInstance interface {
	IsTypeInstance()
}

type GenericMetadata struct {
	Name             string        `json:"name"`
	Prefix           string        `json:"prefix"`
	Path             string        `json:"path"`
	DisplayName      string        `json:"displayName"`
	Description      string        `json:"description"`
	Maintainers      []*Maintainer `json:"maintainers"`
	DocumentationURL string        `json:"documentationURL"`
	SupportURL       string        `json:"supportURL"`
	IconURL          string        `json:"iconURL"`
}

func (GenericMetadata) IsMetadataBaseFields() {}

type Implementation struct {
	Name      string                    `json:"name"`
	Prefix    string                    `json:"prefix"`
	Path      string                    `json:"path"`
	Revisions []*ImplementationRevision `json:"revisions"`
}

type ImplementationAction struct {
	// The Interface or Implementation of a runner, which handles the execution, for example, cap.interface.runner.helm3.run
	RunnerInterface string      `json:"runnerInterface"`
	Args            interface{} `json:"args"`
}

type ImplementationAdditionalInput struct {
	TypeInstances []*InputTypeInstance `json:"typeInstances"`
}

type ImplementationAdditionalOutput struct {
	TypeInstances         []*OutputTypeInstance       `json:"typeInstances"`
	TypeInstanceRelations []*TypeInstanceRelationItem `json:"typeInstanceRelations"`
}

type ImplementationFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
	// If provided, Implementations are filtered by the ones that have satisfied requirements with provided TypeInstance values.
	// For example, to find all Implementations that can be run on a given system, user can provide values of all existing TypeInstances.
	RequirementsSatisfiedBy []*TypeInstanceValue `json:"requirementsSatisfiedBy"`
	Tags                    []*TagFilterInput    `json:"tags"`
}

type ImplementationImport struct {
	InterfaceGroupPath string                        `json:"interfaceGroupPath"`
	Alias              *string                       `json:"alias"`
	AppVersion         *string                       `json:"appVersion"`
	Methods            []*ImplementationImportMethod `json:"methods"`
}

type ImplementationImportMethod struct {
	Name string `json:"name"`
	// If not provided, latest revision for a given Interface is used
	Revision *string `json:"revision"`
}

type ImplementationMetadata struct {
	Name             string         `json:"name"`
	Prefix           *string        `json:"prefix"`
	Path             *string        `json:"path"`
	DisplayName      *string        `json:"displayName"`
	Description      string         `json:"description"`
	Maintainers      []*Maintainer  `json:"maintainers"`
	DocumentationURL *string        `json:"documentationURL"`
	SupportURL       *string        `json:"supportURL"`
	IconURL          *string        `json:"iconURL"`
	Tags             []*TagRevision `json:"tags"`
}

func (ImplementationMetadata) IsMetadataBaseFields() {}

type ImplementationRequirement struct {
	Prefix string                           `json:"prefix"`
	OneOf  []*ImplementationRequirementItem `json:"oneOf"`
	AnyOf  []*ImplementationRequirementItem `json:"anyOf"`
	AllOf  []*ImplementationRequirementItem `json:"allOf"`
}

type ImplementationRequirementItem struct {
	TypeRef *TypeReference `json:"typeRef"`
	// Holds the configuration constraints for the given entry based on Type value.
	// Currently not supported.
	ValueConstraints interface{} `json:"valueConstraints"`
}

type ImplementationRevision struct {
	Metadata *ImplementationMetadata `json:"metadata"`
	Revision string                  `json:"revision"`
}

type ImplementationSpec struct {
	AppVersion       string                          `json:"appVersion"`
	Implements       []*InterfaceReference           `json:"implements"`
	Requires         []*ImplementationRequirement    `json:"requires"`
	Imports          []*ImplementationImport         `json:"imports"`
	Action           *ImplementationAction           `json:"action"`
	AdditionalInput  *ImplementationAdditionalInput  `json:"additionalInput"`
	AdditionalOutput *ImplementationAdditionalOutput `json:"additionalOutput"`
}

type InputParameters struct {
	JSONSchema interface{} `json:"jsonSchema"`
}

type InputTypeInstance struct {
	Name    string                      `json:"name"`
	TypeRef *TypeReference              `json:"typeRef"`
	Verbs   []TypeInstanceOperationVerb `json:"verbs"`
}

func (InputTypeInstance) IsTypeInstance() {}

type Interface struct {
	Name      string               `json:"name"`
	Prefix    string               `json:"prefix"`
	Path      string               `json:"path"`
	Revision  *InterfaceRevision   `json:"revision"`
	Revisions []*InterfaceRevision `json:"revisions"`
}

type InterfaceFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
}

type InterfaceGroup struct {
	Metadata   *GenericMetadata `json:"metadata"`
	Signature  *Signature       `json:"signature"`
	Interfaces []*Interface     `json:"interfaces"`
}

type InterfaceGroupFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
}

type InterfaceInput struct {
	Parameters    *InputParameters     `json:"parameters"`
	TypeInstances []*InputTypeInstance `json:"typeInstances"`
}

type InterfaceOutput struct {
	TypeInstances []*OutputTypeInstance `json:"typeInstances"`
}

type InterfaceReference struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

type InterfaceRevision struct {
	Metadata        *GenericMetadata  `json:"metadata"`
	Revision        string            `json:"revision"`
	Implementations []*Implementation `json:"implementations"`
}

type InterfaceSpec struct {
	Input  *InterfaceInput  `json:"input"`
	Output *InterfaceOutput `json:"output"`
}

type Maintainer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	URL   string `json:"url"`
}

type OutputTypeInstance struct {
	Name    string         `json:"name"`
	TypeRef *TypeReference `json:"typeRef"`
}

func (OutputTypeInstance) IsTypeInstance() {}

type Signature struct {
	Och string `json:"och"`
}

type Tag struct {
	Name           string         `json:"name"`
	Prefix         string         `json:"prefix"`
	Path           string         `json:"path"`
	LatestRevision *TagRevision   `json:"latestRevision"`
	Revision       *TagRevision   `json:"revision"`
	Revisions      []*TagRevision `json:"revisions"`
}

type TagFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
}

type TagFilterInput struct {
	Path string      `json:"path"`
	Rule *FilterRule `json:"rule"`
	// If not provided, latest revision for a given Tag is used
	Revision *string `json:"revision"`
}

type TagRevision struct {
	Metadata  *GenericMetadata `json:"metadata"`
	Revision  string           `json:"revision"`
	Spec      *TagSpec         `json:"spec"`
	Signature *Signature       `json:"signature"`
}

type TagSpec struct {
	AdditionalRefs []string `json:"additionalRefs"`
}

type Type struct {
	Name           string          `json:"name"`
	Prefix         string          `json:"prefix"`
	Path           string          `json:"path"`
	LatestRevision *TypeRevision   `json:"latestRevision"`
	Revision       *TypeRevision   `json:"revision"`
	Revisions      []*TypeRevision `json:"revisions"`
}

type TypeFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
}

type TypeInstanceRelationItem struct {
	TypeInstanceName string `json:"typeInstanceName"`
	// Contains list of Type Instance names, which a given TypeInstance uses (depends on)
	Uses []string `json:"uses"`
}

type TypeInstanceValue struct {
	TypeRef *TypeReferenceInput `json:"typeRef"`
	// Value of the available requirement. If not provided, all valueConstraints conditions are treated as satisfied.
	// Currently not supported.
	Value interface{} `json:"value"`
}

type TypeMetadata struct {
	Name             string         `json:"name"`
	Prefix           *string        `json:"prefix"`
	Path             *string        `json:"path"`
	DisplayName      *string        `json:"displayName"`
	Description      string         `json:"description"`
	Maintainers      []*Maintainer  `json:"maintainers"`
	DocumentationURL *string        `json:"documentationURL"`
	SupportURL       *string        `json:"supportURL"`
	IconURL          *string        `json:"iconURL"`
	Tags             []*TagRevision `json:"tags"`
}

func (TypeMetadata) IsMetadataBaseFields() {}

type TypeReference struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

type TypeReferenceInput struct {
	Path string `json:"path"`
	// If not provided, latest revision for a given Type is used
	Revision *string `json:"revision"`
}

type TypeRevision struct {
	Metadata  *TypeMetadata `json:"metadata"`
	Revision  string        `json:"revision"`
	Spec      *TypeSpec     `json:"spec"`
	Signature *Signature    `json:"signature"`
}

type TypeSpec struct {
	AdditionalRefs []string    `json:"additionalRefs"`
	JSONSchema     interface{} `json:"jsonSchema"`
}

type FilterRule string

const (
	FilterRuleInclude FilterRule = "INCLUDE"
	FilterRuleExclude FilterRule = "EXCLUDE"
)

var AllFilterRule = []FilterRule{
	FilterRuleInclude,
	FilterRuleExclude,
}

func (e FilterRule) IsValid() bool {
	switch e {
	case FilterRuleInclude, FilterRuleExclude:
		return true
	}
	return false
}

func (e FilterRule) String() string {
	return string(e)
}

func (e *FilterRule) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FilterRule(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FilterRule", str)
	}
	return nil
}

func (e FilterRule) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type TypeInstanceOperationVerb string

const (
	TypeInstanceOperationVerbCreate TypeInstanceOperationVerb = "CREATE"
	TypeInstanceOperationVerbGet    TypeInstanceOperationVerb = "GET"
	TypeInstanceOperationVerbList   TypeInstanceOperationVerb = "LIST"
	TypeInstanceOperationVerbUpdate TypeInstanceOperationVerb = "UPDATE"
	TypeInstanceOperationVerbDelete TypeInstanceOperationVerb = "DELETE"
)

var AllTypeInstanceOperationVerb = []TypeInstanceOperationVerb{
	TypeInstanceOperationVerbCreate,
	TypeInstanceOperationVerbGet,
	TypeInstanceOperationVerbList,
	TypeInstanceOperationVerbUpdate,
	TypeInstanceOperationVerbDelete,
}

func (e TypeInstanceOperationVerb) IsValid() bool {
	switch e {
	case TypeInstanceOperationVerbCreate, TypeInstanceOperationVerbGet, TypeInstanceOperationVerbList, TypeInstanceOperationVerbUpdate, TypeInstanceOperationVerbDelete:
		return true
	}
	return false
}

func (e TypeInstanceOperationVerb) String() string {
	return string(e)
}

func (e *TypeInstanceOperationVerb) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TypeInstanceOperationVerb(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TypeInstanceOperationVerb", str)
	}
	return nil
}

func (e TypeInstanceOperationVerb) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
