package action

import (
	"os"

	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/internal/ocftool/action"
)

func NewDelete() *cobra.Command {
	var opts action.DeleteOptions

	cmd := &cobra.Command{
		Use:   "delete ACTION_NAME [ACTION_NAME...]",
		Short: "Deletes the Action",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			for _, name := range args {
				opts.ActionName = name
				err := action.Delete(cmd.Context(), opts, os.Stdout)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "default", "Kubernetes namespace where the Action was created")
	return cmd
}
