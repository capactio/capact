package action

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer/argo"
	"capact.io/capact/pkg/sdk/validation"
	"capact.io/capact/pkg/sdk/validation/action"

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
	Validate      bool

	ParametersFilePath    string
	TypeInstancesFilePath string
	ActionPolicyFilePath  string

	// internal fields
	parameters    string
	typeInstances []types.InputTypeInstanceRef
	policy        *gqlengine.PolicyInput

	// validation specific fields
	areInputParamsRequired        bool
	areInputTypeInstancesRequired bool
	validator                     *action.InputOutputValidator
	ifaceSchemas                  validation.SchemaCollection
	ifaceTypes                    validation.TypeRefCollection
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

func (c *CreateOptions) validate(ctx context.Context) error {
	r := validation.ResultAggregator{}
	err := r.Report(c.validator.ValidateParameters(ctx, c.ifaceSchemas, argo.ToInputParams(c.parameters)))
	if err != nil {
		return err
	}

	err = r.Report(c.validator.ValidateTypeInstances(ctx, c.ifaceTypes, c.typeInstances))
	if err != nil {
		return err
	}

	return r.ErrorOrNil()
}

// resolve resolves the CreateOptions properties with data from different sources.
// If possible starts interactive mode.
func (c *CreateOptions) resolve(ctx context.Context) error {
	if err := c.resolveFromFiles(); err != nil {
		return err
	}

	if c.Interactive {
		return c.resolveWithSurvey()
	}

	c.setDefaults()

	if c.Validate {
		return c.validate(ctx)
	}

	return nil
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

	if c.ParametersFilePath == "" && c.areInputParamsRequired {
		gqlJSON, err := c.askForInputParameters()
		if err != nil {
			return err
		}
		c.parameters = gqlJSON
	}

	if c.TypeInstancesFilePath == "" && c.areInputTypeInstancesRequired {
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
		Parameters:    convertParametersToGQL(c.parameters),
		TypeInstances: convertTypeInstancesRefsToGQL(c.typeInstances),
		ActionPolicy:  c.policy,
	}
}

func (c *CreateOptions) askForInputParameters() (string, error) {
	rawInput := ""
	prompt := &survey.Editor{Message: "Please type Action input parameters in YAML format"}

	valid := []survey.Validator{
		survey.Required,
		isYAML,
	}

	if c.Validate {
		valid = append(valid, validatorAdapter(func(inputParams string) error {
			result, err := c.validator.ValidateParameters(context.Background(), c.ifaceSchemas, argo.ToInputParams(inputParams))
			if err != nil {
				return err
			}
			return result.ErrorOrNil()
		}))
	}

	if err := survey.AskOne(prompt, &rawInput, survey.WithValidator(survey.ComposeValidators(valid...))); err != nil {
		return "", err
	}

	return toInputParameters([]byte(rawInput))
}

func (c *CreateOptions) askForInputTypeInstances() ([]types.InputTypeInstanceRef, error) {
	body, requiredTI := c.getTypeInstancesForEditor()

	// TODO(https://github.com/capactio/capact/issues/438): If input TypeInstances are not required,
	// still ask user whether he wants to specify one.
	// We need to allow to pass additionalTypeInstances
	// which are not specified in a given Interface, e.g. existing database.
	if !requiredTI {
		askAboutTI := &survey.Confirm{Message: "Do you want to provide input TypeInstances?", Default: false}
		if err := survey.AskOne(askAboutTI, &requiredTI); err != nil {
			return nil, err
		}
	}

	if !requiredTI {
		return nil, nil
	}

	valid := []survey.Validator{
		survey.Required,
		isYAML,
	}

	if c.Validate {
		valid = append(valid, validatorAdapter(func(inputParams string) error {
			inputTI, err := toTypeInstance([]byte(inputParams))
			if err != nil {
				return err
			}
			result, err := c.validator.ValidateTypeInstances(context.Background(), c.ifaceTypes, inputTI)
			if err != nil {
				return err
			}
			return result.ErrorOrNil()
		}))
	}

	editor := ""
	prompt := &survey.Editor{
		Message:       "Please type Action input TypeInstance in YAML format",
		Default:       body,
		AppendDefault: true,

		HideDefault: true,
	}
	if err := survey.AskOne(prompt, &editor, survey.WithValidator(survey.ComposeValidators(valid...))); err != nil {
		return nil, err
	}

	return toTypeInstance([]byte(editor))
}

func (c *CreateOptions) getTypeInstancesForEditor() (body string, required bool) {
	if len(c.ifaceTypes) == 0 {
		return heredoc.Doc(`
               # Interface doesn't specify input TypeInstance.

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

func toInputParameters(rawInput []byte) (string, error) {
	converted, err := yaml.YAMLToJSON(rawInput)
	if err != nil {
		return "", err
	}

	return string(converted), nil
}

func toActionPolicy(rawInput []byte) (*gqlengine.PolicyInput, error) {
	policy := &gqlengine.PolicyInput{}

	if err := yaml.UnmarshalStrict(rawInput, policy); err != nil {
		return nil, err
	}

	return policy, nil
}
