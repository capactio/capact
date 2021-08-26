package policy

import (
	"fmt"

	"capact.io/capact/internal/multierror"

	"github.com/pkg/errors"
)

// ValidateTypeInstancesMetadata validates that every TypeInstance has metadata resolved.
func (in *Policy) ValidateTypeInstancesMetadata() error {
	unresolvedTypeInstances := in.typeInstanceIDsWithUnresolvedMetadata()
	return errorOrNil(unresolvedTypeInstances)
}

// ValidateTypeInstanceMetadata validates whether the TypeInstance injection metadata are resolved.
func (in *Rule) ValidateTypeInstanceMetadata() error {
	unresolvedTypeInstances := in.typeInstanceIDsWithUnresolvedMetadata()
	return errorOrNil(unresolvedTypeInstances)
}

// AreTypeInstancesMetadataResolved returns whether every TypeInstance has metadata resolved.
func (in *Policy) AreTypeInstancesMetadataResolved() bool {
	unresolvedTypeInstances := in.typeInstanceIDsWithUnresolvedMetadata()
	return len(unresolvedTypeInstances) == 0
}

type typeInstanceKind string

const (
	requiredTypeInstance   typeInstanceKind = "RequiredTypeInstance"
	additionalTypeInstance typeInstanceKind = "AdditionalTypeInstance"
)

type typeInstanceData struct {
	ID          string
	Name        *string
	Description *string
	Kind        typeInstanceKind
}

func (in *Policy) typeInstanceIDsWithUnresolvedMetadata() []typeInstanceData {
	var tis []typeInstanceData
	for _, rule := range in.Rules {
		for _, ruleItem := range rule.OneOf {
			tis = append(tis, ruleItem.typeInstanceIDsWithUnresolvedMetadata()...)
		}
	}

	return tis
}

func (in *Rule) typeInstanceIDsWithUnresolvedMetadata() []typeInstanceData {
	if in.Inject == nil {
		return nil
	}

	var tis []typeInstanceData

	// Required TypeInstances
	for _, ti := range in.Inject.RequiredTypeInstances {
		if ti.TypeRef != nil && ti.TypeRef.Path != "" && ti.TypeRef.Revision != "" {
			continue
		}

		tis = append(tis, typeInstanceData{
			ID:          ti.ID,
			Description: ti.Description,
			Kind:        requiredTypeInstance,
		})
	}

	// Additional TypeInstances
	for _, ti := range in.Inject.AdditionalTypeInstances {
		if ti.TypeRef != nil && ti.TypeRef.Path != "" && ti.TypeRef.Revision != "" {
			continue
		}

		tiName := ti.Name
		tis = append(tis, typeInstanceData{
			ID:   ti.ID,
			Name: &tiName,
			Kind: additionalTypeInstance,
		})
	}

	return tis
}

func errorOrNil(tis []typeInstanceData) error {
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
