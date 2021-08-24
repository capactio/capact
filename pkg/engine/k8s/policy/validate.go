package policy

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// ValidateTypeInstancesMetadata validates that every TypeInstance has metadata resolved.
func (in *Policy) ValidateTypeInstancesMetadata() error {
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterTypeInstancesWithEmptyTypeRef)
	return validateTypeInstancesMetadata(unresolvedTypeInstances)
}

func validateTypeInstancesMetadata(requiredTypeInstances []RequiredTypeInstanceToInject) error {
	if len(requiredTypeInstances) == 0 {
		return nil
	}

	multiErr := &multierror.Error{}
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
