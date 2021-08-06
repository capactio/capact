package action

import (
	"capact.io/capact/internal/ptr"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/validate"
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/valyala/fastjson"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
	"strings"
)

type hubCli interface {
	ListTypeRefRevisionsJSONSchemas(ctx context.Context, filter gqlpublicapi.TypeFilter) ([]*gqlpublicapi.Type, error)
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error)
}

// InputOutputValidator provides functionality to load and validate Input and Output data
// for Interface and Implementation manifests.
type InputOutputValidator struct {
	hubCli            hubCli
	ParametersSchemas SchemaCollection
}

func NewValidator(hubCli hubCli) *InputOutputValidator {
	return &InputOutputValidator{hubCli: hubCli}
}

type TypeRef struct {
	gqlpublicapi.TypeReference
	Required bool
}

type Schema struct {
	Value    string
	Required bool
}

type (
	SchemaCollection  map[string]Schema
	TypeRefCollection map[string]TypeRef
)

// HasRequiredProp returns true if at least one of schema
// in collection has `required` property at the root level.
func (c *InputOutputValidator) HasRequiredProp(schemas SchemaCollection) (bool, error) {
	foundRequiredProperty := false

	// re-used for parsing multiple json strings.
	// This improves parsing speed by reducing the number
	// of memory allocations.
	var p fastjson.Parser

	for _, schema := range schemas {
		v, err := p.Parse(schema.Value)
		if err != nil { // It's taken from Hub it should be already a valid JSON
			return false, err
		}
		requiredArr := v.GetArray(gojsonschema.KEY_REQUIRED)
		if len(requiredArr) > 0 {
			foundRequiredProperty = true
		}
	}
	return foundRequiredProperty, nil
}

func (c *InputOutputValidator) LoadIfaceInputParametersSchemas(ctx context.Context, iface *gqlpublicapi.InterfaceRevision) (SchemaCollection, error) {
	var (
		parametersSchemas        = SchemaCollection{}
		parametersTypeRef        = TypeRefCollection{}
		indexedRequestedTypeRefs = map[string]struct{}{}
	)

	// 1. Process input parameters
	for _, param := range iface.Spec.Input.Parameters {
		if param.JSONSchema != nil {
			str, ok := param.JSONSchema.(string)
			if !ok {
				return nil, fmt.Errorf("got unexpected JSONSchema type, expected %T, got %T", "", param.JSONSchema)
			}
			parametersSchemas[param.Name] = Schema{
				Value:    str,
				Required: true,
			}
		}
		if param.TypeRef != nil {
			parametersTypeRef[param.Name] = TypeRef{
				TypeReference: *param.TypeRef,
				Required:      true,
			}
			indexedRequestedTypeRefs[c.typeRefToKey(*param.TypeRef)] = struct{}{}
		}
	}

	// 1.2 Resolve input parameters TypeRefs JSONSchemas
	var typeRefsPath []string
	for ref := range indexedRequestedTypeRefs {
		typeRefsPath = append(typeRefsPath, c.keyToTypeRef(ref).Path)
	}
	// No TypeRefs that should be resolved, early return to do not call Hub
	if len(typeRefsPath) == 0 {
		return parametersSchemas, nil
	}

	typeRefsPathFilter := fmt.Sprintf(`(%s)`, strings.Join(typeRefsPath, "|"))
	gotTypes, err := c.hubCli.ListTypeRefRevisionsJSONSchemas(ctx, gqlpublicapi.TypeFilter{
		PathPattern: ptr.String(typeRefsPathFilter),
	})
	if err != nil {
		return nil, err
	}

	indexedTypes := map[string]interface{}{}
	for _, t := range gotTypes {
		for _, rev := range t.Revisions {
			key := fmt.Sprintf("%s:%s", t.Path, rev.Revision)
			indexedTypes[key] = rev.Spec.JSONSchema
		}
	}

	var merr *multierror.Error
	for name, ref := range parametersTypeRef {
		schema, found := indexedTypes[c.typeRefToKey(ref.TypeReference)]
		if !found {
			// It means that Interface manifest refers to not existing TypeRef
			// From our perspective it's error as this should happen that we have invalid manifests in Hub.
			merr = multierror.Append(merr, fmt.Errorf(c.typeRefToKey(ref.TypeReference)))
			continue
		}
		str, ok := schema.(string)
		if !ok {
			return nil, fmt.Errorf("got unexpected JSONSchema type, expected %T, got %T", "", schema)
		}
		parametersSchemas[name] = Schema{
			Value:    str,
			Required: true,
		}
	}

	if err := merr.ErrorOrNil(); err != nil {
		return nil, err
	}

	return parametersSchemas, nil
}

func (c *InputOutputValidator) LoadIfaceInputTypeInstanceRefs(_ context.Context, iface *gqlpublicapi.InterfaceRevision) (TypeRefCollection, error) {
	var typeInstancesTypeRefs = TypeRefCollection{}

	for _, param := range iface.Spec.Input.TypeInstances {
		if param.TypeRef != nil {
			typeInstancesTypeRefs[param.Name] = TypeRef{
				TypeReference: *param.TypeRef,
				Required:      true, // Currently, input TypeInstances are required on Interface and must be passed
			}
		}
	}

	return typeInstancesTypeRefs, nil
}

