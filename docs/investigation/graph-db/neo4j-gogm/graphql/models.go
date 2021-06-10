package graphql

import (
	"fmt"
	"github.com/mindstand/gogm"
	"io"
	"strconv"
)

// inputs

type ImplementationFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
	// If provided, Implementations are filtered by the ones that have satisfied requirements with provided TypeInstance values.
	// For example, to find all Implementations that can be run on a given system, user can provide values of all existing TypeInstances.
	RequirementsSatisfiedBy []*TypeInstanceValue `json:"requirementsSatisfiedBy"`
	Tags                    []*TagFilterInput    `json:"tags"`
}

type InterfaceFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
}

type InterfaceGroupFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
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

type TypeFilter struct {
	PrefixPattern *string `json:"prefixPattern"`
}

type TypeInstanceValue struct {
	TypeRef *TypeReferenceInput `json:"typeRef"`
	// Value of the available requirement. If not provided, all valueConstraints conditions are treated as satisfied.
	// Currently not supported.
	Value interface{} `json:"value"`
}

type TypeReferenceInput struct {
	Path string `json:"path"`
	// If not provided, latest revision for a given Type is used
	Revision *string `json:"revision"`
}

// other types

type MetadataBaseFields interface {
	IsMetadataBaseFields()
}

type TypeInstance interface {
	IsTypeInstance()
}

type GenericMetadata struct {
	gogm.BaseNode `json:"-"`

	InterfaceGroup *InterfaceGroup `json:"-" gogm:"direction=incoming;relationship=describedWith"`
	InterfaceRevision *InterfaceRevision `json:"-" gogm:"direction=incoming;relationship=describedWith"`

	Name             string        `json:"name" gogm:"name=name"`
	Prefix           string       `json:"prefix" gogm:"name=prefix"`
	Path             string       `json:"path" gogm:"name=path"`
	DisplayName      string       `json:"displayName" gogm:"name=displayName"`
	Description      string        `json:"description" gogm:"name=description"`
	Maintainers      []*Maintainer `json:"maintainers" gogm:"direction=outgoing;relationship=maintainedBy"`
	DocumentationURL string       `json:"documentationURL" gogm:"name=documentationURL"`
	SupportURL       string       `json:"supportURL" gogm:"name=supportURL"`
	IconURL          string       `json:"iconURL" gogm:"name=iconURL"`
}

func (GenericMetadata) IsMetadataBaseFields() {}

type Implementation struct {
	gogm.BaseNode `json:"-"`

	Name           string                    `json:"name" gogm:"name=name"`
	Prefix         string                    `json:"prefix" gogm:"name=prefix"`
	Path           string                    `json:"path" gogm:"name=path"`
	LatestRevision *ImplementationRevision   `json:"latestRevision" gogm:"name=latestRevision"`
	Revision       *ImplementationRevision   `json:"revision" gogm:"name=revision"`
	Revisions      []*ImplementationRevision `json:"revisions" gogm:"name=revisions"`
}

type ImplementationAction struct {
	gogm.BaseNode `json:"-"`

	// The Interface or Implementation of a runner, which handles the execution, for example, cap.interface.runner.helm3.run
	RunnerInterface string      `json:"runnerInterface" gogm:"name=runnerInterface"`
	Args            interface{} `json:"args" gogm:"name=args"`
}

type ImplementationAdditionalInput struct {
	gogm.BaseNode `json:"-"`

	TypeInstances []*InputTypeInstance `json:"typeInstances" gogm:"name=typeInstances"`
}

type ImplementationAdditionalOutput struct {
	gogm.BaseNode `json:"-"`

	TypeInstances         []*OutputTypeInstance       `json:"typeInstances" gogm:"name=typeInstances"`
	TypeInstanceRelations []*TypeInstanceRelationItem `json:"typeInstanceRelations" gogm:"name=typeInstanceRelations"`
}

