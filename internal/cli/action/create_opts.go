package action

import (
	"bytes"
	"fmt"
	"io/ioutil"

	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer/argo"
	"capact.io/capact/pkg/validate"
	"capact.io/capact/pkg/validate/action"
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"sigs.k8s.io/yaml"
)

const defaultNamespace = "default"

// CreateOptions holds configuration for creating a given Action.
type CreateOptions struct {
	InterfacePath string
	ActionName    string `survey:"name"`
	Namespace     string
	DryRun        bool
	Interactive   bool

	ParametersFilePath    string
	TypeInstancesFilePath string
	ActionPolicyFilePath  string

	// internal fields
	parameters            *gqlengine.JSON
	typeInstances         []types.InputTypeInstanceRef
	policy                *gqlengine.PolicyInput
	isInputParamsRequired bool
	isInputTypesRequired  bool
	validator             *action.InputOutputValidator
	ifaceSchemas          validate.SchemaCollection
	ifaceTypes            validate.TypeRefCollection
}

// setDefaults defaults not provided options.
func (c *CreateOptions) setDefaults() {
	if c.ActionName == "" {
		c.ActionName = generateDNSName()
	}

	if c.Namespace == "" {
		c.Namespace = defaultNamespace
	}
}

// TODO:
// - try to remove isInputParamsRequired isInputTypesRequired
// - maybe introduce printers for ValidationResults
// - adapter for Survey

func (c *CreateOptions) preValidate() error {
	r := validate.ValidationResultAggregator{}
	if c.TypeInstancesFilePath == "" && c.isInputTypesRequired && !c.Interactive {
		bldr := validate.NewResultBuilder("TypeInstances")
		for tiName, tiType := range c.ifaceTypes {
			bldr.ReportIssue(tiName, "required but missing TypeInstance of type %s:%s", tiType.Path, tiType.Revision)
		}
		if err := r.Report(bldr.Result(), nil); err != nil {
			return err
		}
	}

	if c.ParametersFilePath == "" && c.isInputParamsRequired && !c.Interactive {
		bldr := validate.NewResultBuilder("Parameters")
		bldr.ReportIssue(argo.UserInputName, "Interface requires input parameters but none was specified")
		if err := r.Report(bldr.Result(), nil); err != nil {
			return err
		}
	}

	return r.ErrorOrNil()
}

func (c *CreateOptions) validate() error {
	r := validate.ValidationResultAggregator{}
	if c.parameters != nil {
		err := r.Report(c.validator.ValidateParameters(c.ifaceSchemas, argo.ToInputParams(*c.parameters)))
		if err != nil {
			return err
		}
	}

	if len(c.typeInstances) > 0 {
		err := r.Report(c.validator.ValidateTypeInstances(c.ifaceTypes, c.typeInstances))
		if err != nil {
			return err
		}
	}

	return r.ErrorOrNil()
}

// resolve resolves the CreateOptions properties with data from different sources.
// If possible starts interactive mode.
func (c *CreateOptions) resolve() error {
	if err := c.preValidate(); err != nil {
		return err
	}

	if err := c.resolveFromFiles(); err != nil {
		return err
	}

	if c.Interactive {
		return c.resolveWithSurvey()
	}

	c.setDefaults()

	return c.validate()
}

func (c *CreateOptions) resolveWithSurvey() error {
	var qs []*survey.Question
	if c.ActionName == "" {
		qs = append(qs, actionNameQuestion(generateDNSName()))
	}

	if c.Namespace == "" {
		qs = append(qs, namespaceQuestion())
	}

	if err := survey.Ask(qs, c); err != nil {
		return err
	}

	if c.ParametersFilePath == "" && c.isInputParamsRequired {
		gqlJSON, err := c.askForInputParameters()
		if err != nil {
			return err
		}
		c.parameters = gqlJSON
	}

	if c.TypeInstancesFilePath == "" {
		ti, err := c.askForInputTypeInstances()
		if err != nil {
			return err
		}
		c.typeInstances = ti
	}

	if c.ActionPolicyFilePath == "" {
		policy, err := askForActionPolicy(c.InterfacePath)
		if err != nil {
			return err
		}
		c.policy = policy
	}
	return nil
}

func (c *CreateOptions) resolveFromFiles() error {
	if c.ParametersFilePath != "" {
		rawInput, err := ioutil.ReadFile(c.ParametersFilePath)
		if err != nil {
			return err
		}

		c.parameters, err = toInputParameters(rawInput)
		if err != nil {
			return err
		}
	}

	// We need to allow to pass TypeInstance which are not
	// specified in a given Interface, as for now we don't support
	// advancedRendering so as a workaround we pass them directly
	// to the created Action.
	// TODO(advanced-rendering): validate if user can specify TypeInstances.
	if c.TypeInstancesFilePath != "" {
		rawInput, err := ioutil.ReadFile(c.TypeInstancesFilePath)
		if err != nil {
			return err
		}
		c.typeInstances, err = toTypeInstance(rawInput)
		if err != nil {
			return err
		}
	}

	if c.ActionPolicyFilePath != "" {
		rawInput, err := ioutil.ReadFile(c.ActionPolicyFilePath)
		if err != nil {
			return err
		}
		c.policy, err = toActionPolicy(rawInput)
		if err != nil {
			return err
		}
	}

	return nil
}

