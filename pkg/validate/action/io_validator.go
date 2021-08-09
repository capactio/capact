package action

import (
	"context"
	"fmt"
	"strings"

	"capact.io/capact/internal/ctxutil"
	"capact.io/capact/pkg/sdk/renderer/argo"
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"

	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/validate"
	"github.com/hashicorp/go-multierror"
	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

// HubClient defines external Hub calls used by Validator.
type HubClient interface {
	ListTypeRefRevisionsJSONSchemas(ctx context.Context, filter gqlpublicapi.TypeFilter) ([]*gqlpublicapi.TypeRevision, error)
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error)
}

// InputOutputValidator provides functionality to load and validate Input and Output data for Interface
// and Implementation manifests.
//
// WARNING: This validator calls Public and Local Hubs so it needs to be configured with a proper server.
// If it will be run against different Hubs then for example, it won't be able to
// fetch TypeInstance with a given ID or will fetch a wrong one.
type InputOutputValidator struct {
	hubCli            HubClient
	ParametersSchemas validate.SchemaCollection
}

// NewValidator returns new InputOutputValidator instance.
func NewValidator(hubCli HubClient) *InputOutputValidator {
	return &InputOutputValidator{hubCli: hubCli}
}

// LoadIfaceInputParametersSchemas returns JSONSchemas for all input parameters defined on a given Interface.
// It resolves TypeRefs to a given JSONSchema by calling Hub.
func (c *InputOutputValidator) LoadIfaceInputParametersSchemas(ctx context.Context, iface *gqlpublicapi.InterfaceRevision) (validate.SchemaCollection, error) {
	if c.ifaceHasNoInput(iface) {
		return nil, nil
	}

	var (
		parametersSchemas = validate.SchemaCollection{}
		parametersTypeRef = validate.TypeRefCollection{}
	)

	// 1. Process input parameters
	for _, param := range iface.Spec.Input.Parameters {
		if param.JSONSchema != nil {
			str, ok := param.JSONSchema.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected JSONSchema type, expected %T, got %T", "", param.JSONSchema)
			}
			parametersSchemas[param.Name] = validate.Schema{
				Value:    str,
				Required: true, // Currently, input parameters on Interface are required.
			}
		}
		if param.TypeRef != nil {
			parametersTypeRef[param.Name] = validate.TypeRef{
				TypeRef:  types.TypeRef(*param.TypeRef),
				Required: true, // Currently, input parameters on Interface are required.
			}
		}
	}

	// 2. Resolve input parameters' TypeRefs into JSONSchemas
	resolvedSchemas, err := c.resolveTypeRefsToJSONSchemas(ctx, parametersTypeRef)
	if err != nil {
		return nil, err
	}

	// 3. Merge inlined JSONSchema and resolved  TypeRef into single collection
	allSchemas, err := validate.MergeSchemaCollection(parametersSchemas, resolvedSchemas)
	if err != nil {
		return nil, err
	}

	return allSchemas, nil
}

// LoadIfaceInputTypeInstanceRefs returns input TypeInstances' TypeRefs defined on a given Interface.
func (c *InputOutputValidator) LoadIfaceInputTypeInstanceRefs(_ context.Context, iface *gqlpublicapi.InterfaceRevision) (validate.TypeRefCollection, error) {
	if c.ifaceHasNoInput(iface) {
		return nil, nil
	}

	var typeInstancesTypeRefs = validate.TypeRefCollection{}
	for _, param := range iface.Spec.Input.TypeInstances {
		if param.TypeRef == nil {
			continue
		}
		typeInstancesTypeRefs[param.Name] = validate.TypeRef{
			TypeRef:  types.TypeRef(*param.TypeRef),
			Required: true, // Currently, input TypeInstances are required on Interface
		}
	}

	return typeInstancesTypeRefs, nil
}

// LoadImplInputParametersSchemas returns JSONSchemas for additional parameters defined on a given Implementation.
// It resolves TypeRefs to a given JSONSchema by calling Hub.
func (c *InputOutputValidator) LoadImplInputParametersSchemas(ctx context.Context, impl gqlpublicapi.ImplementationRevision) (validate.SchemaCollection, error) {
	if impl.Spec == nil ||
		impl.Spec.AdditionalInput == nil ||
		impl.Spec.AdditionalInput.Parameters == nil ||
		impl.Spec.AdditionalInput.Parameters.TypeRef == nil {
		return nil, nil
	}

	// Current simplification on Implementation manifest, that only one additional
	// input parameter can be specified.
	in := validate.TypeRefCollection{
		argo.AdditionalInputName: {
			TypeRef:  types.TypeRef(*impl.Spec.AdditionalInput.Parameters.TypeRef),
			Required: false, // additional parameters are not required on Implementation.
		},
	}
	return c.resolveTypeRefsToJSONSchemas(ctx, in)
}