type ImplementationImport struct {
	gogm.BaseNode `json:"-"`

	InterfaceGroupPath string                        `json:"interfaceGroupPath" gogm:"name=interfaceGroupPath"`
	Alias              *string                       `json:"alias" gogm:"name=alias"`
	AppVersion         *string                       `json:"appVersion" gogm:"name=appVersion"`
	Methods            []*ImplementationImportMethod `json:"methods" gogm:"name=methods"`
}

type ImplementationImportMethod struct {
	gogm.BaseNode `json:"-"`

	Name string `json:"name" gogm:"name=name"`
	// If not provided, latest revision for a given Interface is used
	Revision *string `json:"revision" gogm:"name=revision"`
}

type ImplementationMetadata struct {
	gogm.BaseNode `json:"-"`

	Name             string         `json:"name" gogm:"name=name"`
	Prefix           *string        `json:"prefix" gogm:"name=prefix"`
	Path             *string        `json:"path" gogm:"name=path"`
	DisplayName      *string        `json:"displayName" gogm:"name=displayName"`
	Description      string         `json:"description" gogm:"name=description"`
	Maintainers      []*Maintainer  `json:"maintainers" gogm:"name=maintainers"`
	DocumentationURL *string        `json:"documentationURL" gogm:"name=documentationURL"`
	SupportURL       *string        `json:"supportURL" gogm:"name=supportURL"`
	IconURL          *string        `json:"iconURL" gogm:"name=iconURL"`
	Tags             []*TagRevision `json:"tags" gogm:"name=tags"`
}

func (ImplementationMetadata) IsMetadataBaseFields() {}

type ImplementationRequirement struct {
	gogm.BaseNode `json:"-"`

	Prefix string                           `json:"prefix" gogm:"name=prefix"`
	OneOf  []*ImplementationRequirementItem `json:"oneOf" gogm:"name=oneOf"`
	AnyOf  []*ImplementationRequirementItem `json:"anyOf" gogm:"name=anyOf"`
	AllOf  []*ImplementationRequirementItem `json:"allOf" gogm:"name=allOf"`
}

type ImplementationRequirementItem struct {
	gogm.BaseNode `json:"-"`

	TypeRef *TypeReference `json:"typeRef" gogm:"name=typeRef"`
	// Holds the configuration constraints for the given entry based on Type value.
	// Currently not supported.
	ValueConstraints interface{} `json:"valueConstraints" gogm:"name=valueConstraints"`
}

type ImplementationRevision struct {
	gogm.BaseNode `json:"-"`

	Metadata   *ImplementationMetadata `json:"metadata" gogm:"name=metadata"`
	Revision   string                  `json:"revision" gogm:"name=revision"`
	Spec       *ImplementationSpec     `json:"spec" gogm:"name=spec"`
	Interfaces []*Interface            `json:"interfaces" gogm:"name=interfaces"`
	Signature  *Signature              `json:"signature" gogm:"name=signature"`
}

type ImplementationSpec struct {
	gogm.BaseNode `json:"-"`

	AppVersion       string                          `json:"appVersion" gogm:"name=appVersion"`
	Implements       []*InterfaceReference           `json:"implements" gogm:"name=implements"`
	Requires         []*ImplementationRequirement    `json:"requires" gogm:"name=requires"`
	Imports          []*ImplementationImport         `json:"imports" gogm:"name=imports"`
	Action           *ImplementationAction           `json:"action" gogm:"name=action"`
	AdditionalInput  *ImplementationAdditionalInput  `json:"additionalInput" gogm:"name=additionalInput"`
	AdditionalOutput *ImplementationAdditionalOutput `json:"additionalOutput" gogm:"name=additionalOutput"`
}

type InputParameters struct {
	gogm.BaseNode `json:"-"`

	JSONSchema interface{} `json:"jsonSchema" gogm:"name=jsonSchema"`
}

type InputTypeInstance struct {
	gogm.BaseNode `json:"-"`

	Name    string                      `json:"name" gogm:"name=name"`
	TypeRef *TypeReference              `json:"typeRef" gogm:"name=typeRef"`
	Verbs   []TypeInstanceOperationVerb `json:"verbs" gogm:"name=verbs"`
}

