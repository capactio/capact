package action

import (
	"context"
	"io"
	"math/rand"
	"strings"
	"time"

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
	InterfaceName string
	DryRun        bool
}

func Create(ctx context.Context, opts CreateOptions, w io.Writer) error {
	rand.Seed(time.Now().UnixNano())
	answers := struct {
		Name string
	}{}

	qs := []*survey.Question{
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "Please type Action name",
				// must be a DNS-1123 subdomain
				Default: strings.Replace(namesgenerator.GetRandomName(0), "_", "-", 1),
			},
			Validate: survey.ComposeValidators(survey.Required, isDNSSubdomain),
		},
	}
	if err := survey.Ask(qs, &answers); err != nil {
		return err
	}

	ochCli, err := client.NewHub(config.GetDefaultContext())
	if err != nil {
		return err
	}
	latestRev, err := ochCli.InterfaceLatestRevision(ctx, opts.InterfaceName)
	if err != nil {
		return err
	}

	//TODO: we should use JSON schema and ask for a given input parameters
	var input *gqlengine.ActionInputData
	if len(latestRev.Interface.LatestRevision.Spec.Input.Parameters) > 0 {
		params := ""
		prompt := &survey.Editor{Message: "Please type Action input parameters in YAML format"}
		if err := survey.AskOne(prompt, &params); err != nil {
			return err
		}
		converted, _ := yaml.YAMLToJSON([]byte(params))
		p := gqlengine.JSON(converted)
		input = &gqlengine.ActionInputData{
			Parameters: &p,
		}
	}

	actionCli, err := client.NewCluster(config.GetDefaultContext())
	if err != nil {
		return err
	}

	_, err = actionCli.CreateAction(ctx, &gqlengine.ActionDetailsInput{
		Name:  answers.Name,
		Input: input,
		ActionRef: &gqlengine.ManifestReferenceInput{
			Path: opts.InterfaceName,
		},
		DryRun: ptr.Bool(opts.DryRun),
	})

	if err != nil {
		return err
	}

	okCheck := color.New(color.FgGreen).FprintlnFunc()
	okCheck(w, "Action created successfully")

	return nil
}
