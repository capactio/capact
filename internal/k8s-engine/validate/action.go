package validate

import (
	"encoding/json"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo/v2/pkg/client/clientset/versioned"
	"github.com/argoproj/argo/v2/workflow/templateresolution"
	"github.com/argoproj/argo/v2/workflow/validate"
	"github.com/pkg/errors"
)

// ActionValidator provides functionality to statically validate Action definition.
type ActionValidator struct {
	wfCli *wfclientset.Clientset
}

// NewActionValidator returns a new ActionValidator instance.
func NewActionValidator(wfCli *wfclientset.Clientset) *ActionValidator {
	return &ActionValidator{
		wfCli: wfCli,
	}
}

// Validate validates a given Action definition.
func (v *ActionValidator) Validate(action *types.Action, namespace string) error {
	if action == nil {
		return nil
	}

	workflow, err := getWorkflowFromAction(action)
	if err != nil {
		return errors.Wrap(err, "while getting workflow from Action")
	}

	wfTmplGetter := templateresolution.WrapWorkflowTemplateInterface(v.wfCli.ArgoprojV1alpha1().WorkflowTemplates(namespace))
	cwfTmplGetter := templateresolution.WrapClusterWorkflowTemplateInterface(v.wfCli.ArgoprojV1alpha1().ClusterWorkflowTemplates())

	_, err = validate.ValidateWorkflow(wfTmplGetter, cwfTmplGetter, workflow, validate.ValidateOpts{
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
