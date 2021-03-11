package action

import (
	"os"

	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/internal/ocftool/action"
)

func NewRun() *cobra.Command {
	var opts action.RunOptions

	cmd := &cobra.Command{
		Use:   "run ACTION",
		Short: "Run Action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ActionName = args[0]
			return action.Run(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace where Action is created")
	return cmd
}
