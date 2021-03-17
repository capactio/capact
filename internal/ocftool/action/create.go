package action

import (
	"context"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ptr"
	gqlengine "projectvoltron.dev/voltron/pkg/engine/api/graphql"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/fatih/color"
	"sigs.k8s.io/yaml"
)

type CreateOptions struct {
	InterfacePath string
	ActionName    string `survey:"name"`
	Namespace     string
	DryRun        bool
}

type CreateOutput struct {
	Action    *gqlengine.Action
	Namespace string
}

func Create(ctx context.Context, opts CreateOptions, w io.Writer) (*CreateOutput, error) {
	rand.Seed(time.Now().UnixNano())

	// must be a DNS-1123 subdomain
	defaultActionName := strings.Replace(namesgenerator.GetRandomName(0), "_", "-", 1)
	qs := []*survey.Question{
		actionNameQuestion(defaultActionName),
	}
	if opts.Namespace == "" {
		qs = append(qs, namespaceQuestion())
	}

	if err := survey.Ask(qs, &opts); err != nil {
		return nil, err
	}

	input, err := askForInputParams()
	if err != nil {
		return nil, err
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
		Input: input,
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

func askForInputParams() (*gqlengine.ActionInputData, error) {
	gqlJSON, err := askForInputParameters()
	if err != nil {
		return nil, err
	}

	ti, err := askForInputTypeInstances()
	if err != nil {
		return nil, err
	}

	return &gqlengine.ActionInputData{
		Parameters:    gqlJSON,
		TypeInstances: ti,
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
