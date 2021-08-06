package manifestgen

import (
	"errors"
	"log"
	"strings"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/internal/cli/heredoc"
	"github.com/spf13/cobra"
)

var interfaceCfg manifestgen.InterfaceConfig

// NewInterface returns a cobra.Command to bootstrap new Interface manifests.
func NewInterface() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interface [PATH]",
		Short: "Generate new Interface manifests",
		Long:  "Generate new Interface and associated Type manifests",
		Example: heredoc.WithCLIName(`
			# Generate manifests for the cap.interface.database.postgresql.install Interface
			<cli> alpha content interface cap.interface.database.postgresql install`, cli.Name),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("accepts only one argument")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.interface.") || len(strings.Split(path, ".")) < 4 {
				return errors.New(`manifest path must be in format "cap.interface.[PREFIX].[NAME]"`)
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			interfaceCfg.ManifestPath = args[0]

			files, err := manifestgen.GenerateInterfaceManifests(&interfaceCfg)
			if err != nil {
				log.Fatalf("while generating content files: %v", err)
			}

			if err := manifestgen.WriteManifestFiles(manifestOutputDirectory, files, overrideExistingManifest); err != nil {
				log.Fatalf("while writing manifest files: %v", err)
			}
		},
	}

	cmd.Flags().StringVarP(&interfaceCfg.ManifestRevision, "revision", "r", "0.1.0", "Revision of the Interface manifest")

	return cmd
}