// ActionInput returns GraphQL Action input based on the given options.
func (c *CreateOptions) ActionInput() *gqlengine.ActionInputData {
	return &gqlengine.ActionInputData{
		Parameters:    c.parameters,
		TypeInstances: convertTypeInstancesRefsToGQL(c.typeInstances),
		ActionPolicy:  c.policy,
	}
}

func (c *CreateOptions) askForInputParameters() (*gqlengine.JSON, error) {
	rawInput := ""
	prompt := &survey.Editor{Message: "Please type Action input parameters in YAML format"}

	valid := survey.ComposeValidators(
		survey.Required,
		isYAML,
		areParamsValid(func(inputParams string) error {
			result, err := c.validator.ValidateParameters(c.ifaceSchemas, argo.ToInputParams(inputParams))
			if err != nil {
				return err
			}
			return result.ErrorOrNil()
		}),
	)

	if err := survey.AskOne(prompt, &rawInput, survey.WithValidator(valid)); err != nil {
		return nil, err
	}

	return toInputParameters([]byte(rawInput))
}

func (c *CreateOptions) askForInputTypeInstances() ([]types.InputTypeInstanceRef, error) {
	body, requiredTI := c.getTypeInstancesForEditor()

	// If input is not required, still ask user whether he wants to specify one.
	// REASON: we don't support advancedRendering so as a workaround we pass them directly
	// to the created Action.
	// TODO(advanced-rendering): remove me.
	// TODO: or action policy already solves this problem?
	if !requiredTI {
		askAboutTI := &survey.Confirm{Message: "Do you want to provide input TypeInstances?", Default: false}
		if err := survey.AskOne(askAboutTI, &requiredTI); err != nil {
			return nil, err
		}
	}

	if !requiredTI {
		return nil, nil
	}

	editor := ""
	prompt := &survey.Editor{
		Message:       "Please type Action input TypeInstance in YAML format",
		Default:       body,
		AppendDefault: true,

		HideDefault: true,
	}
	if err := survey.AskOne(prompt, &editor, survey.WithValidator(isYAML)); err != nil {
		return nil, err
	}

	return toTypeInstance([]byte(editor))
}

func (c *CreateOptions) getTypeInstancesForEditor() (body string, required bool) {
	if len(c.ifaceTypes) == 0 {
		return heredoc.Doc(`
						# Interface doesn't specify input TypeInstance.
						# You can pass Implementation specific TypeInstance like already existing database. 
						typeInstances:
						  - name: ""
						    id: ""`), false
	}

	out := bytes.Buffer{}
	out.WriteString("typeInstances:")
	for tiName, tiType := range c.ifaceTypes {
		out.WriteString("\n\n")
		out.WriteString(heredoc.Docf(`
						  # TypeInstance ID for %s:%s
						  - name: "%s"
						    id: "" `,
			tiType.Path, tiType.Revision, tiName))
	}
	return out.String(), true
}

func askForActionPolicy(ifacePath string) (*gqlengine.PolicyInput, error) {
	providePolicy := false
	askAboutPolicy := &survey.Confirm{Message: "Do you want to provide one-time Action policy?", Default: false}
	if err := survey.AskOne(askAboutPolicy, &providePolicy); err != nil {
		return nil, err
	}

	if !providePolicy {
		return nil, nil
	}

	editor := ""
	prompt := &survey.Editor{
		Message: "Please type one-time Action policy in YAML format",
		Default: heredoc.Doc(fmt.Sprintf(`
      rules:
        - interface:
            path: "%s"
          oneOf:
            - implementationConstraints:
                path: ""
    `, ifacePath)),
		AppendDefault: true,
		HideDefault:   true,
	}
	if err := survey.AskOne(prompt, &editor, survey.WithValidator(isYAML)); err != nil {
		return nil, err
	}

	return toActionPolicy([]byte(editor))
}

func toTypeInstance(rawInput []byte) ([]types.InputTypeInstanceRef, error) {
	var resp struct {
		TypeInstances []types.InputTypeInstanceRef `json:"typeInstances"`
	}

	if err := yaml.Unmarshal(rawInput, &resp); err != nil {
		return nil, err
	}

	return resp.TypeInstances, nil
}

func toInputParameters(rawInput []byte) (*gqlengine.JSON, error) {
	converted, err := yaml.YAMLToJSON(rawInput)
	if err != nil {
		return nil, err
	}

	gqlJSON := gqlengine.JSON(converted)
	return &gqlJSON, nil
}

func toActionPolicy(rawInput []byte) (*gqlengine.PolicyInput, error) {
	policy := &gqlengine.PolicyInput{}

	if err := yaml.UnmarshalStrict(rawInput, policy); err != nil {
		return nil, err
	}

	return policy, nil
}
