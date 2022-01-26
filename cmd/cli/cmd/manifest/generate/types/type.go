package types

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/manifest/generate/common"
	"capact.io/capact/internal/cli/manifestgen"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/heredoc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewType returns a cobra.Command to bootstrap new Type manifests.
func NewType() *cobra.Command {
	var typeCfg manifestgen.InterfaceConfig

	cmd := &cobra.Command{
		Use:     "type [PATH]",
		Aliases: []string{"types"},
		Short:   "Generate new Type manifests",
		Example: heredoc.WithCLIName(`
			# Generate manifests for the cap.type.database.postgresql.config Type
			<cli> manifest generate type cap.type.database.postgresql.config`, cli.Name),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("accepts one argument: [MANIFEST_PATH]")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.type.") || len(strings.Split(path, ".")) < 4 {
				return errors.New(`manifest path must be in format "cap.type.[PREFIX].[NAME]"`)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			typeCfg.ManifestRef.Path = args[0]
			typeCfg.Metadata = common.GetDefaultInterfaceMetadata()

			manifests, err := manifestgen.GenerateTypeTemplatingConfig(&typeCfg)
			if err != nil {
				return errors.Wrap(err, "while generating content files")
			}

			outputDir, err := cmd.Flags().GetString("output")
			if err != nil {
				return errors.Wrap(err, "while reading output flag")
			}

			overrideManifests, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return errors.Wrap(err, "while reading overwrite flag")
			}

			if err := manifestgen.WriteManifestFiles(outputDir, manifests, overrideManifests); err != nil {
				return errors.Wrap(err, "while writing manifest files")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&typeCfg.ManifestRef.Revision, "revision", "r", "0.1.0", "Revision of the Type manifest")

	return cmd
}
