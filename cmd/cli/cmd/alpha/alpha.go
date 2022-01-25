package alpha

import (
	archiveimages "capact.io/capact/cmd/cli/cmd/alpha/archive-images"
	"capact.io/capact/cmd/cli/cmd/manifest/generate"
	"github.com/spf13/cobra"
)

// NewCmd returns a cobra.Command for operations, which are in alpha version.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha",
		Short: "Alpha features",
		Long:  "Subcommand for alpha features in the CLI",
	}

	cmd.AddCommand(
		generate.NewCmd(),
		archiveimages.NewCmd(),
	)

	return cmd
}
