package action

import (
	"os"

	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/internal/ocftool/action"
)

// TODO (platten): It would be good to merge this together with search. Kind of how you can do `kubectl get pod` or `kubectl get po PODNAME`

func NewGet() *cobra.Command {
	var opts action.GetOptions

	cmd := &cobra.Command{
		Use:   "get ACTION",
		Short: "Displays the details of an Action from the workflow engine",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.ActionName = args[0]
			}
			return action.Get(cmd.Context(), opts, os.Stdout)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace where the Action was created")
	flags.StringVarP(&opts.Output, "output", "o", "table", "Output format. One of:\njson | yaml | table")
	return cmd
}
