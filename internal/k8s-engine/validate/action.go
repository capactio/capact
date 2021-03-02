package validate

import (
	"encoding/json"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/workflow/validate"
	"github.com/pkg/errors"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type ActionValidator struct{}

func NewActionValidator() *ActionValidator {
	return &ActionValidator{}
}

func (v *ActionValidator) Validate(action *types.Action) error {
	workflow, err := getWorkflowFromAction(action)
	if err != nil {
		return errors.Wrap(err, "while getting workflow from Action")
	}

	_, err = validate.ValidateWorkflow(nil, nil, workflow, validate.ValidateOpts{
		Lint: true,
	})
	if err != nil {
		return errors.Wrap(err, "while linting workflow")
	}

	return nil
}

func getWorkflowFromAction(action *types.Action) (*wfv1.Workflow, error) {
	data, err := json.Marshal(action.Args["workflow"])
	if err != nil {
		return nil, err
	}

	workflow := &wfv1.Workflow{}
	err = json.Unmarshal(data, &workflow.Spec)
	if err != nil {
		return nil, err
	}

	return workflow, nil
}
