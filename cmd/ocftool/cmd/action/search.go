package action

import (
	"os"

	"projectvoltron.dev/voltron/internal/ocftool/action"

	"github.com/spf13/cobra"
)

func NewSearch() *cobra.Command {
	var opts action.SearchOptions

	cmd := &cobra.Command{
		Use:   "Search",
		Short: "List Actions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Search(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace where Action is created")
	flags.StringVarP(&opts.Output, "output", "o", "table", "Output format. One of:\njson|yaml|table")
	return cmd
}
