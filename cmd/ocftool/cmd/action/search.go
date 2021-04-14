package action

import (
	"os"

	"capact.io/capact/internal/ocftool/action"

	"github.com/spf13/cobra"
)

func NewSearch() *cobra.Command {
	var opts action.SearchOptions

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Lists the available Actions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Search(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "The Kubernetes namespace where the Action was created")
	flags.StringVarP(&opts.Output, "output", "o", "table", "Output format. One of:\njson | yaml | table")
	return cmd
}
