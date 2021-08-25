package manifestgen

import (
	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/implementation"
	"github.com/spf13/cobra"
)

// NewCmd returns a cobra.Command for content generation operations.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest-gen",
		Short: "Manifests generation",
		Long:  "Subcommand for various manifest generation operations",
	}

	cmd.AddCommand(NewInterface())
	cmd.AddCommand(implementation.NewCmd())

	cmd.PersistentFlags().StringP("output", "o", "generated", "Path to the output directory for the generated manifests")
	cmd.PersistentFlags().Bool("overwrite", false, "Overwrite existing manifest files")

	return cmd
}
