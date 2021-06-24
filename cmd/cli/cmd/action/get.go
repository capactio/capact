package action

import (
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/action"
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/cli/printer"

	"github.com/spf13/cobra"
)

// NewGet returns a cobra.Command for the "action get" command.
func NewGet() *cobra.Command {
	var opts action.GetOptions

	resourcePrinter := printer.NewForResource(os.Stdout, printer.WithJSON(), printer.WithYAML(), printer.WithTable(action.TableDataOnGet))

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Displays one or multiple Actions",
		Example: heredoc.WithCLIName(`
			# Show all Actions in table format
			<cli> action get
			
			# Show the Action "funny-stallman" in JSON format
			<cli> action get funny-stallman -ojson
		`, cli.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ActionNames = args
			return action.Get(cmd.Context(), opts, resourcePrinter)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace where the Action was created")
	resourcePrinter.RegisterFlags(flags)
	return cmd
}
