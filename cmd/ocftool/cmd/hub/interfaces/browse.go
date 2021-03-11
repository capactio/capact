package interfaces

import (
	"context"
	"fmt"
	"io"
	"os"

	"projectvoltron.dev/voltron/internal/ocftool/action"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type browseOptions struct {
	pathPattern string
}

func NewBrowse() *cobra.Command {
	var opts browseOptions

	cmd := &cobra.Command{
		Use:   "browse",
		Short: "Browse provides the ability to search for OCH Interfaces in interactive mode",
		Example: heredoc.Doc(`
			# Start interactive mode
			ocftool hub interfaces browse
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return interactiveSelection(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.pathPattern, "path-pattern", "cap.interface.*", "Pattern of the path of a given Interface, e.g. cap.interface.*")

	return cmd
}

func interactiveSelection(ctx context.Context, opts browseOptions, w io.Writer) error {
	server, err := config.GetDefaultContext()
	if err != nil {
		return err
	}

	cli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	interfaces, err := cli.ListInterfacesWithLatest(ctx, gqlpublicapi.InterfaceFilter{
		PathPattern: &opts.pathPattern,
	})
	if err != nil {
		return err
	}

	interfacePath := ""
	prompt := &survey.Select{
		Message:  "Choose interface to run:",
		PageSize: 20,
	}

	if len(interfaces) == 0 {
		return fmt.Errorf("HUB %s doesn't have any Interfaces", server)
	}

	for _, i := range interfaces {
		if i == nil {
			continue
		}
		prompt.Options = append(prompt.Options, i.Path)
	}

	if err := survey.AskOne(prompt, &interfacePath); err != nil {
		return err
	}

	create := action.CreateOptions{
		InterfacePath: interfacePath,
		DryRun:        false,
	}

	_, err = action.Create(ctx, create, w)
	if err != nil {
		return err
	}

	return nil
}
