package policy

import (
	"capact.io/capact/internal/multierror"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/engine/k8s/policy/metadata"
	hubpublicgraphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"
	"fmt"
	"github.com/pkg/errors"
)

// Validator validates Policy metadata.
type Validator struct {}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) IsTypeRefValidAndEqualToImplReq(typeRef *types.ManifestRef, reqItem *hubpublicgraphql.ImplementationRequirementItem) bool {
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

func (v *Validator) ValidateAdditionalTypeInstances(additionalTIsInPolicy []policy.AdditionalTypeInstanceToInject, implRev hubpublicgraphql.ImplementationRevision) validation.Result {
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

// ValidateTypeInstancesMetadata validates that every TypeInstance has metadata resolved.
func (v *Validator) ValidateTypeInstancesMetadata(in policy.Policy) error {
	unresolvedTypeInstances := metadata.TypeInstanceIDsWithUnresolvedMetadataForPolicy(in)
	return v.errorOrNil(unresolvedTypeInstances)
}

// ValidateTypeInstancesMetadataForRule validates whether the TypeInstance injection metadata are resolved.
func (v *Validator) ValidateTypeInstancesMetadataForRule(in policy.Rule) error {
	unresolvedTypeInstances := metadata.TypeInstanceIDsWithUnresolvedMetadataForRule(in)
	return v.errorOrNil(unresolvedTypeInstances)
}

// AreTypeInstancesMetadataResolved returns whether every TypeInstance has metadata resolved.
func (v *Validator) AreTypeInstancesMetadataResolved(in policy.Policy) bool {
	unresolvedTypeInstances := metadata.TypeInstanceIDsWithUnresolvedMetadataForPolicy(in)
	return len(unresolvedTypeInstances) == 0
}

func (v *Validator) errorOrNil(tis []metadata.TypeInstanceMetadata) error {
	if len(tis) == 0 {
		return nil
	}

	multiErr := multierror.New()
	for _, ti := range tis {
		tiDetails := ""
		if ti.Name != nil {
			tiDetails = fmt.Sprintf("name: %q", *ti.Name)
		} else if ti.Description != nil {
			tiDetails = fmt.Sprintf("description: %q", *ti.Description)
		}

		if tiDetails != "" {
			tiDetails = fmt.Sprintf(" (%s)", tiDetails)
		}

		multiErr = multierror.Append(
			multiErr,
			fmt.Errorf("missing Type reference for %s %q%s", ti.Kind, ti.ID, tiDetails),
		)
	}

	return errors.Wrap(multiErr, "while validating TypeInstance metadata for Policy")
}


// isAdditionalTypeInstanceDefinedInImpl tries to match TypeInstance name and its Type reference against Implementation's `.spec.additionalInput.typeInstances` items.
func (v *Validator) isAdditionalTypeInstanceDefinedInImpl(typeInstance policy.AdditionalTypeInstanceToInject, implRev hubpublicgraphql.ImplementationRevision) bool {
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

func (v *Validator) undefinedAdditionalTIError(typeInstance policy.AdditionalTypeInstanceToInject, implRev hubpublicgraphql.ImplementationRevision) string {
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
