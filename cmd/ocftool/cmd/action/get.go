package action

import (
	"os"

	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/internal/ocftool/action"
)

func NewGet() *cobra.Command {
	var opts action.GetOptions

	cmd := &cobra.Command{
		Use:   "get ACTION",
		Short: "Get Action",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ActionName = args[0]
			return action.Get(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace where Action is created")
	flags.StringVarP(&opts.Output, "output", "o", "table", "Output format. One of:\njson|yaml|table")
	return cmd
}