// LoadImplInputTypeInstanceRefs returns input TypeInstances' TypeRefs defined on a given Implementation.
func (c *InputOutputValidator) LoadImplInputTypeInstanceRefs(_ context.Context, impl gqlpublicapi.ImplementationRevision) (validate.TypeRefCollection, error) {
	if impl.Spec == nil ||
		impl.Spec.AdditionalInput == nil ||
		impl.Spec.AdditionalInput.TypeInstances == nil {
		return nil, nil
	}

	var typeInstancesTypeRefs = validate.TypeRefCollection{}
	for _, param := range impl.Spec.AdditionalInput.TypeInstances {
		if param.TypeRef == nil {
			continue
		}
		typeInstancesTypeRefs[param.Name] = validate.TypeRef{
			TypeRef:  types.TypeRef(*param.TypeRef),
			Required: false, // input TypeInstances are not required on Implementation.
		}
	}

	return typeInstancesTypeRefs, nil
}

// ValidateParameters validates that a given input parameters are valid against JSONSchema defined in SchemaCollection.
func (c *InputOutputValidator) ValidateParameters(ctx context.Context, paramsSchemas validate.SchemaCollection, parameters map[string]string) (validate.ValidationResult, error) {
	resultBldr := validate.NewResultBuilder("Parameters")

	// 1. Check that all required parameters are specified
	for name, schema := range paramsSchemas {
		val, found := parameters[name]
		if schema.Required && (!found || strings.TrimSpace(val) == "") {
			resultBldr.ReportIssue(name, "required but missing input parameters")
		}
	}

	// 2. Validate input parameters against JSONSchema
	for paramName, paramData := range parameters {
		if ctxutil.ShouldExit(ctx) { // validation may cause additional resource usage, so stop if not needed
			return nil, ctx.Err()
		}

		// Ensure that it's in JSON format.
		// It's not a problem if it's already a JSON.
		paramDataJSON, err := yaml.YAMLToJSON([]byte(paramData))
		if err != nil {
			return nil, err
		}

		schema, found := paramsSchemas[paramName]
		if !found {
			resultBldr.ReportIssue(paramName, "JSONSchema was not found")
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

// ValidateTypeInstances validates that a given input TypeInstances has valid TypeRefs.
// It resolves input TypeInstances' TypeRefs by calling Hub.
func (c *InputOutputValidator) ValidateTypeInstances(ctx context.Context, allowedTypes validate.TypeRefCollection, gotTypeInstances []types.InputTypeInstanceRef) (validate.ValidationResult, error) {
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
	resultBldr := validate.NewResultBuilder("TypeInstances")

	// 2.1. Check if specified input TypeInstances were found in Hub
	gotTypes := validate.TypeRefCollection{}
	for _, input := range gotTypeInstances {
		ref, found := gotTypeInstancesTypeRefs[input.ID]
		if !found {
			resultBldr.ReportIssue(input.Name, "TypeInstance was not found in Hub")
			continue
		}
		gotTypes[input.Name] = validate.TypeRef{
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
			// TODO(advanced-rendering): make it optional or maybe policy opt is enough?
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

// HasRequiredProp returns true if at least one of schema in collection has `required` property at the root level.
func (c *InputOutputValidator) HasRequiredProp(schemas validate.SchemaCollection) (bool, error) {
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

func (c *InputOutputValidator) resolveTypeRefsToJSONSchemas(ctx context.Context, inTypeRefs validate.TypeRefCollection) (validate.SchemaCollection, error) {
	// 1. Fetch revisions for given TypeRefs
	var typeRefsPath []string
	for _, ref := range inTypeRefs {
		typeRefsPath = append(typeRefsPath, ref.Path)
	}
	// No TypeRefs that should be resolved, early return to do not call Hub
	if len(typeRefsPath) == 0 {
		return nil, nil
	}

	typeRefsPathFilter := fmt.Sprintf(`(%s)`, strings.Join(typeRefsPath, "|"))
	gotTypes, err := c.hubCli.ListTypeRefRevisionsJSONSchemas(ctx, gqlpublicapi.TypeFilter{
		PathPattern: ptr.String(typeRefsPathFilter),
	})
	if err != nil {
		return nil, errors.Wrap(err, "while fetching JSONSchemas for input TypeRefs")
	}

	indexedTypes := map[string]interface{}{}
	for _, rev := range gotTypes {
		if rev == nil || rev.Spec == nil {
			continue
		}
		key := fmt.Sprintf("%s:%s", rev.Metadata.Path, rev.Revision)
		indexedTypes[key] = rev.Spec.JSONSchema
	}

	var (
		merr    = &multierror.Error{}
		schemas = validate.SchemaCollection{}
	)
	for name, ref := range inTypeRefs {
		refKey := fmt.Sprintf("%s:%s", ref.Path, ref.Revision)
		schema, found := indexedTypes[refKey]
		if !found {
			// It means that manifest refers to not existing TypeRef.
			// From our perspective it's error - we should have invalid manifests in Hub.
			merr = multierror.Append(merr)
			continue
		}
		str, ok := schema.(string)
		if !ok {
			merr = multierror.Append(merr, fmt.Errorf("unexpected JSONSchema type for %s, expected %T, got %T", refKey, "", schema))
			continue
		}
		schemas[name] = validate.Schema{
			Value:    str,
			Required: ref.Required,
		}
	}

	if err := merr.ErrorOrNil(); err != nil {
		return nil, err
	}

	return schemas, nil
}

func (c *InputOutputValidator) ifaceHasNoInput(iface *gqlpublicapi.InterfaceRevision) bool {
	if iface == nil || iface.Spec == nil || iface.Spec.Input == nil {
		return true
	}
	return false
}
