package action

import (
	"os"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/internal/ocftool/action"

	"github.com/spf13/cobra"
)

func NewDelete() *cobra.Command {
	var opts action.DeleteOptions

	cmd := &cobra.Command{
		Use:   "delete ACTION_NAME [ACTION_NAME...]",
		Short: "Deletes the Action",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctxWithNS := namespace.NewContext(cmd.Context(), opts.Namespace)
			for _, name := range args {
				opts.ActionName = name
				err := action.Delete(ctxWithNS, opts, os.Stdout)
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
