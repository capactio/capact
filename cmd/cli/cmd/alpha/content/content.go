package content

import "github.com/spf13/cobra"

// NewCmd returns a cobra.Command for content generation operations.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "content",
		Short: "Content generation",
		Long:  "Subcommand for various content generation operations",
	}

	cmd.AddCommand(NewInterface())
	cmd.AddCommand(NewTerraform())

	return cmd
}
