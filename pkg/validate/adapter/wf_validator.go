package adapter

import (
	"context"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/validate"

	"github.com/pkg/errors"
)

type Validator interface {
	LoadIfaceInputParametersSchemas(context.Context, *gqlpublicapi.InterfaceRevision) (validate.SchemaCollection, error)
	LoadIfaceInputTypeInstanceRefs(context.Context, *gqlpublicapi.InterfaceRevision) (validate.TypeRefCollection, error)
	LoadImplInputParametersSchemas(context.Context, gqlpublicapi.ImplementationRevision) (validate.SchemaCollection, error)
	LoadImplInputTypeInstanceRefs(context.Context, gqlpublicapi.ImplementationRevision) (validate.TypeRefCollection, error)

	ValidateParameters(validate.SchemaCollection, map[string]string) (validate.ValidationResult, error)
	ValidateTypeInstances(validate.TypeRefCollection, []types.InputTypeInstanceRef) (validate.ValidationResult, error)
}

type Workflow struct {
	validator Validator
}

func NewForWorkflow(validator Validator) *Workflow {
	return &Workflow{validator: validator}
}

// 6. Validate given input
// - input "input-parameters"  from k8s secret
// - typeInstances from dedicatedRenderer.inputTypeInstances
// - impl specific parameters from policy - additionalTypeInstances
// - impl specific type instances from policy - additionalParameters
func (w *Workflow) Validate(ctx context.Context,
	iface *gqlpublicapi.InterfaceRevision, impl gqlpublicapi.ImplementationRevision,
	params map[string]string, instances []types.InputTypeInstanceRef,
	additionalParams map[string]string, additionalInstances []types.InputTypeInstanceRef,
) error {
	// 1. Interface
	ifaceTypes, err := w.validator.LoadIfaceInputTypeInstanceRefs(ctx, iface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input TypeInstance types")
	}
	ifaceSchemas, err := w.validator.LoadIfaceInputParametersSchemas(ctx, iface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input parameters JSONSchemas")
	}

	rs := validate.ValidationResultAggregator{}
	err = rs.Report(w.validator.ValidateTypeInstances(ifaceTypes, instances))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}

	err = rs.Report(w.validator.ValidateParameters(ifaceSchemas, params))
	if err != nil {
		return errors.Wrap(err, "while validating parameters")
	}

	// 2. Implementation (first level only)
	if len(additionalInstances) > 0 {
		implTypes, err := w.validator.LoadImplInputTypeInstanceRefs(ctx, impl)
		if err != nil {
			return errors.Wrap(err, "while loading input TypeRefs for Implementation additional TypeInstances")
		}

		err = rs.Report(w.validator.ValidateTypeInstances(implTypes, additionalInstances))
		if err != nil {
			return errors.Wrap(err, "while validating additional TypeInstances")
		}
	}

	if len(additionalParams) > 0 {
		implSchemas, err := w.validator.LoadImplInputParametersSchemas(ctx, impl)
		if err != nil {
			return err
		}

		err = rs.Report(w.validator.ValidateParameters(implSchemas, additionalParams))
		if err != nil {
			return errors.Wrap(err, "while validating additional parameters")
		}
	}

	if err := rs.ErrorOrNil(); err != nil {
		return err
	}
	return nil
}
