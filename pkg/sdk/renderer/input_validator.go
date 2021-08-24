package renderer

import (
	"context"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/pkg/errors"
)

// Validator aggregates method used by workflow validator
type Validator interface {
	LoadIfaceInputParametersSchemas(context.Context, *gqlpublicapi.InterfaceRevision) (validation.SchemaCollection, error)
	LoadIfaceInputTypeInstanceRefs(context.Context, *gqlpublicapi.InterfaceRevision) (validation.TypeRefCollection, error)
	LoadImplInputParametersSchemas(context.Context, gqlpublicapi.ImplementationRevision) (validation.SchemaCollection, error)
	LoadImplInputTypeInstanceRefs(context.Context, gqlpublicapi.ImplementationRevision) (validation.TypeRefCollection, error)

	ValidateParameters(context.Context, validation.SchemaCollection, types.ParametersCollection) (validation.Result, error)
	ValidateTypeInstancesStrict(ctx context.Context, allowedTypes validation.TypeRefCollection, gotTypeInstances []types.InputTypeInstanceRef) (validation.Result, error)
}

// InputValidator provides functionality to validate input data for rendered workflow.
type InputValidator struct {
	validator Validator
}

// NewInputValidator returns a new InputValidator instance.
func NewInputValidator(validator Validator) *InputValidator {
	return &InputValidator{validator: validator}
}

// ValidateInput holds input data for Validate method.
type ValidateInput struct {
	Interface            *gqlpublicapi.InterfaceRevision
	Parameters           types.ParametersCollection
	TypeInstances        []types.InputTypeInstanceRef
	Implementation       gqlpublicapi.ImplementationRevision
	AdditionalParameters types.ParametersCollection
}

// Validate validates both the required and additional input parameters and TypeInstances.
func (w *InputValidator) Validate(ctx context.Context, in ValidateInput) error {
	rs := validation.ResultAggregator{}

	// 1. Validate Interface input parameters
	ifaceSchemas, err := w.validator.LoadIfaceInputParametersSchemas(ctx, in.Interface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input parameters JSONSchemas")
	}

	err = rs.Report(w.validator.ValidateParameters(ctx, ifaceSchemas, in.Parameters))
	if err != nil {
		return errors.Wrap(err, "while validating parameters")
	}

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

	// 3. Validate Interface TypeInstances and additional TypeInstances from Implementation
	ifaceTypes, err := w.validator.LoadIfaceInputTypeInstanceRefs(ctx, in.Interface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input TypeInstance types")
	}
	implTypes, err := w.validator.LoadImplInputTypeInstanceRefs(ctx, in.Implementation)
	if err != nil {
		return errors.Wrap(err, "while loading additional input TypeInstances' TypeRefs")
	}

	allAllowedTypes, err := validation.MergeTypeRefCollection(ifaceTypes, implTypes)
	if err != nil {
		return errors.Wrap(err, "while merging Interface and Implementation TypeInstances' TypeRefs")
	}
	err = rs.Report(w.validator.ValidateTypeInstancesStrict(ctx, allAllowedTypes, in.TypeInstances))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}

	return rs.ErrorOrNil()
}
