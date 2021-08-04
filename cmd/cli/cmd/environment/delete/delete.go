package delete

import "github.com/spf13/cobra"

// NewCmd returns a cobra.Command for Capact environment related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "delete",
		Short: "This command consists of multiple subcommands to delete created Capact environment",
	}

	root.AddCommand(
		NewKind(),
		NewK3d(),
	)
	return root
}
