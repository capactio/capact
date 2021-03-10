package action

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	"projectvoltron.dev/voltron/internal/ptr"
	gqlengine "projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/client"
	"projectvoltron.dev/voltron/pkg/httputil"
	ochclient "projectvoltron.dev/voltron/pkg/och/client/public/generated"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

type CreateOptions struct {
	InterfaceName string
	DryRun        bool
}

func NewCreate() *cobra.Command {
	var opts CreateOptions

	cmd := &cobra.Command{
		Use:   "create INTERFACE",
		Short: "List OCH Interfaces",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.InterfaceName = args[0]
			return Create(cmd.Context(), opts, os.Stdout)
		},
	}
	flags := cmd.Flags()

	flags.BoolVarP(&opts.DryRun, "dry-run", "", false, "Specifies whether the Action performs server-side test without actually running the Action")

	return cmd
}

// TODO export to `internal/ocftool/action`
func Create(ctx context.Context, opts CreateOptions, w io.Writer) error {
	rand.Seed(time.Now().UnixNano())
	answers := struct {
		Name       string
		Parameters string `survey:"input-parameters"`
	}{}

	qs := []*survey.Question{
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "Please type Action name",

				// invalid value: "gallant_mahavira": a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.',
				// and must start and end with an alphanumeric character (e.g. 'example.com',
				// regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')
				Default: strings.Replace(namesgenerator.GetRandomName(0), "_", "-", 1),
			},
			Validate: survey.Required,
		},
	}
	if err := survey.Ask(qs, &answers); err != nil {
		return err
	}

	ochCli, err := getOCHClient(config.GetDefaultContext())
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

	actionCli, err := getActionClient(config.GetDefaultContext())
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

// TODO: move it from here
func getActionClient(server string) (*client.Client, error) {
	store := credstore.NewOCH()
	user, pass, err := store.Get(server)
	if err != nil {
		return nil, err
	}

	httpClient := httputil.NewClient(30*time.Second, false,
		httputil.WithBasicAuth(user, pass))

	return client.New(fmt.Sprintf("%s/graphql", server), httpClient), nil
}

// TODO: move it from here
func getOCHClient(server string) (*ochclient.Client, error) {
	store := credstore.NewOCH()
	user, pass, err := store.Get(server)
	if err != nil {
		return nil, err
	}

	httpClient := httputil.NewClient(30*time.Second, false,
		httputil.WithBasicAuth(user, pass))

	return ochclient.NewClient(httpClient, fmt.Sprintf("%s/graphql", server)), nil
}
