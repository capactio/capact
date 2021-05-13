package action

import (
	"os"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/action"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/spf13/cobra"
)

func NewGet() *cobra.Command {
	var opts action.GetOptions

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Lists the available Actions",
		Example: heredoc.WithCLIName(`
			# Show all Actions in table format
			<cli> action get
			
			# Show the Action "funny-wozniak" in JSON format
			<cli> action get funny-wozniak -ojson
		`, cli.Name),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ActionName = args[0]
			}

			return action.Get(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "The Kubernetes namespace where the Action was created")
	flags.StringVarP(&opts.Output, "output", "o", "table", "Output format. One of:\njson | yaml | table")
	return cmd
}
