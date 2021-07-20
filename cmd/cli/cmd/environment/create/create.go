package create

import "github.com/spf13/cobra"

// NewCmd returns a cobra.Command for Capact environment creation operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "create",
		Short: "This command consists of multiple subcommands to create a Capact environment",
	}

	root.AddCommand(
		NewKind(),
		NewK3D(),
	)
	return root
}
