package policy

import (
	"fmt"

	"capact.io/capact/internal/multierror"

	"github.com/pkg/errors"
)

// ValidateTypeInstancesMetadata validates that every TypeInstance has metadata resolved.
func (in *Policy) ValidateTypeInstancesMetadata() error {
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterReqTypeInstancesWithEmptyTypeRef)
	return validateTypeInstancesMetadata(unresolvedTypeInstances)
}

// ValidateTypeInstanceMetadata validates whether the TypeInstance injection metadata are resolved.
func (in *Rule) ValidateTypeInstanceMetadata() error {
	//var unresolvedTypeInstances []/
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterReqTypeInstancesWithEmptyTypeRef)
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterReqTypeInstancesWithEmptyTypeRef)

	return validateTypeInstancesMetadata(unresolvedTypeInstances)
}




func (in *Rule) filterRequiredTypeInstances(filterFn func(ti RequiredTypeInstanceToInject) bool) []RequiredTypeInstanceToInject {
	if in.Inject == nil {
		return nil
	}

	var typeInstances []RequiredTypeInstanceToInject
	for _, tiToInject := range in.Inject.RequiredTypeInstances {
		if !filterFn(tiToInject) {
			continue
		}

		typeInstances = append(typeInstances, tiToInject)
	}

	return typeInstances
}

func (in *Rule) filterAdditionalTypeInstances(filterFn func(ti AdditionalTypeInstanceToInject) bool) []AdditionalTypeInstanceToInject {
	if in.Inject == nil {
		return nil
	}

	var typeInstances []AdditionalTypeInstanceToInject
	for _, tiToInject := range in.Inject.AdditionalTypeInstances {
		if !filterFn(tiToInject) {
			continue
		}

		typeInstances = append(typeInstances, tiToInject)
	}

	return typeInstances
}



func (in *Policy) filterRequiredTypeInstances(filterFn func(ti RequiredTypeInstanceToInject) bool) []RequiredTypeInstanceToInject {
	var typeInstances []RequiredTypeInstanceToInject
	for _, rule := range in.Rules {
		for _, ruleItem := range rule.OneOf {
			typeInstances = append(typeInstances, ruleItem.filterRequiredTypeInstances(filterFn)...)
		}
	}

	return typeInstances
}

func (in *Policy) filterAdditionalTypeInstances(filterFn func(ti AdditionalTypeInstanceToInject) bool) []AdditionalTypeInstanceToInject {
	var typeInstances []RequiredTypeInstanceToInject
	for _, rule := range in.Rules {
		for _, ruleItem := range rule.OneOf {
			typeInstances = append(typeInstances, ruleItem.filterRequiredTypeInstances(filterFn)...)
		}
	}

	return typeInstances
}

var filterReqTypeInstancesWithEmptyTypeRef = func(ti RequiredTypeInstanceToInject) bool {
	return ti.TypeRef == nil || ti.TypeRef.Path == "" || ti.TypeRef.Revision == ""
}

func validateTypeInstancesMetadata(requiredTypeInstances []RequiredTypeInstanceToInject) error {
	if len(requiredTypeInstances) == 0 {
		return nil
	}

	multiErr := multierror.New()
	for _, ti := range requiredTypeInstances {
		tiDesc := ""
		if ti.Description != nil {
			tiDesc = *ti.Description
		}

		multiErr = multierror.Append(
			multiErr,
			fmt.Errorf("missing Type reference for TypeInstance %q (description: %q)", ti.ID, tiDesc),
		)
	}

	return errors.Wrap(multiErr, "while validating TypeInstance metadata for Policy")
}