func (InputTypeInstance) IsTypeInstance() {}

type Interface struct {
	gogm.BaseNode `json:"-"`

	InterfaceGroup *InterfaceGroup `json:"-" gogm:"direction=incoming;relationship=contains"`

	Name   string `json:"name" gogm:"name=name"`
	Prefix string `json:"prefix" gogm:"name=prefix"`
	Path   string `json:"path" gogm:"name=path"`
	//LatestRevision *InterfaceRevision   `json:"latestRevision" gogm:"direction=outgoing;relationship=latest_revision"`
	//Revision       *InterfaceRevision   `json:"revision" gogm:"direction=outgoing;relationship=revision"`
	Revisions      []*InterfaceRevision `json:"revisions" gogm:"direction=outgoing;relationship=revision"`
}

type InterfaceGroup struct {
	gogm.BaseNode `json:"-"`

	Metadata   *GenericMetadata `json:"metadata" gogm:"direction=outgoing;relationship=describedWith"`
	Signature  *Signature       `json:"signature" gogm:"direction=outgoing;relationship=signedWith"`
	Interfaces []*Interface     `json:"interfaces" gogm:"direction=outgoing;relationship=contains"`
}

type InterfaceInput struct {
	gogm.BaseNode `json:"-"`

	Parameters    *InputParameters     `json:"parameters" gogm:"name=parameters"`
	TypeInstances []*InputTypeInstance `json:"typeInstances" gogm:"name=typeInstances"`
}

type InterfaceOutput struct {
	gogm.BaseNode `json:"-"`

	TypeInstances []*OutputTypeInstance `json:"typeInstances" gogm:"name=typeInstances"`
}

type InterfaceReference struct {
	gogm.BaseNode `json:"-"`

	Path     string `json:"path" gogm:"name=path"`
	Revision string `json:"revision" gogm:"name=revision"`
}

type InterfaceRevision struct {
	gogm.BaseNode `json:"-"`

	Interface      *Interface `json:"-" gogm:"direction=incoming;relationship=revision"`

	Metadata *GenericMetadata `json:"metadata" gogm:"direction=outgoing;relationship=describedWith"`
	Revision string           `json:"revision" gogm:"name=revision"`
	//Spec     *InterfaceSpec   `json:"spec" gogm:"name=spec"`
	// List Implementations for a given Interface
	//Implementations []*Implementation `json:"implementations" gogm:"name=implementations"`
	//Signature       *Signature        `json:"signature" gogm:"name=signature"`
}

type InterfaceSpec struct {
	gogm.BaseNode `json:"-"`

	Input  *InterfaceInput  `json:"input" gogm:"name=input"`
	Output *InterfaceOutput `json:"output" gogm:"name=output"`
}

type Maintainer struct {
	gogm.BaseNode `json:"-"`

	GenericMetadata *GenericMetadata `json:"-" gogm:"direction=incoming;relationship=maintainedBy"`

	Name  string `json:"name" gogm:"name=name"`
	Email string  `json:"email" gogm:"name=email"`
	URL   string `json:"url" gogm:"name=url"`
}

type OutputTypeInstance struct {
	gogm.BaseNode `json:"-"`

	Name    string         `json:"name" gogm:"name=name"`
	TypeRef *TypeReference `json:"typeRef" gogm:"name=typeRef"`
}

func (OutputTypeInstance) IsTypeInstance() {}

type Signature struct {
	gogm.BaseNode `json:"-"`

	InterfaceGroup *InterfaceGroup `json:"-" gogm:"direction=incoming;relationship=signedWith"`

	Och string `json:"och" gogm:"name=och"`
}

type Tag struct {
	gogm.BaseNode `json:"-"`

	Name           string         `json:"name" gogm:"name=name"`
	Prefix         string         `json:"prefix" gogm:"name=prefix"`
	Path           string         `json:"path" gogm:"name=path"`
	LatestRevision *TagRevision   `json:"latestRevision" gogm:"name=latestRevision"`
	Revision       *TagRevision   `json:"revision" gogm:"name=revision"`
	Revisions      []*TagRevision `json:"revisions" gogm:"name=revisions"`
}

