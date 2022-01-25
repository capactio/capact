package manifest

import (
	"capact.io/capact/cmd/cli/cmd/manifest/generate"
	"github.com/spf13/cobra"
)

// NewCmd returns a cobra.Command for manifest related operations.
func NewCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "manifest",
		Aliases: []string{"mf", "manifests"},
		Short:   "This command consists of multiple subcommands to interact with OCF manifests",
	}

	root.AddCommand(
		NewValidate(),
		generate.NewCmd(),
	)
	return root
}
