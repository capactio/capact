package attribute

import (
	"strings"

	"capact.io/capact/cmd/cli/cmd/alpha/manifest-gen/common"
	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/alpha/manifestgen"
	"capact.io/capact/internal/cli/heredoc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewAttribute returns a cobra.Command to bootstrap new Attribute manifests.
func NewAttribute() *cobra.Command {
	var attributeCfg manifestgen.AttributeConfig

	cmd := &cobra.Command{
		Use:     "attribute [PATH]",
		Aliases: []string{"attributes"},
		Short:   "Generate new Attribute manifests",
		Example: heredoc.WithCLIName(`
			# Generate manifests for the cap.attribute.cloud.provider.aws Attribute
			<cli> alpha manifest-gen attribute cap.attribute.cloud.provider.aws`, cli.Name),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("accepts only one argument")
			}

			path := args[0]
			if !strings.HasPrefix(path, "cap.attribute.") || len(strings.Split(path, ".")) < 4 {
				return errors.New(`manifest path must be in format "cap.attribute.[PREFIX].[NAME]"`)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			attributeCfg.ManifestPath = args[0]
			attributeCfg.ManifestMetadata = common.GetDefaultMetadata()

			files, err := manifestgen.GenerateAttributeTemplatingConfig(&attributeCfg)
			if err != nil {
				return errors.Wrap(err, "while generating attribute templating config")
			}

			outputDir, err := cmd.Flags().GetString("output")
			if err != nil {
				return errors.Wrap(err, "while reading output flag")
			}

			overrideManifests, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return errors.Wrap(err, "while reading overwrite flag")
			}

			if err := manifestgen.WriteManifestFiles(outputDir, files, overrideManifests); err != nil {
				return errors.Wrap(err, "while writing manifest files")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&attributeCfg.ManifestRevision, "revision", "r", "0.1.0", "Revision of the Attribute manifest")

	return cmd
}

// GenerateAttributeFile generates new Attribute file.
func GenerateAttributeFile(opts common.ManifestGenOptions) (map[string]string, error) {
	var attributeCfg manifestgen.AttributeConfig
	attributeCfg.ManifestPath = common.CreateManifestPath(common.AttributeManifest, opts.ManifestPath)
	attributeCfg.ManifestMetadata = opts.Metadata
	attributeCfg.ManifestRevision = opts.Revision
	files, err := manifestgen.GenerateAttributeTemplatingConfig(&attributeCfg)
	if err != nil {
		return nil, errors.Wrap(err, "while generating attribute content file")
	}
	return files, nil
}
