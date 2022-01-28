package interfaceio

import (
	"context"
	"fmt"

	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
	"github.com/xeipuuv/gojsonschema"
)

// HubClient defines external Hub calls used by Validator.
type HubClient interface {
	ListTypes(ctx context.Context, opts ...public.TypeOption) ([]*gqlpublicapi.Type, error)
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error)
}

// Validator provides functionality to load and validate Input and Output data for Interface manifest.
//
// WARNING: This validator calls Public and Local Hubs so it needs to be configured with a proper server.
// If it will be run against different Hubs then for example, it won't be able to
// fetch TypeInstance with a given ID or will fetch a wrong one.
type Validator struct {
	hubCli            HubClient
	ParametersSchemas validation.SchemaCollection
}

// NewValidator returns new Validator instance.
func NewValidator(hubCli HubClient) *Validator {
	return &Validator{hubCli: hubCli}
}

// LoadInputParametersSchemas returns JSONSchemas for all input parameters defined on a given Interface.
// It resolves TypeRefs to a given JSONSchema by calling Hub.
func (c *Validator) LoadInputParametersSchemas(ctx context.Context, iface *gqlpublicapi.InterfaceRevision) (validation.SchemaCollection, error) {
	if c.ifaceHasNoInput(iface) {
		return nil, nil
	}

	var (
		parametersSchemas = validation.SchemaCollection{}
		parametersTypeRef = validation.TypeRefCollection{}
	)

	// 1. Process input parameters
	for _, param := range iface.Spec.Input.Parameters {
		if param.JSONSchema != nil {
			str, ok := param.JSONSchema.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected JSONSchema type, expected %T, got %T", "", param.JSONSchema)
			}
			parametersSchemas[param.Name] = validation.Schema{
				Value:    str,
				Required: true, // Currently, input parameters on Interface are required.
			}
		}
		if param.TypeRef != nil {
			parametersTypeRef[param.Name] = validation.TypeRef{
				TypeRef:  types.TypeRef(*param.TypeRef),
				Required: true, // Currently, input parameters on Interface are required.
			}
		}
	}

	// 2. Resolve input parameters' TypeRefs into JSONSchemas
	resolvedSchemas, err := validation.ResolveTypeRefsToJSONSchemas(ctx, c.hubCli, parametersTypeRef)
	if err != nil {
		return nil, err
	}

	// 3. Merge inlined JSONSchema and resolved TypeRef into single collection
	allSchemas, err := validation.MergeSchemaCollection(parametersSchemas, resolvedSchemas)
	if err != nil {
		return nil, err
	}

	return allSchemas, nil
}

// LoadInputTypeInstanceRefs returns input TypeInstances' TypeRefs defined on a given Interface.
func (c *Validator) LoadInputTypeInstanceRefs(_ context.Context, iface *gqlpublicapi.InterfaceRevision) (validation.TypeRefCollection, error) {
	if c.ifaceHasNoInput(iface) {
		return nil, nil
	}

	var typeInstancesTypeRefs = validation.TypeRefCollection{}
	for _, param := range iface.Spec.Input.TypeInstances {
		if param.TypeRef == nil {
			continue
		}
		typeInstancesTypeRefs[param.Name] = validation.TypeRef{
			TypeRef:  types.TypeRef(*param.TypeRef),
			Required: true, // Currently, input TypeInstances are required on Interface
		}
	}

	return typeInstancesTypeRefs, nil
}

// ValidateParameters validates that a given input parameters are valid against JSONSchema defined in SchemaCollection.
func (c *Validator) ValidateParameters(ctx context.Context, paramsSchemas validation.SchemaCollection, parameters types.ParametersCollection) (validation.Result, error) {
	return validation.ValidateParameters(ctx, "Parameters", paramsSchemas, parameters)
}

// ValidateTypeInstances validates that a given input TypeInstances has valid TypeRefs.
//
// It resolves input TypeInstances' TypeRefs by calling Hub.
func (c *Validator) ValidateTypeInstances(ctx context.Context, allowedTypes validation.TypeRefCollection, gotTypeInstances []types.InputTypeInstanceRef) (validation.Result, error) {
	// 1. Resolve TypeRef for given TypeInstances
	var ids []string
	indexedInputTINames := map[string]struct{}{}
	for _, input := range gotTypeInstances {
		ids = append(ids, input.ID)
		indexedInputTINames[input.Name] = struct{}{}
	}

	gotTypeInstancesTypeRefs, err := c.hubCli.FindTypeInstancesTypeRef(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, "while resolving input TypeInstances' TypeRefs")
	}

	// 2. Validation
	resultBldr := validation.NewResultBuilder("TypeInstances")

	// 2.1. Check if specified input TypeInstances were found in Hub
	gotTypes := validation.TypeRefCollection{}
	for _, input := range gotTypeInstances {
		ref, found := gotTypeInstancesTypeRefs[input.ID]
		if !found {
			resultBldr.ReportIssue(input.Name, "TypeInstance was not found in Hub")
			continue
		}
		gotTypes[input.Name] = validation.TypeRef{
			TypeRef: types.TypeRef(ref),
		}
	}

	// 2.2. Check that all required TypeInstances are specified
	for name, ref := range allowedTypes {
		// Needs to check input typeInstance and not those found in Hub
		// As here we check whether we got this input at all.
		_, found := indexedInputTINames[name]
		if ref.Required && !found {
			resultBldr.ReportIssue(name, "required but missing TypeInstance of type %s:%s", ref.Path, ref.Revision)
		}
	}

	// 2.2. Check if given TypeInstances match expected TypeRefs
	for name, gotTypeRef := range gotTypes {
		expTypeRef, found := allowedTypes[name]
		if !found {
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

// HasRequiredProp returns true if at least one of schema in collection has `required` property at the root level.
func (c *Validator) HasRequiredProp(schemas validation.SchemaCollection) (bool, error) {
	// re-used for parsing multiple json strings.
	// This improves parsing speed by reducing the number
	// of memory allocations.
	var p fastjson.Parser

	for name, schema := range schemas {
		v, err := p.Parse(schema.Value)
		if err != nil { // It's taken from Hub it should be already a valid JSON
			return false, errors.Wrapf(err, "while parsing JSONSchema for %q", name)
		}
		requiredArr := v.GetArray(gojsonschema.KEY_REQUIRED)
		if len(requiredArr) > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (c *Validator) ifaceHasNoInput(iface *gqlpublicapi.InterfaceRevision) bool {
	if iface == nil || iface.Spec == nil || iface.Spec.Input == nil {
		return true
	}
	return false
}
