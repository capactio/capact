package action

import (
	"context"
	"io"
	"math/rand"
	"strings"
	"time"

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

	input, err := askForInputParams(ctx, opts.InterfacePath)
	if err != nil {
		return nil, err
	}

	actionCli, err := client.NewCluster(config.GetDefaultContext())
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
	okCheck(w, "Action created successfully")

	return &CreateOutput{
		Action:    act,
		Namespace: opts.Namespace,
	}, nil
}

func askForInputParams(ctx context.Context, interfacePath string) (*gqlengine.ActionInputData, error) {
	ochCli, err := client.NewHub(config.GetDefaultContext())
	if err != nil {
		return nil, err
	}
	latestRev, err := ochCli.InterfaceLatestRevision(ctx, interfacePath)
	if err != nil {
		return nil, err
	}

	if len(latestRev.Interface.LatestRevision.Spec.Input.Parameters) == 0 {
		return nil, nil
	}

	editor := ""
	prompt := &survey.Editor{Message: "Please type Action input parameters in YAML format"}
	if err := survey.AskOne(prompt, &editor, survey.WithValidator(isYAML)); err != nil {
		return nil, err
	}
	converted, err := yaml.YAMLToJSON([]byte(editor))
	if err != nil {
		return nil, err
	}

	gqlJSON := gqlengine.JSON(converted)
	return &gqlengine.ActionInputData{
		Parameters: &gqlJSON,
	}, nil
}
