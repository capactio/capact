package action

import (
	"os"

	"github.com/spf13/cobra"
	"projectvoltron.dev/voltron/internal/ocftool/action"
)

func NewCreate() *cobra.Command {
	var opts action.CreateOptions

	cmd := &cobra.Command{
		Use:   "create INTERFACE",
		Short: "Create Action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.InterfacePath = args[0]
			_, err := action.Create(cmd.Context(), opts, os.Stdout)
			return err
		},
	}
	flags := cmd.Flags()

	flags.BoolVarP(&opts.DryRun, "dry-run", "", false, "Specifies whether the Action performs server-side test without actually running the Action")

	return cmd
}
