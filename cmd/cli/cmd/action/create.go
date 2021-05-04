package action

import (
	"os"

	"capact.io/capact/internal/cli/action"

	"github.com/MakeNowJust/heredoc"
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
	flags.StringVar(&opts.ActionName, "name", "", "The Action name. By default, a random name is generated.")
	flags.StringVar(&opts.ParametersFilePath, "parameters-from-file", "", "The Action input parameters in YAML format")
	flags.StringVar(&opts.TypeInstancesFilePath, "type-instances-from-file", "", heredoc.Doc(`The Action input TypeInstances in YAML format. Example:
						typeInstances:
						  - name: "config"
						    id: "ABCD-1234-EFGH-4567"`))
	flags.BoolVarP(&opts.Interactive, "interactive", "i", false, "Toggle interactive prompting in the terminal")
	flags.BoolVar(&opts.DryRun, "dry-run", false, "Specifies whether the Action performs server-side test without actually running the Action")
	// TODO: add support for creating an action directly from an implementation
	return cmd
}
