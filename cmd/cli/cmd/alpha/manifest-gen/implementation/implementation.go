package implementation

import "github.com/spf13/cobra"

// NewCmd returns a cobra.Command for Implementation manifest generation operations.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "implementation",
		Aliases: []string{"impl", "implementations"},
		Short:   "Generate new Implementation manifests",
		Long:    "Generate new Implementation manifests for various tools.",
	}

	cmd.AddCommand(NewTerraform())
	cmd.AddCommand(NewHelm())

	return cmd
}
