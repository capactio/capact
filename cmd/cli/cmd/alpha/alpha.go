package alpha

import (
	archiveimages "capact.io/capact/cmd/cli/cmd/alpha/archive-images"
	manifestgen "capact.io/capact/cmd/cli/cmd/alpha/manifest-gen"
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
		manifestgen.NewCmd(),
		archiveimages.NewCmd(),
	)

	return cmd
}
