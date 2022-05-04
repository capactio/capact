package action

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	gqlengine "capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer/argo"
	"capact.io/capact/pkg/sdk/validation"
	"capact.io/capact/pkg/sdk/validation/interfaceio"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/pkg/errors"
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
	parameters    json.RawMessage
	typeInstances []types.InputTypeInstanceRef
	policy        *gqlengine.PolicyInput

	// validation specific fields
	areInputParamsRequired        bool
	areInputTypeInstancesRequired bool
	validator                     *interfaceio.Validator
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

	parameters, err := argo.ToParametersCollection(c.parameters)
	if err != nil {
		return errors.Wrap(err, "while getting parameters collection")
	}

	err = r.Report(c.validator.ValidateParameters(ctx, c.ifaceSchemas, parameters))
	if err != nil {
		return errors.Wrap(err, "while validating parameters collection")
	}

	err = r.Report(c.validator.ValidateTypeInstances(ctx, c.ifaceTypes, c.typeInstances))
	if err != nil {
		return errors.Wrap(err, "while validating TypeInstances")
	}

	return r.ErrorOrNil()
}

// resolve resolves the CreateOptions properties with data from different sources.
// If possible starts interactive mode.
func (c *CreateOptions) resolve(ctx context.Context) error {
	if err := c.resolveFromFiles(); err != nil {
		return errors.Wrap(err, "while resolving properties")
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
		return errors.Wrap(err, "while asking for Action name and namespace")
	}

	if c.ParametersFilePath == "" && c.areInputParamsRequired {
		gqlJSON, err := c.askForInputParameters()
		if err != nil {
			return errors.Wrap(err, "while asking for input parameters")
		}
		c.parameters = gqlJSON
	}

	if c.TypeInstancesFilePath == "" && c.areInputTypeInstancesRequired {
		ti, err := c.askForInputTypeInstances()
		if err != nil {
			return errors.Wrap(err, "while asking for input TypeInstances")
		}
		c.typeInstances = ti
	}

	if c.ActionPolicyFilePath == "" {
		policy, err := askForActionPolicy(c.InterfacePath)
		if err != nil {
			return errors.Wrap(err, "while asking for Action policy")
		}
		c.policy = policy
	}
	return nil
}

func (c *CreateOptions) resolveFromFiles() error {
	if c.ParametersFilePath != "" {
		yamlInputParameters, err := ioutil.ReadFile(c.ParametersFilePath)
		if err != nil {
			return errors.Wrap(err, "while reading Action input parameters")
		}

		c.parameters, err = yaml.YAMLToJSON(yamlInputParameters)
		if err != nil {
			return errors.Wrap(err, "while converting YAML Action input parameters to JSON")
		}
	}

	if c.TypeInstancesFilePath != "" {
		rawInput, err := ioutil.ReadFile(c.TypeInstancesFilePath)
		if err != nil {
			return errors.Wrap(err, "while reading Action input TypeInstances file")
		}
		c.typeInstances, err = toTypeInstance(rawInput)
		if err != nil {
			return errors.Wrap(err, "while unmarshaling Action input TypeInstances file")
		}
	}

	if c.ActionPolicyFilePath != "" {
		rawInput, err := ioutil.ReadFile(c.ActionPolicyFilePath)
		if err != nil {
			return errors.Wrap(err, "while reading Action policy file")
		}
		c.policy, err = toActionPolicy(rawInput)
		if err != nil {
			return errors.Wrap(err, "while unmarshaling Action policy file")
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

func (c *CreateOptions) askForInputParameters() (json.RawMessage, error) {
	editor := ""
	prompt := &survey.Editor{
		Default:       c.getParametersForEditor(),
		Message:       "Please type Action input parameters in YAML format",
		AppendDefault: true,
		HideDefault:   true,
		FileName:      "*.yaml",
	}

	valid := []survey.Validator{
		survey.Required,
		isYAML,
	}

	if c.Validate {
		valid = append(valid, validatorAdapter(func(inputParams string) error {
			jsonInputParameters, err := yaml.YAMLToJSON([]byte(inputParams))
			if err != nil {
				return errors.Wrap(err, "while converting YAML to JSON")
			}

			parameters, err := argo.ToParametersCollection(jsonInputParameters)
			if err != nil {
				return errors.Wrap(err, "while getting parameters collection")
			}

			result, err := c.validator.ValidateParameters(context.Background(), c.ifaceSchemas, parameters)
			if err != nil {
				return errors.Wrap(err, "while validating parameters collection")
			}
			return result.ErrorOrNil()
		}))
	}

	if err := survey.AskOne(prompt, &editor, survey.WithValidator(survey.ComposeValidators(valid...))); err != nil {
		return nil, err
	}

	return yaml.YAMLToJSON([]byte(editor))
}

func (c *CreateOptions) getParametersForEditor() (body string) {
	out := bytes.Buffer{}
	for name := range c.ifaceSchemas {
		out.WriteString(heredoc.Docf(`
               %s:
                 # put data for %s here`, name, name))
		out.WriteString("\n")
	}
	return out.String()
}

func (c *CreateOptions) askForInputTypeInstances() ([]types.InputTypeInstanceRef, error) {
	editor := ""
	prompt := &survey.Editor{
		Message:       "Please type Action input TypeInstance in YAML format",
		Default:       c.getTypeInstancesForEditor(),
		AppendDefault: true,
		HideDefault:   true,
		FileName:      "*.yaml",
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

	if err := survey.AskOne(prompt, &editor, survey.WithValidator(survey.ComposeValidators(valid...))); err != nil {
		return nil, err
	}

	return toTypeInstance([]byte(editor))
}

func (c *CreateOptions) getTypeInstancesForEditor() (body string) {
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
	return out.String()
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
      interface:
        rules:
          - interface:
              path: "%s"
            oneOf:
              - implementationConstraints:
                  path: ""
    `, ifacePath)),
		AppendDefault: true,
		HideDefault:   true,
		FileName:      "*.yaml",
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

func toActionPolicy(rawInput []byte) (*gqlengine.PolicyInput, error) {
	policy := &gqlengine.PolicyInput{}

	if err := yaml.UnmarshalStrict(rawInput, policy); err != nil {
		return nil, err
	}

	return policy, nil
}
