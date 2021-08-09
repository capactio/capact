package action

import (
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/validate"
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/valyala/fastjson"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
	"strings"
)

type HubClient interface {
	ListTypeRefRevisionsJSONSchemas(ctx context.Context, filter gqlpublicapi.TypeFilter) ([]*gqlpublicapi.Type, error)
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error)
}

// InputOutputValidator provides functionality to load and validate Input and Output data
// for Interface and Implementation manifests.
type InputOutputValidator struct {
	hubCli            HubClient
	ParametersSchemas SchemaCollection
}

func NewValidator(hubCli HubClient) *InputOutputValidator {
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
			return true, nil
		}
	}
	return false, nil
}

func (c *InputOutputValidator) LoadIfaceInputParametersSchemas(ctx context.Context, iface *gqlpublicapi.InterfaceRevision) (SchemaCollection, error) {
	var (
		parametersSchemas = SchemaCollection{}
		parametersTypeRef = TypeRefCollection{}
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
		}
	}

	// 2. Resolve input parameters TypeRefs JSONSchemas
	schemas, err := c.resolveTypeRefsToJSONSchemas(ctx, parametersTypeRef)
	if err != nil {
		return nil, err
	}

	// 3. Merge schemas
	for key, val := range schemas {
		parametersSchemas[key] = val
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

func (c *InputOutputValidator) ValidateTypeInstances(allowedTypes TypeRefCollection, gotTypeInstances []types.InputTypeInstanceRef) (validate.ValidationResult, error) {
	// 1. Resolve TypeRef for given Types
	var ids []string
	for _, input := range gotTypeInstances {
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
	for name, ref := range allowedTypes {
		_, found := gotTypes[name]
		if !found && ref.Required {
			resultBldr.ReportIssue(name, "input TypeInstance was not found but it's required")
		}
	}

	// 2.2. Check if given TypeRefs match expected ones
	for name, gotTypeRef := range gotTypes {
		expTypeRef, found := allowedTypes[name]
		if !found {
			// TODO: make it optional or maybe policy opt is enough?
			// (reason allow to skip, e.g. Interface doesn't specify impl TI
			resultBldr.ReportIssue(name, "TypeInstance was not found in manifest definition")
			continue
		}

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

func (*InputOutputValidator) typeRefToKey(ref TypeRef) string {
	return fmt.Sprintf("%s:%s", ref.Path, ref.Revision)
}

func (c *InputOutputValidator) LoadImplInputParametersSchemas(ctx context.Context, impl gqlpublicapi.ImplementationRevision) (SchemaCollection, error) {
	if impl.Spec == nil ||
		impl.Spec.AdditionalInput == nil ||
		impl.Spec.AdditionalInput.Parameters == nil ||
		impl.Spec.AdditionalInput.Parameters.TypeRef == nil {
		return nil, nil
	}

	// called `additional-parameters`
	in := TypeRefCollection{
		"additional-parameters": { // TODO: const
			TypeReference: *impl.Spec.AdditionalInput.Parameters.TypeRef,
			Required:      false, // Parameters on Implementation are not required.
		},
	}
	return c.resolveTypeRefsToJSONSchemas(ctx, in)
}

func (c *InputOutputValidator) resolveTypeRefsToJSONSchemas(ctx context.Context, inTypeRefs TypeRefCollection) (SchemaCollection, error) {
	var (
		typeRefsPath = []string{}
		schemas      = SchemaCollection{}
	)
	for _, ref := range inTypeRefs {
		typeRefsPath = append(typeRefsPath, ref.Path)
	}
	// No TypeRefs that should be resolved, early return to do not call Hub
	if len(typeRefsPath) == 0 {
		return schemas, nil
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
	for name, ref := range inTypeRefs {
		refKey := c.typeRefToKey(ref)
		schema, found := indexedTypes[refKey]
		if !found {
			// It means that manifest refers to not existing TypeRef
			// From our perspective it's error as this should happen that we have invalid manifests in Hub.
			merr = multierror.Append(merr)
			continue
		}
		str, ok := schema.(string)
		if !ok {
			merr = multierror.Append(merr, fmt.Errorf("unexpected JSONSchema type for %s, expected %T, got %T", refKey, "", schema))
			continue

		}
		schemas[name] = Schema{
			Value:    str,
			Required: ref.Required,
		}
	}

	if err := merr.ErrorOrNil(); err != nil {
		return nil, err
	}
	return schemas, nil
}

func (c *InputOutputValidator) LoadImplInputTypeInstanceRefs(_ context.Context, impl gqlpublicapi.ImplementationRevision) (TypeRefCollection, error) {
	if impl.Spec == nil ||
		impl.Spec.AdditionalInput == nil ||
		impl.Spec.AdditionalInput.TypeInstances == nil {
		return nil, nil
	}

	var typeInstancesTypeRefs = TypeRefCollection{}

	for _, param := range impl.Spec.AdditionalInput.TypeInstances {
		if param.TypeRef != nil {
			typeInstancesTypeRefs[param.Name] = TypeRef{
				TypeReference: *param.TypeRef,
				Required:      false, // The Implementaiton input TypeInstances are not required.
			}
		}
	}

	return typeInstancesTypeRefs, nil
}
