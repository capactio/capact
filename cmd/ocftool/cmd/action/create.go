package action

import (
	"os"

	"capact.io/capact/internal/ocftool/action"
	"github.com/spf13/cobra"
)

func NewCreate() *cobra.Command {
	var opts action.CreateOptions

	cmd := &cobra.Command{
		Use:   "create INTERFACE",
		Short: "Creates/renders a new Action with a specified Interface",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.InterfacePath = args[0]
			_, err := action.Create(cmd.Context(), opts, os.Stdout)
			return err
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Namespace, "namespace", "n", "", "Kubernetes namespace where the Action is to be created")
	flags.BoolVarP(&opts.DryRun, "dry-run", "", false, "Specifies whether the Action performs server-side test without actually running the Action")
	// TODO: add support for creating an action directly from an implementation
	return cmd
}
