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
	ValidateTypeInstances(ctx context.Context, allowedTypes validation.TypeRefCollection, gotTypeInstances []types.InputTypeInstanceRef) (validation.Result, error)
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
}

// Validate validates required input parameters and TypeInstances against the Interface.
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

	// 2. Validate Interface TypeInstances
	ifaceTypes, err := w.validator.LoadIfaceInputTypeInstanceRefs(ctx, in.Interface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input TypeInstance types")
	}

	err = rs.Report(w.validator.ValidateTypeInstances(ctx, ifaceTypes, in.TypeInstances))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}

	return rs.ErrorOrNil()
}
