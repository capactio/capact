package action

import (
	"capact.io/capact/pkg/engine/k8s/policy"
	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"
	"fmt"
)

func foo() {
	// 2. Validate Implementation additional parameters only if specified
	if len(in.AdditionalParameters) > 0 {
		implSchemas, err := w.validator.LoadImplInputParametersSchemas(ctx, in.Implementation)
		if err != nil {
			return errors.Wrap(err, "while loading additional input parameters")
		}

		err = rs.Report(w.validator.ValidateParameters(ctx, implSchemas, in.AdditionalParameters))
		if err != nil {
			return errors.Wrap(err, "while validating additional parameters")
		}
	}

	allAllowedTypes, err := validation.MergeTypeRefCollection(ifaceTypes, implTypes)
	if err != nil {
		return errors.Wrap(err, "while merging Interface and Implementation TypeInstances' TypeRefs")
	}

	// Validate impl TypeInstances
	implTypes, err := w.validator.LoadImplInputTypeInstanceRefs(ctx, in.Implementation)
	if err != nil {
		return errors.Wrap(err, "while loading additional input TypeInstances' TypeRefs")
	}
}


// TODO: Policy validator?

func (c *InputOutputValidator) IsTypeRefValidAndEqualToImplReq(typeRef *types.ManifestRef, reqItem *hubpublicgraphql.ImplementationRequirementItem) bool {
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

func (c *InputOutputValidator) ValidateAdditionalTypeInstances(additionalTIsInPolicy []policy.AdditionalTypeInstanceToInject, implRev hubpublicgraphql.ImplementationRevision) validation.Result {
	resultBldr := validation.NewResultBuilder("AdditionalTypeInstance")

	for _, typeInstance := range additionalTIsInPolicy {
		if exists := c.isAdditionalTypeInstanceDefinedInImpl(typeInstance, implRev); !exists {
			resultBldr.ReportIssue(typeInstance.Name, c.undefinedAdditionalTIError(typeInstance, implRev))
			undefinedAdditionalTIsErr = multierror.Append(undefinedAdditionalTIsErr, e.undefinedAdditionalTIError(typeInstance, implRev))
			continue
		}
	}



	return resultBldr.Result(), nil
}

// isAdditionalTypeInstanceDefinedInImpl tries to match TypeInstance name and its Type reference against Implementation's `.spec.additionalInput.typeInstances` items.
func (c *InputOutputValidator) isAdditionalTypeInstanceDefinedInImpl(typeInstance policy.AdditionalTypeInstanceToInject, implRev hubpublicgraphql.ImplementationRevision) bool {
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


func (c *InputOutputValidator) undefinedAdditionalTIError(typeInstance policy.AdditionalTypeInstanceToInject, implRev hubpublicgraphql.ImplementationRevision) string {
	implPath := ""
	if implRev.Metadata != nil {
		implPath = implRev.Metadata.Path
	}

	tiTypeRef := ""
	if typeInstance.TypeRef != nil {
		tiTypeRef = fmt.Sprintf("%s:%s", typeInstance.TypeRef.Path, typeInstance.TypeRef.Revision)
	}

	return fmt.Sprintf(`TypeInstance (Type reference: %q) was not found in Implementation %q`, tiTypeRef, implPath)
}
