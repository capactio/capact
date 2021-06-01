package action

import (
	"fmt"

	gqlengine "capact.io/capact/pkg/engine/api/graphql"

	"io/ioutil"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"sigs.k8s.io/yaml"
)

const DefaultNamespace = "default"

type CreateOptions struct {
	InterfacePath string
	ActionName    string `survey:"name"`
	Namespace     string
	DryRun        bool
	Interactive   bool

	ParametersFilePath    string
	TypeInstancesFilePath string
	UserPolicyFilePath    string

	parameters    *gqlengine.JSON
	typeInstances []*gqlengine.InputTypeInstanceData
	policy        *gqlengine.PolicyInput
}

func (c *CreateOptions) SetDefaults() {
	if c.ActionName == "" {
		c.ActionName = generateDNSName()
	}

	if c.Namespace == "" {
		c.Namespace = DefaultNamespace
	}
}

func (c *CreateOptions) Resolve() error {
	if err := c.resolveFromFiles(); err != nil {
		return err
	}

	if c.Interactive {
		return c.resolveWithSurvey()
	}

	c.SetDefaults()
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

	if c.ParametersFilePath == "" {
		gqlJSON, err := askForInputParameters()
		if err != nil {
			return err
		}
		c.parameters = gqlJSON
	}

	if c.TypeInstancesFilePath == "" {
		ti, err := askForInputTypeInstances()
		if err != nil {
			return err
		}
		c.typeInstances = ti
	}

	if c.UserPolicyFilePath == "" {
		policy, err := askForUserPolicy(c.InterfacePath)
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

	if c.UserPolicyFilePath != "" {
		rawInput, err := ioutil.ReadFile(c.UserPolicyFilePath)
		if err != nil {
			return err
		}
		c.policy, err = toUserPolicy(rawInput)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CreateOptions) ActionInput() *gqlengine.ActionInputData {
	return &gqlengine.ActionInputData{
		Parameters:    c.parameters,
		TypeInstances: c.typeInstances,
		UserPolicy:    c.policy,
	}
}

// TODO: ask only if input-parameters are defined, add support for JSON Schema
func askForInputParameters() (*gqlengine.JSON, error) {
	provideInput := false
	askAboutTI := &survey.Confirm{Message: "Do you want to provide input parameters?", Default: false}
	if err := survey.AskOne(askAboutTI, &provideInput); err != nil {
		return nil, err
	}

	if !provideInput {
		return nil, nil
	}

	rawInput := ""
	prompt := &survey.Editor{Message: "Please type Action input parameters in YAML format"}
	if err := survey.AskOne(prompt, &rawInput, survey.WithValidator(isYAML)); err != nil {
		return nil, err
	}

	return toInputParameters([]byte(rawInput))
}

func askForInputTypeInstances() ([]*gqlengine.InputTypeInstanceData, error) {
	provideTI := false
	askAboutTI := &survey.Confirm{Message: "Do you want to provide input TypeInstances?", Default: false}
	if err := survey.AskOne(askAboutTI, &provideTI); err != nil {
		return nil, err
	}

	if !provideTI {
		return nil, nil
	}

	editor := ""
	prompt := &survey.Editor{
		Message: "Please type Action input TypeInstance in YAML format",
		Default: heredoc.Doc(`
						typeInstances:
						  - name: ""
						    id: ""`),
		AppendDefault: true,

		HideDefault: true,
	}
	if err := survey.AskOne(prompt, &editor, survey.WithValidator(isYAML)); err != nil {
		return nil, err
	}

	return toTypeInstance([]byte(editor))
}

func askForUserPolicy(ifacePath string) (*gqlengine.PolicyInput, error) {
	providePolicy := false
	askAboutPolicy := &survey.Confirm{Message: "Do you want to provide one-time user policy?", Default: false}
	if err := survey.AskOne(askAboutPolicy, &providePolicy); err != nil {
		return nil, err
	}

	if !providePolicy {
		return nil, nil
	}

	editor := ""
	prompt := &survey.Editor{
		Message: "Please type one-time user policy in YAML format",
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

	return toUserPolicy([]byte(editor))
}

func toTypeInstance(rawInput []byte) ([]*gqlengine.InputTypeInstanceData, error) {
	var resp struct {
		TypeInstances []*gqlengine.InputTypeInstanceData `json:"typeInstances"`
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

func toUserPolicy(rawInput []byte) (*gqlengine.PolicyInput, error) {
	policy := &gqlengine.PolicyInput{}

	if err := yaml.Unmarshal(rawInput, policy); err != nil {
		return nil, err
	}

	return policy, nil
}
