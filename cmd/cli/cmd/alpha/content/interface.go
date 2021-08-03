package content

import (
	"log"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/alpha/content"
	"capact.io/capact/internal/cli/heredoc"
	"github.com/spf13/cobra"
)

var interfaceCfg content.InterfaceConfig

// NewInterface returns a cobra.Command to bootstrap new Interface manifests.
func NewInterface() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interface [PREFIX] [NAME]",
		Short: "Bootstrap new Interface manifests",
		Long:  "Bootstrap new Interface and associated Type manifests",
		Example: heredoc.WithCLIName(`
			# Bootstrap manifests for the cap.interface.database.postgresql.install Interface
			<cli> alpha content interface database.postgresql install`, cli.Name),
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			interfaceCfg.ManifestsPrefix = args[0]
			interfaceCfg.ManifestName = args[1]

			files, err := content.GenerateInterfaceManifests(&interfaceCfg)
			if err != nil {
				log.Fatalf("while generating content files: %v", err)
			}

			if err := writeManifestFiles(files); err != nil {
				log.Fatalf("while writing manifest files: %v", err)
			}
		},
	}

	return cmd
}