type TagRevision struct {
	Metadata  *GenericMetadata `json:"metadata" gogm:"name=metadata"`
	Revision  string           `json:"revision" gogm:"name=revision"`
	Spec      *TagSpec         `json:"spec" gogm:"name=spec"`
	Signature *Signature       `json:"signature" gogm:"name=signature"`
}

type TagSpec struct {
	AdditionalRefs []string `json:"additionalRefs" gogm:"name=additionalRefs"`
}

type Type struct {
	Name           string          `json:"name" gogm:"name=name"`
	Prefix         string          `json:"prefix" gogm:"name=prefix"`
	Path           string          `json:"path" gogm:"name=path"`
	LatestRevision *TypeRevision   `json:"latestRevision" gogm:"name=latestRevision"`
	Revision       *TypeRevision   `json:"revision" gogm:"name=revision"`
	Revisions      []*TypeRevision `json:"revisions" gogm:"name=revisions"`
}

type TypeInstanceRelationItem struct {
	TypeInstanceName string `json:"typeInstanceName" gogm:"name=typeInstanceName"`
	// Contains list of Type Instance names, which a given TypeInstance uses (depends on)
	Uses []string `json:"uses" gogm:"name=uses"`
}

type TypeMetadata struct {
	Name             string         `json:"name" gogm:"name=name"`
	Prefix           *string        `json:"prefix" gogm:"name=prefix"`
	Path             *string        `json:"path" gogm:"name=path"`
	DisplayName      *string        `json:"displayName" gogm:"name=displayName"`
	Description      string         `json:"description" gogm:"name=description"`
	Maintainers      []*Maintainer  `json:"maintainers" gogm:"name=maintainers"`
	DocumentationURL *string        `json:"documentationURL" gogm:"name=documentationURL"`
	SupportURL       *string        `json:"supportURL" gogm:"name=supportURL"`
	IconURL          *string        `json:"iconURL" gogm:"name=iconURL"`
	Tags             []*TagRevision `json:"tags" gogm:"name=tags"`
}

func (TypeMetadata) IsMetadataBaseFields() {}

type TypeReference struct {
	Path     string `json:"path" gogm:"name=path"`
	Revision string `json:"revision" gogm:"name=revision"`
}

type TypeRevision struct {
	Metadata  *TypeMetadata `json:"metadata" gogm:"name=metadata"`
	Revision  string        `json:"revision" gogm:"name=revision"`
	Spec      *TypeSpec     `json:"spec" gogm:"name=spec"`
	Signature *Signature    `json:"signature" gogm:"name=signature"`
}

type TypeSpec struct {
	AdditionalRefs []string    `json:"additionalRefs" gogm:"name=additionalRefs"`
	JSONSchema     interface{} `json:"jsonSchema" gogm:"name=jsonSchema"`
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

type SemVerTaggingStrategyTags string

const (
	SemVerTaggingStrategyTagsStable SemVerTaggingStrategyTags = "STABLE"
	SemVerTaggingStrategyTagsEdge   SemVerTaggingStrategyTags = "EDGE"
)

var AllSemVerTaggingStrategyTags = []SemVerTaggingStrategyTags{
	SemVerTaggingStrategyTagsStable,
	SemVerTaggingStrategyTagsEdge,
}

func (e SemVerTaggingStrategyTags) IsValid() bool {
	switch e {
	case SemVerTaggingStrategyTagsStable, SemVerTaggingStrategyTagsEdge:
		return true
	}
	return false
}

func (e SemVerTaggingStrategyTags) String() string {
	return string(e)
}

func (e *SemVerTaggingStrategyTags) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SemVerTaggingStrategyTags(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SemVerTaggingStrategyTags", str)
	}
	return nil
}

func (e SemVerTaggingStrategyTags) MarshalGQL(w io.Writer) {
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
