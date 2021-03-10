package action

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	"projectvoltron.dev/voltron/internal/ptr"
	gqlengine "projectvoltron.dev/voltron/pkg/engine/api/graphql"

	"github.com/AlecAivazis/survey/v2"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	"projectvoltron.dev/voltron/pkg/engine/client"
	"projectvoltron.dev/voltron/pkg/httputil"
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
	answers := struct {
		Name string
	}{}

	qs := []*survey.Question{
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "Please type Action name",
				Default: namesgenerator.GetRandomName(0),
			},
			Validate: survey.Required,
		},
	}

	if err := survey.Ask(qs, &answers); err != nil {
		return err
	}

	// TODO: we should use JSON schema and ask for a given input parameters
	//ochCli, err := getOCHClient(config.GetDefaultContext())
	//if err != nil {
	//	return err
	//}
	//latestRev, err := ochCli.InterfaceLatestRevision(ctx, opts.interfaceName)
	//if err != nil {
	//	return err
	//}

	//params := latestRev.Interface.LatestRevision.Spec.Input.Parameters

	actionCli, err := getActionClient(config.GetDefaultContext())
	if err != nil {
		return err
	}
	_ = actionCli

	_, err = actionCli.CreateAction(ctx, &gqlengine.ActionDetailsInput{
		Name:  answers.Name,
		Input: nil,
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
