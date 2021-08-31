package renderer

import (
	"context"

	"capact.io/capact/pkg/hub/client"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"

	"github.com/pkg/errors"
)

// InterfaceIOValidator aggregates methods used for Interface validation
type InterfaceIOValidator interface {
	LoadInputParametersSchemas(context.Context, *gqlpublicapi.InterfaceRevision) (validation.SchemaCollection, error)
	LoadInputTypeInstanceRefs(context.Context, *gqlpublicapi.InterfaceRevision) (validation.TypeRefCollection, error)
	ValidateParameters(context.Context, validation.SchemaCollection, types.ParametersCollection) (validation.Result, error)
	ValidateTypeInstances(ctx context.Context, allowedTypes validation.TypeRefCollection, gotTypeInstances []types.InputTypeInstanceRef) (validation.Result, error)
}

// WorkflowInputValidator provides functionality to validate input data for rendered workflow.
type WorkflowInputValidator struct {
	interfaceValidator InterfaceIOValidator
	policyValidator    client.PolicyIOValidator
}

// NewWorkflowInputValidator returns a new WorkflowInputValidator instance.
func NewWorkflowInputValidator(interfaceValidator InterfaceIOValidator, policyValidator client.PolicyIOValidator) *WorkflowInputValidator {
	return &WorkflowInputValidator{
		interfaceValidator: interfaceValidator,
		policyValidator:    policyValidator,
	}
}

// InterfaceInput holds input data for Validate method.
type InterfaceInput struct {
	Interface     *gqlpublicapi.InterfaceRevision
	Parameters    types.ParametersCollection
	TypeInstances []types.InputTypeInstanceRef
}

// ValidateInterfaceInput validates required input parameters and TypeInstances against the Interface.
func (w *WorkflowInputValidator) ValidateInterfaceInput(ctx context.Context, in InterfaceInput) error {
	rs := validation.ResultAggregator{}

	// 1. Validate Interface input parameters
	ifaceSchemas, err := w.interfaceValidator.LoadInputParametersSchemas(ctx, in.Interface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input parameters JSONSchemas")
	}

	err = rs.Report(w.interfaceValidator.ValidateParameters(ctx, ifaceSchemas, in.Parameters))
	if err != nil {
		return errors.Wrap(err, "while validating parameters")
	}

	// 2. Validate Interface TypeInstances
	ifaceTypes, err := w.interfaceValidator.LoadInputTypeInstanceRefs(ctx, in.Interface)
	if err != nil {
		return errors.Wrap(err, "while loading Interface input TypeInstance types")
	}

	err = rs.Report(w.interfaceValidator.ValidateTypeInstances(ctx, ifaceTypes, in.TypeInstances))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}

	return rs.ErrorOrNil()
}

// PolicyValidator returns Policy-related validator for a given workflow.
func (w *WorkflowInputValidator) PolicyValidator() client.PolicyIOValidator {
	return w.policyValidator
}
