package action

import (
	"context"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	"capact.io/capact/internal/ocftool/client"
	"capact.io/capact/internal/ocftool/config"
	"capact.io/capact/internal/ptr"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

type CreateOptions struct {
	InterfacePath      string
	ActionName         string `survey:"name"`
	Namespace          string
	DryRun             bool
	Interactive        bool
	ParametersFilePath string
}

func (c *CreateOptions) SetDefaults() {
	rand.Seed(time.Now().UnixNano())

	// must be a DNS-1123 subdomain
	if c.ActionName == "" {
		c.ActionName = strings.Replace(namesgenerator.GetRandomName(0), "_", "-", 1)
	}
}

func (c *CreateOptions) Validate() error {
	if c.Interactive {
		if c.Namespace == "" {
			return errors.New("must provide namespace when not running interactively")
		}
		if c.ActionName == "" {
			return errors.New("must provide Action name when not running interactively")
		}
	}
	return nil
}

type CreateOutput struct {
	Action    *gqlengine.Action
	Namespace string
}

func Create(ctx context.Context, opts CreateOptions, w io.Writer) (*CreateOutput, error) {
	opts.SetDefaults()

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	inputData := &gqlengine.ActionInputData{}

	if opts.Interactive {
		qs := []*survey.Question{
			actionNameQuestion(opts.ActionName),
		}

		if opts.Namespace == "" {
			qs = append(qs, namespaceQuestion())
		}

		if err := survey.Ask(qs, &opts); err != nil {
			return nil, err
		}

		if opts.ParametersFilePath == "" {
			gqlJSON, err := askForInputParameters()
			if err != nil {
				return nil, err
			}
			inputData.Parameters = gqlJSON
		}

		ti, err := askForInputTypeInstances()
		if err != nil {
			return nil, err
		}
		inputData.TypeInstances = ti
	}

	if opts.ParametersFilePath != "" {
		rawInput, err := ioutil.ReadFile(opts.ParametersFilePath)
		if err != nil {
			return nil, err
		}
		converted, err := yaml.YAMLToJSON(rawInput)
		if err != nil {
			return nil, err
		}

		gqlJSON := gqlengine.JSON(converted)
		inputData.Parameters = &gqlJSON
	}

	server, err := config.GetDefaultContext()
	if err != nil {
		return nil, err
	}

	actionCli, err := client.NewCluster(server)
	if err != nil {
		return nil, err
	}

	ctxWithNs := namespace.NewContext(ctx, opts.Namespace)
	act, err := actionCli.CreateAction(ctxWithNs, &gqlengine.ActionDetailsInput{
		Name:  opts.ActionName,
		Input: inputData,
		ActionRef: &gqlengine.ManifestReferenceInput{
			Path: opts.InterfacePath,
		},
		DryRun: ptr.Bool(opts.DryRun),
	})
	if err != nil {
		return nil, err
	}

	okCheck := color.New(color.FgGreen).FprintlnFunc()
	okCheck(w, "Action created successfully\n")

	return &CreateOutput{
		Action:    act,
		Namespace: opts.Namespace,
	}, nil
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

	converted, err := yaml.YAMLToJSON([]byte(rawInput))
	if err != nil {
		return nil, err
	}

	gqlJSON := gqlengine.JSON(converted)
	return &gqlJSON, nil
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

	var resp struct {
		TypeInstances []*gqlengine.InputTypeInstanceData `json:"typeInstances"`
	}

	if err := yaml.Unmarshal([]byte(editor), &resp); err != nil {
		return nil, err
	}

	return resp.TypeInstances, nil
}
