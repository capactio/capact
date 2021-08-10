package facade

import (
	"context"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/validate"

	"github.com/pkg/errors"
)

// Validator aggregates method used by workflow validator facade
type Validator interface {
	LoadIfaceInputParametersSchemas(context.Context, *gqlpublicapi.InterfaceRevision) (validate.SchemaCollection, error)
	LoadIfaceInputTypeInstanceRefs(context.Context, *gqlpublicapi.InterfaceRevision) (validate.TypeRefCollection, error)
	LoadImplInputParametersSchemas(context.Context, gqlpublicapi.ImplementationRevision) (validate.SchemaCollection, error)
	LoadImplInputTypeInstanceRefs(context.Context, gqlpublicapi.ImplementationRevision) (validate.TypeRefCollection, error)

	ValidateParameters(context.Context, validate.SchemaCollection, map[string]string) (validate.ValidationResult, error)
	ValidateTypeInstancesStrict(ctx context.Context, allowedTypes validate.TypeRefCollection, gotTypeInstances []types.InputTypeInstanceRef) (validate.ValidationResult, error)
}

// Workflow provides facade to simplify validator usage in render engine.
type Workflow struct {
	validator Validator
}

// NewForWorkflow returns a new Workflow instance.
func NewForWorkflow(validator Validator) *Workflow {
	return &Workflow{validator: validator}
}

// WorkflowValidateInput holds input data for Validate method.
type WorkflowValidateInput struct {
	Interface            *gqlpublicapi.InterfaceRevision
	Parameters           map[string]string
	TypeInstances        []types.InputTypeInstanceRef
	Implementation       gqlpublicapi.ImplementationRevision
	AdditionalParameters map[string]string
}

// Validate validates both the required and additional input parameters and TypeInstances.
//
func (w *Workflow) Validate(ctx context.Context, in WorkflowValidateInput) error {
	rs := validate.ValidationResultAggregator{}

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

	allAllowedTypes, err := validate.MergeTypeRefCollection(ifaceTypes, implTypes)
	if err != nil {
		return errors.Wrap(err, "while merging Interface and Implementation TypeInstances' TypeRefs")
	}
	err = rs.Report(w.validator.ValidateTypeInstancesStrict(ctx, allAllowedTypes, in.TypeInstances))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}

	return rs.ErrorOrNil()
}
