package interfaces

import (
	"context"
	"fmt"
	"io"
	"os"

	"capact.io/capact/internal/cli/action"
	"capact.io/capact/internal/cli/client"
	"capact.io/capact/internal/cli/config"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type browseOptions struct {
	pathPattern string
}

// NewBrowse returns a cobra.Command for browsing available Interfaces in a Public Hub.
func NewBrowse() *cobra.Command {
	var opts browseOptions

	cmd := &cobra.Command{
		Use:   "browse",
		Short: "Browse provides the ability to browse through the available OCF Interfaces in interactive mode. Optionally create a Target Action.",
		Example: heredoc.Doc(`
			# Browse (and optionally create an Action) from the available OCF Interfaces.
			<cli> hub interfaces browse
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return interactiveSelection(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.pathPattern, "path-pattern", "cap.interface.*", "The pattern of the path of a given Interface, e.g. cap.interface.*")

	return cmd
}

func interactiveSelection(ctx context.Context, opts browseOptions, w io.Writer) error {
	server := config.GetDefaultContext()

	cli, err := client.NewHub(server)
	if err != nil {
		return err
	}

	interfaces, err := cli.ListInterfaces(ctx, public.WithIfaceFilter(gqlpublicapi.InterfaceFilter{
		PathPattern: &opts.pathPattern,
	}))
	if err != nil {
		return err
	}

	interfacePath := ""
	prompt := &survey.Select{
		Message:  "Choose interface to run: ",
		PageSize: 20,
	}

	if len(interfaces) == 0 {
		return fmt.Errorf("Hub %s doesn't have any Interfaces", server)
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
		Interactive:   true,
	}

	_, err = action.Create(ctx, create, w)
	if err != nil {
		return err
	}

	return nil
}
