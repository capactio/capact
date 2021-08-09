package facade

import (
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/validate"
	"context"

	"github.com/pkg/errors"
)

// Validator aggregates method used by workflow validator facade
type Validator interface {
	LoadIfaceInputParametersSchemas(context.Context, *gqlpublicapi.InterfaceRevision) (validate.SchemaCollection, error)
	LoadIfaceInputTypeInstanceRefs(context.Context, *gqlpublicapi.InterfaceRevision) (validate.TypeRefCollection, error)
	LoadImplInputParametersSchemas(context.Context, gqlpublicapi.ImplementationRevision) (validate.SchemaCollection, error)
	LoadImplInputTypeInstanceRefs(context.Context, gqlpublicapi.ImplementationRevision) (validate.TypeRefCollection, error)

	ValidateParameters(context.Context, validate.SchemaCollection, map[string]string) (validate.ValidationResult, error)
	ValidateTypeInstances(context.Context, validate.TypeRefCollection, []types.InputTypeInstanceRef) (validate.ValidationResult, error)
}

// Workflow provides facade to simplify validator usage in render engine and unit-testing.
type Workflow struct {
	validator Validator
}

// NewForWorkflow returns a new Workflow instance.
func NewForWorkflow(validator Validator) *Workflow {
	return &Workflow{validator: validator}
}

type WorkflowValidateInput struct {
	Interface               *gqlpublicapi.InterfaceRevision
	Parameters              map[string]string
	TypeInstances           []types.InputTypeInstanceRef
	Implementation          gqlpublicapi.ImplementationRevision
	AdditionalParameters    map[string]string
	AdditionalTypeInstances []types.InputTypeInstanceRef
}

// Validate validates both the required and additional input parameters and TypeInstances.
func (w *Workflow) Validate(ctx context.Context, in WorkflowValidateInput) error {
	// 1. Interface
	ifaceTypes, err := w.validator.LoadIfaceInputTypeInstanceRefs(ctx, in.Interface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input TypeInstance types")
	}
	ifaceSchemas, err := w.validator.LoadIfaceInputParametersSchemas(ctx, in.Interface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input parameters JSONSchemas")
	}

	rs := validate.ValidationResultAggregator{}
	err = rs.Report(w.validator.ValidateTypeInstances(ctx, ifaceTypes, in.TypeInstances))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}

	err = rs.Report(w.validator.ValidateParameters(ctx, ifaceSchemas, in.Parameters))
	if err != nil {
		return errors.Wrap(err, "while validating parameters")
	}

	// 2. Implementation (first level only)
	if len(in.AdditionalTypeInstances) > 0 {
		implTypes, err := w.validator.LoadImplInputTypeInstanceRefs(ctx, in.Implementation)
		if err != nil {
			return errors.Wrap(err, "while loading additional input TypeInstances' TypeRefs")
		}

		err = rs.Report(w.validator.ValidateTypeInstances(ctx, implTypes, in.AdditionalTypeInstances))
		if err != nil {
			return errors.Wrap(err, "while validating additional TypeInstances")
		}
	}

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

	if err := rs.ErrorOrNil(); err != nil {
		return err
	}
	return nil
}
