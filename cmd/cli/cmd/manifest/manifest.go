package manifest

import "github.com/spf13/cobra"

// NewCmd returns a cobra.Command for manifest related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "manifest",
		Aliases: []string{"mf"},
		Short:   "This command consists of multiple subcommands to interact with OCF manifests",
	}

	root.AddCommand(
		NewValidate(),
	)
	return root
}
