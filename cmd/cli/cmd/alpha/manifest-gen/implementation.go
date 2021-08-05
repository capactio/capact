package manifestgen

import "github.com/spf13/cobra"

// NewImplementation returns a cobra.Command for Implementation manifest generation operations.
func NewImplementation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "implementation",
		Short: "Bootstrap new Implementation manifests",
		Long:  "Bootstrap new Implementation manifests for various tools.",
	}

	cmd.AddCommand(NewTerraform())

	return cmd
}