func (c *InputOutputValidator) ValidateParameter(paramsSchemas SchemaCollection, name, value string) (validate.ValidationResult, error) {
	in := map[string]string{
		name: value,
	}
	return c.ValidateParameters(paramsSchemas, in)
}

// ValidateParameters validate that a given input parameters are valid against JSONSchema defined in SchemaCollection.
func (c *InputOutputValidator) ValidateParameters(paramsSchemas SchemaCollection, parameters map[string]string) (validate.ValidationResult, error) {
	resultBldr := validate.NewResultBuilder("Parameters")

	// 2.2. Check that all required typeRef from collection are passed
	for name := range paramsSchemas {
		_, found := parameters[name]
		if !found {
			resultBldr.ReportIssue(name, "not found but it's required")
		}
	}

	for paramName, paramData := range parameters {
		paramDataJSON, err := yaml.YAMLToJSON([]byte(paramData))
		if err != nil {
			return nil, err
		}

		schema, found := paramsSchemas[paramName]
		if !found {
			resultBldr.ReportIssue(paramName, fmt.Sprintf("JSONSchema was not found"))
			continue
		}

		schemaLoader := gojsonschema.NewStringLoader(schema.Value)
		dataLoader := gojsonschema.NewBytesLoader(paramDataJSON)

		result, err := gojsonschema.Validate(schemaLoader, dataLoader)
		if err != nil {
			return nil, err
		}

		if !result.Valid() {
			for _, err := range result.Errors() {
				resultBldr.ReportIssue(paramName, err.String())
			}
		}
	}

	return resultBldr.Result(), nil
}

func (c *InputOutputValidator) ValidateTypeInstances(allowedTypes TypeRefCollection, gotTypeInstances []*gqlengine.InputTypeInstanceData) (validate.ValidationResult, error) {
	// 1. Resolve TypeRef for given Types
	var ids []string
	for _, input := range gotTypeInstances {
		if input == nil {
			continue
		}
		ids = append(ids, input.ID)
	}

	gotTypeInstancesTypeRefs, err := c.hubCli.FindTypeInstancesTypeRef(context.TODO(), ids)
	if err != nil {
		return nil, err
	}

	// 2. Validation
	resultBldr := validate.NewResultBuilder("TypeInstances")

	// 2.1. Check if specified input TypeInstance were found in Hub
	gotTypes := TypeRefCollection{}
	for _, input := range gotTypeInstances {
		ref, found := gotTypeInstancesTypeRefs[input.ID]
		if !found {
			resultBldr.ReportIssue(input.Name, "TypeInstance was not found in Hub")
			continue
		}
		gotTypes[input.Name] = TypeRef{
			TypeReference: gqlpublicapi.TypeReference(ref),
		}
	}

	// 2.2. Check that all required typeRef from collection are passed
	for name := range allowedTypes {
		_, found := gotTypes[name]
		if !found {
			resultBldr.ReportIssue(name, "input TypeInstance was not found but it's required")
		}
	}

	// 2.2. Check if given TypeRefs match expected ones
	for name, gotTypeRef := range gotTypes {
		expTypeRef, found := allowedTypes[name]
		if !found { // TODO
			// 2.2 (optional) Check if given TypeRefs are allowed
			// allow to skip, e.g. Interface doesn't specify impl TI
			resultBldr.ReportIssue(name, "TypeInstance was not found in manifest definition")
			continue
		}

		//if expTypeRef.Revision != gotTypeRef.Revision || expTypeRef.Path != gotTypeRef.Path {
		//	bldr.ReportIssue(fmt.Sprintf("Input %q TypeInstance must be of Type '%s:%s' but it's '%s:%s'",
		//		input.Name, expTypeRef.Path, expTypeRef.Revision, gotTypeRef.Path, gotTypeRef.Revision))
		//}
		if expTypeRef.Path != gotTypeRef.Path {
			resultBldr.ReportIssue(name, "must be of Type %q but it's %q",
				expTypeRef.Path, gotTypeRef.Path)
		}
		if expTypeRef.Revision != gotTypeRef.Revision {
			resultBldr.ReportIssue(name, "must be in Revision %q but it's %q",
				expTypeRef.Revision, gotTypeRef.Revision)
		}
	}

	return resultBldr.Result(), nil
}

func (*InputOutputValidator) typeRefToKey(ref gqlpublicapi.TypeReference) string {
	return fmt.Sprintf("%s:%s", ref.Path, ref.Revision)
}

func (*InputOutputValidator) keyToTypeRef(ref string) gqlpublicapi.TypeReference {
	out := strings.SplitN(ref, ":", 2)
	refType := gqlpublicapi.TypeReference{Path: out[0]}

	if len(out) == 2 {
		refType.Revision = out[1]
	}

	return refType
}
