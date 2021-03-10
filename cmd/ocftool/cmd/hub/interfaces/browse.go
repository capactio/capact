package interfaces

import (
	"context"
	"fmt"
	"io"
	"os"

	"projectvoltron.dev/voltron/internal/ocftool/action"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type browseOptions struct {
	pathPrefix string
}

func NewBrowse() *cobra.Command {
	var opts browseOptions

	cmd := &cobra.Command{
		Use:   "browse",
		Short: "Browse provides the ability to search for OCH Interfaces in interactive mode",
		Example: heredoc.Doc(`
			# Start interactive mode
			ocftool hub interfaces search
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return interactiveSelection(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.pathPrefix, "path-prefix", "cap.interface.*", "Pattern of the path of a given Interface, e.g. cap.interface.*")

	return cmd
}

func interactiveSelection(ctx context.Context, opts browseOptions, w io.Writer) error {
	url := config.GetDefaultContext()
	cli, err := client.NewHub(url)
	if err != nil {
		return err
	}

	interfaces, err := cli.InterfacesWithPrefixFilter(ctx, opts.pathPrefix)
	if err != nil {
		return err
	}

	interfaceName := ""
	prompt := &survey.Select{
		Message:  "Choose interface to run:",
		PageSize: 20,
	}

	if len(interfaces.Interfaces) == 0 {
		return fmt.Errorf("HUB %s doesn't have any Interfaces", url)
	}

	for _, i := range interfaces.Interfaces {
		prompt.Options = append(prompt.Options, i.Path)
	}

	if err := survey.AskOne(prompt, &interfaceName); err != nil {
		return err
	}

	create := action.CreateOptions{
		InterfaceName: interfaceName,
		DryRun:        false,
	}
	return action.Create(ctx, create, w)
}
