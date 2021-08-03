package content

import "github.com/spf13/cobra"

var manifestOutputDirectory string

// NewCmd returns a cobra.Command for content generation operations.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "content",
		Short: "Content generation",
		Long:  "Subcommand for various content generation operations",
	}

	cmd.AddCommand(NewInterface())
	cmd.AddCommand(NewTerraform())

	cmd.PersistentFlags().StringVarP(&manifestOutputDirectory, "output", "o", "generated", "Path to the output directory for the generated manifests")

	return cmd
}
