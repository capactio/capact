package argoactions

import (
	"fmt"
	"io"
	"strconv"
)

type AttributeFilterInput struct {
	Path string      `json:"path"`
	Rule *FilterRule `json:"rule"`
	// If not provided, any revision of the Attribute applies to this filter
	Revision *string `json:"revision"`
}

type AttributeReference struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

type AttributeReferenceInput struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

type CreateTypeInstanceInput struct {
	// Used to define the relationships, between the created TypeInstances
	Alias      *string                    `json:"alias"`
	TypeRef    *TypeReferenceInput        `json:"typeRef"`
	Attributes []*AttributeReferenceInput `json:"attributes"`
	Value      interface{}                `json:"value"`
}

type CreateTypeInstanceOutput struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
}

type CreateTypeInstancesInput struct {
	TypeInstances []*CreateTypeInstanceInput       `json:"typeInstances"`
	UsesRelations []*TypeInstanceUsesRelationInput `json:"usesRelations"`
}

type TypeInstance struct {
	ResourceVersion int                   `json:"resourceVersion"`
	Metadata        *TypeInstanceMetadata `json:"metadata"`
	Spec            *TypeInstanceSpec     `json:"spec"`
	Uses            []*TypeInstance       `json:"uses"`
	UsedBy          []*TypeInstance       `json:"usedBy"`
}

type TypeInstanceFilter struct {
	Attributes []*AttributeFilterInput `json:"attributes"`
	TypeRef    *TypeRefFilterInput     `json:"typeRef"`
}

type TypeInstanceInstrumentation struct {
	Metrics *TypeInstanceInstrumentationMetrics `json:"metrics"`
	Health  *TypeInstanceInstrumentationHealth  `json:"health"`
}

type TypeInstanceInstrumentationHealth struct {
	URL    *string                                  `json:"url"`
	Method *HTTPRequestMethod                       `json:"method"`
	Status *TypeInstanceInstrumentationHealthStatus `json:"status"`
}

type TypeInstanceInstrumentationMetrics struct {
	Endpoint   *string                                        `json:"endpoint"`
	Regex      *string                                        `json:"regex"`
	Dashboards []*TypeInstanceInstrumentationMetricsDashboard `json:"dashboards"`
}

type TypeInstanceInstrumentationMetricsDashboard struct {
	URL string `json:"url"`
}

type TypeInstanceMetadata struct {
	ID         string                `json:"id"`
	Attributes []*AttributeReference `json:"attributes"`
}

type TypeInstanceSpec struct {
	TypeRef         *TypeReference               `json:"typeRef"`
	Value           interface{}                  `json:"value"`
	Instrumentation *TypeInstanceInstrumentation `json:"instrumentation"`
}

type TypeInstanceUsesRelationInput struct {
	// Can be existing TypeInstance ID or alias of a TypeInstance from typeInstances list
	From string `json:"from"`
	// Can be existing TypeInstance ID or alias of a TypeInstance from typeInstances list
	To string `json:"to"`
}

type TypeRefFilterInput struct {
	Path string `json:"path"`
	// If not provided, it returns TypeInstances for all revisions of given Type
	Revision *string `json:"revision"`
}

type TypeReference struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

type TypeReferenceInput struct {
	Path     string `json:"path"`
	Revision string `json:"revision"`
}

type UpdateTypeInstanceInput struct {
	TypeRef         *TypeReferenceInput        `json:"typeRef"`
	Attributes      []*AttributeReferenceInput `json:"attributes"`
	Value           interface{}                `json:"value"`
	ResourceVersion int                        `json:"resourceVersion"`
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

type HTTPRequestMethod string

const (
	HTTPRequestMethodGet  HTTPRequestMethod = "GET"
	HTTPRequestMethodPost HTTPRequestMethod = "POST"
)

var AllHTTPRequestMethod = []HTTPRequestMethod{
	HTTPRequestMethodGet,
	HTTPRequestMethodPost,
}

func (e HTTPRequestMethod) IsValid() bool {
	switch e {
	case HTTPRequestMethodGet, HTTPRequestMethodPost:
		return true
	}
	return false
}

func (e HTTPRequestMethod) String() string {
	return string(e)
}

func (e *HTTPRequestMethod) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = HTTPRequestMethod(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid HTTPRequestMethod", str)
	}
	return nil
}

func (e HTTPRequestMethod) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type TypeInstanceInstrumentationHealthStatus string

const (
	TypeInstanceInstrumentationHealthStatusUnknown TypeInstanceInstrumentationHealthStatus = "UNKNOWN"
	TypeInstanceInstrumentationHealthStatusReady   TypeInstanceInstrumentationHealthStatus = "READY"
	TypeInstanceInstrumentationHealthStatusFailing TypeInstanceInstrumentationHealthStatus = "FAILING"
)

var AllTypeInstanceInstrumentationHealthStatus = []TypeInstanceInstrumentationHealthStatus{
	TypeInstanceInstrumentationHealthStatusUnknown,
	TypeInstanceInstrumentationHealthStatusReady,
	TypeInstanceInstrumentationHealthStatusFailing,
}

func (e TypeInstanceInstrumentationHealthStatus) IsValid() bool {
	switch e {
	case TypeInstanceInstrumentationHealthStatusUnknown, TypeInstanceInstrumentationHealthStatusReady, TypeInstanceInstrumentationHealthStatusFailing:
		return true
	}
	return false
}

func (e TypeInstanceInstrumentationHealthStatus) String() string {
	return string(e)
}

func (e *TypeInstanceInstrumentationHealthStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TypeInstanceInstrumentationHealthStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TypeInstanceInstrumentationHealthStatus", str)
	}
	return nil
}

func (e TypeInstanceInstrumentationHealthStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
