package policy

import (
	"context"
	"fmt"

	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/engine/k8s/policy/metadata"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"
)

// HubClient defines external Hub calls used by Validator.
type HubClient interface {
	ListTypes(ctx context.Context, opts ...public.TypeOption) ([]*gqlpublicapi.Type, error)
}

// Validator validates Policy metadata.
type Validator struct {
	hubCli HubClient
}

// NewValidator returns new Validator instance.
func NewValidator(hubCli HubClient) *Validator {
	return &Validator{hubCli: hubCli}
}

// LoadAdditionalInputParametersSchemas returns JSONSchemas for additional parameters defined on a given Implementation.
// It resolves TypeRefs to a given JSONSchema by calling Hub.
func (v *Validator) LoadAdditionalInputParametersSchemas(ctx context.Context, impl gqlpublicapi.ImplementationRevision) (validation.SchemaCollection, error) {
	if !v.hasImplAdditionalInputParams(impl) {
		return nil, nil
	}

	var paramsTypeRefs = validation.TypeRefCollection{}
	for _, param := range impl.Spec.AdditionalInput.Parameters {
		if param.TypeRef == nil {
			continue
		}
		paramsTypeRefs[param.Name] = validation.TypeRef{
			TypeRef:  types.TypeRef(*param.TypeRef),
			Required: false, // additional parameters are not required on Implementation.
		}
	}

	return validation.ResolveTypeRefsToJSONSchemas(ctx, v.hubCli, paramsTypeRefs)
}

// IsTypeRefInjectableAndEqualToImplReq returns boolean value if a given Type reference matches the one in Implementation requirement item and can be injected.
func (v *Validator) IsTypeRefInjectableAndEqualToImplReq(typeRef *types.ManifestRef, reqItem *gqlpublicapi.ImplementationRequirementItem) bool {
	// check requirement item valid
	if reqItem == nil || reqItem.TypeRef == nil || reqItem.Alias == nil {
		return false
	}

	// check RequiredTypeInstance to inject valid
	if typeRef == nil {
		return false
	}

	// check path
	if typeRef.Path != reqItem.TypeRef.Path {
		return false
	}

	// check revision (if provided)
	if typeRef.Revision != reqItem.TypeRef.Revision {
		return false
	}

	return true
}

// ValidateAdditionalTypeInstances validates additional input TypeInstances.
func (v *Validator) ValidateAdditionalTypeInstances(additionalTIsInPolicy []policy.AdditionalTypeInstanceToInject, implRev gqlpublicapi.ImplementationRevision) validation.Result {
	resultBldr := validation.NewResultBuilder("AdditionalTypeInstance")

	for _, typeInstance := range additionalTIsInPolicy {
		exists := v.isAdditionalTypeInstanceDefinedInImpl(typeInstance, implRev)
		if exists {
			continue
		}
		resultBldr.ReportIssue(typeInstance.Name, v.undefinedAdditionalTIError(typeInstance, implRev))
	}

	return resultBldr.Result()
}

// ValidateAdditionalInputParameters validates additional input parameters.
func (v *Validator) ValidateAdditionalInputParameters(ctx context.Context, paramsSchemas validation.SchemaCollection, parameters types.ParametersCollection) (validation.Result, error) {
	return validation.ValidateParameters(ctx, "AdditionalParameters", paramsSchemas, parameters)
}

// ValidateTypeInstancesMetadata validates that every TypeInstance has metadata resolved.
func (v *Validator) ValidateTypeInstancesMetadata(in policy.Policy) validation.Result {
	unresolvedTypeInstances := metadata.TypeInstanceIDsWithUnresolvedMetadataForPolicy(in)
	return v.validationResultForTIMetadata(unresolvedTypeInstances)
}

// ValidateTypeInstancesMetadataForRule validates whether the TypeInstance injection metadata are resolved.
func (v *Validator) ValidateTypeInstancesMetadataForRule(in policy.Rule) validation.Result {
	unresolvedTypeInstances := metadata.TypeInstanceIDsWithUnresolvedMetadataForRule(in)
	return v.validationResultForTIMetadata(unresolvedTypeInstances)
}

// AreTypeInstancesMetadataResolved returns whether every TypeInstance has metadata resolved.
func (v *Validator) AreTypeInstancesMetadataResolved(in policy.Policy) bool {
	unresolvedTypeInstances := metadata.TypeInstanceIDsWithUnresolvedMetadataForPolicy(in)
	return len(unresolvedTypeInstances) == 0
}

func (v *Validator) hasImplAdditionalInputParams(impl gqlpublicapi.ImplementationRevision) bool {
	if impl.Spec == nil || impl.Spec.AdditionalInput == nil || impl.Spec.AdditionalInput.Parameters == nil {
		return false
	}

	return true
}

func (v *Validator) validationResultForTIMetadata(tis []metadata.TypeInstanceMetadata) validation.Result {
	if len(tis) == 0 {
		return validation.Result{}
	}

	resultBldr := validation.NewResultBuilder("Metadata for")

	for _, ti := range tis {
		resultBldr.ReportIssue(string(ti.Kind), "missing Type reference for %s", ti.String(false))
	}

	return resultBldr.Result()
}

// isAdditionalTypeInstanceDefinedInImpl tries to match TypeInstance name and its Type reference against Implementation's `.spec.additionalInput.typeInstances` items.
func (v *Validator) isAdditionalTypeInstanceDefinedInImpl(typeInstance policy.AdditionalTypeInstanceToInject, implRev gqlpublicapi.ImplementationRevision) bool {
	if typeInstance.TypeRef == nil || implRev.Spec == nil || implRev.Spec.AdditionalInput == nil || len(implRev.Spec.AdditionalInput.TypeInstances) == 0 {
		return false
	}

	for _, additionalTi := range implRev.Spec.AdditionalInput.TypeInstances {
		if additionalTi == nil || additionalTi.TypeRef == nil {
			continue
		}

		if additionalTi.Name != typeInstance.Name ||
			additionalTi.TypeRef.Path != typeInstance.TypeRef.Path ||
			additionalTi.TypeRef.Revision != typeInstance.TypeRef.Revision {
			continue
		}

		return true
	}

	return false
}

func (v *Validator) undefinedAdditionalTIError(typeInstance policy.AdditionalTypeInstanceToInject, implRev gqlpublicapi.ImplementationRevision) string {
	implPath := ""
	if implRev.Metadata != nil {
		implPath = implRev.Metadata.Path
	}

	tiTypeRef := ""
	if typeInstance.TypeRef != nil {
		tiTypeRef = fmt.Sprintf("%s:%s", typeInstance.TypeRef.Path, typeInstance.TypeRef.Revision)
	}

	return fmt.Sprintf(`cannot find such definition with exact name and Type reference %q in Implementation %q`, tiTypeRef, implPath)
}
