package manifestgen

import "github.com/spf13/cobra"

// NewImplementation returns a cobra.Command for Implementation manifest generation operations.
func NewImplementation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "implementation",
		Short: "Generate new Implementation manifests",
		Long:  "Generate new Implementation manifests for various tools.",
	}

	cmd.AddCommand(NewTerraform())

	return cmd
}
