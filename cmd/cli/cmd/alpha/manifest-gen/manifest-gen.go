package manifestgen

import "github.com/spf13/cobra"

var (
	manifestOutputDirectory  string
	overrideExistingManifest bool
)

// NewCmd returns a cobra.Command for content generation operations.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest-gen",
		Short: "Manifests generation",
		Long:  "Subcommand for various manifest generation operations",
	}

	cmd.AddCommand(NewInterface())
	cmd.AddCommand(NewImplementation())

	cmd.PersistentFlags().StringVarP(&manifestOutputDirectory, "output", "o", "generated", "Path to the output directory for the generated manifests")
	cmd.PersistentFlags().BoolVar(&overrideExistingManifest, "override", false, "Override existing manifest files")

	return cmd
}
