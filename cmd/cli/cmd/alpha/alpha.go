package alpha

import "github.com/spf13/cobra"

// NewCmd returns a cobra.Command for operations, which are in alpha version.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha",
		Short: "Alpha features",
		Long:  "Use alpha features in the CLI",
	}

	cmd.AddCommand(NewTerraform())

	return cmd
}
